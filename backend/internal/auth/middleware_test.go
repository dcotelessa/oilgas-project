// backend/internal/auth/middleware_test.go
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCK SERVICE IMPLEMENTATION
// ============================================================================

// MockService implements the auth Service interface for testing
type MockService struct {
	mock.Mock
}

func (m *MockService) Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	args := m.Called(ctx, req)
	if response := args.Get(0); response != nil {
		return response.(*LoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if response := args.Get(0); response != nil {
		return response.(*LoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockService) ValidateToken(ctx context.Context, token string) (*User, *TenantAccess, error) {
	args := m.Called(ctx, token)
	user := args.Get(0)
	tenantAccess := args.Get(1)
	
	if user != nil && tenantAccess != nil {
		return user.(*User), tenantAccess.(*TenantAccess), args.Error(2)
	}
	return nil, nil, args.Error(2)
}

func (m *MockService) CreateCustomerContact(ctx context.Context, req CreateCustomerContactRequest) (*User, error) {
	args := m.Called(ctx, req)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) CreateEnterpriseUser(ctx context.Context, req CreateEnterpriseUserRequest) (*User, error) {
	args := m.Called(ctx, req)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, userID int, updates UserUpdates) (*User, error) {
	args := m.Called(ctx, userID, updates)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) DeactivateUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockService) CheckPermission(ctx context.Context, check UserPermissionCheck) error {
	args := m.Called(ctx, check)
	return args.Error(0)
}

func (m *MockService) CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error {
	args := m.Called(ctx, userID, tenantID, yardLocation)
	return args.Error(0)
}

func (m *MockService) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	if perms := args.Get(0); perms != nil {
		return perms.([]Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetEnterpriseContext(ctx context.Context, userID int) (*EnterpriseContext, error) {
	args := m.Called(ctx, userID)
	if context := args.Get(0); context != nil {
		return context.(*EnterpriseContext), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetCustomerAccessContext(ctx context.Context, userID int) (*CustomerAccessContext, error) {
	args := m.Called(ctx, userID)
	if context := args.Get(0); context != nil {
		return context.(*CustomerAccessContext), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetUser(ctx context.Context, userID int) (*User, error) {
	args := m.Called(ctx, userID)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	args := m.Called(ctx, tenantID)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error) {
	args := m.Called(ctx, filters)
	users := args.Get(0)
	total := args.Get(1)
	if users != nil && total != nil {
		return users.([]User), total.(int), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockService) ListAllUsers(ctx context.Context, filters AdminUserFilters) ([]User, int, error) {
	args := m.Called(ctx, filters)
	users := args.Get(0)
	total := args.Get(1)
	if users != nil && total != nil {
		return users.([]User), total.(int), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockService) GetUserStats(ctx context.Context, req UserStatsRequest) (*UserStats, error) {
	args := m.Called(ctx, req)
	if stats := args.Get(0); stats != nil {
		return stats.(*UserStats), args.Error(1)
	}
	return nil, args.Error(1)
}

// ============================================================================
// TEST FIXTURES
// ============================================================================

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestUser() *User {
	customerID := 123
	return &User{
		ID:               1,
		Username:         "contact@customer.com",
		Email:            "contact@customer.com",
		FullName:         "Test Contact",
		Role:             RoleCustomerContact,
		AccessLevel:      1,
		IsEnterpriseUser: false,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleCustomerContact,
				YardAccess: []YardAccess{
					{
						YardLocation:       "houston_north",
						CanViewWorkOrders:  true,
						CanViewInventory:   true,
						CanCreateWorkOrders: true,
					},
				},
			},
		},
		PrimaryTenantID: "houston",
		CustomerID:      &customerID,
		ContactType:     ContactPrimary,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestEnterpriseUser() *User {
	return &User{
		ID:               2,
		Username:         "admin@company.com",
		Email:            "admin@company.com",
		FullName:         "Enterprise Admin",
		Role:             RoleEnterpriseAdmin,
		AccessLevel:      5,
		IsEnterpriseUser: true,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleEnterpriseAdmin,
				Permissions: []Permission{
					PermissionCrossTenantView,
					PermissionUserManagement,
				},
			},
			{
				TenantID: "longbeach",
				Role:     RoleEnterpriseAdmin,
				Permissions: []Permission{
					PermissionCrossTenantView,
					PermissionUserManagement,
				},
			},
		},
		PrimaryTenantID: "houston",
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestTenantAccess() *TenantAccess {
	return &TenantAccess{
		TenantID: "houston",
		Role:     RoleCustomerContact,
		YardAccess: []YardAccess{
			{
				YardLocation:       "houston_north",
				CanViewWorkOrders:  true,
				CanViewInventory:   true,
				CanCreateWorkOrders: true,
			},
		},
	}
}

// ============================================================================
// REQUIRE AUTH MIDDLEWARE TESTS
// ============================================================================

func TestRequireAuth_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	testTenantAccess := createTestTenantAccess()
	
	mockService.On("ValidateToken", mock.Anything, "valid_token").Return(testUser, testTenantAccess, nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireAuth_NoToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("ValidateToken", mock.Anything, "invalid_token").Return(nil, nil, ErrInvalidCredentials)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("ValidateToken", mock.Anything, "expired_token").Return(nil, nil, ErrTokenExpired)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer expired_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireAuth_QueryToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	testTenantAccess := createTestTenantAccess()
	
	mockService.On("ValidateToken", mock.Anything, "query_token").Return(testUser, testTenantAccess, nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test?token=query_token", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// REQUIRE ROLE MIDDLEWARE TESTS
// ============================================================================

func TestRequireRole_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.RequireRole(RoleCustomerContact, RoleOperator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_InsufficientRole(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser() // Customer contact
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.RequireRole(RoleAdmin, RoleEnterpriseAdmin))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NoUserContext(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.Use(middleware.RequireRole(RoleCustomerContact))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// REQUIRE PERMISSION MIDDLEWARE TESTS
// ============================================================================

func TestRequirePermission_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	expectedCheck := UserPermissionCheck{
		UserID:     testUser.ID,
		TenantID:   "houston",
		Permission: PermissionInventoryRead,
	}
	
	mockService.On("CheckPermission", mock.Anything, expectedCheck).Return(nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("user_id", testUser.ID)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionInventoryRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequirePermission_Denied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	expectedCheck := UserPermissionCheck{
		UserID:     testUser.ID,
		TenantID:   "houston",
		Permission: PermissionUserManagement,
	}
	
	mockService.On("CheckPermission", mock.Anything, expectedCheck).Return(ErrPermissionDenied)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("user_id", testUser.ID)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionUserManagement))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// REQUIRE YARD ACCESS MIDDLEWARE TESTS
// ============================================================================

func TestRequireYardAccess_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("CheckYardAccess", mock.Anything, 1, "houston", "houston_north").Return(nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequireYardAccess())
	router.GET("/test/:yardLocation", func(c *gin.Context) {
		yardLocation, exists := GetYardLocationFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston_north", yardLocation)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test/houston_north", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireYardAccess_Denied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("CheckYardAccess", mock.Anything, 1, "houston", "houston_south").Return(ErrYardAccessDenied)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequireYardAccess())
	router.GET("/test/:yardLocation", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test/houston_south", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireYardAccess_NoYardInURL(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequireYardAccess())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// REQUIRE CUSTOMER CONTACT MIDDLEWARE TESTS
// ============================================================================

func TestRequireCustomerContact_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.RequireCustomerContact())
	router.GET("/test", func(c *gin.Context) {
		customerID, exists := c.Get("customer_id")
		assert.True(t, exists)
		assert.Equal(t, *testUser.CustomerID, customerID)
		
		contactType, exists := c.Get("contact_type")
		assert.True(t, exists)
		assert.Equal(t, testUser.ContactType, contactType)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireCustomerContact_NotCustomerContact(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	enterpriseUser := createTestEnterpriseUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", enterpriseUser)
		c.Next()
	})
	router.Use(middleware.RequireCustomerContact())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ============================================================================
// REQUIRE ENTERPRISE ACCESS MIDDLEWARE TESTS
// ============================================================================

func TestRequireEnterpriseAccess_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	enterpriseUser := createTestEnterpriseUser()
	enterpriseContext := &EnterpriseContext{
		UserID:            enterpriseUser.ID,
		AccessibleTenants: enterpriseUser.TenantAccess,
		IsEnterpriseAdmin: true,
		CrossTenantPerms:  []Permission{PermissionCrossTenantView, PermissionUserManagement},
	}
	
	mockService.On("GetEnterpriseContext", mock.Anything, enterpriseUser.ID).Return(enterpriseContext, nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", enterpriseUser)
		c.Next()
	})
	router.Use(middleware.RequireEnterpriseAccess())
	router.GET("/test", func(c *gin.Context) {
		context, exists := c.Get("enterprise_context")
		assert.True(t, exists)
		assert.Equal(t, enterpriseContext, context)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRequireEnterpriseAccess_Denied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser() // Regular customer contact
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.RequireEnterpriseAccess())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ============================================================================
// TENANT CONTEXT MIDDLEWARE TESTS
// ============================================================================

func TestSetTenantContext_FromHeader(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.SetTenantContext())
	router.GET("/test", func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, "houston", tenantID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "houston")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetTenantContext_FromQuery(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.SetTenantContext())
	router.GET("/test", func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, "houston", tenantID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test?tenant_id=houston", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetTenantContext_UsePrimaryTenant(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.SetTenantContext())
	router.GET("/test", func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, testUser.PrimaryTenantID, tenantID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil) // No tenant specified
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================================
// VALIDATE TENANT ACCESS MIDDLEWARE TESTS
// ============================================================================

func TestValidateTenantAccess_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("tenant_id", "houston") // Current tenant
		c.Next()
	})
	router.Use(middleware.ValidateTenantAccess())
	router.GET("/test/:tenantId", func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, "houston", tenantID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test/houston", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateTenantAccess_AccessDenied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser() // Only has access to houston
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.ValidateTenantAccess())
	router.GET("/test/:tenantId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test/longbeach", nil) // Different tenant
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ============================================================================
// INTEGRATION TESTS - MIDDLEWARE COMPOSITION
// ============================================================================

func TestMiddlewareComposition_CustomerContactWorkflow(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	testTenantAccess := createTestTenantAccess()
	
	mockService.On("ValidateToken", mock.Anything, "valid_token").Return(testUser, testTenantAccess, nil)
	mockService.On("CheckYardAccess", mock.Anything, testUser.ID, "houston", "houston_north").Return(nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.Use(middleware.RequireCustomerContact())
	router.Use(middleware.RequireYardAccess())
	router.GET("/customer/work-orders/:yardLocation", func(c *gin.Context) {
		// Verify all contexts are set
		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)
		
		customerID, exists := c.Get("customer_id")
		assert.True(t, exists)
		assert.Equal(t, *testUser.CustomerID, customerID)
		
		yardLocation, exists := GetYardLocationFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston_north", yardLocation)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/customer/work-orders/houston_north", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestMiddlewareComposition_EnterpriseAdminWorkflow(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	enterpriseUser := createTestEnterpriseUser()
	testTenantAccess := &TenantAccess{
		TenantID: "houston",
		Role:     RoleEnterpriseAdmin,
	}
	
	enterpriseContext := &EnterpriseContext{
		UserID:            enterpriseUser.ID,
		AccessibleTenants: enterpriseUser.TenantAccess,
		IsEnterpriseAdmin: true,
		CrossTenantPerms:  []Permission{PermissionCrossTenantView, PermissionUserManagement},
	}
	
	mockService.On("ValidateToken", mock.Anything, "admin_token").Return(enterpriseUser, testTenantAccess, nil)
	mockService.On("GetEnterpriseContext", mock.Anything, enterpriseUser.ID).Return(enterpriseContext, nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.Use(middleware.RequireEnterpriseAccess())
	router.GET("/admin/users", func(c *gin.Context) {
		// Verify enterprise context is set
		context, exists := GetEnterpriseContextFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, enterpriseContext.UserID, context.UserID)
		assert.True(t, context.IsEnterpriseAdmin)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer admin_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ============================================================================
// TOKEN EXTRACTION TESTS
// ============================================================================

func TestExtractToken_FromHeader(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		token := middleware.extractToken(c)
		assert.Equal(t, "header_token", token)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer header_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractToken_FromQuery(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		token := middleware.extractToken(c)
		assert.Equal(t, "query_token", token)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	
	req := httptest.NewRequest("GET", "/test?token=query_token", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractToken_NoToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		token := middleware.extractToken(c)
		assert.Empty(t, token)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================================
// CONTEXT HELPER TESTS
// ============================================================================

func TestContextHelpers(t *testing.T) {
	router := setupTestRouter()
	
	testUser := createTestUser()
	
	router.GET("/test", func(c *gin.Context) {
		// Set test context
		c.Set("user", testUser)
		c.Set("user_id", testUser.ID)
		c.Set("tenant_id", "houston")
		c.Set("customer_id", *testUser.CustomerID)
		c.Set("yard_location", "houston_north")
		
		// Test GetUserFromContext
		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.ID)
		
		// Test GetUserIDFromContext
		userID, exists := GetUserIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, userID)
		
		// Test GetTenantIDFromContext
		tenantID, exists := GetTenantIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston", tenantID)
		
		// Test GetCustomerIDFromContext
		customerID, exists := GetCustomerIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, *testUser.CustomerID, customerID)
		
		// Test GetYardLocationFromContext
		yardLocation, exists := GetYardLocationFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston_north", yardLocation)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================================
// PERFORMANCE BENCHMARKS
// ============================================================================

func BenchmarkRequireAuth(b *testing.B) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	testTenantAccess := createTestTenantAccess()
	
	mockService.On("ValidateToken", mock.Anything, "valid_token").Return(testUser, testTenantAccess, nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPermissionCheck(b *testing.B) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	expectedCheck := UserPermissionCheck{
		UserID:     testUser.ID,
		TenantID:   "houston",
		Permission: PermissionInventoryRead,
	}
	
	mockService.On("CheckPermission", mock.Anything, expectedCheck).Return(nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("user_id", testUser.ID)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionInventoryRead))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
	}
}

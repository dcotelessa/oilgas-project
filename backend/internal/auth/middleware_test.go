// backend/internal/auth/middleware_test.go
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service for middleware testing
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
	return args.Get(0).([]Permission), args.Error(1)
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
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockService) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockService) SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockService) GetUserStats(ctx context.Context) (*UserStats, error) {
	args := m.Called(ctx)
	if stats := args.Get(0); stats != nil {
		return stats.(*UserStats), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) InvalidateUserSessions(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockService) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test helper functions
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestTenantAccess() *TenantAccess {
	return &TenantAccess{
		TenantID: "houston",
		Role:     RoleCustomerContact,
		Permissions: []Permission{
			PermissionWorkOrderRead,
			PermissionInventoryRead,
		},
		YardAccess: []YardAccess{
			{
				YardLocation:        "houston_north",
				CanViewWorkOrders:   true,
				CanViewInventory:    true,
				CanCreateWorkOrders: true,
			},
		},
		CanRead:  true,
		CanWrite: true,
	}
}

// RequireAuth middleware tests
func TestRequireAuth_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	testTenantAccess := createTestTenantAccess()
	
	mockService.On("ValidateToken", mock.Anything, "valid_token").Return(testUser, testTenantAccess, nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		// Verify user context is set
		user, exists := c.Get("user")
		assert.True(t, exists)
		assert.Equal(t, testUser, user)
		
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, userID)
		
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, testTenantAccess.TenantID, tenantID)
		
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
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// RequireEnterpriseAccess middleware tests
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

// ValidateTenantAccess middleware tests
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

func TestValidateTenantAccess_Denied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.ValidateTenantAccess())
	router.GET("/test/:tenantId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test/longbeach", nil) // User doesn't have access
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// CustomerContactFilter middleware tests
func TestCustomerContactFilter_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.CustomerContactFilter())
	router.GET("/customers/:customerId", func(c *gin.Context) {
		customerFilter, exists := c.Get("customer_filter")
		assert.True(t, exists)
		assert.Equal(t, *testUser.CustomerID, customerFilter)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/customers/123", nil) // Same as user's customer ID
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomerContactFilter_AccessDenied(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	testUser := createTestUser()
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Next()
	})
	router.Use(middleware.CustomerContactFilter())
	router.GET("/customers/:customerId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/customers/456", nil) // Different customer ID
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// SetTenantContext middleware tests
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

// Integration tests for middleware composition
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
		user, _ := GetUserFromContext(c)
		assert.NotNil(t, user)
		assert.True(t, user.IsCustomerContact())
		
		customerID, exists := GetCustomerIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, *testUser.CustomerID, customerID)
		
		yardLocation, exists := GetYardLocationFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston_north", yardLocation)
		
		c.JSON(http.StatusOK, gin.H{"message": "customer contact has yard access"})
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
		TenantID:    "houston",
		Role:        RoleEnterpriseAdmin,
		Permissions: []Permission{PermissionCrossTenantView, PermissionUserManagement},
		CanRead:     true,
		CanWrite:    true,
		CanDelete:   true,
	}
	
	enterpriseContext := &EnterpriseContext{
		UserID:            enterpriseUser.ID,
		AccessibleTenants: enterpriseUser.TenantAccess,
		IsEnterpriseAdmin: true,
		CrossTenantPerms:  []Permission{PermissionCrossTenantView, PermissionUserManagement},
	}
	
	mockService.On("ValidateToken", mock.Anything, "admin_token").Return(enterpriseUser, testTenantAccess, nil)
	mockService.On("GetEnterpriseContext", mock.Anything, enterpriseUser.ID).Return(enterpriseContext, nil)
	mockService.On("CheckPermission", mock.Anything, mock.MatchedBy(func(check UserPermissionCheck) bool {
		return check.Permission == PermissionUserManagement
	})).Return(nil)
	
	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.Use(middleware.RequireEnterpriseAccess())
	router.Use(middleware.RequirePermission(PermissionUserManagement))
	router.GET("/admin/users", func(c *gin.Context) {
		// Verify enterprise context is set
		enterpriseCtx, exists := GetEnterpriseContextFromContext(c)
		assert.True(t, exists)
		assert.True(t, enterpriseCtx.IsEnterpriseAdmin)
		
		user, _ := GetUserFromContext(c)
		assert.NotNil(t, user)
		assert.True(t, user.CanPerformCrossTenantOperation())
		
		c.JSON(http.StatusOK, gin.H{"message": "enterprise admin has user management access"})
	})
	
	req := httptest.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer admin_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Context utility function tests
func TestGetUserFromContext_Success(t *testing.T) {
	router := setupTestRouter()
	testUser := createTestUser()
	
	router.GET("/test", func(c *gin.Context) {
		c.Set("user", testUser)
		
		user, exists := GetUserFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, testUser, user)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserFromContext_NotFound(t *testing.T) {
	router := setupTestRouter()
	
	router.GET("/test", func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		assert.False(t, exists)
		assert.Nil(t, user)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTenantIDFromContext_Success(t *testing.T) {
	router := setupTestRouter()
	
	router.GET("/test", func(c *gin.Context) {
		c.Set("tenant_id", "houston")
		
		tenantID, exists := GetTenantIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, "houston", tenantID)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test token extraction from different sources
func TestExtractToken_FromAuthorizationHeader(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		token := middleware.extractToken(c)
		assert.Equal(t, "test_token", token)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test_token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractToken_FromQueryParameter(t *testing.T) {
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

// RequireRole middleware tests
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
	
	testUser := createTestUser()
	
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

// RequirePermission middleware tests
func TestRequirePermission_Success(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("CheckPermission", mock.Anything, mock.MatchedBy(func(check UserPermissionCheck) bool {
		return check.UserID == 1 && 
			   check.TenantID == "houston" && 
			   check.Permission == PermissionWorkOrderRead
	})).Return(nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionWorkOrderRead))
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
	
	mockService.On("CheckPermission", mock.Anything, mock.MatchedBy(func(check UserPermissionCheck) bool {
		return check.UserID == 1 && 
			   check.TenantID == "houston" && 
			   check.Permission == PermissionCustomerDelete
	})).Return(ErrPermissionDenied)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionCustomerDelete))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

// RequireYardAccess middleware tests
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
		yardLocation, exists := c.Get("yard_location")
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

// RequireCustomerContact middleware tests
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

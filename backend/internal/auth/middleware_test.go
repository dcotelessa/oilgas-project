// backend/internal/auth/middleware_test.go
// Fixed with complete MockService
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
// COMPLETE MOCK SERVICE IMPLEMENTATION
// ============================================================================

// MockService implements the complete auth Service interface for testing
type MockService struct {
	mock.Mock
}

// Authentication methods
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

// User management methods
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

func (m *MockService) DeactivateUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockService) GetUserByID(ctx context.Context, userID int) (*User, error) {
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

// Search and filtering methods
func (m *MockService) SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error) {
	args := m.Called(ctx, filters)
	users := args.Get(0)
	total := args.Get(1)
	if users != nil && total != nil {
		return users.([]User), total.(int), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockService) GetUserStats(ctx context.Context) (*UserStats, error) {
	args := m.Called(ctx)
	if stats := args.Get(0); stats != nil {
		return stats.(*UserStats), args.Error(1)
	}
	return nil, args.Error(1)
}

// Tenant access management methods
func (m *MockService) UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// Permission checking methods (missing methods that were causing errors)
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

// Enterprise operations methods (missing methods that were causing errors)
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
		user, exists := c.Get("user")
		assert.True(t, exists)
		assert.Equal(t, testUser.ID, user.(*User).ID)
		
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
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mockService := new(MockService)
	middleware := NewMiddleware(mockService)
	
	mockService.On("ValidateToken", mock.Anything, "invalid_token").Return(nil, nil, ErrInvalidToken)
	
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
		Permission: PermissionViewInventory,
	}
	
	mockService.On("CheckPermission", mock.Anything, expectedCheck).Return(nil)
	
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", testUser)
		c.Set("user_id", testUser.ID)
		c.Set("tenant_id", "houston")
		c.Next()
	})
	router.Use(middleware.RequirePermission(PermissionViewInventory))
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
	
	mockService.On("CheckYardAccess", mock.Anything, 1, "houston", "houston_north").Return(ErrYardAccessDenied)
	
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
	
	req := httptest.NewRequest("GET", "/test/houston_north", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

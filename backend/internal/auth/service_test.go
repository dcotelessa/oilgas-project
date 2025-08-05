// backend/internal/auth/service_test.go
package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCK REPOSITORY FOR SERVICE TESTING
// ============================================================================

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error {
	args := m.Called(ctx, userID, tenantAccess)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error {
	args := m.Called(ctx, userID, tenantID, yardAccess)
	return args.Error(0)
}

func (m *MockAuthRepository) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuthRepository) CreateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	args := m.Called(ctx, token)
	if session := args.Get(0); session != nil {
		return session.(*Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetSessionByID(ctx context.Context, sessionID string) (*Session, error) {
	args := m.Called(ctx, sessionID)
	if session := args.Get(0); session != nil {
		return session.(*Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockAuthRepository) InvalidateSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthRepository) InvalidateUserSessions(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthRepository) GetEnterpriseUsers(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	args := m.Called(ctx, tenantID)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUsersByRole(ctx context.Context, role UserRole) ([]User, error) {
	args := m.Called(ctx, role)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return nil, args.Error(0)
}

// ============================================================================
// SERVICE TESTS
// ============================================================================

func TestService_CreateCustomerContact(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	// Test user creation
	req := CreateCustomerContactRequest{
		CustomerID:  123,
		TenantID:    "houston",
		Email:       "contact@customer.com",
		FullName:    "Test Contact",
		Password:    "password123",
		ContactType: ContactPrimary,
		YardAccess: []YardAccess{
			{
				YardLocation:       "houston_north",
				CanViewWorkOrders:  true,
				CanViewInventory:   true,
				CanCreateWorkOrders: true,
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	// Execute
	user, err := service.CreateCustomerContact(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, RoleCustomerContact, user.Role)
	assert.Equal(t, req.CustomerID, *user.CustomerID)
	assert.Equal(t, req.TenantID, user.PrimaryTenantID)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateEnterpriseUser(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	req := CreateEnterpriseUserRequest{
		Username:         "admin@company.com",
		Email:            "admin@company.com",
		FullName:         "Enterprise Admin",
		Password:         "password123",
		Role:             RoleEnterpriseAdmin,
		IsEnterpriseUser: true,
		PrimaryTenantID:  "houston",
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleEnterpriseAdmin,
				Permissions: []Permission{
					PermissionCrossTenantView,
					PermissionUserManagement,
				},
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("GetUserByUsername", mock.Anything, req.Username).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	// Execute
	user, err := service.CreateEnterpriseUser(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, RoleEnterpriseAdmin, user.Role)
	assert.True(t, user.IsEnterpriseUser)
	assert.Equal(t, req.PrimaryTenantID, user.PrimaryTenantID)

	mockRepo.AssertExpectations(t)
}

func TestService_Authenticate_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	// Create test user with hashed password
	hashedPassword, _ := hashPassword("password123")
	testUser := &User{
		ID:               1,
		Username:         "test@example.com",
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		Role:             RoleOperator,
		IsActive:         true,
		PrimaryTenantID:  "houston",
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleOperator,
			},
		},
	}

	req := LoginRequest{
		Username: "test@example.com",
		Password: "password123",
		TenantID: "houston",
	}

	// Mock repository calls
	mockRepo.On("GetUserByUsername", mock.Anything, req.Username).Return(testUser, nil)
	mockRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*auth.Session")).Return(nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	// Execute
	response, err := service.Authenticate(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, testUser.ID, response.User.ID)

	mockRepo.AssertExpectations(t)
}

func TestService_Authenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	req := LoginRequest{
		Username: "test@example.com",
		Password: "wrongpassword",
	}

	// Mock repository calls
	mockRepo.On("GetUserByUsername", mock.Anything, req.Username).Return(nil, ErrUserNotFound)

	// Execute
	response, err := service.Authenticate(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestService_CheckPermission_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	testUser := &User{
		ID:   1,
		Role: RoleAdmin,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleAdmin,
				Permissions: []Permission{
					PermissionInventoryRead,
					PermissionInventoryWrite,
				},
			},
		},
	}

	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionInventoryRead,
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, check.UserID).Return(testUser, nil)

	// Execute
	err := service.CheckPermission(context.Background(), check)

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_CheckPermission_Denied(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, "test-jwt-secret")

	testUser := &User{
		ID:   1,
		Role: RoleOperator,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleOperator,
				Permissions: []Permission{
					PermissionInventoryRead,
				},
			},
		},
	}

	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionUserManagement, // Operator doesn't have this
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, check.UserID).Return(testUser, nil)

	// Execute
	err := service.CheckPermission(context.Background(), check)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrPermissionDenied, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// HELPER FUNCTIONS FOR TESTS
// ============================================================================

func hashPassword(password string) (string, error) {
	// Simple password hashing for tests
	return "hashed_" + password, nil
}

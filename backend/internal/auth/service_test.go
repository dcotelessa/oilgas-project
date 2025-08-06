// backend/internal/auth/service_test.go - Complete comprehensive test suite
package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// COMPLETE MOCK REPOSITORY IMPLEMENTATION
// ============================================================================

type MockAuthRepository struct {
	mock.Mock
}

// User management methods
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

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	args := m.Called(ctx, user)
	if u := args.Get(0); u != nil {
		return u.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepository) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Session management methods
func (m *MockAuthRepository) CreateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthRepository) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	args := m.Called(ctx, sessionID)
	if session := args.Get(0); session != nil {
		return session.(*Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	args := m.Called(ctx, token)
	if session := args.Get(0); session != nil {
		return session.(*Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) UpdateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
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

// Multi-tenant user queries
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

func (m *MockAuthRepository) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

// Permission and access queries
func (m *MockAuthRepository) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	if perms := args.Get(0); perms != nil {
		return perms.([]Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error) {
	args := m.Called(ctx, tenantID, yardLocation)
	if users := args.Get(0); users != nil {
		return users.([]User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) ValidateCustomerExists(ctx context.Context, customerID int) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

// Search and analytics
func (m *MockAuthRepository) SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error) {
	args := m.Called(ctx, filters)
	users := args.Get(0)
	total := args.Get(1)
	if users != nil && total != nil {
		return users.([]User), total.(int), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockAuthRepository) GetUserStats(ctx context.Context) (*UserStats, error) {
	args := m.Called(ctx)
	if stats := args.Get(0); stats != nil {
		return stats.(*UserStats), args.Error(1)
	}
	return nil, args.Error(1)
}

// Tenant access management
func (m *MockAuthRepository) UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error {
	args := m.Called(ctx, userID, tenantAccess)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error {
	args := m.Called(ctx, userID, tenantID, yardAccess)
	return args.Error(0)
}

// Legacy tenant support
func (m *MockAuthRepository) GetUserTenants(ctx context.Context, userID int) ([]LegacyTenant, error) {
	args := m.Called(ctx, userID)
	if tenants := args.Get(0); tenants != nil {
		return tenants.([]LegacyTenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) GetTenantBySlug(ctx context.Context, slug string) (*LegacyTenant, error) {
	args := m.Called(ctx, slug)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*LegacyTenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthRepository) ListTenants(ctx context.Context) ([]LegacyTenant, error) {
	args := m.Called(ctx)
	if tenants := args.Get(0); tenants != nil {
		return tenants.([]LegacyTenant), args.Error(1)
	}
	return nil, args.Error(1)
}

// Transaction support
func (m *MockAuthRepository) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return nil, args.Error(0)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func stringPtr(s string) *string {
	return &s
}

func rolePtr(r UserRole) *UserRole {
	return &r
}

func boolPtr(b bool) *bool {
	return &b
}

// ============================================================================
// SERVICE TESTS - USER MANAGEMENT
// ============================================================================

func TestService_GetUserStats_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	expectedStats := &UserStats{
		TotalUsers:       100,
		ActiveUsers:      95,
		CustomerContacts: 25,
		EnterpriseUsers:  5,
	}

	// Mock repository calls
	mockRepo.On("GetUserStats", mock.Anything).Return(expectedStats, nil)

	// Execute
	stats, err := service.GetUserStats(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalUsers)
	assert.Equal(t, 95, stats.ActiveUsers)
	assert.Equal(t, 25, stats.CustomerContacts)
	assert.Equal(t, 5, stats.EnterpriseUsers)

	mockRepo.AssertExpectations(t)
}

func TestService_GetUsersByTenant_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	expectedUsers := []User{
		{ID: 1, Email: "user1@houston.com", PrimaryTenantID: "houston"},
		{ID: 2, Email: "user2@houston.com", PrimaryTenantID: "houston"},
	}

	// Mock repository calls
	mockRepo.On("GetUsersByTenant", mock.Anything, "houston").Return(expectedUsers, nil)

	// Execute
	users, err := service.GetUsersByTenant(context.Background(), "houston")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, expectedUsers[0].Email, users[0].Email)
	assert.Equal(t, expectedUsers[1].Email, users[1].Email)

	mockRepo.AssertExpectations(t)
}

func TestService_GetCustomerContacts_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	customerID := 123
	expectedContacts := []User{
		{
			ID:          1,
			Email:       "primary@customer.com",
			Role:        RoleCustomerContact,
			CustomerID:  &customerID,
			ContactType: ContactPrimary,
		},
		{
			ID:          2,
			Email:       "billing@customer.com",
			Role:        RoleCustomerContact,
			CustomerID:  &customerID,
			ContactType: ContactBilling,
		},
	}

	// Mock repository calls
	mockRepo.On("GetCustomerContacts", mock.Anything, customerID).Return(expectedContacts, nil)

	// Execute
	contacts, err := service.GetCustomerContacts(context.Background(), customerID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 2, len(contacts))
	assert.Equal(t, RoleCustomerContact, contacts[0].Role)
	assert.Equal(t, RoleCustomerContact, contacts[1].Role)
	assert.Equal(t, ContactPrimary, contacts[0].ContactType)
	assert.Equal(t, ContactBilling, contacts[1].ContactType)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// SERVICE TESTS - TENANT ACCESS MANAGEMENT
// ============================================================================

func TestService_UpdateUserTenantAccess_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	req := UpdateUserTenantAccessRequest{
		UserID: 1,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleManager,
				Permissions: []Permission{
					PermissionViewInventory,
					PermissionCreateWorkOrder,
					PermissionApproveWorkOrder,
				},
			},
		},
	}

	// Mock repository calls
	mockRepo.On("UpdateUserTenantAccess", mock.Anything, req.UserID, req.TenantAccess).Return(nil)

	// Execute
	err := service.UpdateUserTenantAccess(context.Background(), req)

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUserYardAccess_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	req := UpdateUserYardAccessRequest{
		UserID:   1,
		TenantID: "houston",
		YardAccess: []YardAccess{
			{
				YardLocation:       "houston_north",
				CanViewWorkOrders:  true,
				CanCreateWorkOrders: true,
				CanApproveOrders:   false,
				CanViewInventory:   true,
				CanManageTransport: false,
				CanExportData:      true,
			},
		},
	}

	// Mock repository calls
	mockRepo.On("UpdateUserYardAccess", mock.Anything, req.UserID, req.TenantID, req.YardAccess).Return(nil)

	// Execute
	err := service.UpdateUserYardAccess(context.Background(), req)

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_GetUserPermissions_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	expectedPermissions := []Permission{
		PermissionViewInventory,
		PermissionCreateWorkOrder,
		PermissionApproveWorkOrder,
	}

	testUser := &User{
		ID: 1,
		TenantAccess: []TenantAccess{
			{
				TenantID:    "houston",
				Permissions: expectedPermissions,
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	// Execute
	permissions, err := service.GetUserPermissions(context.Background(), 1, "houston")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, len(expectedPermissions), len(permissions))
	assert.Contains(t, permissions, PermissionViewInventory)
	assert.Contains(t, permissions, PermissionCreateWorkOrder)
	assert.Contains(t, permissions, PermissionApproveWorkOrder)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestService_UpdateUser_UserNotFound(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	updates := UserUpdates{
		Email: stringPtr("new@example.com"),
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 999).Return(nil, ErrUserNotFound)

	// Execute
	user, err := service.UpdateUser(context.Background(), 999, updates)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateEnterpriseUser_InvalidRole(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	req := CreateEnterpriseUserRequest{
		Username:         "admin@company.com",
		Email:            "admin@company.com",
		FullName:         "Enterprise Admin",
		Password:         "password123",
		Role:             "INVALID_ROLE", // Invalid role
		IsEnterpriseUser: true,
		PrimaryTenantID:  "houston",
		TenantAccess:     []TenantAccess{},
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)

	// Execute
	user, err := service.CreateEnterpriseUser(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidUserRole, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// INTEGRATION TESTS - ADMIN CONTACT REGISTRATION FLOW
// ============================================================================

func TestService_AdminRegisterCustomerContact_CompleteFlow(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	// Step 1: Admin authentication
	adminUser := &User{
		ID:               1,
		Role:             RoleAdmin,
		IsEnterpriseUser: true,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleAdmin,
				Permissions: []Permission{
					PermissionUserManagement,
				},
			},
		},
	}

	// Step 2: Admin creates customer contact
	contactReq := CreateCustomerContactRequest{
		CustomerID:  456,
		TenantID:    "houston",
		Email:       "newcontact@customer.com",
		FullName:    "New Customer Contact",
		Password:    "tempPassword123",
		ContactType: ContactPrimary,
		YardAccess: []YardAccess{
			{
				YardLocation:       "houston_north",
				CanViewWorkOrders:  true,
				CanCreateWorkOrders: true,
				CanViewInventory:   true,
			},
		},
	}

	newContact := &User{
		ID:              2,
		Email:           contactReq.Email,
		FullName:        contactReq.FullName,
		Role:            RoleCustomerContact,
		CustomerID:      &contactReq.CustomerID,
		PrimaryTenantID: contactReq.TenantID,
	}

	// Mock repository calls for admin permission check
	adminCheck := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionUserManagement,
	}
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(adminUser, nil)

	// Mock repository calls for contact creation
	mockRepo.On("GetUserByEmail", mock.Anything, contactReq.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(newContact, nil)

	// Execute admin permission check
	err := service.CheckPermission(context.Background(), adminCheck)
	assert.NoError(t, err)

	// Execute contact creation
	createdContact, err := service.CreateCustomerContact(context.Background(), contactReq)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdContact)
	assert.Equal(t, RoleCustomerContact, createdContact.Role)
	assert.Equal(t, contactReq.CustomerID, *createdContact.CustomerID)
	assert.Equal(t, contactReq.Email, createdContact.Email)

	mockRepo.AssertExpectations(t)
}

func TestService_AdminBulkContactRegistration_MultipleContacts(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	// Simulate admin registering multiple contacts for a customer
	customerID := 789
	contacts := []CreateCustomerContactRequest{
		{
			CustomerID:  customerID,
			TenantID:    "houston",
			Email:       "primary@customer.com",
			FullName:    "Primary Contact",
			Password:    "password123",
			ContactType: ContactPrimary,
		},
		{
			CustomerID:  customerID,
			TenantID:    "houston",
			Email:       "billing@customer.com",
			FullName:    "Billing Contact",
			Password:    "password456",
			ContactType: ContactBilling,
		},
		{
			CustomerID:  customerID,
			TenantID:    "houston",
			Email:       "shipping@customer.com",
			FullName:    "Shipping Contact",
			Password:    "password789",
			ContactType: ContactShipping,
		},
	}

	// Mock repository calls for each contact creation
	for i, req := range contacts {
		expectedUser := &User{
			ID:              i + 1,
			Email:           req.Email,
			FullName:        req.FullName,
			Role:            RoleCustomerContact,
			CustomerID:      &req.CustomerID,
			ContactType:     req.ContactType,
			PrimaryTenantID: req.TenantID,
		}

		mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)
		mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(expectedUser, nil)
	}

	// Execute creation of all contacts
	var createdContacts []*User
	for _, req := range contacts {
		contact, err := service.CreateCustomerContact(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, contact)
		createdContacts = append(createdContacts, contact)
	}

	// Verify all contacts were created with correct properties
	assert.Equal(t, 3, len(createdContacts))
	assert.Equal(t, ContactPrimary, createdContacts[0].ContactType)
	assert.Equal(t, ContactBilling, createdContacts[1].ContactType)
	assert.Equal(t, ContactShipping, createdContacts[2].ContactType)

	// All should belong to the same customer
	for _, contact := range createdContacts {
		assert.Equal(t, customerID, *contact.CustomerID)
		assert.Equal(t, RoleCustomerContact, contact.Role)
	}

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// PERFORMANCE BENCHMARKS
// ============================================================================

func BenchmarkService_CheckPermission(b *testing.B) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	testUser := &User{
		ID: 1,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Permissions: []Permission{
					PermissionViewInventory,
					PermissionCreateWorkOrder,
				},
			},
		},
	}

	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionViewInventory,
	}

	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CheckPermission(context.Background(), check)
	}
}

func BenchmarkService_Authenticate(b *testing.B) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	hashedPassword, _ := hashPassword("password123")
	testUser := &User{
		ID:              1,
		Email:           "test@example.com",
		PasswordHash:    hashedPassword,
		Role:            RoleOperator,
		IsActive:        true,
		PrimaryTenantID: "houston",
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleOperator,
			},
		},
	}

	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		TenantID: "houston",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(testUser, nil)
	mockRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*auth.Session")).Return(nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Authenticate(context.Background(), req)
	}
}CreateCustomerContact_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

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

	expectedUser := &User{
		ID:              1,
		Email:           req.Email,
		FullName:        req.FullName,
		Role:            RoleCustomerContact,
		CustomerID:      &req.CustomerID,
		PrimaryTenantID: req.TenantID,
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(expectedUser, nil)

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

func TestService_CreateCustomerContact_UserExists(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	req := CreateCustomerContactRequest{
		CustomerID: 123,
		TenantID:   "houston",
		Email:      "existing@customer.com",
		FullName:   "Existing User",
		Password:   "password123",
	}

	existingUser := &User{
		ID:    1,
		Email: "existing@customer.com",
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	// Execute
	user, err := service.CreateCustomerContact(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateEnterpriseUser_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

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

	expectedUser := &User{
		ID:               1,
		Username:         req.Username,
		Email:            req.Email,
		FullName:         req.FullName,
		Role:             req.Role,
		IsEnterpriseUser: req.IsEnterpriseUser,
		PrimaryTenantID:  req.PrimaryTenantID,
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(expectedUser, nil)

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

func TestService_UpdateUser_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	existingUser := &User{
		ID:       1,
		Email:    "old@example.com",
		FullName: "Old Name",
		Role:     RoleOperator,
		IsActive: true,
	}

	updates := UserUpdates{
		Email:    stringPtr("new@example.com"),
		FullName: stringPtr("New Name"),
		Role:     rolePtr(RoleManager),
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(existingUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	// Execute
	result, err := service.UpdateUser(context.Background(), 1, updates)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new@example.com", result.Email)
	assert.Equal(t, "New Name", result.FullName)
	assert.Equal(t, RoleManager, result.Role)

	mockRepo.AssertExpectations(t)
}

func TestService_DeactivateUser_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	existingUser := &User{
		ID:       1,
		Email:    "user@example.com",
		IsActive: true,
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(existingUser, nil)
	mockRepo.On("InvalidateUserSessions", mock.Anything, 1).Return(nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	// Execute
	err := service.DeactivateUser(context.Background(), 1)

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// SERVICE TESTS - AUTHENTICATION
// ============================================================================

func TestService_Authenticate_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	// Create test user with hashed password
	hashedPassword, _ := hashPassword("password123")
	testUser := &User{
		ID:              1,
		Username:        "test@example.com",
		Email:           "test@example.com",
		PasswordHash:    hashedPassword,
		Role:            RoleOperator,
		IsActive:        true,
		PrimaryTenantID: "houston",
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleOperator,
			},
		},
	}

	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		TenantID: "houston",
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(testUser, nil)
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
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	req := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(nil, ErrUserNotFound)

	// Execute
	response, err := service.Authenticate(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestService_Authenticate_UserInactive(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	hashedPassword, _ := hashPassword("password123")
	inactiveUser := &User{
		ID:           1,
		Email:        "inactive@example.com",
		PasswordHash: hashedPassword,
		IsActive:     false, // User is inactive
	}

	req := LoginRequest{
		Email:    "inactive@example.com",
		Password: "password123",
	}

	// Mock repository calls
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(inactiveUser, nil)

	// Execute
	response, err := service.Authenticate(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrUserInactive, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// SERVICE TESTS - PERMISSION & ACCESS CONTROL
// ============================================================================

func TestService_CheckPermission_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	testUser := &User{
		ID:   1,
		Role: RoleAdmin,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleAdmin,
				Permissions: []Permission{
					PermissionUserManagement,
					PermissionApproveWorkOrder,
				},
			},
		},
	}

	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionUserManagement,
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	// Execute
	err := service.CheckPermission(context.Background(), check)

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_CheckPermission_Denied(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	testUser := &User{
		ID:   1,
		Role: RoleOperator,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				Role:     RoleOperator,
				Permissions: []Permission{
					PermissionViewInventory,
				},
			},
		},
	}

	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionUserManagement, // User doesn't have this permission
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	// Execute
	err := service.CheckPermission(context.Background(), check)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrPermissionDenied, err)

	mockRepo.AssertExpectations(t)
}

func TestService_CheckYardAccess_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	testUser := &User{
		ID: 1,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				YardAccess: []YardAccess{
					{
						YardLocation:     "houston_north",
						CanViewInventory: true,
					},
				},
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	// Execute
	err := service.CheckYardAccess(context.Background(), 1, "houston", "houston_north")

	// Assertions
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_CheckYardAccess_Denied(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	testUser := &User{
		ID: 1,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				YardAccess: []YardAccess{
					{
						YardLocation:     "houston_north",
						CanViewInventory: true,
					},
				},
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	// Execute - try to access yard user doesn't have access to
	err := service.CheckYardAccess(context.Background(), 1, "houston", "houston_south")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrYardAccessDenied, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// SERVICE TESTS - ENTERPRISE OPERATIONS
// ============================================================================

func TestService_GetEnterpriseContext_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	enterpriseUser := &User{
		ID:               2,
		Role:             RoleEnterpriseAdmin,
		IsEnterpriseUser: true,
		TenantAccess: []TenantAccess{
			{TenantID: "houston", Role: RoleEnterpriseAdmin},
			{TenantID: "longbeach", Role: RoleEnterpriseAdmin},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 2).Return(enterpriseUser, nil)

	// Execute
	context, err := service.GetEnterpriseContext(context.Background(), 2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, context)
	assert.Equal(t, 2, context.UserID)
	assert.True(t, context.IsEnterpriseAdmin)
	assert.Equal(t, 2, len(context.AccessibleTenants))

	mockRepo.AssertExpectations(t)
}

func TestService_GetEnterpriseContext_AccessDenied(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	regularUser := &User{
		ID:               1,
		Role:             RoleOperator,
		IsEnterpriseUser: false,
		TenantAccess: []TenantAccess{
			{TenantID: "houston", Role: RoleOperator},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(regularUser, nil)

	// Execute
	context, err := service.GetEnterpriseContext(context.Background(), 1)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrEnterpriseAccessDenied, err)
	assert.Nil(t, context)

	mockRepo.AssertExpectations(t)
}

func TestService_GetCustomerAccessContext_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	customerID := 123
	customerContactUser := &User{
		ID:         1,
		Role:       RoleCustomerContact,
		CustomerID: &customerID,
		TenantAccess: []TenantAccess{
			{
				TenantID: "houston",
				YardAccess: []YardAccess{
					{
						YardLocation:       "houston_north",
						CanViewWorkOrders:  true,
						CanCreateWorkOrders: true,
						CanViewInventory:   true,
					},
				},
			},
		},
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(customerContactUser, nil)

	// Execute
	context, err := service.GetCustomerAccessContext(context.Background(), 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, context)
	assert.Equal(t, customerID, context.CustomerID)
	assert.Equal(t, 1, len(context.AccessibleYards))
	assert.Equal(t, 1, len(context.TenantAccess))

	mockRepo.AssertExpectations(t)
}

func TestService_GetCustomerAccessContext_NotCustomerContact(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	enterpriseUser := &User{
		ID:               1,
		Role:             RoleAdmin,
		IsEnterpriseUser: true,
		CustomerID:       nil, // Not a customer contact
	}

	// Mock repository calls
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(enterpriseUser, nil)

	// Execute
	context, err := service.GetCustomerAccessContext(context.Background(), 1)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrNotCustomerContact, err)
	assert.Nil(t, context)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// SERVICE TESTS - SEARCH & ANALYTICS
// ============================================================================

func TestService_SearchUsers_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	service := NewService(mockRepo, []byte("test-jwt-secret"), time.Hour)

	filters := UserSearchFilters{
		Query:    "test",
		TenantID: "houston",
		Role:     RoleOperator,
		Limit:    10,
		Offset:   0,
	}

	expectedUsers := []User{
		{ID: 1, Email: "test1@example.com", Role: RoleOperator},
		{ID: 2, Email: "test2@example.com", Role: RoleOperator},
	}

	// Mock repository calls
	mockRepo.On("SearchUsers", mock.Anything, filters).Return(expectedUsers, 2, nil)

	// Execute
	users, total, err := service.SearchUsers(context.Background(), filters)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, 2, total)
	assert.Equal(t, expectedUsers[0].Email, users[0].Email)

	mockRepo.AssertExpectations(t)
}

func TestService_

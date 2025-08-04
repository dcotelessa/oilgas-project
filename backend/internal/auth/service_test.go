// backend/internal/auth/service_test.go
package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		user.ID = 1 // Simulate database assignment
	}
	return args.Error(0)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error {
	args := m.Called(ctx, userID, tenantAccess)
	return args.Error(0)
}

func (m *MockRepository) UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error {
	args := m.Called(ctx, userID, tenantID, yardAccess)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) CreateSession(ctx context.Context, session *Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	args := m.Called(ctx, token)
	if session := args.Get(0); session != nil {
		return session.(*Session), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockRepository) InvalidateSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRepository) InvalidateUserSessions(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) GetEnterpriseUsers(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) GetUsersByRole(ctx context.Context, role UserRole) ([]User, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	return args.Get(0).([]Permission), args.Error(1)
}

func (m *MockRepository) GetRolePermissions(ctx context.Context, role UserRole) ([]Permission, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]Permission), args.Error(1)
}

func (m *MockRepository) GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error) {
	args := m.Called(ctx, tenantID, yardLocation)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockRepository) ValidateCustomerExists(ctx context.Context, customerID int) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockRepository) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return nil, args.Error(0)
}

func (m *MockRepository) GetUserStats(ctx context.Context) (*UserStats, error) {
	args := m.Called(ctx)
	if stats := args.Get(0); stats != nil {
		return stats.(*UserStats), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]User), args.Error(1)
}

// Test helper functions
func createTestUser() *User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	customerID := 123
	
	return &User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		FullName:     "Test User",
		PasswordHash: string(hashedPassword),
		Role:         RoleCustomerContact,
		AccessLevel:  1,
		CustomerID:   &customerID,
		ContactType:  ContactPrimary,
		PrimaryTenantID: "houston",
		TenantAccess: TenantAccessList{
			{
				TenantID: "houston",
				Role:     RoleCustomerContact,
				Permissions: []Permission{PermissionWorkOrderRead, PermissionInventoryRead},
				YardAccess: []YardAccess{
					{
						YardLocation:        "houston_north",
						CanViewWorkOrders:   true,
						CanViewInventory:    true,
						CanCreateWorkOrders: true,
						CanApproveOrders:    false,
					},
				},
				CanRead:  true,
				CanWrite: true,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
	}
}

func createTestEnterpriseUser() *User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
	
	return &User{
		ID:               2,
		Username:         "admin",
		Email:            "admin@company.com",
		FullName:         "Enterprise Admin",
		PasswordHash:     string(hashedPassword),
		Role:             RoleEnterpriseAdmin,
		AccessLevel:      5,
		IsEnterpriseUser: true,
		PrimaryTenantID:  "houston",
		TenantAccess: TenantAccessList{
			{
				TenantID:    "houston",
				Role:        RoleEnterpriseAdmin,
				Permissions: []Permission{PermissionCrossTenantView, PermissionUserManagement},
				CanRead:     true,
				CanWrite:    true,
				CanDelete:   true,
				CanApprove:  true,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
	}
}

// Authentication tests
func TestAuthenticate_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil)
	mockRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*auth.Session")).Return(nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	
	req := LoginRequest{
		Username: "testuser",
		Password: "testpassword",
		TenantID: "houston",
	}
	
	response, err := service.Authenticate(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, testUser.Username, response.User.Username)
	assert.Equal(t, "houston", response.TenantContext.TenantID)
	assert.NotEmpty(t, response.Token)
	
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil)
	
	req := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
		TenantID: "houston",
	}
	
	response, err := service.Authenticate(context.Background(), req)
	
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)
	
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_InactiveUser(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	testUser.IsActive = false
	mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil)
	
	req := LoginRequest{
		Username: "testuser",
		Password: "testpassword",
		TenantID: "houston",
	}
	
	response, err := service.Authenticate(context.Background(), req)
	
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrUserInactive, err)
	
	mockRepo.AssertExpectations(t)
}

// Customer contact creation tests
func TestCreateCustomerContact_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	mockRepo.On("ValidateCustomerExists", mock.Anything, 123).Return(nil)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	
	req := CreateCustomerContactRequest{
		CustomerID:  123,
		TenantID:    "houston",
		Email:       "contact@customer.com",
		FullName:    "Customer Contact",
		Password:    "password123",
		ContactType: ContactPrimary,
		YardAccess: []YardAccess{
			{
				YardLocation:        "houston_north",
				CanViewWorkOrders:   true,
				CanViewInventory:    true,
				CanCreateWorkOrders: true,
			},
		},
	}
	
	user, err := service.CreateCustomerContact(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, RoleCustomerContact, user.Role)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, &req.CustomerID, user.CustomerID)
	assert.Equal(t, req.ContactType, user.ContactType)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateCustomerContact_CustomerNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	mockRepo.On("ValidateCustomerExists", mock.Anything, 123).Return(ErrCustomerNotFound)
	
	req := CreateCustomerContactRequest{
		CustomerID:  123,
		TenantID:    "houston",
		Email:       "contact@customer.com",
		FullName:    "Customer Contact",
		Password:    "password123",
		ContactType: ContactPrimary,
	}
	
	user, err := service.CreateCustomerContact(context.Background(), req)
	
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrCustomerNotFound, err)
	
	mockRepo.AssertExpectations(t)
}

// Enterprise user tests
func TestCreateEnterpriseUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	
	req := CreateEnterpriseUserRequest{
		Username:         "enterprise_admin",
		Email:            "admin@company.com",
		FullName:         "Enterprise Admin",
		Password:         "adminpassword",
		Role:             RoleEnterpriseAdmin,
		IsEnterpriseUser: true,
		PrimaryTenantID:  "houston",
		TenantAccess: TenantAccessList{
			{
				TenantID:    "houston",
				Role:        RoleEnterpriseAdmin,
				Permissions: []Permission{PermissionCrossTenantView},
				CanRead:     true,
				CanWrite:    true,
			},
		},
	}
	
	user, err := service.CreateEnterpriseUser(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, RoleEnterpriseAdmin, user.Role)
	assert.True(t, user.IsEnterpriseUser)
	assert.Equal(t, 5, user.AccessLevel)
	
	mockRepo.AssertExpectations(t)
}

// Permission checking tests
func TestCheckPermission_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionWorkOrderRead,
	}
	
	err := service.CheckPermission(context.Background(), check)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCheckPermission_Denied(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	check := UserPermissionCheck{
		UserID:     1,
		TenantID:   "houston",
		Permission: PermissionCustomerDelete, // Not granted to customer contact
	}
	
	err := service.CheckPermission(context.Background(), check)
	
	assert.Error(t, err)
	assert.Equal(t, ErrPermissionDenied, err)
	mockRepo.AssertExpectations(t)
}

// Yard access tests
func TestCheckYardAccess_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	err := service.CheckYardAccess(context.Background(), 1, "houston", "houston_north")
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCheckYardAccess_Denied(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	err := service.CheckYardAccess(context.Background(), 1, "houston", "houston_south")
	
	assert.Error(t, err)
	assert.Equal(t, ErrYardAccessDenied, err)
	mockRepo.AssertExpectations(t)
}

// Enterprise context tests
func TestGetEnterpriseContext_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestEnterpriseUser()
	mockRepo.On("GetUserByID", mock.Anything, 2).Return(testUser, nil)
	
	context, err := service.GetEnterpriseContext(context.Background(), 2)
	
	assert.NoError(t, err)
	assert.NotNil(t, context)
	assert.True(t, context.IsEnterpriseAdmin)
	assert.Equal(t, 2, context.UserID)
	assert.Contains(t, context.CrossTenantPerms, PermissionCrossTenantView)
	
	mockRepo.AssertExpectations(t)
}

func TestGetEnterpriseContext_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser() // Regular customer contact
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	context, err := service.GetEnterpriseContext(context.Background(), 1)
	
	assert.Error(t, err)
	assert.Nil(t, context)
	assert.Equal(t, ErrEnterpriseAccessDenied, err)
	
	mockRepo.AssertExpectations(t)
}

// Customer access context tests
func TestGetCustomerAccessContext_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestUser()
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	
	context, err := service.GetCustomerAccessContext(context.Background(), 1)
	
	assert.NoError(t, err)
	assert.NotNil(t, context)
	assert.Equal(t, *testUser.CustomerID, context.CustomerID)
	assert.Len(t, context.AccessibleYards, 1)
	assert.Equal(t, "houston_north", context.AccessibleYards[0].YardLocation)
	
	mockRepo.AssertExpectations(t)
}

func TestGetCustomerAccessContext_NotCustomerContact(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	testUser := createTestEnterpriseUser()
	mockRepo.On("GetUserByID", mock.Anything, 2).Return(testUser, nil)
	
	context, err := service.GetCustomerAccessContext(context.Background(), 2)
	
	assert.Error(t, err)
	assert.Nil(t, context)
	assert.Equal(t, ErrNotCustomerContact, err)
	
	mockRepo.AssertExpectations(t)
}

// User model helper method tests
func TestUser_HasAccessToTenant(t *testing.T) {
	testUser := createTestUser()
	
	// Should have access to houston
	assert.True(t, testUser.HasAccessToTenant("houston"))
	
	// Should not have access to longbeach
	assert.False(t, testUser.HasAccessToTenant("longbeach"))
	
	// Enterprise user should have access to any tenant
	enterpriseUser := createTestEnterpriseUser()
	assert.True(t, enterpriseUser.HasAccessToTenant("any_tenant"))
}

func TestUser_HasAccessToYard(t *testing.T) {
	testUser := createTestUser()
	
	// Should have access to houston_north
	assert.True(t, testUser.HasAccessToYard("houston", "houston_north"))
	
	// Should not have access to houston_south
	assert.False(t, testUser.HasAccessToYard("houston", "houston_south"))
	
	// Should not have access to different tenant
	assert.False(t, testUser.HasAccessToYard("longbeach", "longbeach_main"))
}

func TestUser_HasPermissionInTenant(t *testing.T) {
	testUser := createTestUser()
	
	// Should have work order read permission
	assert.True(t, testUser.HasPermissionInTenant("houston", PermissionWorkOrderRead))
	
	// Should not have delete permission
	assert.False(t, testUser.HasPermissionInTenant("houston", PermissionCustomerDelete))
	
	// System admin should have all permissions
	systemAdmin := createTestEnterpriseUser()
	systemAdmin.Role = RoleSystemAdmin
	assert.True(t, systemAdmin.HasPermissionInTenant("houston", PermissionCustomerDelete))
}

func TestUser_IsCustomerContact(t *testing.T) {
	testUser := createTestUser()
	assert.True(t, testUser.IsCustomerContact())
	
	enterpriseUser := createTestEnterpriseUser()
	assert.False(t, enterpriseUser.IsCustomerContact())
}

func TestUser_CanPerformCrossTenantOperation(t *testing.T) {
	testUser := createTestUser()
	assert.False(t, testUser.CanPerformCrossTenantOperation())
	
	enterpriseUser := createTestEnterpriseUser()
	assert.True(t, enterpriseUser.CanPerformCrossTenantOperation())
}

// Yard access validation tests
func TestValidateYardAccess_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret").(*service)
	
	yardAccess := []YardAccess{
		{
			YardLocation:        "houston_north",
			CanViewWorkOrders:   true,
			CanViewInventory:    true,
			CanCreateWorkOrders: false,
		},
	}
	
	err := service.validateYardAccess(yardAccess)
	assert.NoError(t, err)
}

func TestValidateYardAccess_EmptyLocation(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret").(*service)
	
	yardAccess := []YardAccess{
		{
			YardLocation:        "", // Empty location
			CanViewWorkOrders:   true,
			CanViewInventory:    true,
		},
	}
	
	err := service.validateYardAccess(yardAccess)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidYardAccess, err)
}

func TestValidateYardAccess_NoPermissions(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret").(*service)
	
	yardAccess := []YardAccess{
		{
			YardLocation:        "houston_north",
			CanViewWorkOrders:   false,
			CanViewInventory:    false,
			CanCreateWorkOrders: false,
			CanApproveOrders:    false,
			CanManageTransport:  false,
			CanExportData:       false,
		},
	}
	
	err := service.validateYardAccess(yardAccess)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidYardAccess, err)
}

// Integration tests
func TestMultiTenantCustomerContactAccess(t *testing.T) {
	// Test scenario: Customer contact with access to multiple yards within tenant
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	// Create customer contact with multi-yard access
	req := CreateCustomerContactRequest{
		CustomerID:  123,
		TenantID:    "houston",
		Email:       "contact@customer.com",
		FullName:    "Customer Contact",
		Password:    "password123",
		ContactType: ContactPrimary,
		YardAccess: []YardAccess{
			{
				YardLocation:        "houston_north",
				CanViewWorkOrders:   true,
				CanViewInventory:    true,
				CanCreateWorkOrders: true,
			},
			{
				YardLocation:        "houston_south",
				CanViewWorkOrders:   true,
				CanViewInventory:    false,
				CanCreateWorkOrders: false,
			},
		},
	}
	
	mockRepo.On("ValidateCustomerExists", mock.Anything, 123).Return(nil)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	
	user, err := service.CreateCustomerContact(context.Background(), req)
	
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Len(t, user.TenantAccess[0].YardAccess, 2)
	
	// Test yard access for different yards
	assert.True(t, user.HasAccessToYard("houston", "houston_north"))
	assert.True(t, user.HasAccessToYard("houston", "houston_south"))
	assert.False(t, user.HasAccessToYard("houston", "houston_west"))
	
	// Test specific yard permissions
	yardPermissions := user.TenantAccess[0].YardAccess
	northYard := yardPermissions[0]
	southYard := yardPermissions[1]
	
	assert.True(t, northYard.CanCreateWorkOrders)
	assert.False(t, southYard.CanCreateWorkOrders)
	assert.True(t, northYard.CanViewInventory)
	assert.False(t, southYard.CanViewInventory)
	
	mockRepo.AssertExpectations(t)
}

func TestEnterpriseAdminCrossTenantAccess(t *testing.T) {
	// Test scenario: Enterprise admin accessing multiple tenants
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, "test-secret")
	
	enterpriseUser := createTestEnterpriseUser()
	
	// Add access to multiple tenants
	enterpriseUser.TenantAccess = append(enterpriseUser.TenantAccess, TenantAccess{
		TenantID:    "longbeach",
		Role:        RoleEnterpriseAdmin,
		Permissions: []Permission{PermissionCrossTenantView, PermissionUserManagement},
		CanRead:     true,
		CanWrite:    true,
		CanDelete:   true,
	})
	
	mockRepo.On("GetUserByID", mock.Anything, 2).Return(enterpriseUser, nil)
	
	// Test cross-tenant access
	context, err := service.GetEnterpriseContext(context.Background(), 2)
	
	assert.NoError(t, err)
	assert.NotNil(t, context)
	assert.True(t, context.IsEnterpriseAdmin)
	assert.Len(t, context.AccessibleTenants, 2)
	assert.Contains(t, context.CrossTenantPerms, PermissionCrossTenantView)
	assert.Contains(t, context.CrossTenantPerms, PermissionUserManagement)
	
	mockRepo.AssertExpectations(t)
}

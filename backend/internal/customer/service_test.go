// backend/internal/customer/service_test.go
package customer

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// Mock repository for testing
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) CreateCustomer(ctx context.Context, customer *Customer) (*Customer, error) {
    args := m.Called(ctx, customer)
    return args.Get(0).(*Customer), args.Error(1)
}

func (m *MockRepository) GetCustomerByID(ctx context.Context, tenantID string, id int) (*Customer, error) {
    args := m.Called(ctx, tenantID, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Customer), args.Error(1)
}

func (m *MockRepository) UpdateCustomer(ctx context.Context, customer *Customer) error {
    args := m.Called(ctx, customer)
    return args.Error(0)
}

func (m *MockRepository) DeleteCustomer(ctx context.Context, tenantID string, id int) error {
    args := m.Called(ctx, tenantID, id)
    return args.Error(0)
}

func (m *MockRepository) ListCustomers(ctx context.Context, tenantID string, filters CustomerSearchFilters) ([]Customer, int, error) {
    args := m.Called(ctx, tenantID, filters)
    return args.Get(0).([]Customer), args.Int(1), args.Error(2)
}

func (m *MockRepository) SearchCustomers(ctx context.Context, tenantID, query string) ([]Customer, error) {
    args := m.Called(ctx, tenantID, query)
    return args.Get(0).([]Customer), args.Error(1)
}

func (m *MockRepository) GetCustomerWithContacts(ctx context.Context, tenantID string, customerID int) (*CustomerWithContacts, error) {
    args := m.Called(ctx, tenantID, customerID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*CustomerWithContacts), args.Error(1)
}

func (m *MockRepository) ValidateCustomerExists(ctx context.Context, tenantID string, customerID int) error {
    args := m.Called(ctx, tenantID, customerID)
    return args.Error(0)
}

func (m *MockRepository) GetCustomerAnalytics(ctx context.Context, tenantID string) (*CustomerAnalytics, error) {
    args := m.Called(ctx, tenantID)
    return args.Get(0).(*CustomerAnalytics), args.Error(1)
}

func (m *MockRepository) CreateCustomerContact(ctx context.Context, contact *CustomerAuthContact) (*CustomerAuthContact, error) {
    args := m.Called(ctx, contact)
    return args.Get(0).(*CustomerAuthContact), args.Error(1)
}

func (m *MockRepository) UpdateContactPermissions(ctx context.Context, customerID, userID int, permissions []string) error {
    args := m.Called(ctx, customerID, userID, permissions)
    return args.Error(0)
}

func (m *MockRepository) DeactivateCustomerContact(ctx context.Context, customerID, userID int) error {
    args := m.Called(ctx, customerID, userID)
    return args.Error(0)
}

// Test suite setup
func setupTestService() *Service {
    mockRepo := new(MockRepository)
    return NewService(mockRepo)
}

// Customer CRUD Tests
func TestService_CreateCustomer(t *testing.T) {
    tests := []struct {
        name     string
        tenantID string
        request  CreateCustomerRequest
        mockSetup func(*MockRepository)
        wantError bool
        errorMsg  string
    }{
        {
            name:     "successful customer creation",
            tenantID: "test-tenant",
            request: CreateCustomerRequest{
                Name:        "Test Company",
                CompanyCode: stringPtr("TC001"),
                Email:       stringPtr("test@company.com"),
            },
            mockSetup: func(m *MockRepository) {
                expectedCustomer := &Customer{
                    ID:          1,
                    TenantID:    "test-tenant",
                    Name:        "Test Company",
                    CompanyCode: stringPtr("TC001"),
                    Email:       stringPtr("test@company.com"),
                    IsActive:    true,
                    CreatedAt:   time.Now(),
                    UpdatedAt:   time.Now(),
                }
                m.On("CreateCustomer", mock.Anything, mock.MatchedBy(func(c *Customer) bool {
                    return c.Name == "Test Company" && 
                           c.TenantID == "test-tenant" && 
                           c.IsActive == true
                })).Return(expectedCustomer, nil)
            },
            wantError: false,
        },
        {
            name:     "empty customer name",
            tenantID: "test-tenant",
            request: CreateCustomerRequest{
                Name: "",
            },
            mockSetup: func(m *MockRepository) {},
            wantError: true,
            errorMsg:  "customer name is required",
        },
        {
            name:     "empty tenant ID",
            tenantID: "",
            request: CreateCustomerRequest{
                Name: "Test Company",
            },
            mockSetup: func(m *MockRepository) {},
            wantError: true,
            errorMsg:  "tenant ID is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tt.mockSetup(mockRepo)
            
            service := NewService(mockRepo)
            ctx := context.Background()

            customer, err := service.CreateCustomer(ctx, tt.tenantID, tt.request)

            if tt.wantError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
                assert.Nil(t, customer)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, customer)
                assert.Equal(t, tt.request.Name, customer.Name)
                assert.Equal(t, tt.tenantID, customer.TenantID)
                assert.True(t, customer.IsActive)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}

func TestService_GetCustomer(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    expectedCustomer := &Customer{
        ID:       1,
        TenantID: "test-tenant",
        Name:     "Test Company",
        IsActive: true,
    }

    mockRepo.On("GetCustomerByID", ctx, "test-tenant", 1).Return(expectedCustomer, nil)

    customer, err := service.GetCustomer(ctx, "test-tenant", 1)

    assert.NoError(t, err)
    assert.NotNil(t, customer)
    assert.Equal(t, expectedCustomer.Name, customer.Name)
    mockRepo.AssertExpectations(t)
}

func TestService_GetCustomer_NotFound(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    mockRepo.On("GetCustomerByID", ctx, "test-tenant", 999).Return(nil, ErrCustomerNotFound)

    customer, err := service.GetCustomer(ctx, "test-tenant", 999)

    assert.Error(t, err)
    assert.Nil(t, customer)
    assert.Equal(t, ErrCustomerNotFound, err)
    mockRepo.AssertExpectations(t)
}

// Customer search and filtering tests
func TestService_SearchCustomers(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    expectedCustomers := []Customer{
        {ID: 1, Name: "Alpha Company", TenantID: "test-tenant"},
        {ID: 2, Name: "Beta Corporation", TenantID: "test-tenant"},
    }

    mockRepo.On("SearchCustomers", ctx, "test-tenant", "company").Return(expectedCustomers, nil)

    customers, err := service.SearchCustomers(ctx, "test-tenant", "company")

    assert.NoError(t, err)
    assert.Len(t, customers, 2)
    assert.Equal(t, "Alpha Company", customers[0].Name)
    mockRepo.AssertExpectations(t)
}

func TestService_ListCustomers_WithFilters(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    filters := CustomerSearchFilters{
        Query:    "test",
        IsActive: boolPtr(true),
        Page:     1,
        Limit:    10,
    }

    expectedCustomers := []Customer{
        {ID: 1, Name: "Test Company", TenantID: "test-tenant", IsActive: true},
    }

    mockRepo.On("ListCustomers", ctx, "test-tenant", filters).Return(expectedCustomers, 1, nil)

    customers, total, err := service.ListCustomers(ctx, "test-tenant", filters)

    assert.NoError(t, err)
    assert.Len(t, customers, 1)
    assert.Equal(t, 1, total)
    assert.Equal(t, "Test Company", customers[0].Name)
    mockRepo.AssertExpectations(t)
}

// Contact management tests
func TestService_RegisterCustomerContact(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    // Mock auth service should be injected in real implementation
    request := RegisterContactRequest{
        CustomerID:      1,
        Email:           "contact@customer.com",
        FullName:        "John Doe",
        ContactType:     "PRIMARY",
        YardPermissions: []string{"main-yard", "storage-yard"},
    }

    expectedContact := &CustomerAuthContact{
        ID:              1,
        CustomerID:      1,
        AuthUserID:      123,
        ContactType:     "PRIMARY",
        YardPermissions: []string{"main-yard", "storage-yard"},
        IsActive:        true,
    }

    mockRepo.On("ValidateCustomerExists", ctx, "test-tenant", 1).Return(nil)
    mockRepo.On("CreateCustomerContact", ctx, mock.MatchedBy(func(c *CustomerAuthContact) bool {
        return c.CustomerID == 1 && c.ContactType == "PRIMARY"
    })).Return(expectedContact, nil)

    // This test assumes auth service integration - will need to mock that too
    response, err := service.RegisterCustomerContact(ctx, "test-tenant", request)

    // Basic validation for now - full test requires auth service mock
    assert.NoError(t, err)
    assert.NotNil(t, response)
    mockRepo.AssertExpected(t, "ValidateCustomerExists")
}

// Customer analytics tests
func TestService_GetCustomerAnalytics(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    expectedAnalytics := &CustomerAnalytics{
        TotalCustomers:        100,
        ActiveCustomers:       85,
        CustomersWithContacts: 45,
        TotalContacts:         78,
        ContactsByType: map[string]int{
            "PRIMARY": 45,
            "BILLING": 20,
            "SHIPPING": 13,
        },
    }

    mockRepo.On("GetCustomerAnalytics", ctx, "test-tenant").Return(expectedAnalytics, nil)

    analytics, err := service.GetCustomerAnalytics(ctx, "test-tenant")

    assert.NoError(t, err)
    assert.NotNil(t, analytics)
    assert.Equal(t, 100, analytics.TotalCustomers)
    assert.Equal(t, 85, analytics.ActiveCustomers)
    assert.Equal(t, 45, analytics.ContactsByType["PRIMARY"])
    mockRepo.AssertExpectations(t)
}

// Multi-tenant isolation tests
func TestService_MultiTenantIsolation(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    // Test that operations are properly isolated by tenant
    mockRepo.On("GetCustomerByID", ctx, "tenant-a", 1).Return(&Customer{
        ID: 1, TenantID: "tenant-a", Name: "Tenant A Customer",
    }, nil)
    
    mockRepo.On("GetCustomerByID", ctx, "tenant-b", 1).Return(nil, ErrCustomerNotFound)

    // Same customer ID, different tenants should return different results
    customerA, errA := service.GetCustomer(ctx, "tenant-a", 1)
    customerB, errB := service.GetCustomer(ctx, "tenant-b", 1)

    assert.NoError(t, errA)
    assert.NotNil(t, customerA)
    assert.Equal(t, "tenant-a", customerA.TenantID)

    assert.Error(t, errB)
    assert.Nil(t, customerB)
    assert.Equal(t, ErrCustomerNotFound, errB)

    mockRepo.AssertExpectations(t)
}

// Edge case and validation tests
func TestService_ValidateCustomerExists(t *testing.T) {
    tests := []struct {
        name       string
        tenantID   string
        customerID int
        mockSetup  func(*MockRepository)
        wantError  bool
    }{
        {
            name:       "customer exists",
            tenantID:   "test-tenant",
            customerID: 1,
            mockSetup: func(m *MockRepository) {
                m.On("ValidateCustomerExists", mock.Anything, "test-tenant", 1).Return(nil)
            },
            wantError: false,
        },
        {
            name:       "customer does not exist",
            tenantID:   "test-tenant",
            customerID: 999,
            mockSetup: func(m *MockRepository) {
                m.On("ValidateCustomerExists", mock.Anything, "test-tenant", 999).Return(ErrCustomerNotFound)
            },
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tt.mockSetup(mockRepo)
            
            service := NewService(mockRepo)
            ctx := context.Background()

            err := service.ValidateCustomerExists(ctx, tt.tenantID, tt.customerID)

            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}

// Benchmark tests for performance validation
func BenchmarkService_GetCustomer(b *testing.B) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    customer := &Customer{
        ID: 1, TenantID: "test-tenant", Name: "Benchmark Customer",
    }

    mockRepo.On("GetCustomerByID", mock.Anything, mock.Anything, mock.Anything).Return(customer, nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = service.GetCustomer(ctx, "test-tenant", 1)
    }
}

// Helper functions
func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool { return &b }

// Test coverage: TestMain to setup and teardown
func TestMain(m *testing.M) {
    // Setup test database if needed for integration tests
    // For now, we're using mocks for unit tests
    m.Run()
}

func BenchmarkCache_GoInternal(b *testing.B) {
    cache := cache.NewTenantAwareCache(cache.CacheConfig{
        DefaultTTL: 15 * time.Minute,
        MaxEntries: 10000,
    })
    
    customer := &Customer{ID: 1, Name: "Test Co"}
    cache.CacheCustomer("houston", customer)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = cache.GetCustomer("houston", 1)
    }
}

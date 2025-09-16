// backend/internal/customer/service_test.go
package customer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetCustomerByID(ctx context.Context, tenantID string, id int) (*Customer, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *mockRepository) SearchCustomers(ctx context.Context, tenantID string, filters SearchFilters) ([]Customer, int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Get(0).([]Customer), args.Get(1).(int), args.Error(2)
}

func (m *mockRepository) CreateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	args := m.Called(ctx, tenantID, customer)
	if args.Error(0) == nil {
		customer.ID = 1
		customer.CreatedAt = time.Now()
		customer.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockRepository) UpdateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	args := m.Called(ctx, tenantID, customer)
	if args.Error(0) == nil {
		customer.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockRepository) DeleteCustomer(ctx context.Context, tenantID string, id int) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *mockRepository) GetCustomerContacts(ctx context.Context, tenantID string, customerID int) ([]CustomerContact, error) {
	args := m.Called(ctx, tenantID, customerID)
	return args.Get(0).([]CustomerContact), args.Error(1)
}

func (m *mockRepository) AddCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error {
	args := m.Called(ctx, tenantID, contact)
	if args.Error(0) == nil {
		contact.ID = 1
		contact.CreatedAt = time.Now()
		contact.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockRepository) UpdateCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error {
	args := m.Called(ctx, tenantID, contact)
	return args.Error(0)
}

func (m *mockRepository) RemoveCustomerContact(ctx context.Context, tenantID string, customerID, authUserID int) error {
	args := m.Called(ctx, tenantID, customerID, authUserID)
	return args.Error(0)
}

func (m *mockRepository) GetCustomerAnalytics(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	args := m.Called(ctx, tenantID, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CustomerAnalytics), args.Error(1)
}

type mockCacheService struct {
	mock.Mock
}

func (m *mockCacheService) GetCustomer(tenantID string, customerID int) (*Customer, bool) {
	args := m.Called(tenantID, customerID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*Customer), args.Bool(1)
}

func (m *mockCacheService) CacheCustomer(tenantID string, customer *Customer) {
	m.Called(tenantID, customer)
}

func (m *mockCacheService) InvalidateCustomer(tenantID string, customerID int) {
	m.Called(tenantID, customerID)
}

func (m *mockCacheService) InvalidateCustomerSearch(tenantID string, filters SearchFilters) {
	m.Called(tenantID, filters)
}

type CustomerServiceTestSuite struct {
	suite.Suite
	service  Service
	repo     *mockRepository
	cache    *mockCacheService
	ctx      context.Context
	tenantID string
}

func (suite *CustomerServiceTestSuite) SetupTest() {
	suite.repo = &mockRepository{}
	suite.cache = &mockCacheService{}
	suite.service = NewService(suite.repo, nil, suite.cache)
	suite.ctx = context.Background()
	suite.tenantID = "longbeach"
}

func TestCustomerServiceSuite(t *testing.T) {
	suite.Run(t, new(CustomerServiceTestSuite))
}

func (suite *CustomerServiceTestSuite) TestGetCustomer_Success_CacheHit() {
	customerID := 1
	expectedCustomer := &Customer{
		ID:       customerID,
		TenantID: suite.tenantID,
		Name:     "Test Company",
		Status:   StatusActive,
	}

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(expectedCustomer, true)

	result, err := suite.service.GetCustomer(suite.ctx, suite.tenantID, customerID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedCustomer, result)
	suite.cache.AssertExpectations(suite.T())
	suite.repo.AssertNotCalled(suite.T(), "GetCustomerByID")
}

func (suite *CustomerServiceTestSuite) TestGetCustomer_Success_CacheMiss() {
	customerID := 1
	expectedCustomer := &Customer{
		ID:       customerID,
		TenantID: suite.tenantID,
		Name:     "Test Company",
		Status:   StatusActive,
	}

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(nil, false)
	suite.repo.On("GetCustomerByID", suite.ctx, suite.tenantID, customerID).Return(expectedCustomer, nil)
	suite.cache.On("CacheCustomer", suite.tenantID, expectedCustomer).Return()

	result, err := suite.service.GetCustomer(suite.ctx, suite.tenantID, customerID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedCustomer, result)
	suite.cache.AssertExpectations(suite.T())
	suite.repo.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestGetCustomer_ValidationErrors() {
	testCases := []struct {
		name        string
		tenantID    string
		customerID  int
		expectError string
	}{
		{
			name:        "empty tenant ID",
			tenantID:    "",
			customerID:  1,
			expectError: "invalid tenant: tenant ID is required",
		},
		{
			name:        "invalid customer ID",
			tenantID:    suite.tenantID,
			customerID:  0,
			expectError: "invalid customer ID: 0",
		},
		{
			name:        "negative customer ID",
			tenantID:    suite.tenantID,
			customerID:  -1,
			expectError: "invalid customer ID: -1",
		},
		{
			name:        "tenant ID too long",
			tenantID:    string(make([]byte, 101)),
			customerID:  1,
			expectError: "tenant ID too long",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := suite.service.GetCustomer(suite.ctx, tc.tenantID, tc.customerID)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

func (suite *CustomerServiceTestSuite) TestGetCustomer_RepositoryError() {
	customerID := 1
	repoError := errors.New("database connection failed")

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(nil, false)
	suite.repo.On("GetCustomerByID", suite.ctx, suite.tenantID, customerID).Return(nil, repoError)

	result, err := suite.service.GetCustomer(suite.ctx, suite.tenantID, customerID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to get customer 1")
	suite.cache.AssertExpectations(suite.T())
	suite.repo.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestCreateCustomer_Success() {
	companyCode := "TEST123"
	taxID := "123456789"
	street := "123 Main St"
	city := "Houston"
	state := "TX"
	zipCode := "77001"
	
	customer := &Customer{
		Name:           "Test Company",
		CompanyCode:    &companyCode,
		TaxID:          &taxID,
		PaymentTerms:   "NET30",
		BillingStreet:  &street,
		BillingCity:    &city,
		BillingState:   &state,
		BillingZip:     &zipCode,
		BillingCountry: "US",
	}

	suite.repo.On("CreateCustomer", suite.ctx, suite.tenantID, customer).Return(nil)
	suite.cache.On("CacheCustomer", suite.tenantID, customer).Return()

	err := suite.service.CreateCustomer(suite.ctx, suite.tenantID, customer)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.tenantID, customer.TenantID)
	assert.Equal(suite.T(), StatusActive, customer.Status)
	assert.True(suite.T(), customer.IsActive)
	assert.Equal(suite.T(), "NET30", customer.PaymentTerms)
	assert.Equal(suite.T(), "US", customer.BillingCountry)
	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestCreateCustomer_ValidationErrors() {
	testCases := []struct {
		name        string
		tenantID    string
		customer    *Customer
		expectError string
	}{
		{
			name:     "nil customer",
			tenantID: suite.tenantID,
			customer: nil,
			expectError: "validation failed: customer is required",
		},
		{
			name:     "empty customer name",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "",
			},
			expectError: "customer name is required",
		},
		{
			name:     "whitespace only customer name",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "   ",
			},
			expectError: "customer name is required",
		},
		{
			name:     "customer name too long",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: string(make([]byte, 256)),
			},
			expectError: "customer name too long",
		},
		{
			name:     "invalid company code format",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "Test Company",
				CompanyCode: func() *string { s := "invalid-code!"; return &s }(),
			},
			expectError: "company code must be alphanumeric",
		},
		{
			name:     "company code too short",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "Test Company",
				CompanyCode: func() *string { s := "A"; return &s }(),
			},
			expectError: "company code must be alphanumeric, 2-50 characters",
		},
		{
			name:     "company code too long",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "Test Company",
				CompanyCode: func() *string { s := string(make([]byte, 51)); return &s }(),
			},
			expectError: "company code must be alphanumeric, 2-50 characters",
		},
		{
			name:     "invalid status",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "Test Company",
				Status: "invalid-status",
			},
			expectError: "invalid status: invalid-status",
		},
		{
			name:     "billing country not 2 characters",
			tenantID: suite.tenantID,
			customer: &Customer{
				Name: "Test Company",
				BillingCountry: "USA",
			},
			expectError: "billing country must be 2-character code",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.service.CreateCustomer(suite.ctx, tc.tenantID, tc.customer)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

func (suite *CustomerServiceTestSuite) TestCreateCustomer_RepositoryError() {
	customer := &Customer{
		Name: "Test Company",
	}
	repoError := errors.New("constraint violation")

	suite.repo.On("CreateCustomer", suite.ctx, suite.tenantID, customer).Return(repoError)

	err := suite.service.CreateCustomer(suite.ctx, suite.tenantID, customer)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create customer")
	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertNotCalled(suite.T(), "CacheCustomer")
}

func (suite *CustomerServiceTestSuite) TestUpdateCustomer_Success() {
	customer := &Customer{
		ID:       1,
		TenantID: suite.tenantID,
		Name:     "Updated Company",
		Status:   StatusActive,
	}

	suite.repo.On("UpdateCustomer", suite.ctx, suite.tenantID, customer).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customer.ID).Return()

	err := suite.service.UpdateCustomer(suite.ctx, suite.tenantID, customer)

	assert.NoError(suite.T(), err)
	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestUpdateCustomer_ValidationErrors() {
	testCases := []struct {
		name        string
		customer    *Customer
		expectError string
	}{
		{
			name: "invalid customer ID",
			customer: &Customer{
				ID:   0,
				Name: "Test Company",
			},
			expectError: "invalid customer ID: 0",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.service.UpdateCustomer(suite.ctx, suite.tenantID, tc.customer)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

func (suite *CustomerServiceTestSuite) TestDeleteCustomer_Success() {
	customerID := 1

	suite.repo.On("DeleteCustomer", suite.ctx, suite.tenantID, customerID).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customerID).Return()

	err := suite.service.DeleteCustomer(suite.ctx, suite.tenantID, customerID)

	assert.NoError(suite.T(), err)
	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestSearchCustomers_Success() {
	filters := SearchFilters{
		Name:   "Test Company",
		Status: []Status{StatusActive},
		Limit:  10,
		Offset: 0,
	}

	expectedCustomers := []Customer{
		{ID: 1, Name: "Test Company 1", Status: StatusActive},
		{ID: 2, Name: "Test Company 2", Status: StatusActive},
	}

	suite.repo.On("SearchCustomers", suite.ctx, suite.tenantID, filters).Return(expectedCustomers, 2, nil)

	result, total, err := suite.service.SearchCustomers(suite.ctx, suite.tenantID, filters)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedCustomers, result)
	assert.Equal(suite.T(), 2, total)
	suite.repo.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestSearchCustomers_ValidationErrors() {
	testCases := []struct {
		name        string
		filters     SearchFilters
		expectError string
	}{
		{
			name: "limit too high",
			filters: SearchFilters{
				Limit: 1001,
			},
			expectError: "limit must be between 0 and 1000",
		},
		{
			name: "negative offset",
			filters: SearchFilters{
				Offset: -1,
			},
			expectError: "offset must be non-negative",
		},
		{
			name: "invalid status",
			filters: SearchFilters{
				Status: []Status{"invalid"},
			},
			expectError: "invalid status in filter: invalid",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, total, err := suite.service.SearchCustomers(suite.ctx, suite.tenantID, tc.filters)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Zero(t, total)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

func (suite *CustomerServiceTestSuite) TestRegisterCustomerContact_Success() {
	customerID := 1
	fullName := "John Doe"
	email := "john@company.com"
	
	contact := &CustomerContact{
		CustomerID:  customerID,
		AuthUserID:  123,
		ContactType: ContactTypePrimary,
		FullName:    &fullName,
		Email:       &email,
	}

	existingCustomer := &Customer{
		ID:       customerID,
		TenantID: suite.tenantID,
		Name:     "Test Company",
	}

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(existingCustomer, true)
	suite.repo.On("AddCustomerContact", suite.ctx, suite.tenantID, contact).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customerID).Return()

	err := suite.service.RegisterCustomerContact(suite.ctx, suite.tenantID, customerID, contact)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), customerID, contact.CustomerID)
	assert.True(suite.T(), contact.IsActive)
	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestRegisterCustomerContact_CustomerNotFound() {
	contact := &CustomerContact{
		CustomerID: 999,
		AuthUserID: 123,
	}

	suite.cache.On("GetCustomer", suite.tenantID, 999).Return(nil, false)
	suite.repo.On("GetCustomerByID", suite.ctx, suite.tenantID, 999).Return(nil, errors.New("customer not found"))

	err := suite.service.RegisterCustomerContact(suite.ctx, suite.tenantID, 999, contact)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "customer not found")
}

func (suite *CustomerServiceTestSuite) TestMultiTenantIsolation() {
	tenantA := "tenant-a"
	tenantB := "tenant-b"
	customerID := 1

	customerA := &Customer{
		ID:       customerID,
		TenantID: tenantA,
		Name:     "Company A",
	}
	
	customerB := &Customer{
		ID:       customerID,
		TenantID: tenantB,
		Name:     "Company B",
	}

	suite.cache.On("GetCustomer", tenantA, customerID).Return(nil, false)
	suite.cache.On("GetCustomer", tenantB, customerID).Return(nil, false)
	suite.repo.On("GetCustomerByID", suite.ctx, tenantA, customerID).Return(customerA, nil)
	suite.repo.On("GetCustomerByID", suite.ctx, tenantB, customerID).Return(customerB, nil)
	suite.cache.On("CacheCustomer", tenantA, customerA).Return()
	suite.cache.On("CacheCustomer", tenantB, customerB).Return()

	resultA, err := suite.service.GetCustomer(suite.ctx, tenantA, customerID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Company A", resultA.Name)
	assert.Equal(suite.T(), tenantA, resultA.TenantID)

	resultB, err := suite.service.GetCustomer(suite.ctx, tenantB, customerID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Company B", resultB.Name)
	assert.Equal(suite.T(), tenantB, resultB.TenantID)

	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestCacheInvalidation_UpdateOperations() {
	customerID := 1
	customer := &Customer{
		ID:       customerID,
		TenantID: suite.tenantID,
		Name:     "Test Company",
	}

	suite.repo.On("UpdateCustomer", suite.ctx, suite.tenantID, customer).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customerID).Return()

	err := suite.service.UpdateCustomer(suite.ctx, suite.tenantID, customer)

	assert.NoError(suite.T(), err)
	suite.cache.AssertExpectations(suite.T())

	suite.repo.On("DeleteCustomer", suite.ctx, suite.tenantID, customerID).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customerID).Return()

	err = suite.service.DeleteCustomer(suite.ctx, suite.tenantID, customerID)

	assert.NoError(suite.T(), err)
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestConcurrentOperations() {
	customerID := 1
	done := make(chan bool, 2)
	
	customer := &Customer{
		ID:       customerID,
		TenantID: suite.tenantID,
		Name:     "Concurrent Test Company",
	}

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(nil, false).Maybe()
	suite.repo.On("GetCustomerByID", suite.ctx, suite.tenantID, customerID).Return(customer, nil).Maybe()
	suite.cache.On("CacheCustomer", suite.tenantID, customer).Return().Maybe()

	go func() {
		result, err := suite.service.GetCustomer(suite.ctx, suite.tenantID, customerID)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)
		done <- true
	}()

	go func() {
		result, err := suite.service.GetCustomer(suite.ctx, suite.tenantID, customerID)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)
		done <- true
	}()

	<-done
	<-done
}

func (suite *CustomerServiceTestSuite) TestErrorHandlingWithContext() {
	customerID := 1
	ctxWithTimeout, cancel := context.WithTimeout(suite.ctx, time.Millisecond)
	defer cancel()

	suite.cache.On("GetCustomer", suite.tenantID, customerID).Return(nil, false)
	suite.repo.On("GetCustomerByID", ctxWithTimeout, suite.tenantID, customerID).Return(nil, context.DeadlineExceeded)

	result, err := suite.service.GetCustomer(ctxWithTimeout, suite.tenantID, customerID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to get customer")
}

func (suite *CustomerServiceTestSuite) TestEndToEndWorkflow() {
	companyCode := "E2E123"
	taxID := "987654321"
	street := "456 Test Ave"
	city := "Dallas"
	state := "TX"
	zip := "75001"
	fullName := "Jane Manager"
	email := "jane@e2e.com"

	customer := &Customer{
		Name:           "End-to-End Test Company",
		CompanyCode:    &companyCode,
		TaxID:          &taxID,
		BillingStreet:  &street,
		BillingCity:    &city,
		BillingState:   &state,
		BillingZip:     &zip,
		BillingCountry: "US",
	}

	suite.repo.On("CreateCustomer", suite.ctx, suite.tenantID, customer).Return(nil)
	suite.cache.On("CacheCustomer", suite.tenantID, customer).Return()

	err := suite.service.CreateCustomer(suite.ctx, suite.tenantID, customer)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, customer.ID)

	contact := &CustomerContact{
		CustomerID:  customer.ID,
		AuthUserID:  456,
		ContactType: ContactTypePrimary,
		FullName:    &fullName,
		Email:       &email,
	}

	suite.cache.On("GetCustomer", suite.tenantID, customer.ID).Return(customer, true)
	suite.repo.On("AddCustomerContact", suite.ctx, suite.tenantID, contact).Return(nil)
	suite.cache.On("InvalidateCustomer", suite.tenantID, customer.ID).Return()

	err = suite.service.RegisterCustomerContact(suite.ctx, suite.tenantID, customer.ID, contact)
	assert.NoError(suite.T(), err)

	contacts := []CustomerContact{*contact}
	suite.repo.On("GetCustomerContacts", suite.ctx, suite.tenantID, customer.ID).Return(contacts, nil)

	resultContacts, err := suite.service.GetCustomerContacts(suite.ctx, suite.tenantID, customer.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), resultContacts, 1)
	assert.Equal(suite.T(), contact.Email, resultContacts[0].Email)

	analytics := &CustomerAnalytics{
		CustomerID:      customer.ID,
		TotalWorkOrders: 0,
		ActiveOrders:    0,
		TotalRevenue:    0,
		AvgOrderValue:   0,
	}
	suite.repo.On("GetCustomerAnalytics", suite.ctx, suite.tenantID, customer.ID).Return(analytics, nil)

	resultAnalytics, err := suite.service.GetCustomerAnalytics(suite.ctx, suite.tenantID, customer.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), analytics, resultAnalytics)

	suite.repo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}
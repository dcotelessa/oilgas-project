// backend/internal/customer/handlers_test.go
// Test suite using service interface
package customer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockService implements CustomerService interface for testing
type MockService struct {
	customers   []Customer
	tenantID    string
	shouldError bool
}

func NewMockService() *MockService {
	return &MockService{
		customers: []Customer{
			{
				CustomerID:   1,
				Customer:     "Test Company",
				TenantID:     "test-tenant",
				BillingCity:  stringPtr("Houston"),
				BillingState: stringPtr("TX"),
				Phone:        stringPtr("555-0123"),
				Email:        stringPtr("test@company.com"),
				Deleted:      false,
			},
		},
		tenantID: "test-tenant",
	}
}

// Implement CustomerService interface
func (m *MockService) GetCustomersByTenant(tenantID string) ([]Customer, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	var result []Customer
	for _, customer := range m.customers {
		if customer.TenantID == tenantID && !customer.Deleted {
			result = append(result, customer)
		}
	}
	return result, nil
}

func (m *MockService) GetCustomerByID(tenantID string, customerID int) (*Customer, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	
	for _, customer := range m.customers {
		if customer.CustomerID == customerID && customer.TenantID == tenantID && !customer.Deleted {
			return &customer, nil
		}
	}
	return nil, assert.AnError
}

func (m *MockService) SearchCustomers(filter CustomerFilter) ([]Customer, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	return m.customers, nil
}

func (m *MockService) CreateCustomer(customer *Customer) error {
	if m.shouldError {
		return assert.AnError
	}
	customer.CustomerID = len(m.customers) + 1
	m.customers = append(m.customers, *customer)
	return nil
}

func (m *MockService) UpdateCustomer(customer *Customer) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockService) SoftDeleteCustomer(tenantID string, customerID int) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockService) GetCustomerStats(tenantID string) (map[string]interface{}, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	return map[string]interface{}{"total_customers": 1}, nil
}

func (m *MockService) SetTenantContext(tenantID string) error {
	if m.shouldError {
		return assert.AnError
	}
	m.tenantID = tenantID
	return nil
}

func (m *MockService) GetCurrentTenant() (string, error) {
	if m.shouldError {
		return "", assert.AnError
	}
	return m.tenantID, nil
}

func (m *MockService) ValidateCustomer(customer *Customer) error {
	if customer.Customer == "" {
		return assert.AnError
	}
	return nil
}

func TestGetCustomers(t *testing.T) {
	mockService := NewMockService()
	handler := NewHandler(mockService)

	req := httptest.NewRequest("GET", "/customers?tenant_id=test-tenant", nil)
	w := httptest.NewRecorder()

	handler.GetCustomers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

func TestCreateCustomer(t *testing.T) {
	mockService := NewMockService()
	handler := NewHandler(mockService)

	customerData := Customer{
		Customer: "New Company",
	}

	jsonData, _ := json.Marshal(customerData)
	req := httptest.NewRequest("POST", "/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant")
	
	w := httptest.NewRecorder()

	handler.CreateCustomer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck(t *testing.T) {
	mockService := NewMockService()
	handler := NewHandler(mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

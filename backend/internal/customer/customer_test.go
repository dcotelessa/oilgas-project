// backend/internal/customer/service_test.go
package customer

import (
	"context"
	"testing"
	"time"
)

// MockRepository for testing
type MockRepository struct {
	customers  map[string][]Customer
	analytics  map[int]*CustomerAnalytics
	nextID     int
	shouldFail bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		customers: make(map[string][]Customer),
		analytics: make(map[int]*CustomerAnalytics),
		nextID:    1,
	}
}

func (m *MockRepository) GetAllForTenant(ctx context.Context, tenantID string, filters CustomerFilters) ([]Customer, error) {
	if m.shouldFail {
		return nil, ErrCustomerNotFound
	}
	customers, exists := m.customers[tenantID]
	if !exists {
		return []Customer{}, nil
	}
	
	// Simple filtering for tests
	var filtered []Customer
	for _, customer := range customers {
		if filters.Query != "" {
			if !contains(customer.Customer, filters.Query) &&
			   (customer.Contact == nil || !contains(*customer.Contact, filters.Query)) {
				continue
			}
		}
		filtered = append(filtered, customer)
	}
	return filtered, nil
}

func (m *MockRepository) GetByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error) {
	if m.shouldFail {
		return nil, ErrCustomerNotFound
	}
	customers, exists := m.customers[tenantID]
	if !exists {
		return nil, ErrCustomerNotFound
	}
	
	for _, customer := range customers {
		if customer.CustomerID == id {
			return &customer, nil
		}
	}
	return nil, ErrCustomerNotFound
}

func (m *MockRepository) SearchForTenant(ctx context.Context, tenantID, query string) ([]Customer, error) {
	return m.GetAllForTenant(ctx, tenantID, CustomerFilters{Query: query})
}

func (m *MockRepository) GetCountForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (int, error) {
	customers, err := m.GetAllForTenant(ctx, tenantID, filters)
	return len(customers), err
}

func (m *MockRepository) GetAnalyticsForTenant(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	analytics, exists := m.analytics[customerID]
	if !exists {
		return &CustomerAnalytics{CustomerID: customerID}, nil
	}
	return analytics, nil
}

func (m *MockRepository) CreateForTenant(ctx context.Context, tenantID string, customer *Customer) error {
	if m.shouldFail {
		return ErrCustomerExists
	}
	
	customer.CustomerID = m.nextID
	customer.TenantID = tenantID
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()
	m.nextID++
	
	if m.customers[tenantID] == nil {
		m.customers[tenantID] = []Customer{}
	}
	m.customers[tenantID] = append(m.customers[tenantID], *customer)
	return nil
}

func (m *MockRepository) UpdateForTenant(ctx context.Context, tenantID string, customer *Customer) error {
	if m.shouldFail {
		return ErrCustomerNotFound
	}
	
	customers, exists := m.customers[tenantID]
	if !exists {
		return ErrCustomerNotFound
	}
	
	for i, existing := range customers {
		if existing.CustomerID == customer.CustomerID {
			customer.UpdatedAt = time.Now()
			m.customers[tenantID][i] = *customer
			return nil
		}
	}
	return ErrCustomerNotFound
}

func (m *MockRepository) DeleteForTenant(ctx context.Context, tenantID string, id int) error {
	if m.shouldFail {
		return ErrCustomerNotFound
	}
	
	customers, exists := m.customers[tenantID]
	if !exists {
		return ErrCustomerNotFound
	}
	
	for i, customer := range customers {
		if customer.CustomerID == id {
			m.customers[tenantID][i].Deleted = true
			return nil
		}
	}
	return ErrCustomerNotFound
}

func (m *MockRepository) GetCustomersByIDs(ctx context.Context, customerIDs []int) ([]Customer, error) {
	var result []Customer
	for tenantID, customers := range m.customers {
		for _, customer := range customers {
			for _, id := range customerIDs {
				if customer.CustomerID == id && !customer.Deleted {
					result = append(result, customer)
					break
				}
			}
		}
	}
	return result, nil
}

func (m *MockRepository) GetCustomerSummaryByTenant(ctx context.Context) (map[string]int, error) {
	summary := make(map[string]int)
	for tenantID, customers := range m.customers {
		count := 0
		for _, customer := range customers {
			if !customer.Deleted {
				count++
			}
		}
		summary[tenantID] = count
	}
	return summary, nil
}

// Helper function
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     (str[:len(substr)] == substr || 
		      str[len(str)-len(substr):] == substr ||
		      findInString(str, substr))))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test Cases
func TestCustomerService_CreateCustomer(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()
	tenantID := "tenant-123"

	tests := []struct {
		name    string
		request CreateCustomerRequest
		wantErr bool
		errType error
	}{
		{
			name: "valid customer creation",
			request: CreateCustomerRequest{
				Customer:       "Test Oil Company",
				BillingAddress: stringPtr("123 Main St"),
				BillingCity:    stringPtr("Houston"),
				BillingState:   stringPtr("TX"),
				Email:          stringPtr("test@testoil.com"),
				Phone:          stringPtr("555-123-4567"),
			},
			wantErr: false,
		},
		{
			name: "missing customer name",
			request: CreateCustomerRequest{
				Customer: "",
				Email:    stringPtr("test@testoil.com"),
			},
			wantErr: true,
			errType: ErrCustomerNameRequired,
		},
		{
			name: "invalid email",
			request: CreateCustomerRequest{
				Customer: "Test Company",
				Email:    stringPtr("invalid-email"),
			},
			wantErr: true,
		},
		{
			name: "invalid state code",
			request: CreateCustomerRequest{
				Customer:     "Test Company",
				BillingState: stringPtr("XX"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := service.CreateCustomerForTenant(ctx, tenantID, tt.request)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("expected error type %v, got %v", tt.errType, err)
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if customer == nil {
				t.Errorf("expected customer but got nil")
				return
			}
			
			if customer.Customer != tt.request.Customer {
				t.Errorf("expected customer name %s, got %s", tt.request.Customer, customer.Customer)
			}
			
			if customer.TenantID != tenantID {
				t.Errorf("expected tenant ID %s, got %s", tenantID, customer.TenantID)
			}
		})
	}
}

func TestCustomerService_GetCustomersForTenant(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()
	tenantID := "tenant-123"

	// Setup test data
	testCustomers := []Customer{
		{
			CustomerID: 1,
			Customer:   "Alpha Oil Corp",
			Contact:    stringPtr("John Doe"),
			Email:      stringPtr("john@alphaoil.com"),
			TenantID:   tenantID,
			CreatedAt:  time.Now(),
		},
		{
			CustomerID: 2,
			Customer:   "Beta Gas LLC",
			Contact:    stringPtr("Jane Smith"),
			Email:      stringPtr("jane@betagas.com"),
			TenantID:   tenantID,
			CreatedAt:  time.Now(),
		},
	}
	
	repo.customers[tenantID] = testCustomers

	tests := []struct {
		name        string
		filters     CustomerFilters
		wantCount   int
		wantErr     bool
	}{
		{
			name:      "get all customers",
			filters:   CustomerFilters{},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "search by name",
			filters: CustomerFilters{
				Query: "Alpha",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "search with no results",
			filters: CustomerFilters{
				Query: "NonExistent",
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.GetCustomersForTenant(ctx, tenantID, tt.filters)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(response.Customers) != tt.wantCount {
				t.Errorf("expected %d customers, got %d", tt.wantCount, len(response.Customers))
			}
			
			if response.Total != tt.wantCount {
				t.Errorf("expected total %d, got %d", tt.wantCount, response.Total)
			}
		})
	}
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()
	tenantID := "tenant-123"

	// Setup existing customer
	existingCustomer := Customer{
		CustomerID: 1,
		Customer:   "Original Name",
		Email:      stringPtr("original@email.com"),
		TenantID:   tenantID,
		CreatedAt:  time.Now(),
	}
	repo.customers[tenantID] = []Customer{existingCustomer}

	tests := []struct {
		name    string
		id      int
		request UpdateCustomerRequest
		wantErr bool
	}{
		{
			name: "update customer name",
			id:   1,
			request: UpdateCustomerRequest{
				Customer: stringPtr("Updated Name"),
			},
			wantErr: false,
		},
		{
			name: "update with invalid email",
			id:   1,
			request: UpdateCustomerRequest{
				Email: stringPtr("invalid-email"),
			},
			wantErr: true,
		},
		{
			name: "update non-existent customer",
			id:   999,
			request: UpdateCustomerRequest{
				Customer: stringPtr("New Name"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := service.UpdateCustomerForTenant(ctx, tenantID, tt.id, tt.request)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if customer == nil {
				t.Errorf("expected customer but got nil")
				return
			}
			
			if tt.request.Customer != nil && customer.Customer != *tt.request.Customer {
				t.Errorf("expected customer name %s, got %s", *tt.request.Customer, customer.Customer)
			}
		})
	}
}

func TestCustomerService_TenantIsolation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	ctx := context.Background()
	
	tenant1 := "tenant-1"
	tenant2 := "tenant-2"

	// Create customers for different tenants
	customer1 := Customer{
		CustomerID: 1,
		Customer:   "Tenant 1 Customer",
		TenantID:   tenant1,
		CreatedAt:  time.Now(),
	}
	customer2 := Customer{
		CustomerID: 2,
		Customer:   "Tenant 2 Customer",
		TenantID:   tenant2,
		CreatedAt:  time.Now(),
	}
	
	repo.customers[tenant1] = []Customer{customer1}
	repo.customers[tenant2] = []Customer{customer2}

	// Test tenant 1 can only see their customers
	response1, err := service.GetCustomersForTenant(ctx, tenant1, CustomerFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(response1.Customers) != 1 {
		t.Errorf("tenant 1 should see 1 customer, got %d", len(response1.Customers))
	}
	
	if response1.Customers[0].Customer != "Tenant 1 Customer" {
		t.Errorf("tenant 1 should see their customer, got %s", response1.Customers[0].Customer)
	}

	// Test tenant 2 can only see their customers
	response2, err := service.GetCustomersForTenant(ctx, tenant2, CustomerFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(response2.Customers) != 1 {
		t.Errorf("tenant 2 should see 1 customer, got %d", len(response2.Customers))
	}
	
	if response2.Customers[0].Customer != "Tenant 2 Customer" {
		t.Errorf("tenant 2 should see their customer, got %s", response2.Customers[0].Customer)
	}

	// Test tenant 1 cannot access tenant 2's customer
	_, err = service.GetCustomerByIDForTenant(ctx, tenant1, 2)
	if err != ErrCustomerNotFound {
		t.Errorf("expected customer not found error, got %v", err)
	}
}

func TestCustomerService_Validation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	tests := []struct {
		name     string
		customer Customer
		wantErr  bool
	}{
		{
			name: "valid customer",
			customer: Customer{
				Customer:     "Valid Company",
				Email:        stringPtr("valid@email.com"),
				Phone:        stringPtr("555-123-4567"),
				BillingState: stringPtr("TX"),
			},
			wantErr: false,
		},
		{
			name: "empty customer name",
			customer: Customer{
				Customer: "",
			},
			wantErr: true,
		},
		{
			name: "customer name too long",
			customer: Customer{
				Customer: string(make([]byte, 256)), // 256 characters
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			customer: Customer{
				Customer: "Valid Company",
				Email:    stringPtr("invalid-email"),
			},
			wantErr: true,
		},
		{
			name: "invalid state code",
			customer: Customer{
				Customer:     "Valid Company",
				BillingState: stringPtr("XX"),
			},
			wantErr: true,
		},
		{
			name: "negative credit limit",
			customer: Customer{
				Customer:    "Valid Company",
				CreditLimit: float64Ptr(-1000.0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCustomer(&tt.customer)
			
			if tt.wantErr && err == nil {
				t.Errorf("expected validation error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

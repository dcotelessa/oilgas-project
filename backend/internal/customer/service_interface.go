// backend/internal/customer/service_interface.go
// Service interface for testing and dependency injection
package customer

// CustomerService defines the interface for customer operations
type CustomerService interface {
	GetCustomersByTenant(tenantID string) ([]Customer, error)
	GetCustomerByID(tenantID string, customerID int) (*Customer, error)
	SearchCustomers(filter CustomerFilter) ([]Customer, error)
	CreateCustomer(customer *Customer) error
	UpdateCustomer(customer *Customer) error
	SoftDeleteCustomer(tenantID string, customerID int) error
	GetCustomerStats(tenantID string) (map[string]interface{}, error)
	SetTenantContext(tenantID string) error
	GetCurrentTenant() (string, error)
	ValidateCustomer(customer *Customer) error
}

// Note: The concrete service implementation should implement this interface

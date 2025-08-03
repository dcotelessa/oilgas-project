// backend/internal/customer/repository.go
// Customer repository interface for clean architecture
package customer

import "context"

// Repository defines the interface for customer data access
type Repository interface {
	// GetByID retrieves a customer by ID within tenant context
	GetByID(ctx context.Context, tenantID string, customerID int) (*Customer, error)
	
	// GetByTenant retrieves all customers for a tenant
	GetByTenant(ctx context.Context, tenantID string) ([]Customer, error)
	
	// Search performs filtered search with pagination
	Search(ctx context.Context, filter CustomerFilter) ([]Customer, error)
	
	// Create creates a new customer
	Create(ctx context.Context, customer *Customer) error
	
	// Update updates an existing customer
	Update(ctx context.Context, customer *Customer) error
	
	// SoftDelete marks a customer as deleted
	SoftDelete(ctx context.Context, tenantID string, customerID int) error
	
	// GetStats returns customer statistics for a tenant
	GetStats(ctx context.Context, tenantID string) (map[string]interface{}, error)
	
	// SetTenantContext sets the database tenant context
	SetTenantContext(ctx context.Context, tenantID string) error
}

// TenantManager handles tenant context operations
type TenantManager interface {
	// SetContext sets the current tenant context
	SetContext(tenantID string) error
	
	// GetCurrentContext returns the current tenant
	GetCurrentContext() (string, error)
	
	// ClearContext clears the tenant context
	ClearContext() error
}

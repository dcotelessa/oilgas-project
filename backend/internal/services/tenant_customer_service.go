// backend/internal/services/tenant_customer_service.go
// Enhanced CustomerService with tenant isolation
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/repository"
)

// TenantCustomerService extends CustomerService with tenant capabilities
type TenantCustomerService struct {
	*CustomerService // Embed existing service
	tenantRepo       repository.TenantCustomerRepository
}

func NewTenantCustomerService(repo repository.CustomerRepository, tenantRepo repository.TenantCustomerRepository) *TenantCustomerService {
	return &TenantCustomerService{
		CustomerService: NewCustomerService(repo),
		tenantRepo:      tenantRepo,
	}
}

// Tenant-aware methods
func (s *TenantCustomerService) GetAllCustomersForTenant(ctx context.Context, tenantID string) ([]repository.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	return s.tenantRepo.GetAllForTenant(ctx, tenantID)
}

func (s *TenantCustomerService) GetCustomerByIDForTenant(ctx context.Context, tenantID string, customerID int) (*repository.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}
	return s.tenantRepo.GetByIDForTenant(ctx, tenantID, customerID)
}

func (s *TenantCustomerService) SearchCustomersForTenant(ctx context.Context, tenantID, query string) ([]repository.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	return s.tenantRepo.SearchForTenant(ctx, tenantID, query)
}

func (s *TenantCustomerService) GetCustomersWithFiltersForTenant(ctx context.Context, tenantID string, filters repository.CustomerFilters) ([]repository.Customer, int, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, 0, err
	}
	if filters.Limit <= 0 {
		filters.Limit = 50 // Default limit
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000 // Cap at 1000
	}
	
	customers, err := s.tenantRepo.GetAllWithFiltersForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total, err := s.tenantRepo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return customers, 0, err // Return customers even if count fails
	}
	
	return customers, total, nil
}

func (s *TenantCustomerService) GetCustomerRelatedDataForTenant(ctx context.Context, tenantID string, customerID int) (*repository.CustomerRelatedData, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}
	return s.tenantRepo.GetRelatedDataForTenant(ctx, tenantID, customerID)
}

// Tenant-aware CRUD operations
func (s *TenantCustomerService) CreateCustomerForTenant(ctx context.Context, tenantID string, customer *repository.Customer) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Set tenant ID in customer record
	customer.TenantID = tenantID
	
	return s.tenantRepo.CreateForTenant(ctx, tenantID, customer)
}

func (s *TenantCustomerService) UpdateCustomerForTenant(ctx context.Context, tenantID string, customer *repository.Customer) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if customer.CustomerID <= 0 {
		return fmt.Errorf("invalid customer ID: %d", customer.CustomerID)
	}
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Ensure tenant ID matches
	customer.TenantID = tenantID
	
	return s.tenantRepo.UpdateForTenant(ctx, tenantID, customer)
}

func (s *TenantCustomerService) DeleteCustomerForTenant(ctx context.Context, tenantID string, customerID int) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	if customerID <= 0 {
		return fmt.Errorf("invalid customer ID: %d", customerID)
	}
	return s.tenantRepo.DeleteForTenant(ctx, tenantID, customerID)
}

func (s *TenantCustomerService) validateTenantID(tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return fmt.Errorf("tenant ID must be between 2 and 20 characters")
	}
	return nil
}

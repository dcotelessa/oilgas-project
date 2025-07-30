// backend/internal/services/tenant_customer_service.go
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type TenantCustomerService struct {
	tenantRepo repository.TenantCustomerRepository
}

func NewTenantCustomerService(tenantRepo repository.TenantCustomerRepository) *TenantCustomerService {
	return &TenantCustomerService{
		tenantRepo: tenantRepo,
	}
}

func (s *TenantCustomerService) GetCustomersForTenant(ctx context.Context, tenantID string) ([]models.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	
	// Use default empty filters when none provided
	filters := models.CustomerFilters{
		Limit: 100, // Default limit
	}
	
	return s.tenantRepo.GetAllForTenant(ctx, tenantID, filters)
}

func (s *TenantCustomerService) GetCustomersWithFilters(ctx context.Context, tenantID string, filters models.CustomerFilters) ([]models.Customer, int, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, 0, err
	}
	
	// Set default limit if not provided
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000
	}
	
	customers, err := s.tenantRepo.GetAllForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total, err := s.tenantRepo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return customers, 0, err // Return customers even if count fails
	}
	
	return customers, total, nil
}

func (s *TenantCustomerService) GetCustomerByID(ctx context.Context, tenantID string, customerID int) (*models.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	
	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}
	
	return s.tenantRepo.GetByIDForTenant(ctx, tenantID, customerID)
}

func (s *TenantCustomerService) SearchCustomers(ctx context.Context, tenantID, query string) ([]models.Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	
	return s.tenantRepo.SearchForTenant(ctx, tenantID, query)
}

func (s *TenantCustomerService) GetCustomerRelatedData(ctx context.Context, tenantID string, customerID int) (*models.CustomerRelatedData, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	
	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}
	
	return s.tenantRepo.GetRelatedDataForTenant(ctx, tenantID, customerID)
}

func (s *TenantCustomerService) CreateCustomer(ctx context.Context, tenantID string, customer *models.Customer) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}
	
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Set tenant ID
	customer.TenantID = tenantID
	
	return s.tenantRepo.CreateForTenant(ctx, tenantID, customer)
}

func (s *TenantCustomerService) UpdateCustomer(ctx context.Context, tenantID string, customer *models.Customer) error {
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

func (s *TenantCustomerService) DeleteCustomer(ctx context.Context, tenantID string, customerID int) error {
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

func (s *TenantCustomerService) validateCustomer(customer *models.Customer) error {
	if strings.TrimSpace(customer.Customer) == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(customer.Customer) > 255 {
		return fmt.Errorf("customer name too long (max 255 characters)")
	}
	return nil
}

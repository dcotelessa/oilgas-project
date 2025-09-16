// backend/internal/customer/service.go
package customer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"oilgas-backend/internal/auth"
)

type CacheService interface {
	GetCustomer(tenantID string, customerID int) (*Customer, bool)
	CacheCustomer(tenantID string, customer *Customer)
	InvalidateCustomer(tenantID string, customerID int)
	InvalidateCustomerSearch(tenantID string, filters SearchFilters)
}

type Service interface {
	GetCustomer(ctx context.Context, tenantID string, id int) (*Customer, error)
	SearchCustomers(ctx context.Context, tenantID string, filters SearchFilters) ([]Customer, int, error)
	CreateCustomer(ctx context.Context, tenantID string, customer *Customer) error
	UpdateCustomer(ctx context.Context, tenantID string, customer *Customer) error
	DeleteCustomer(ctx context.Context, tenantID string, id int) error
	
	RegisterCustomerContact(ctx context.Context, tenantID string, customerID int, contact *CustomerContact) error
	GetCustomerContacts(ctx context.Context, tenantID string, customerID int) ([]CustomerContact, error)
	UpdateCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error
	RemoveCustomerContact(ctx context.Context, tenantID string, customerID, authUserID int) error
	
	GetCustomerAnalytics(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error)
}

type service struct {
	repo      Repository
	authSvc   *auth.AuthService
	cache     CacheService
}

func NewService(repo Repository, authSvc *auth.AuthService, cache CacheService) Service {
	return &service{
		repo:    repo,
		authSvc: authSvc,
		cache:   cache,
	}
}

func (s *service) GetCustomer(ctx context.Context, tenantID string, id int) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}
	
	if id <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", id)
	}
	
	if customer, found := s.cache.GetCustomer(tenantID, id); found {
		return customer, nil
	}
	
	customer, err := s.repo.GetCustomerByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer %d: %w", id, err)
	}
	
	s.cache.CacheCustomer(tenantID, customer)
	return customer, nil
}

func (s *service) RegisterCustomerContact(ctx context.Context, tenantID string, customerID int, contact *CustomerContact) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}
	
	if err := s.validateCustomerContact(contact); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	_, err := s.GetCustomer(ctx, tenantID, customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}
	
	contact.CustomerID = customerID
	contact.IsActive = true
	
	if err := s.repo.AddCustomerContact(ctx, tenantID, contact); err != nil {
		return fmt.Errorf("failed to add customer contact: %w", err)
	}
	
	s.cache.InvalidateCustomer(tenantID, customerID)
	return nil
}

func (s *service) SearchCustomers(ctx context.Context, tenantID string, filters SearchFilters) ([]Customer, int, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, 0, fmt.Errorf("invalid tenant: %w", err)
	}

	if err := s.validateSearchFilters(&filters); err != nil {
		return nil, 0, fmt.Errorf("validation failed: %w", err)
	}

	customers, total, err := s.repo.SearchCustomers(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	return customers, total, nil
}

func (s *service) CreateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	customer.TenantID = tenantID
	customer.IsActive = true
	if customer.Status == "" {
		customer.Status = StatusActive
	}
	if customer.PaymentTerms == "" {
		customer.PaymentTerms = "NET30"
	}
	if customer.BillingCountry == "" {
		customer.BillingCountry = "US"
	}

	if err := s.repo.CreateCustomer(ctx, tenantID, customer); err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	s.cache.CacheCustomer(tenantID, customer)
	return nil
}

func (s *service) UpdateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	if customer.ID <= 0 {
		return fmt.Errorf("invalid customer ID: %d", customer.ID)
	}

	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.UpdateCustomer(ctx, tenantID, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	s.cache.InvalidateCustomer(tenantID, customer.ID)
	return nil
}

func (s *service) DeleteCustomer(ctx context.Context, tenantID string, id int) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	if id <= 0 {
		return fmt.Errorf("invalid customer ID: %d", id)
	}

	if err := s.repo.DeleteCustomer(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	s.cache.InvalidateCustomer(tenantID, id)
	return nil
}

func (s *service) GetCustomerContacts(ctx context.Context, tenantID string, customerID int) ([]CustomerContact, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}

	contacts, err := s.repo.GetCustomerContacts(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer contacts: %w", err)
	}

	return contacts, nil
}

func (s *service) UpdateCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	if contact.ID <= 0 {
		return fmt.Errorf("invalid contact ID: %d", contact.ID)
	}

	if err := s.validateCustomerContact(contact); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.UpdateCustomerContact(ctx, tenantID, contact); err != nil {
		return fmt.Errorf("failed to update customer contact: %w", err)
	}

	s.cache.InvalidateCustomer(tenantID, contact.CustomerID)
	return nil
}

func (s *service) RemoveCustomerContact(ctx context.Context, tenantID string, customerID, authUserID int) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	if customerID <= 0 {
		return fmt.Errorf("invalid customer ID: %d", customerID)
	}

	if authUserID <= 0 {
		return fmt.Errorf("invalid auth user ID: %d", authUserID)
	}

	if err := s.repo.RemoveCustomerContact(ctx, tenantID, customerID, authUserID); err != nil {
		return fmt.Errorf("failed to remove customer contact: %w", err)
	}

	s.cache.InvalidateCustomer(tenantID, customerID)
	return nil
}

func (s *service) GetCustomerAnalytics(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, fmt.Errorf("invalid tenant: %w", err)
	}

	if customerID <= 0 {
		return nil, fmt.Errorf("invalid customer ID: %d", customerID)
	}

	analytics, err := s.repo.GetCustomerAnalytics(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer analytics: %w", err)
	}

	return analytics, nil
}

var (
	companyCodeRegex = regexp.MustCompile(`^[A-Za-z0-9]{2,50}$`)
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func (s *service) validateTenantID(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if len(tenantID) > 100 {
		return fmt.Errorf("tenant ID too long: %d characters", len(tenantID))
	}
	return nil
}

func (s *service) validateCustomer(customer *Customer) error {
	if customer == nil {
		return fmt.Errorf("customer is required")
	}

	if strings.TrimSpace(customer.Name) == "" {
		return fmt.Errorf("customer name is required")
	}

	if len(customer.Name) > 255 {
		return fmt.Errorf("customer name too long: %d characters", len(customer.Name))
	}

	if customer.CompanyCode != nil && *customer.CompanyCode != "" {
		if !companyCodeRegex.MatchString(*customer.CompanyCode) {
			return fmt.Errorf("company code must be alphanumeric, 2-50 characters")
		}
	}

	if customer.Status != "" {
		switch customer.Status {
		case StatusActive, StatusInactive, StatusSuspended:
		default:
			return fmt.Errorf("invalid status: %s", customer.Status)
		}
	}

	if customer.TaxID != nil && len(*customer.TaxID) > 50 {
		return fmt.Errorf("tax ID too long: %d characters", len(*customer.TaxID))
	}

	if len(customer.PaymentTerms) > 50 {
		return fmt.Errorf("payment terms too long: %d characters", len(customer.PaymentTerms))
	}

	if customer.BillingStreet != nil && len(*customer.BillingStreet) > 255 {
		return fmt.Errorf("billing street too long: %d characters", len(*customer.BillingStreet))
	}

	if customer.BillingCity != nil && len(*customer.BillingCity) > 100 {
		return fmt.Errorf("billing city too long: %d characters", len(*customer.BillingCity))
	}

	if customer.BillingState != nil && len(*customer.BillingState) > 10 {
		return fmt.Errorf("billing state too long: %d characters", len(*customer.BillingState))
	}

	if customer.BillingZip != nil && len(*customer.BillingZip) > 20 {
		return fmt.Errorf("billing zip code too long: %d characters", len(*customer.BillingZip))
	}

	if len(customer.BillingCountry) != 2 {
		return fmt.Errorf("billing country must be 2-character code")
	}

	return nil
}

func (s *service) validateCustomerContact(contact *CustomerContact) error {
	if contact == nil {
		return fmt.Errorf("customer contact is required")
	}

	if contact.AuthUserID <= 0 {
		return fmt.Errorf("auth user ID is required")
	}

	switch contact.ContactType {
	case ContactTypePrimary, ContactTypeBilling, ContactTypeShipping, ContactTypeApprover:
	default:
		return fmt.Errorf("invalid contact type: %s", contact.ContactType)
	}

	if contact.FullName != nil && len(*contact.FullName) > 255 {
		return fmt.Errorf("full name too long: %d characters", len(*contact.FullName))
	}

	if contact.Email != nil && *contact.Email != "" {
		if !emailRegex.MatchString(*contact.Email) {
			return fmt.Errorf("invalid email format: %s", *contact.Email)
		}
		if len(*contact.Email) > 255 {
			return fmt.Errorf("email too long: %d characters", len(*contact.Email))
		}
	}

	return nil
}

func (s *service) validateSearchFilters(filters *SearchFilters) error {
	if filters == nil {
		return nil
	}

	if len(filters.Name) > 255 {
		return fmt.Errorf("name filter too long: %d characters", len(filters.Name))
	}

	if filters.CompanyCode != "" && !companyCodeRegex.MatchString(filters.CompanyCode) {
		return fmt.Errorf("invalid company code format in filter")
	}

	if len(filters.TaxID) > 50 {
		return fmt.Errorf("tax ID filter too long: %d characters", len(filters.TaxID))
	}

	if filters.Limit < 0 || filters.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}

	if filters.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	for _, status := range filters.Status {
		switch status {
		case StatusActive, StatusInactive, StatusSuspended:
		default:
			return fmt.Errorf("invalid status in filter: %s", status)
		}
	}

	return nil
}

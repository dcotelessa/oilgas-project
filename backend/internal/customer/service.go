// backend/internal/customer/service.go
package customer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetCustomersForTenant retrieves customers with filtering and analytics
func (s *Service) GetCustomersForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (*CustomerSearchResponse, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	// Set default pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	customers, err := s.repo.GetAllForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	total, err := s.repo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer count: %w", err)
	}

	page := 1
	if filters.Offset > 0 && filters.Limit > 0 {
		page = (filters.Offset / filters.Limit) + 1
	}

	hasMore := filters.Offset+len(customers) < total

	return &CustomerSearchResponse{
		Customers: customers,
		Total:     total,
		Page:      page,
		PageSize:  len(customers),
		HasMore:   hasMore,
	}, nil
}

// GetCustomerByIDForTenant retrieves a single customer with analytics
func (s *Service) GetCustomerByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, ErrInvalidCustomerID
	}

	customer, err := s.repo.GetByIDForTenant(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// SearchCustomersForTenant provides enhanced search with ranking
func (s *Service) SearchCustomersForTenant(ctx context.Context, tenantID, query string) ([]Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if len(query) > 100 {
		return nil, fmt.Errorf("search query too long (max 100 characters)")
	}

	customers, err := s.repo.SearchForTenant(ctx, tenantID, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return customers, nil
}

// GetCustomerAnalyticsForTenant provides detailed analytics
func (s *Service) GetCustomerAnalyticsForTenant(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	if customerID <= 0 {
		return nil, ErrInvalidCustomerID
	}

	// Verify customer exists and belongs to tenant
	_, err := s.repo.GetByIDForTenant(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}

	analytics, err := s.repo.GetAnalyticsForTenant(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// CreateCustomerForTenant creates a new customer with validation
func (s *Service) CreateCustomerForTenant(ctx context.Context, tenantID string, req CreateCustomerRequest) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	customer := &Customer{
		Customer:                req.Customer,
		BillingAddress:          req.BillingAddress,
		BillingCity:             req.BillingCity,
		BillingState:            req.BillingState,
		BillingZipcode:          req.BillingZipcode,
		Contact:                 req.Contact,
		Phone:                   req.Phone,
		Fax:                     req.Fax,
		Email:                   req.Email,
		PreferredPaymentTerms:   req.PreferredPaymentTerms,
		PreferredShippingMethod: req.PreferredShippingMethod,
		DefaultPORequired:       req.DefaultPORequired,
		CreditLimit:             req.CreditLimit,
		TenantID:                tenantID,
	}

	if err := s.validateCustomer(customer); err != nil {
		return nil, err
	}

	if err := s.repo.CreateForTenant(ctx, tenantID, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// UpdateCustomerForTenant updates an existing customer
func (s *Service) UpdateCustomerForTenant(ctx context.Context, tenantID string, id int, req UpdateCustomerRequest) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, ErrInvalidCustomerID
	}

	// Get existing customer
	customer, err := s.repo.GetByIDForTenant(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Customer != nil {
		customer.Customer = *req.Customer
	}
	if req.BillingAddress != nil {
		customer.BillingAddress = req.BillingAddress
	}
	if req.BillingCity != nil {
		customer.BillingCity = req.BillingCity
	}
	if req.BillingState != nil {
		customer.BillingState = req.BillingState
	}
	if req.BillingZipcode != nil {
		customer.BillingZipcode = req.BillingZipcode
	}
	if req.Contact != nil {
		customer.Contact = req.Contact
	}
	if req.Phone != nil {
		customer.Phone = req.Phone
	}
	if req.Fax != nil {
		customer.Fax = req.Fax
	}
	if req.Email != nil {
		customer.Email = req.Email
	}
	if req.PreferredPaymentTerms != nil {
		customer.PreferredPaymentTerms = req.PreferredPaymentTerms
	}
	if req.PreferredShippingMethod != nil {
		customer.PreferredShippingMethod = req.PreferredShippingMethod
	}
	if req.DefaultPORequired != nil {
		customer.DefaultPORequired = *req.DefaultPORequired
	}
	if req.CreditLimit != nil {
		customer.CreditLimit = req.CreditLimit
	}

	if err := s.validateCustomer(customer); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateForTenant(ctx, tenantID, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// UpdateCustomerContactsForTenant updates contact information
func (s *Service) UpdateCustomerContactsForTenant(ctx context.Context, tenantID string, customerID int, req UpdateContactsRequest) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}

	if customerID <= 0 {
		return ErrInvalidCustomerID
	}

	// Verify customer exists
	_, err := s.repo.GetByIDForTenant(ctx, tenantID, customerID)
	if err != nil {
		return err
	}

	// Validate contacts
	if req.PrimaryContact != nil {
		if err := s.validateContact(req.PrimaryContact); err != nil {
			return fmt.Errorf("invalid primary contact: %w", err)
		}
	}

	if req.BillingContact != nil {
		if err := s.validateContact(req.BillingContact); err != nil {
			return fmt.Errorf("invalid billing contact: %w", err)
		}
	}

	// TODO: Store contacts in separate table or JSON column
	// For now, this would require database schema updates
	return fmt.Errorf("contact updates not yet implemented - requires schema enhancement")
}

// DeleteCustomerForTenant soft deletes a customer
func (s *Service) DeleteCustomerForTenant(ctx context.Context, tenantID string, id int) error {
	if err := s.validateTenantID(tenantID); err != nil {
		return err
	}

	if id <= 0 {
		return ErrInvalidCustomerID
	}

	// TODO: Check for active work orders before allowing deletion
	// For now, proceed with soft delete
	if err := s.repo.DeleteForTenant(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// Enterprise methods for cross-tenant operations (admin only)
func (s *Service) GetCustomersByIDs(ctx context.Context, customerIDs []int) ([]Customer, error) {
	if len(customerIDs) == 0 {
		return []Customer{}, nil
	}

	if len(customerIDs) > 100 {
		return nil, fmt.Errorf("too many customer IDs (max 100)")
	}

	return s.repo.GetCustomersByIDs(ctx, customerIDs)
}

func (s *Service) GetCustomerSummaryByTenant(ctx context.Context) (map[string]int, error) {
	return s.repo.GetCustomerSummaryByTenant(ctx)
}

// Validation methods
func (s *Service) validateTenantID(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if len(tenantID) > 50 {
		return fmt.Errorf("tenant ID too long")
	}
	return nil
}

func (s *Service) validateCustomer(customer *Customer) error {
	if err := customer.Validate(); err != nil {
		return err
	}

	// Additional business logic validation
	if customer.BillingState != nil && *customer.BillingState != "" {
		if !s.isValidStateCode(*customer.BillingState) {
			return fmt.Errorf("invalid state code: %s", *customer.BillingState)
		}
	}

	if customer.Email != nil && *customer.Email != "" {
		if !s.isValidEmail(*customer.Email) {
			return fmt.Errorf("invalid email format")
		}
	}

	if customer.Phone != nil && *customer.Phone != "" {
		if !s.isValidPhone(*customer.Phone) {
			return fmt.Errorf("invalid phone format")
		}
	}

	if customer.CreditLimit != nil && *customer.CreditLimit < 0 {
		return fmt.Errorf("credit limit cannot be negative")
	}

	return nil
}

func (s *Service) validateContact(contact *Contact) error {
	if contact.Name == "" {
		return fmt.Errorf("contact name is required")
	}
	if len(contact.Name) > 255 {
		return fmt.Errorf("contact name too long")
	}
	if contact.Email != nil && *contact.Email != "" {
		if !s.isValidEmail(*contact.Email) {
			return fmt.Errorf("invalid email format")
		}
	}
	if contact.Phone != nil && *contact.Phone != "" {
		if !s.isValidPhone(*contact.Phone) {
			return fmt.Errorf("invalid phone format")
		}
	}
	return nil
}

// Helper validation functions
func (s *Service) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}// backend/internal/customer/service.go
package customer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetCustomersForTenant retrieves customers with filtering and analytics
func (s *Service) GetCustomersForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (*CustomerSearchResponse, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	// Set default pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	customers, err := s.repo.GetAllForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	total, err := s.repo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer count: %w", err)
	}

	page := 1
	if filters.Offset > 0 && filters.Limit > 0 {
		page = (filters.Offset / filters.Limit) + 1
	}

	hasMore := filters.Offset+len(customers) < total

	return &CustomerSearchResponse{
		Customers: customers,
		Total:     total,
		Page:      page,
		PageSize:  len(customers),
		HasMore:   hasMore,
	}, nil
}

// GetCustomerByIDForTenant retrieves a single customer with analytics
func (s *Service) GetCustomerByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, ErrInvalidCustomerID
	}

	customer, err := s.repo.GetByIDForTenant(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// SearchCustomersForTenant provides enhanced search with ranking
func (s *Service) SearchCustomersForTenant(ctx context.Context, tenantID, query string) ([]Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
)
	return emailRegex.MatchString(email)
}

func (s *Service) isValidPhone(phone string) bool {
	// Basic phone validation - can be enhanced
	phoneRegex := regexp.MustCompile(`^[\d\s\-\(\)\+\.]{10,20}// backend/internal/customer/service.go
package customer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetCustomersForTenant retrieves customers with filtering and analytics
func (s *Service) GetCustomersForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (*CustomerSearchResponse, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	// Set default pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	customers, err := s.repo.GetAllForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	total, err := s.repo.GetCountForTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer count: %w", err)
	}

	page := 1
	if filters.Offset > 0 && filters.Limit > 0 {
		page = (filters.Offset / filters.Limit) + 1
	}

	hasMore := filters.Offset+len(customers) < total

	return &CustomerSearchResponse{
		Customers: customers,
		Total:     total,
		Page:      page,
		PageSize:  len(customers),
		HasMore:   hasMore,
	}, nil
}

// GetCustomerByIDForTenant retrieves a single customer with analytics
func (s *Service) GetCustomerByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, ErrInvalidCustomerID
	}

	customer, err := s.repo.GetByIDForTenant(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// SearchCustomersForTenant provides enhanced search with ranking
func (s *Service) SearchCustomersForTenant(ctx context.Context, tenantID, query string) ([]Customer, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
)
	return phoneRegex.MatchString(phone)
}

func (s *Service) isValidStateCode(state string) bool {
	// US state codes validation
	validStates := map[string]bool{
		"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true, "CT": true, "DE": true,
		"FL": true, "GA": true, "HI": true, "ID": true, "IL": true, "IN": true, "IA": true, "KS": true,
		"KY": true, "LA": true, "ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
		"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true, "NM": true, "NY": true,
		"NC": true, "ND": true, "OH": true, "OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
		"SD": true, "TN": true, "TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
		"WI": true, "WY": true, "DC": true,
	}
	return validStates[strings.ToUpper(state)]
}query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}

	if len(

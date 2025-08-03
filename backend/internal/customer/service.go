// backend/internal/customer/service.go
// Customer domain service aligned with existing Customer struct
package customer

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Service handles customer business logic with tenant isolation
type Service struct {
	db *sqlx.DB
}

// NewService creates a new customer service instance
func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// CustomerFilter represents filtering options for customer queries
type CustomerFilter struct {
	TenantID       string
	State          string
	Search         string
	IncludeDeleted bool
	Limit          int
	Offset         int
}

// SetTenantContext sets the database tenant context for row-level security
func (s *Service) SetTenantContext(tenantID string) error {
	query := `SELECT set_tenant_context($1)`
	_, err := s.db.Exec(query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// GetCurrentTenant returns the current tenant context
func (s *Service) GetCurrentTenant() (string, error) {
	var tenantID string
	query := `SELECT get_current_tenant()`
	err := s.db.Get(&tenantID, query)
	if err != nil {
		return "", fmt.Errorf("failed to get current tenant: %w", err)
	}
	return tenantID, nil
}

// GetCustomersByTenant retrieves customers for a specific tenant
func (s *Service) GetCustomersByTenant(tenantID string) ([]Customer, error) {
	if err := s.SetTenantContext(tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			customer_id, customer, billing_address, billing_city, billing_state, 
			billing_zipcode, contact, phone, fax, email, tenant_id,
			deleted, created_at, updated_at
		FROM store.customers_standardized 
		WHERE deleted = false
		ORDER BY customer`

	var customers []Customer
	err := s.db.Select(&customers, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers for tenant %s: %w", tenantID, err)
	}

	return customers, nil
}

// GetCustomerByID retrieves a specific customer by ID within tenant context
func (s *Service) GetCustomerByID(tenantID string, customerID int) (*Customer, error) {
	if err := s.SetTenantContext(tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			customer_id, customer, billing_address, billing_city, billing_state, 
			billing_zipcode, contact, phone, fax, email, tenant_id,
			deleted, created_at, updated_at
		FROM store.customers_standardized 
		WHERE customer_id = $1 AND deleted = false`

	var customer Customer
	err := s.db.Get(&customer, query, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer %d not found for tenant %s", customerID, tenantID)
		}
		return nil, fmt.Errorf("failed to get customer %d: %w", customerID, err)
	}

	return &customer, nil
}

// SearchCustomers performs filtered search with pagination
func (s *Service) SearchCustomers(filter CustomerFilter) ([]Customer, error) {
	if err := s.SetTenantContext(filter.TenantID); err != nil {
		return nil, err
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT 
			customer_id, customer, billing_address, billing_city, billing_state, 
			billing_zipcode, contact, phone, fax, email, tenant_id,
			deleted, created_at, updated_at
		FROM store.customers_standardized WHERE 1=1`

	// Add filters
	if !filter.IncludeDeleted {
		conditions = append(conditions, "deleted = false")
	}

	if filter.State != "" {
		conditions = append(conditions, fmt.Sprintf("billing_state = $%d", argIndex))
		args = append(args, filter.State)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(customer) LIKE $%d OR LOWER(contact) LIKE $%d)", argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Build final query
	query := baseQuery
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY customer"

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++

		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
		}
	}

	var customers []Customer
	err := s.db.Select(&customers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	return customers, nil
}

// CreateCustomer creates a new customer with validation
func (s *Service) CreateCustomer(customer *Customer) error {
	if err := s.ValidateCustomer(customer); err != nil {
		return err
	}

	if err := s.SetTenantContext(customer.TenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO store.customers_standardized (
			customer, billing_address, billing_city, billing_state, billing_zipcode,
			contact, phone, fax, email, tenant_id
		) VALUES (
			:customer, :billing_address, :billing_city, :billing_state, :billing_zipcode,
			:contact, :phone, :fax, :email, :tenant_id
		) RETURNING customer_id`

	rows, err := s.db.NamedQuery(query, customer)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&customer.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to get new customer ID: %w", err)
		}
	}

	return nil
}

// UpdateCustomer updates an existing customer
func (s *Service) UpdateCustomer(customer *Customer) error {
	if err := s.ValidateCustomer(customer); err != nil {
		return err
	}

	if err := s.SetTenantContext(customer.TenantID); err != nil {
		return err
	}

	query := `
		UPDATE store.customers_standardized SET
			customer = :customer,
			billing_address = :billing_address,
			billing_city = :billing_city,
			billing_state = :billing_state,
			billing_zipcode = :billing_zipcode,
			contact = :contact,
			phone = :phone,
			fax = :fax,
			email = :email
		WHERE customer_id = :customer_id AND tenant_id = :tenant_id`

	result, err := s.db.NamedExec(query, customer)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer %d not found for tenant %s", customer.CustomerID, customer.TenantID)
	}

	return nil
}

// SoftDeleteCustomer marks a customer as deleted (soft delete)
func (s *Service) SoftDeleteCustomer(tenantID string, customerID int) error {
	if err := s.SetTenantContext(tenantID); err != nil {
		return err
	}

	query := `UPDATE store.customers_standardized SET deleted = true WHERE customer_id = $1`
	result, err := s.db.Exec(query, customerID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer %d not found for tenant %s", customerID, tenantID)
	}

	return nil
}

// GetCustomerStats returns statistics about customers for a tenant
func (s *Service) GetCustomerStats(tenantID string) (map[string]interface{}, error) {
	if err := s.SetTenantContext(tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			COUNT(*) as total_customers,
			COUNT(CASE WHEN deleted = false THEN 1 END) as active_customers,
			COUNT(CASE WHEN email IS NOT NULL AND email != '' THEN 1 END) as customers_with_email,
			COUNT(DISTINCT billing_state) as states_covered
		FROM store.customers_standardized`

	var stats struct {
		TotalCustomers     int `db:"total_customers"`
		ActiveCustomers    int `db:"active_customers"`
		CustomersWithEmail int `db:"customers_with_email"`
		StatesCovered      int `db:"states_covered"`
	}

	err := s.db.Get(&stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}

	return map[string]interface{}{
		"total_customers":      stats.TotalCustomers,
		"active_customers":     stats.ActiveCustomers,
		"customers_with_email": stats.CustomersWithEmail,
		"states_covered":       stats.StatesCovered,
	}, nil
}

// ValidateCustomer validates customer data
func (s *Service) ValidateCustomer(customer *Customer) error {
	var errors []string

	if strings.TrimSpace(customer.Customer) == "" {
		errors = append(errors, "customer name is required")
	}

	if strings.TrimSpace(customer.TenantID) == "" {
		errors = append(errors, "tenant ID is required")
	}

	// Validate email format if provided
	if customer.Email != nil && *customer.Email != "" && !isValidEmail(*customer.Email) {
		errors = append(errors, "invalid email address format")
	}

	// Validate state code if provided
	if customer.BillingState != nil && *customer.BillingState != "" && len(*customer.BillingState) > 2 {
		errors = append(errors, "billing state should be 2-letter code")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// isValidEmail performs comprehensive email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	// Must contain exactly one @
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}
	
	// Split on @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	local := parts[0]
	domain := parts[1]
	
	// Local part validations
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	
	// Domain part validations
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	
	// Domain must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}
	
	// Domain cannot start or end with dot
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	
	// Domain cannot have consecutive dots
	if strings.Contains(domain, "..") {
		return false
	}
	
	// Split domain parts
	domainParts := strings.Split(domain, ".")
	for _, part := range domainParts {
		if len(part) == 0 {
			return false
		}
	}
	
	// Basic character validation (simplified)
	for _, char := range email {
		if char < 32 || char > 126 {
			return false // Non-printable ASCII
		}
	}
	
	return true
}
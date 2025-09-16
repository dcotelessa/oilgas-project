// backend/internal/customer/repository.go
package customer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	
	"oilgas-backend/internal/shared/database"
)

type Repository interface {
	GetCustomerByID(ctx context.Context, tenantID string, id int) (*Customer, error)
	SearchCustomers(ctx context.Context, tenantID string, filters SearchFilters) ([]Customer, int, error)
	CreateCustomer(ctx context.Context, tenantID string, customer *Customer) error
	UpdateCustomer(ctx context.Context, tenantID string, customer *Customer) error
	DeleteCustomer(ctx context.Context, tenantID string, id int) error
	
	GetCustomerContacts(ctx context.Context, tenantID string, customerID int) ([]CustomerContact, error)
	AddCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error
	UpdateCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error
	RemoveCustomerContact(ctx context.Context, tenantID string, customerID, authUserID int) error
	
	GetCustomerAnalytics(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error)
}

type repository struct {
	dbManager *database.DatabaseManager
}

func NewRepository(dbManager *database.DatabaseManager) Repository {
	return &repository{dbManager: dbManager}
}

func (r *repository) GetCustomerByID(ctx context.Context, tenantID string, id int) (*Customer, error) {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	query := `
		SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms,
		       billing_street, billing_city, billing_state, billing_zip_code, billing_country,
		       is_active, created_at, updated_at
		FROM store.customers 
		WHERE id = $1 AND tenant_id = $2 AND is_active = true`
	
	var c Customer
	err = db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.CompanyCode, &c.Status,
		&c.TaxID, &c.PaymentTerms,
		&c.BillingStreet, &c.BillingCity, &c.BillingState, &c.BillingZip, &c.BillingCountry,
		&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	
	return &c, nil
}

func (r *repository) CreateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	query := `
		INSERT INTO store.customers (
			tenant_id, name, company_code, status, tax_id, payment_terms,
			billing_street, billing_city, billing_state, billing_zip_code, billing_country,
			is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`
	
	err = db.QueryRowContext(ctx, query,
		tenantID, customer.Name, customer.CompanyCode, customer.Status,
		customer.TaxID, customer.PaymentTerms,
		customer.BillingStreet, customer.BillingCity, customer.BillingState, 
		customer.BillingZip, customer.BillingCountry, true,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	
	customer.TenantID = tenantID
	customer.IsActive = true
	return nil
}

func (r *repository) GetCustomerContacts(ctx context.Context, tenantID string, customerID int) ([]CustomerContact, error) {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	query := `
		SELECT id, customer_id, auth_user_id, contact_type, is_primary, is_active,
		       full_name, email, created_at, updated_at
		FROM store.customer_contacts
		WHERE customer_id = $1 AND is_active = true
		ORDER BY is_primary DESC, created_at ASC`
	
	rows, err := db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer contacts: %w", err)
	}
	defer rows.Close()
	
	var contacts []CustomerContact
	for rows.Next() {
		var contact CustomerContact
		err := rows.Scan(
			&contact.ID, &contact.CustomerID, &contact.AuthUserID, &contact.ContactType,
			&contact.IsPrimary, &contact.IsActive,
			&contact.FullName, &contact.Email, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer contact: %w", err)
		}
		contacts = append(contacts, contact)
	}
	
	return contacts, nil
}

func (r *repository) GetCustomerAnalytics(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var exists bool
	err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM store.customers WHERE id = $1 AND tenant_id = $2)", customerID, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("customer not found")
	}
	
	return &CustomerAnalytics{
		CustomerID:      customerID,
		TotalWorkOrders: 0,
		ActiveOrders:    0,
		TotalRevenue:    0,
		AvgOrderValue:   0,
		LastOrderDate:   nil,
	}, nil
}

func (r *repository) SearchCustomers(ctx context.Context, tenantID string, filters SearchFilters) ([]Customer, int, error) {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tenant database: %w", err)
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID)
	argIndex++

	conditions = append(conditions, "is_active = true")

	if filters.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Name+"%")
		argIndex++
	}

	if filters.CompanyCode != "" {
		conditions = append(conditions, fmt.Sprintf("company_code = $%d", argIndex))
		args = append(args, filters.CompanyCode)
		argIndex++
	}

	if len(filters.Status) > 0 {
		statusPlaceholders := make([]string, len(filters.Status))
		for i, status := range filters.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	if filters.TaxID != "" {
		conditions = append(conditions, fmt.Sprintf("tax_id = $%d", argIndex))
		args = append(args, filters.TaxID)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")
	
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM store.customers WHERE %s", whereClause)
	var total int
	err = db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, company_code, status, tax_id, payment_terms,
		       billing_street, billing_city, billing_state, billing_zip_code, billing_country,
		       is_active, created_at, updated_at
		FROM store.customers
		WHERE %s
		ORDER BY name ASC`, whereClause)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var c Customer
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.CompanyCode, &c.Status,
			&c.TaxID, &c.PaymentTerms,
			&c.BillingStreet, &c.BillingCity, &c.BillingState, &c.BillingZip, &c.BillingCountry,
			&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

func (r *repository) UpdateCustomer(ctx context.Context, tenantID string, customer *Customer) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	query := `
		UPDATE store.customers 
		SET name = $3, company_code = $4, status = $5, tax_id = $6, payment_terms = $7,
		    billing_street = $8, billing_city = $9, billing_state = $10, 
		    billing_zip_code = $11, billing_country = $12, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND is_active = true
		RETURNING updated_at`

	err = db.QueryRowContext(ctx, query,
		customer.ID, tenantID, customer.Name, customer.CompanyCode, customer.Status,
		customer.TaxID, customer.PaymentTerms,
		customer.BillingStreet, customer.BillingCity, customer.BillingState,
		customer.BillingZip, customer.BillingCountry,
	).Scan(&customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("customer not found")
		}
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

func (r *repository) DeleteCustomer(ctx context.Context, tenantID string, id int) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	result, err := db.ExecContext(ctx,
		"UPDATE store.customers SET is_active = false, updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND is_active = true",
		id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

func (r *repository) AddCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	var customerExists bool
	err = db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM store.customers WHERE id = $1 AND tenant_id = $2 AND is_active = true)",
		contact.CustomerID, tenantID).Scan(&customerExists)
	if err != nil || !customerExists {
		return fmt.Errorf("customer not found")
	}

	query := `
		INSERT INTO store.customer_contacts (
			customer_id, auth_user_id, contact_type, is_primary, is_active, full_name, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err = db.QueryRowContext(ctx, query,
		contact.CustomerID, contact.AuthUserID, contact.ContactType,
		contact.IsPrimary, true, contact.FullName, contact.Email,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to add customer contact: %w", err)
	}

	contact.IsActive = true
	return nil
}

func (r *repository) UpdateCustomerContact(ctx context.Context, tenantID string, contact *CustomerContact) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	query := `
		UPDATE store.customer_contacts 
		SET contact_type = $3, is_primary = $4, full_name = $5, email = $6, updated_at = NOW()
		WHERE id = $1 AND customer_id IN (
			SELECT id FROM store.customers WHERE tenant_id = $2 AND is_active = true
		) AND is_active = true
		RETURNING updated_at`

	err = db.QueryRowContext(ctx, query,
		contact.ID, tenantID, contact.ContactType, contact.IsPrimary,
		contact.FullName, contact.Email,
	).Scan(&contact.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("customer contact not found")
		}
		return fmt.Errorf("failed to update customer contact: %w", err)
	}

	return nil
}

func (r *repository) RemoveCustomerContact(ctx context.Context, tenantID string, customerID, authUserID int) error {
	db, err := r.dbManager.GetTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	result, err := db.ExecContext(ctx, `
		UPDATE store.customer_contacts 
		SET is_active = false, updated_at = NOW()
		WHERE customer_id = $1 AND auth_user_id = $2 
		AND customer_id IN (
			SELECT id FROM store.customers WHERE tenant_id = $3 AND is_active = true
		) AND is_active = true`,
		customerID, authUserID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to remove customer contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer contact not found")
	}

	return nil
}

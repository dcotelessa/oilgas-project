// backend/internal/customer/repository.go
package customer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetAllForTenant retrieves customers with filtering and analytics
func (r *Repository) GetAllForTenant(ctx context.Context, tenantID string, filters CustomerFilters) ([]Customer, error) {
	query := `
		SELECT 
			c.customer_id, c.customer, c.billing_address, c.billing_city, 
			c.billing_state, c.billing_zipcode, c.contact, c.phone, c.fax, 
			c.email, c.tenant_id, c.preferred_payment_terms, 
			c.preferred_shipping_method, c.default_po_required, c.credit_limit,
			c.imported_at, c.deleted, c.created_at, c.updated_at,
			COUNT(DISTINCT wo.work_order) as work_order_count,
			COALESCE(SUM(wo.total_amount), 0) as total_revenue,
			MAX(wo.created_at) as last_activity
		FROM customers c
		LEFT JOIN work_orders wo ON wo.customer_id = c.customer_id AND wo.tenant_id = c.tenant_id
		WHERE c.tenant_id = $1 AND c.deleted = false`

	args := []interface{}{tenantID}
	argIndex := 2

	// Apply filters
	if filters.Query != "" {
		query += fmt.Sprintf(" AND (c.customer ILIKE $%d OR c.contact ILIKE $%d OR c.email ILIKE $%d)", argIndex, argIndex, argIndex)
		searchTerm := "%" + filters.Query + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	if filters.State != "" {
		query += fmt.Sprintf(" AND c.billing_state = $%d", argIndex)
		args = append(args, filters.State)
		argIndex++
	}

	if filters.HasOrders != nil {
		if *filters.HasOrders {
			query += " AND wo.customer_id IS NOT NULL"
		} else {
			query += " AND wo.customer_id IS NULL"
		}
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND c.created_at >= $%d", argIndex)
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND c.created_at <= $%d", argIndex)
		args = append(args, filters.DateTo)
		argIndex++
	}

	// Group by for aggregations
	query += ` GROUP BY c.customer_id, c.customer, c.billing_address, c.billing_city, 
		c.billing_state, c.billing_zipcode, c.contact, c.phone, c.fax, 
		c.email, c.tenant_id, c.preferred_payment_terms, 
		c.preferred_shipping_method, c.default_po_required, c.credit_limit,
		c.imported_at, c.deleted, c.created_at, c.updated_at`

	// Apply sorting
	orderBy := "c.customer ASC"
	if filters.SortBy != "" {
		direction := "ASC"
		if strings.ToUpper(filters.SortOrder) == "DESC" {
			direction = "DESC"
		}
		
		switch filters.SortBy {
		case "name":
			orderBy = fmt.Sprintf("c.customer %s", direction)
		case "created_at":
			orderBy = fmt.Sprintf("c.created_at %s", direction)
		case "work_orders":
			orderBy = fmt.Sprintf("work_order_count %s", direction)
		case "revenue":
			orderBy = fmt.Sprintf("total_revenue %s", direction)
		case "last_activity":
			orderBy = fmt.Sprintf("last_activity %s", direction)
		}
	}
	query += " ORDER BY " + orderBy

	// Apply pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var customers []Customer
	err := r.db.SelectContext(ctx, &customers, query, args...)
	return customers, err
}

// GetByIDForTenant retrieves a customer by ID with analytics
func (r *Repository) GetByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error) {
	query := `
		SELECT 
			c.customer_id, c.customer, c.billing_address, c.billing_city, 
			c.billing_state, c.billing_zipcode, c.contact, c.phone, c.fax, 
			c.email, c.tenant_id, c.preferred_payment_terms, 
			c.preferred_shipping_method, c.default_po_required, c.credit_limit,
			c.imported_at, c.deleted, c.created_at, c.updated_at,
			COUNT(DISTINCT wo.work_order) as work_order_count,
			COALESCE(SUM(wo.total_amount), 0) as total_revenue,
			MAX(wo.created_at) as last_activity,
			COUNT(DISTINCT CASE WHEN wo.status IN ('pending', 'in_progress') THEN wo.work_order END) as active_jobs
		FROM customers c
		LEFT JOIN work_orders wo ON wo.customer_id = c.customer_id AND wo.tenant_id = c.tenant_id
		WHERE c.customer_id = $1 AND c.tenant_id = $2 AND c.deleted = false
		GROUP BY c.customer_id, c.customer, c.billing_address, c.billing_city, 
			c.billing_state, c.billing_zipcode, c.contact, c.phone, c.fax, 
			c.email, c.tenant_id, c.preferred_payment_terms, 
			c.preferred_shipping_method, c.default_po_required, c.credit_limit,
			c.imported_at, c.deleted, c.created_at, c.updated_at`

	var customer Customer
	err := r.db.GetContext(ctx, &customer, query, id, tenantID)
	if err == sql.ErrNoRows {
		return nil, ErrCustomerNotFound
	}
	return &customer, err
}

// SearchForTenant provides enhanced search capabilities
func (r *Repository) SearchForTenant(ctx context.Context, tenantID, query string) ([]Customer, error) {
	searchQuery := `
		SELECT 
			customer_id, customer, billing_address, billing_city, 
			billing_state, billing_zipcode, contact, phone, fax, 
			email, tenant_id, preferred_payment_terms, 
			preferred_shipping_method, default_po_required, credit_limit,
			imported_at, deleted, created_at, updated_at
		FROM customers 
		WHERE tenant_id = $1 AND deleted = false
		AND (
			customer ILIKE $2 OR 
			contact ILIKE $2 OR 
			email ILIKE $2 OR
			billing_city ILIKE $2 OR
			phone ILIKE $2
		)
		ORDER BY 
			CASE WHEN customer ILIKE $3 THEN 1 ELSE 2 END,
			customer ASC
		LIMIT 20`

	searchTerm := "%" + query + "%"
	exactTerm := query + "%"

	var customers []Customer
	err := r.db.SelectContext(ctx, &customers, searchQuery, tenantID, searchTerm, exactTerm)
	return customers, err
}

// GetCountForTenant returns filtered count
func (r *Repository) GetCountForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (int, error) {
	query := `
		SELECT COUNT(DISTINCT c.customer_id)
		FROM customers c
		LEFT JOIN work_orders wo ON wo.customer_id = c.customer_id AND wo.tenant_id = c.tenant_id
		WHERE c.tenant_id = $1 AND c.deleted = false`

	args := []interface{}{tenantID}
	argIndex := 2

	// Apply same filters as GetAllForTenant
	if filters.Query != "" {
		query += fmt.Sprintf(" AND (c.customer ILIKE $%d OR c.contact ILIKE $%d OR c.email ILIKE $%d)", argIndex, argIndex, argIndex)
		searchTerm := "%" + filters.Query + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	if filters.State != "" {
		query += fmt.Sprintf(" AND c.billing_state = $%d", argIndex)
		args = append(args, filters.State)
		argIndex++
	}

	if filters.HasOrders != nil {
		if *filters.HasOrders {
			query += " AND wo.customer_id IS NOT NULL"
		} else {
			query += " AND wo.customer_id IS NULL"
		}
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND c.created_at >= $%d", argIndex)
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND c.created_at <= $%d", argIndex)
		args = append(args, filters.DateTo)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// GetAnalyticsForTenant provides detailed customer analytics
func (r *Repository) GetAnalyticsForTenant(ctx context.Context, tenantID string, customerID int) (*CustomerAnalytics, error) {
	analytics := &CustomerAnalytics{CustomerID: customerID}

	// Get basic metrics
	err := r.db.GetContext(ctx, analytics, `
		SELECT 
			COUNT(*) as total_work_orders,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_orders,
			COUNT(CASE WHEN status IN ('pending', 'in_progress') THEN 1 END) as pending_orders,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(AVG(total_amount), 0) as average_job_value,
			MAX(created_at) as last_order_date
		FROM work_orders 
		WHERE customer_id = $1 AND tenant_id = $2`,
		customerID, tenantID)
	
	if err != nil {
		return nil, err
	}

	// Get recent work orders
	err = r.db.SelectContext(ctx, &analytics.RecentWorkOrders, `
		SELECT work_order, service_date, status, total_amount as total_value, description
		FROM work_orders 
		WHERE customer_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC 
		LIMIT 10`,
		customerID, tenantID)
	
	if err != nil {
		return nil, err
	}

	// Get monthly revenue (last 12 months)
	err = r.db.SelectContext(ctx, &analytics.MonthlyRevenue, `
		SELECT 
			TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') as month,
			COALESCE(SUM(total_amount), 0) as revenue,
			COUNT(*) as orders
		FROM work_orders 
		WHERE customer_id = $1 AND tenant_id = $2 
		AND created_at >= NOW() - INTERVAL '12 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month DESC`,
		customerID, tenantID)

	return analytics, err
}

// CreateForTenant creates a new customer
func (r *Repository) CreateForTenant(ctx context.Context, tenantID string, customer *Customer) error {
	customer.TenantID = tenantID
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()

	query := `
		INSERT INTO customers (
			customer, billing_address, billing_city, billing_state, billing_zipcode,
			contact, phone, fax, email, tenant_id, preferred_payment_terms,
			preferred_shipping_method, default_po_required, credit_limit,
			deleted, created_at, updated_at
		) VALUES (
			:customer, :billing_address, :billing_city, :billing_state, :billing_zipcode,
			:contact, :phone, :fax, :email, :tenant_id, :preferred_payment_terms,
			:preferred_shipping_method, :default_po_required, :credit_limit,
			false, :created_at, :updated_at
		) RETURNING customer_id`

	rows, err := r.db.NamedQueryContext(ctx, query, customer)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&customer.CustomerID)
	}
	return fmt.Errorf("failed to get customer ID")
}

// UpdateForTenant updates an existing customer
func (r *Repository) UpdateForTenant(ctx context.Context, tenantID string, customer *Customer) error {
	customer.UpdatedAt = time.Now()

	query := `
		UPDATE customers SET
			customer = :customer,
			billing_address = :billing_address,
			billing_city = :billing_city,
			billing_state = :billing_state,
			billing_zipcode = :billing_zipcode,
			contact = :contact,
			phone = :phone,
			fax = :fax,
			email = :email,
			preferred_payment_terms = :preferred_payment_terms,
			preferred_shipping_method = :preferred_shipping_method,
			default_po_required = :default_po_required,
			credit_limit = :credit_limit,
			updated_at = :updated_at
		WHERE customer_id = :customer_id AND tenant_id = :tenant_id AND deleted = false`

	result, err := r.db.NamedExecContext(ctx, query, customer)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// DeleteForTenant soft deletes a customer
func (r *Repository) DeleteForTenant(ctx context.Context, tenantID string, id int) error {
	query := `
		UPDATE customers 
		SET deleted = true, updated_at = NOW()
		WHERE customer_id = $1 AND tenant_id = $2 AND deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// Enterprise queries for cross-tenant operations (admin only)
func (r *Repository) GetCustomersByIDs(ctx context.Context, customerIDs []int) ([]Customer, error) {
	if len(customerIDs) == 0 {
		return []Customer{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT customer_id, customer, billing_city, billing_state, 
			   tenant_id, contact, phone, email, created_at
		FROM customers 
		WHERE customer_id IN (?) AND deleted = false
		ORDER BY customer`, customerIDs)
	
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var customers []Customer
	err = r.db.SelectContext(ctx, &customers, query, args...)
	return customers, err
}

func (r *Repository) GetCustomerSummaryByTenant(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT tenant_id, COUNT(*) as count
		FROM customers 
		WHERE deleted = false
		GROUP BY tenant_id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]int)
	for rows.Next() {
		var tenantID string
		var count int
		if err := rows.Scan(&tenantID, &count); err != nil {
			return nil, err
		}
		summary[tenantID] = count
	}

	return summary, rows.Err()
}

// internal/repository/postgres/customer.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/models"
)

type CustomerRepo struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) repository.CustomerRepository {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) GetAll(ctx context.Context) ([]models.Customer, error) {
	query := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state, 
		       billing_zipcode, contact, phone, fax, email, deleted, created_at
		FROM store.customers 
		WHERE deleted = false 
		ORDER BY customer`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity,
			&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, &c.Fax, 
			&c.Email, &c.Deleted, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	
	return customers, rows.Err()
}

func (r *CustomerRepo) GetByID(ctx context.Context, id int) (*models.Customer, error) {
	query := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state,
		       billing_zipcode, contact, phone, fax, email, deleted, created_at
		FROM store.customers 
		WHERE customer_id = $1 AND deleted = false`
	
	var c models.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity,
		&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, 
		&c.Fax, &c.Email, &c.Deleted, &c.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer %d: %w", id, err)
	}
	
	return &c, nil
}

func (r *CustomerRepo) Search(ctx context.Context, query string) ([]models.Customer, error) {
	searchQuery := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state,
		       billing_zipcode, contact, phone, fax, email, deleted, created_at
		FROM store.customers 
		WHERE deleted = false 
		  AND (customer ILIKE $1 OR contact ILIKE $1 OR email ILIKE $1)
		ORDER BY customer`
	
	searchTerm := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity,
			&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, &c.Fax,
			&c.Email, &c.Deleted, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	
	return customers, rows.Err()
}

func (r *CustomerRepo) Create(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO store.customers (customer, billing_address, billing_city, billing_state,
			billing_zipcode, contact, phone, fax, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING customer_id, created_at`
	
	err := r.db.QueryRowContext(ctx, query, customer.Customer, customer.BillingAddress,
		customer.BillingCity, customer.BillingState, customer.BillingZipcode,
		customer.Contact, customer.Phone, customer.Fax, customer.Email).
		Scan(&customer.CustomerID, &customer.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	
	return nil
}

func (r *CustomerRepo) Update(ctx context.Context, customer *models.Customer) error {
	query := `
		UPDATE store.customers 
		SET customer = $2, billing_address = $3, billing_city = $4, billing_state = $5,
		    billing_zipcode = $6, contact = $7, phone = $8, fax = $9, email = $10
		WHERE customer_id = $1 AND deleted = false`
	
	result, err := r.db.ExecContext(ctx, query, customer.CustomerID, customer.Customer,
		customer.BillingAddress, customer.BillingCity, customer.BillingState,
		customer.BillingZipcode, customer.Contact, customer.Phone, customer.Fax, customer.Email)
	
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("customer %d not found or already deleted", customer.CustomerID)
	}
	
	return nil
}

func (r *CustomerRepo) Delete(ctx context.Context, id int) error {
	query := `UPDATE store.customers SET deleted = true WHERE customer_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("customer %d not found", id)
	}
	
	return nil
}

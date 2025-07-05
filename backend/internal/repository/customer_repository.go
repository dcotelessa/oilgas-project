// backend/internal/repository/customer_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) GetAll(ctx context.Context) ([]models.Customer, error) {
	query := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state, 
		       billing_zipcode, contact, phone, fax, email, 
		       color1, color2, color3, color4, color5,
		       loss1, loss2, loss3, loss4, loss5,
		       wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
		       wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
		       deleted, created_at
		FROM store.customers 
		WHERE deleted = false 
		ORDER BY customer
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(
			&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity, 
			&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, 
			&c.Fax, &c.Email,
			&c.Color1, &c.Color2, &c.Color3, &c.Color4, &c.Color5,
			&c.Loss1, &c.Loss2, &c.Loss3, &c.Loss4, &c.Loss5,
			&c.WSColor1, &c.WSColor2, &c.WSColor3, &c.WSColor4, &c.WSColor5,
			&c.WSLoss1, &c.WSLoss2, &c.WSLoss3, &c.WSLoss4, &c.WSLoss5,
			&c.Deleted, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, rows.Err()
}

func (r *customerRepository) GetByID(ctx context.Context, id int) (*models.Customer, error) {
	query := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state, 
		       billing_zipcode, contact, phone, fax, email,
		       color1, color2, color3, color4, color5,
		       loss1, loss2, loss3, loss4, loss5,
		       wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
		       wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
		       deleted, created_at
		FROM store.customers 
		WHERE customer_id = $1 AND deleted = false
	`
	
	var c models.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity, 
		&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, 
		&c.Fax, &c.Email,
		&c.Color1, &c.Color2, &c.Color3, &c.Color4, &c.Color5,
		&c.Loss1, &c.Loss2, &c.Loss3, &c.Loss4, &c.Loss5,
		&c.WSColor1, &c.WSColor2, &c.WSColor3, &c.WSColor4, &c.WSColor5,
		&c.WSLoss1, &c.WSLoss2, &c.WSLoss3, &c.WSLoss4, &c.WSLoss5,
		&c.Deleted, &c.CreatedAt,
	)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	
	return &c, nil
}

func (r *customerRepository) Create(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO store.customers (
			customer, billing_address, billing_city, billing_state, 
			billing_zipcode, contact, phone, fax, email,
			color1, color2, color3, color4, color5,
			loss1, loss2, loss3, loss4, loss5,
			wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
			wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
			deleted, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 
		          $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
		          $20, $21, $22, $23, $24, $25, $26, $27, $28, $29,
		          false, NOW())
		RETURNING customer_id, created_at
	`
	
	err := r.db.QueryRow(ctx, query,
		customer.Customer, customer.BillingAddress, customer.BillingCity, 
		customer.BillingState, customer.BillingZipcode, customer.Contact, 
		customer.Phone, customer.Fax, customer.Email,
		customer.Color1, customer.Color2, customer.Color3, customer.Color4, customer.Color5,
		customer.Loss1, customer.Loss2, customer.Loss3, customer.Loss4, customer.Loss5,
		customer.WSColor1, customer.WSColor2, customer.WSColor3, customer.WSColor4, customer.WSColor5,
		customer.WSLoss1, customer.WSLoss2, customer.WSLoss3, customer.WSLoss4, customer.WSLoss5,
	).Scan(&customer.CustomerID, &customer.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	
	return nil
}

func (r *customerRepository) Update(ctx context.Context, customer *models.Customer) error {
	query := `
		UPDATE store.customers 
		SET customer = $2, billing_address = $3, billing_city = $4, billing_state = $5,
		    billing_zipcode = $6, contact = $7, phone = $8, fax = $9, email = $10,
		    color1 = $11, color2 = $12, color3 = $13, color4 = $14, color5 = $15,
		    loss1 = $16, loss2 = $17, loss3 = $18, loss4 = $19, loss5 = $20,
		    wscolor1 = $21, wscolor2 = $22, wscolor3 = $23, wscolor4 = $24, wscolor5 = $25,
		    wsloss1 = $26, wsloss2 = $27, wsloss3 = $28, wsloss4 = $29, wsloss5 = $30
		WHERE customer_id = $1 AND deleted = false
	`
	
	result, err := r.db.Exec(ctx, query, customer.CustomerID,
		customer.Customer, customer.BillingAddress, customer.BillingCity, 
		customer.BillingState, customer.BillingZipcode, customer.Contact, 
		customer.Phone, customer.Fax, customer.Email,
		customer.Color1, customer.Color2, customer.Color3, customer.Color4, customer.Color5,
		customer.Loss1, customer.Loss2, customer.Loss3, customer.Loss4, customer.Loss5,
		customer.WSColor1, customer.WSColor2, customer.WSColor3, customer.WSColor4, customer.WSColor5,
		customer.WSLoss1, customer.WSLoss2, customer.WSLoss3, customer.WSLoss4, customer.WSLoss5,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found or already deleted")
	}
	
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE store.customers SET deleted = true WHERE customer_id = $1 AND deleted = false`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found or already deleted")
	}
	
	return nil
}

func (r *customerRepository) ExistsByName(ctx context.Context, name string, excludeID ...int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM store.customers WHERE LOWER(customer) = LOWER($1) AND deleted = false`
	args := []interface{}{name}
	
	if len(excludeID) > 0 && excludeID[0] != 0 {
		query += " AND customer_id != $2"
		args = append(args, excludeID[0])
	}
	query += ")"
	
	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (r *customerRepository) HasActiveInventory(ctx context.Context, customerID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM store.inventory 
			WHERE customer_id = $1 AND deleted = false
		)
	`
	
	var hasInventory bool
	err := r.db.QueryRow(ctx, query, customerID).Scan(&hasInventory)
	return hasInventory, err
}

func (r *customerRepository) GetTotalCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM store.customers WHERE deleted = false`
	
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *customerRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error) {
	searchTerm := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	
	whereClause := `
		WHERE deleted = false AND (
			LOWER(customer) LIKE $1 OR
			LOWER(contact) LIKE $1 OR
			LOWER(phone) LIKE $1 OR
			LOWER(email) LIKE $1 OR
			LOWER(billing_city) LIKE $1 OR
			LOWER(billing_state) LIKE $1
		)
	`
	
	// Count total
	countQuery := "SELECT COUNT(*) FROM store.customers " + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}
	
	// Get results
	searchQuery := `
		SELECT customer_id, customer, billing_address, billing_city, billing_state, 
		       billing_zipcode, contact, phone, fax, email, 
		       color1, color2, color3, color4, color5,
		       loss1, loss2, loss3, loss4, loss5,
		       wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
		       wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
		       deleted, created_at
		FROM store.customers ` + whereClause + `
		ORDER BY 
			CASE WHEN LOWER(customer) LIKE $1 THEN 1
			     WHEN LOWER(contact) LIKE $1 THEN 2
			     ELSE 3 END,
			customer
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()
	
	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(
			&c.CustomerID, &c.Customer, &c.BillingAddress, &c.BillingCity, 
			&c.BillingState, &c.BillingZipcode, &c.Contact, &c.Phone, 
			&c.Fax, &c.Email,
			&c.Color1, &c.Color2, &c.Color3, &c.Color4, &c.Color5,
			&c.Loss1, &c.Loss2, &c.Loss3, &c.Loss4, &c.Loss5,
			&c.WSColor1, &c.WSColor2, &c.WSColor3, &c.WSColor4, &c.WSColor5,
			&c.WSLoss1, &c.WSLoss2, &c.WSLoss3, &c.WSLoss4, &c.WSLoss5,
			&c.Deleted, &c.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	
	return customers, total, rows.Err()
}

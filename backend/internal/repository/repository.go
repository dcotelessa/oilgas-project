// backend/internal/repository/repository.go
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"oilgas-backend/internal/models"
	"oilgas-backend/pkg/validation"
)

type Repositories struct {
	Customer  CustomerRepository
	Inventory InventoryRepository
}

func New(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Customer:  NewCustomerRepository(db),
		Inventory: NewInventoryRepository(db),
	}
}

// Customer repository
type CustomerRepository interface {
	// Basic CRUD operations
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, id int) (*models.Customer, error)
	Create(ctx context.Context, req *validation.CustomerValidation) (*models.Customer, error)
	Update(ctx context.Context, id int, req *validation.CustomerValidation) (*models.Customer, error)
	Delete(ctx context.Context, id int) error
	
	// Validation and checks
	ExistsByName(ctx context.Context, name string, excludeID ...int) (bool, error)
	HasActiveInventory(ctx context.Context, customerID int) (bool, error)
}

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
		
		// Use pgtype.Text for nullable fields
		var billingAddress, billingCity, billingState, billingZipcode pgtype.Text
		var contact, phone, fax, email pgtype.Text
		var color1, color2, color3, color4, color5 pgtype.Text
		var loss1, loss2, loss3, loss4, loss5 pgtype.Text
		var wscolor1, wscolor2, wscolor3, wscolor4, wscolor5 pgtype.Text
		var wsloss1, wsloss2, wsloss3, wsloss4, wsloss5 pgtype.Text
		
		err := rows.Scan(
			&c.CustomerID, &c.CustomerName,
			&billingAddress, &billingCity, &billingState, &billingZipcode,
			&contact, &phone, &fax, &email,
			&color1, &color2, &color3, &color4, &color5,
			&loss1, &loss2, &loss3, &loss4, &loss5,
			&wscolor1, &wscolor2, &wscolor3, &wscolor4, &wscolor5,
			&wsloss1, &wsloss2, &wsloss3, &wsloss4, &wsloss5,
			&c.Deleted, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		
		// Convert pgtype.Text to string
		c.BillingAddress = pgTextToString(billingAddress)
		c.BillingCity = pgTextToString(billingCity)
		c.BillingState = pgTextToString(billingState)
		c.BillingZipcode = pgTextToString(billingZipcode)
		c.Contact = pgTextToString(contact)
		c.Phone = pgTextToString(phone)
		c.Fax = pgTextToString(fax)
		c.Email = pgTextToString(email)
		c.Color1 = pgTextToString(color1)
		c.Color2 = pgTextToString(color2)
		c.Color3 = pgTextToString(color3)
		c.Color4 = pgTextToString(color4)
		c.Color5 = pgTextToString(color5)
		c.Loss1 = pgTextToString(loss1)
		c.Loss2 = pgTextToString(loss2)
		c.Loss3 = pgTextToString(loss3)
		c.Loss4 = pgTextToString(loss4)
		c.Loss5 = pgTextToString(loss5)
		c.WSColor1 = pgTextToString(wscolor1)
		c.WSColor2 = pgTextToString(wscolor2)
		c.WSColor3 = pgTextToString(wscolor3)
		c.WSColor4 = pgTextToString(wscolor4)
		c.WSColor5 = pgTextToString(wscolor5)
		c.WSLoss1 = pgTextToString(wsloss1)
		c.WSLoss2 = pgTextToString(wsloss2)
		c.WSLoss3 = pgTextToString(wsloss3)
		c.WSLoss4 = pgTextToString(wsloss4)
		c.WSLoss5 = pgTextToString(wsloss5)
		
		customers = append(customers, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return customers, nil
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
	
	// Use pgtype.Text for nullable fields
	var billingAddress, billingCity, billingState, billingZipcode pgtype.Text
	var contact, phone, fax, email pgtype.Text
	var color1, color2, color3, color4, color5 pgtype.Text
	var loss1, loss2, loss3, loss4, loss5 pgtype.Text
	var wscolor1, wscolor2, wscolor3, wscolor4, wscolor5 pgtype.Text
	var wsloss1, wsloss2, wsloss3, wsloss4, wsloss5 pgtype.Text
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.CustomerID, &c.CustomerName,
		&billingAddress, &billingCity, &billingState, &billingZipcode,
		&contact, &phone, &fax, &email,
		&color1, &color2, &color3, &color4, &color5,
		&loss1, &loss2, &loss3, &loss4, &loss5,
		&wscolor1, &wscolor2, &wscolor3, &wscolor4, &wscolor5,
		&wsloss1, &wsloss2, &wsloss3, &wsloss4, &wsloss5,
		&c.Deleted, &c.CreatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	
	// Convert pgtype.Text to string
	c.BillingAddress = pgTextToString(billingAddress)
	c.BillingCity = pgTextToString(billingCity)
	c.BillingState = pgTextToString(billingState)
	c.BillingZipcode = pgTextToString(billingZipcode)
	c.Contact = pgTextToString(contact)
	c.Phone = pgTextToString(phone)
	c.Fax = pgTextToString(fax)
	c.Email = pgTextToString(email)
	c.Color1 = pgTextToString(color1)
	c.Color2 = pgTextToString(color2)
	c.Color3 = pgTextToString(color3)
	c.Color4 = pgTextToString(color4)
	c.Color5 = pgTextToString(color5)
	c.Loss1 = pgTextToString(loss1)
	c.Loss2 = pgTextToString(loss2)
	c.Loss3 = pgTextToString(loss3)
	c.Loss4 = pgTextToString(loss4)
	c.Loss5 = pgTextToString(loss5)
	c.WSColor1 = pgTextToString(wscolor1)
	c.WSColor2 = pgTextToString(wscolor2)
	c.WSColor3 = pgTextToString(wscolor3)
	c.WSColor4 = pgTextToString(wscolor4)
	c.WSColor5 = pgTextToString(wscolor5)
	c.WSLoss1 = pgTextToString(wsloss1)
	c.WSLoss2 = pgTextToString(wsloss2)
	c.WSLoss3 = pgTextToString(wsloss3)
	c.WSLoss4 = pgTextToString(wsloss4)
	c.WSLoss5 = pgTextToString(wsloss5)
	
	return &c, nil
}

func (r *customerRepository) Create(ctx context.Context, req *validation.CustomerValidation) (*models.Customer, error) {
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
		RETURNING customer_id, customer, billing_address, billing_city, billing_state, 
		          billing_zipcode, contact, phone, fax, email,
		          color1, color2, color3, color4, color5,
		          loss1, loss2, loss3, loss4, loss5,
		          wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
		          wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
		          deleted, created_at
	`
	
	var c models.Customer
	
	// Use pgtype.Text for nullable fields in the RETURNING clause
	var billingAddress, billingCity, billingState, billingZipcode pgtype.Text
	var contact, phone, fax, email pgtype.Text
	var color1, color2, color3, color4, color5 pgtype.Text
	var loss1, loss2, loss3, loss4, loss5 pgtype.Text
	var wscolor1, wscolor2, wscolor3, wscolor4, wscolor5 pgtype.Text
	var wsloss1, wsloss2, wsloss3, wsloss4, wsloss5 pgtype.Text
	
	err := r.db.QueryRow(ctx, query,
		req.CustomerName, req.Address, req.City, req.State, req.Zip,
		req.Contact, req.Phone, req.Fax, req.Email,
		req.Color1, req.Color2, req.Color3, req.Color4, req.Color5,
		req.Loss1, req.Loss2, req.Loss3, req.Loss4, req.Loss5,
		req.WSColor1, req.WSColor2, req.WSColor3, req.WSColor4, req.WSColor5,
		req.WSLoss1, req.WSLoss2, req.WSLoss3, req.WSLoss4, req.WSLoss5,
	).Scan(
		&c.CustomerID, &c.CustomerName,
		&billingAddress, &billingCity, &billingState, &billingZipcode,
		&contact, &phone, &fax, &email,
		&color1, &color2, &color3, &color4, &color5,
		&loss1, &loss2, &loss3, &loss4, &loss5,
		&wscolor1, &wscolor2, &wscolor3, &wscolor4, &wscolor5,
		&wsloss1, &wsloss2, &wsloss3, &wsloss4, &wsloss5,
		&c.Deleted, &c.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	
	// Convert pgtype.Text to string
	c.BillingAddress = pgTextToString(billingAddress)
	c.BillingCity = pgTextToString(billingCity)
	c.BillingState = pgTextToString(billingState)
	c.BillingZipcode = pgTextToString(billingZipcode)
	c.Contact = pgTextToString(contact)
	c.Phone = pgTextToString(phone)
	c.Fax = pgTextToString(fax)
	c.Email = pgTextToString(email)
	c.Color1 = pgTextToString(color1)
	c.Color2 = pgTextToString(color2)
	c.Color3 = pgTextToString(color3)
	c.Color4 = pgTextToString(color4)
	c.Color5 = pgTextToString(color5)
	c.Loss1 = pgTextToString(loss1)
	c.Loss2 = pgTextToString(loss2)
	c.Loss3 = pgTextToString(loss3)
	c.Loss4 = pgTextToString(loss4)
	c.Loss5 = pgTextToString(loss5)
	c.WSColor1 = pgTextToString(wscolor1)
	c.WSColor2 = pgTextToString(wscolor2)
	c.WSColor3 = pgTextToString(wscolor3)
	c.WSColor4 = pgTextToString(wscolor4)
	c.WSColor5 = pgTextToString(wscolor5)
	c.WSLoss1 = pgTextToString(wsloss1)
	c.WSLoss2 = pgTextToString(wsloss2)
	c.WSLoss3 = pgTextToString(wsloss3)
	c.WSLoss4 = pgTextToString(wsloss4)
	c.WSLoss5 = pgTextToString(wsloss5)
	
	return &c, nil
}

func (r *customerRepository) Update(ctx context.Context, id int, req *validation.CustomerValidation) (*models.Customer, error) {
	query := `
		UPDATE store.customers 
		SET customer = $2, billing_address = $3, billing_city = $4, billing_state = $5,
		    billing_zipcode = $6, contact = $7, phone = $8, fax = $9, email = $10,
		    color1 = $11, color2 = $12, color3 = $13, color4 = $14, color5 = $15,
		    loss1 = $16, loss2 = $17, loss3 = $18, loss4 = $19, loss5 = $20,
		    wscolor1 = $21, wscolor2 = $22, wscolor3 = $23, wscolor4 = $24, wscolor5 = $25,
		    wsloss1 = $26, wsloss2 = $27, wsloss3 = $28, wsloss4 = $29, wsloss5 = $30
		WHERE customer_id = $1 AND deleted = false
		RETURNING customer_id, customer, billing_address, billing_city, billing_state, 
		          billing_zipcode, contact, phone, fax, email,
		          color1, color2, color3, color4, color5,
		          loss1, loss2, loss3, loss4, loss5,
		          wscolor1, wscolor2, wscolor3, wscolor4, wscolor5,
		          wsloss1, wsloss2, wsloss3, wsloss4, wsloss5,
		          deleted, created_at
	`
	
	var c models.Customer
	
	// Use pgtype.Text for nullable fields in the RETURNING clause
	var billingAddress, billingCity, billingState, billingZipcode pgtype.Text
	var contact, phone, fax, email pgtype.Text
	var color1, color2, color3, color4, color5 pgtype.Text
	var loss1, loss2, loss3, loss4, loss5 pgtype.Text
	var wscolor1, wscolor2, wscolor3, wscolor4, wscolor5 pgtype.Text
	var wsloss1, wsloss2, wsloss3, wsloss4, wsloss5 pgtype.Text
	
	err := r.db.QueryRow(ctx, query, id,
		req.CustomerName, req.Address, req.City, req.State, req.Zip,
		req.Contact, req.Phone, req.Fax, req.Email,
		req.Color1, req.Color2, req.Color3, req.Color4, req.Color5,
		req.Loss1, req.Loss2, req.Loss3, req.Loss4, req.Loss5,
		req.WSColor1, req.WSColor2, req.WSColor3, req.WSColor4, req.WSColor5,
		req.WSLoss1, req.WSLoss2, req.WSLoss3, req.WSLoss4, req.WSLoss5,
	).Scan(
		&c.CustomerID, &c.CustomerName,
		&billingAddress, &billingCity, &billingState, &billingZipcode,
		&contact, &phone, &fax, &email,
		&color1, &color2, &color3, &color4, &color5,
		&loss1, &loss2, &loss3, &loss4, &loss5,
		&wscolor1, &wscolor2, &wscolor3, &wscolor4, &wscolor5,
		&wsloss1, &wsloss2, &wsloss3, &wsloss4, &wsloss5,
		&c.Deleted, &c.CreatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}
	
	// Convert pgtype.Text to string
	c.BillingAddress = pgTextToString(billingAddress)
	c.BillingCity = pgTextToString(billingCity)
	c.BillingState = pgTextToString(billingState)
	c.BillingZipcode = pgTextToString(billingZipcode)
	c.Contact = pgTextToString(contact)
	c.Phone = pgTextToString(phone)
	c.Fax = pgTextToString(fax)
	c.Email = pgTextToString(email)
	c.Color1 = pgTextToString(color1)
	c.Color2 = pgTextToString(color2)
	c.Color3 = pgTextToString(color3)
	c.Color4 = pgTextToString(color4)
	c.Color5 = pgTextToString(color5)
	c.Loss1 = pgTextToString(loss1)
	c.Loss2 = pgTextToString(loss2)
	c.Loss3 = pgTextToString(loss3)
	c.Loss4 = pgTextToString(loss4)
	c.Loss5 = pgTextToString(loss5)
	c.WSColor1 = pgTextToString(wscolor1)
	c.WSColor2 = pgTextToString(wscolor2)
	c.WSColor3 = pgTextToString(wscolor3)
	c.WSColor4 = pgTextToString(wscolor4)
	c.WSColor5 = pgTextToString(wscolor5)
	c.WSLoss1 = pgTextToString(wsloss1)
	c.WSLoss2 = pgTextToString(wsloss2)
	c.WSLoss3 = pgTextToString(wsloss3)
	c.WSLoss4 = pgTextToString(wsloss4)
	c.WSLoss5 = pgTextToString(wsloss5)
	
	return &c, nil
}

func (r *customerRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE store.customers 
		SET deleted = true 
		WHERE customer_id = $1 AND deleted = false
	`
	
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
	if err != nil {
		return false, fmt.Errorf("failed to check if customer exists: %w", err)
	}
	
	return exists, nil
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
	if err != nil {
		return false, fmt.Errorf("failed to check customer inventory: %w", err)
	}
	
	return hasInventory, nil
}

// Helper function to convert pgtype.Text to string (empty string if NULL)
func pgTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// Inventory repository - basic implementation
type InventoryRepository interface {
	GetByID(ctx context.Context, id int) (*models.InventoryItem, error)
	Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error)
	Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error)
	Delete(ctx context.Context, id int) error
	GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error)
	GetSummary(ctx context.Context) (*models.InventorySummary, error)
}

type inventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) InventoryRepository {
	return &inventoryRepository{db: db}
}

// Basic implementations for inventory (placeholder - you'll need to implement these based on your inventory table structure)
func (r *inventoryRepository) GetByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	// TODO: Implement based on your inventory table structure
	return nil, fmt.Errorf("inventory GetByID not implemented")
}

func (r *inventoryRepository) Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// TODO: Implement based on your inventory table structure
	return nil, fmt.Errorf("inventory Create not implemented")
}

func (r *inventoryRepository) Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// TODO: Implement based on your inventory table structure
	return nil, fmt.Errorf("inventory Update not implemented")
}

func (r *inventoryRepository) Delete(ctx context.Context, id int) error {
	// TODO: Implement based on your inventory table structure
	return fmt.Errorf("inventory Delete not implemented")
}

func (r *inventoryRepository) GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error) {
	// TODO: Implement based on your inventory table structure
	return nil, 0, fmt.Errorf("inventory GetFiltered not implemented")
}

func (r *inventoryRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error) {
	// TODO: Implement based on your inventory table structure
	return nil, 0, fmt.Errorf("inventory Search not implemented")
}

func (r *inventoryRepository) GetSummary(ctx context.Context) (*models.InventorySummary, error) {
	// TODO: Implement based on your inventory table structure
	return nil, fmt.Errorf("inventory GetSummary not implemented")
}

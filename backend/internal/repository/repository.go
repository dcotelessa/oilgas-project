// backend/internal/repository/repository.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	
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
	err := r.db.QueryRow(ctx, query,
		req.CustomerName, req.Address, req.City, req.State, req.Zip,
		req.Contact, req.Phone, req.Fax, req.Email,
		req.Color1, req.Color2, req.Color3, req.Color4, req.Color5,
		req.Loss1, req.Loss2, req.Loss3, req.Loss4, req.Loss5,
		req.WSColor1, req.WSColor2, req.WSColor3, req.WSColor4, req.WSColor5,
		req.WSLoss1, req.WSLoss2, req.WSLoss3, req.WSLoss4, req.WSLoss5,
	).Scan(
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
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	
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
	err := r.db.QueryRow(ctx, query, id,
		req.CustomerName, req.Address, req.City, req.State, req.Zip,
		req.Contact, req.Phone, req.Fax, req.Email,
		req.Color1, req.Color2, req.Color3, req.Color4, req.Color5,
		req.Loss1, req.Loss2, req.Loss3, req.Loss4, req.Loss5,
		req.WSColor1, req.WSColor2, req.WSColor3, req.WSColor4, req.WSColor5,
		req.WSLoss1, req.WSLoss2, req.WSLoss3, req.WSLoss4, req.WSLoss5,
	).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}
	
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
			WHERE custid = $1 AND deleted = false
		)
	`
	
	var hasInventory bool
	err := r.db.QueryRow(ctx, query, customerID).Scan(&hasInventory)
	if err != nil {
		return false, fmt.Errorf("failed to check customer inventory: %w", err)
	}
	
	return hasInventory, nil
}

// Inventory repository - core methods only
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

func (r *inventoryRepository) GetByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	query := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory 
		WHERE id = $1 AND deleted = false
	`

	var item models.InventoryItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
		&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
		&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
		&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
		&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
		&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
		&item.OrderedBy, &item.Deleted, &item.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory item not found")
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return &item, nil
}

func (r *inventoryRepository) Create(ctx context.Context, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// First, get customer name
	var customerName string
	err := r.db.QueryRow(ctx, "SELECT customer FROM store.customers WHERE customer_id = $1", req.CustomerID).Scan(&customerName)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	query := `
		INSERT INTO store.inventory (
			customer_id, customer, joints, size, weight, grade, connection, 
			color, location, ctd, w_string, deleted, created_at, date_in
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, false, NOW(), NOW()
		) RETURNING id, username, work_order, r_number, customer_id, customer, joints, 
		           rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		           customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		           well_out, lease_out, trucking, trailer, location, notes, pcode, 
		           cn, ordered_by, deleted, created_at
	`

	var item models.InventoryItem
	err = r.db.QueryRow(ctx, query,
		req.CustomerID, customerName, req.Joints, req.Size, req.Weight,
		req.Grade, req.Connection, req.Color, req.Location,
		false, false, // ctd, w_string defaults
	).Scan(
		&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
		&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
		&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
		&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
		&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
		&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
		&item.OrderedBy, &item.Deleted, &item.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	return &item, nil
}

func (r *inventoryRepository) Update(ctx context.Context, id int, req *validation.InventoryValidation) (*models.InventoryItem, error) {
	// Check if item exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get customer name
	var customerName string
	err = r.db.QueryRow(ctx, "SELECT customer FROM store.customers WHERE customer_id = $1", req.CustomerID).Scan(&customerName)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	query := `
		UPDATE store.inventory SET
			customer_id = $2, customer = $3, joints = $4, size = $5, weight = $6,
			grade = $7, connection = $8, color = $9, location = $10
		WHERE id = $1 AND deleted = false
		RETURNING id, username, work_order, r_number, customer_id, customer, joints, 
		          rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		          customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		          well_out, lease_out, trucking, trailer, location, notes, pcode, 
		          cn, ordered_by, deleted, created_at
	`

	var item models.InventoryItem
	err = r.db.QueryRow(ctx, query, id,
		req.CustomerID, customerName, req.Joints, req.Size, req.Weight,
		req.Grade, req.Connection, req.Color, req.Location,
	).Scan(
		&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
		&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
		&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
		&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
		&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
		&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
		&item.OrderedBy, &item.Deleted, &item.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	return &item, nil
}

func (r *inventoryRepository) Delete(ctx context.Context, id int) error {
	// Soft delete
	query := "UPDATE store.inventory SET deleted = true WHERE id = $1 AND deleted = false"
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory item not found")
	}

	return nil
}

func (r *inventoryRepository) GetFiltered(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error) {
	// Build dynamic WHERE clause
	whereClause := "WHERE deleted = false"
	args := []interface{}{}
	argIndex := 1

	if customerID, ok := filters["customer_id"]; ok {
		whereClause += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, customerID)
		argIndex++
	}

	if grade, ok := filters["grade"]; ok {
		whereClause += fmt.Sprintf(" AND UPPER(grade) = UPPER($%d)", argIndex)
		args = append(args, grade)
		argIndex++
	}

	if size, ok := filters["size"]; ok {
		whereClause += fmt.Sprintf(" AND size = $%d", argIndex)
		args = append(args, size)
		argIndex++
	}

	if location, ok := filters["location"]; ok {
		whereClause += fmt.Sprintf(" AND UPPER(location) LIKE UPPER($%d)", argIndex)
		args = append(args, "%"+location.(string)+"%")
		argIndex++
	}

	if rack, ok := filters["rack"]; ok {
		whereClause += fmt.Sprintf(" AND UPPER(rack) LIKE UPPER($%d)", argIndex)
		args = append(args, "%"+rack.(string)+"%")
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM store.inventory " + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory: %w", err)
	}

	// Get filtered records
	query := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory ` + whereClause + `
		ORDER BY date_in DESC, id DESC
		LIMIT $%d OFFSET $%d
	`

	query = fmt.Sprintf(query, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query inventory: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(
			&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
			&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
			&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
			&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
			&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
			&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
			&item.OrderedBy, &item.Deleted, &item.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return items, total, nil
}

func (r *inventoryRepository) Search(ctx context.Context, searchQuery string, limit, offset int) ([]models.InventoryItem, int, error) {
	// Simple search across multiple fields
	whereClause := `
		WHERE deleted = false AND (
			UPPER(customer) LIKE UPPER($1) OR
			UPPER(grade) LIKE UPPER($1) OR  
			UPPER(size) LIKE UPPER($1) OR
			UPPER(location) LIKE UPPER($1) OR
			UPPER(rack) LIKE UPPER($1) OR
			UPPER(notes) LIKE UPPER($1) OR
			UPPER(customer_po) LIKE UPPER($1) OR
			work_order LIKE UPPER($1)
		)
	`

	searchTerm := "%" + strings.TrimSpace(searchQuery) + "%"

	// Count total
	countQuery := "SELECT COUNT(*) FROM store.inventory " + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get results
	query := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory ` + whereClause + `
		ORDER BY 
			CASE WHEN UPPER(customer) LIKE UPPER($1) THEN 1
			     WHEN UPPER(grade) LIKE UPPER($1) THEN 2
			     WHEN UPPER(customer_po) LIKE UPPER($1) THEN 3
			     ELSE 4 END,
			date_in DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, searchTerm, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search inventory: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(
			&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
			&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
			&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
			&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
			&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
			&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
			&item.OrderedBy, &item.Deleted, &item.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *inventoryRepository) GetSummary(ctx context.Context) (*models.InventorySummary, error) {
	summary := &models.InventorySummary{
		ItemsByGrade:    make(map[string]int),
		ItemsByCustomer: make(map[string]int),
		ItemsByLocation: make(map[string]int),
		LastUpdated:     time.Now(),
	}

	// Total items and joints
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(joints), 0) 
		FROM store.inventory 
		WHERE deleted = false
	`).Scan(&summary.TotalItems, &summary.TotalJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Items by grade
	rows, err := r.db.Query(ctx, `
		SELECT grade, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND grade IS NOT NULL
		GROUP BY grade 
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by grade: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var grade string
		var count int
		if err := rows.Scan(&grade, &count); err == nil {
			summary.ItemsByGrade[grade] = count
		}
	}

	// Items by customer
	rows, err = r.db.Query(ctx, `
		SELECT customer, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND customer IS NOT NULL
		GROUP BY customer 
		ORDER BY COUNT(*) DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by customer: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var customer string
		var count int
		if err := rows.Scan(&customer, &count); err == nil {
			summary.ItemsByCustomer[customer] = count
		}
	}

	// Items by location
	rows, err = r.db.Query(ctx, `
		SELECT location, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND location IS NOT NULL
		GROUP BY location 
		ORDER BY COUNT(*) DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var location string
		var count int
		if err := rows.Scan(&location, &count); err == nil {
			summary.ItemsByLocation[location] = count
		}
	}

	// Recent activity (last 10 items)
	rows, err = r.db.Query(ctx, `
		SELECT id, customer, joints, size, grade, location, date_in
		FROM store.inventory 
		WHERE deleted = false 
		ORDER BY date_in DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.ID, &item.Customer, &item.Joints, &item.Size, &item.Grade, &item.Location, &item.DateIn)
		if err == nil {
			summary.RecentActivity = append(summary.RecentActivity, item)
		}
	}

	return summary, nil
}

// backend/internal/repository/postgres/inventory.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"oilgas-backend/internal/models"
)

type InventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepo(db *sql.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) GetAll(ctx context.Context, filters models.InventoryFilters) ([]models.InventoryItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, tenant_id, deleted, created_at
		FROM store.inventory 
		WHERE deleted = false`
	
	var args []interface{}
	argCount := 0
	
	// Build WHERE conditions dynamically
	if filters.CustomerID != nil {
		argCount++
		query += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, *filters.CustomerID)
	}
	
	if filters.Grade != nil && *filters.Grade != "" {  // Fixed: check pointer and dereference
		argCount++
		query += fmt.Sprintf(" AND grade = $%d", argCount)
		args = append(args, *filters.Grade)
	}
	
	if filters.Size != nil && *filters.Size != "" {    // Fixed: check pointer and dereference
		argCount++
		query += fmt.Sprintf(" AND size = $%d", argCount)
		args = append(args, *filters.Size)
	}
	
	if filters.Location != nil && *filters.Location != "" {  // Fixed: check pointer and dereference
		argCount++
		query += fmt.Sprintf(" AND location = $%d", argCount)
		args = append(args, *filters.Location)
	}
	
	if filters.WorkOrder != "" {
		argCount++
		query += fmt.Sprintf(" AND work_order = $%d", argCount)
		args = append(args, filters.WorkOrder)
	}
	
	if filters.Available != nil {
		if *filters.Available {
			query += " AND date_out IS NULL"  // Available items have no date_out
		} else {
			query += " AND date_out IS NOT NULL"  // Unavailable items have date_out
		}
	}
	
	if filters.DateFrom != nil {
		argCount++
		query += fmt.Sprintf(" AND date_in >= $%d", argCount)
		args = append(args, *filters.DateFrom)
	}
	
	if filters.DateTo != nil {
		argCount++
		query += fmt.Sprintf(" AND date_in <= $%d", argCount)
		args = append(args, *filters.DateTo)
	}
	
	if filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (customer ILIKE $%d OR work_order ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm)
	}
	
	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}
	
	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
			&item.LeaseOut, &item.Location, &item.Notes, &item.TenantID, &item.Deleted, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *InventoryRepo) GetByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, tenant_id, deleted, created_at
		FROM store.inventory 
		WHERE id = $1 AND deleted = false`
	
	var item models.InventoryItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
		&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
		&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
		&item.LeaseOut, &item.Location, &item.Notes, &item.TenantID, &item.Deleted, &item.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory item %d: %w", id, err)
	}
	
	return &item, nil
}

func (r *InventoryRepo) GetByWorkOrder(ctx context.Context, workOrder string) ([]models.InventoryItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, tenant_id, deleted, created_at
		FROM store.inventory 
		WHERE work_order = $1 AND deleted = false
		ORDER BY created_at`
	
	rows, err := r.db.QueryContext(ctx, query, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory by work order: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
			&item.LeaseOut, &item.Location, &item.Notes, &item.TenantID, &item.Deleted, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *InventoryRepo) GetAvailable(ctx context.Context) ([]models.InventoryItem, error) {
	filters := models.InventoryFilters{
		Available: &[]bool{true}[0], // Available items have no date_out
	}
	return r.GetAll(ctx, filters)
}

func (r *InventoryRepo) Search(ctx context.Context, query string) ([]models.InventoryItem, error) {
	filters := models.InventoryFilters{
		Search: query,
		Limit:  100, // Default search limit
	}
	return r.GetAll(ctx, filters)
}

func (r *InventoryRepo) Create(ctx context.Context, item *models.InventoryItem) error {
	// Implementation for Create
	return fmt.Errorf("Create method not implemented yet")
}

func (r *InventoryRepo) Update(ctx context.Context, item *models.InventoryItem) error {
	// Implementation for Update  
	return fmt.Errorf("Update method not implemented yet")
}

func (r *InventoryRepo) Delete(ctx context.Context, id int) error {
	// Implementation for Delete
	return fmt.Errorf("Delete method not implemented yet")
}

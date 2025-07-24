// internal/repository/postgres/inventory.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"oilgas-backend/internal/repository"
)

type InventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) repository.InventoryRepository {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) GetAll(ctx context.Context, filters repository.InventoryFilters) ([]repository.InventoryItem, error) {
	baseQuery := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, deleted, created_at
		FROM store.inventory 
		WHERE deleted = false`
	
	var conditions []string
	var args []interface{}
	argCount := 0

	if filters.CustomerID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argCount))
		args = append(args, *filters.CustomerID)
	}
	
	if filters.Grade != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("grade = $%d", argCount))
		args = append(args, *filters.Grade)
	}
	
	if filters.Size != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("size = $%d", argCount))
		args = append(args, *filters.Size)
	}
	
	if filters.Location != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("location = $%d", argCount))
		args = append(args, *filters.Location)
	}
	
	if filters.Available != nil && *filters.Available {
		conditions = append(conditions, "date_out IS NULL")
	}
	
	if filters.DateFrom != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("date_in >= $%d", argCount))
		args = append(args, *filters.DateFrom)
	}
	
	if filters.DateTo != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("date_in <= $%d", argCount))
		args = append(args, *filters.DateTo)
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
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

	var items []repository.InventoryItem
	for rows.Next() {
		var item repository.InventoryItem
		err := rows.Scan(&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
			&item.LeaseOut, &item.Location, &item.Notes, &item.Deleted, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *InventoryRepo) GetByID(ctx context.Context, id int) (*repository.InventoryItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, deleted, created_at
		FROM store.inventory 
		WHERE id = $1 AND deleted = false`
	
	var item repository.InventoryItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
		&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
		&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
		&item.LeaseOut, &item.Location, &item.Notes, &item.Deleted, &item.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory item %d: %w", id, err)
	}
	
	return &item, nil
}

func (r *InventoryRepo) GetByWorkOrder(ctx context.Context, workOrder string) ([]repository.InventoryItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, deleted, created_at
		FROM store.inventory 
		WHERE work_order = $1 AND deleted = false
		ORDER BY created_at`
	
	rows, err := r.db.QueryContext(ctx, query, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory by work order: %w", err)
	}
	defer rows.Close()

	var items []repository.InventoryItem
	for rows.Next() {
		var item repository.InventoryItem
		err := rows.Scan(&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
			&item.LeaseOut, &item.Location, &item.Notes, &item.Deleted, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *InventoryRepo) GetAvailable(ctx context.Context) ([]repository.InventoryItem, error) {
	filters := repository.InventoryFilters{
		Available: &[]bool{true}[0], // Available items have no date_out
	}
	return r.GetAll(ctx, filters)
}

func (r *InventoryRepo) Search(ctx context.Context, query string) ([]repository.InventoryItem, error) {
	searchQuery := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade,
		       connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
		       location, notes, deleted, created_at
		FROM store.inventory 
		WHERE deleted = false 
		  AND (work_order ILIKE $1 OR customer ILIKE $1 OR size ILIKE $1 OR 
		       grade ILIKE $1 OR location ILIKE $1 OR notes ILIKE $1)
		ORDER BY created_at DESC`
	
	searchTerm := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search inventory: %w", err)
	}
	defer rows.Close()

	var items []repository.InventoryItem
	for rows.Next() {
		var item repository.InventoryItem
		err := rows.Scan(&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn, &item.WellOut,
			&item.LeaseOut, &item.Location, &item.Notes, &item.Deleted, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *InventoryRepo) Create(ctx context.Context, item *repository.InventoryItem) error {
	query := `
		INSERT INTO store.inventory (work_order, customer_id, customer, joints, size, weight,
			grade, connection, date_in, date_out, well_in, lease_in, well_out, lease_out,
			location, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at`
	
	err := r.db.QueryRowContext(ctx, query, item.WorkOrder, item.CustomerID, item.Customer,
		item.Joints, item.Size, item.Weight, item.Grade, item.Connection,
		item.DateIn, item.DateOut, item.WellIn, item.LeaseIn, item.WellOut,
		item.LeaseOut, item.Location, item.Notes).
		Scan(&item.ID, &item.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}
	
	return nil
}

func (r *InventoryRepo) Update(ctx context.Context, item *repository.InventoryItem) error {
	query := `
		UPDATE store.inventory 
		SET work_order = $2, customer_id = $3, customer = $4, joints = $5, size = $6,
		    weight = $7, grade = $8, connection = $9, date_in = $10, date_out = $11,
		    well_in = $12, lease_in = $13, well_out = $14, lease_out = $15,
		    location = $16, notes = $17
		WHERE id = $1 AND deleted = false`
	
	result, err := r.db.ExecContext(ctx, query, item.ID, item.WorkOrder, item.CustomerID,
		item.Customer, item.Joints, item.Size, item.Weight, item.Grade, item.Connection,
		item.DateIn, item.DateOut, item.WellIn, item.LeaseIn, item.WellOut,
		item.LeaseOut, item.Location, item.Notes)
	
	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("inventory item %d not found or already deleted", item.ID)
	}
	
	return nil
}

func (r *InventoryRepo) Delete(ctx context.Context, id int) error {
	query := `UPDATE store.inventory SET deleted = true WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("inventory item %d not found", id)
	}
	
	return nil
}

// backend/internal/repository/received_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

type receivedRepository struct {
	db *pgxpool.Pool
}

func NewReceivedRepository(db *pgxpool.Pool) ReceivedRepository {
	return &receivedRepository{db: db}
}

func (r *receivedRepository) GetByID(ctx context.Context, id int) (*models.ReceivedItem, error) {
	query := `
		SELECT id, work_order, customer_id, customer, joints, rack, size_id, size, weight, 
		       grade, connection, ctd, w_string, well, lease, ordered_by, notes, customer_po, 
		       date_received, background, norm, services, bill_to_id, entered_by, when_entered, 
		       trucking, trailer, in_production, inspected_date, threading_date, 
		       straighten_required, excess_material, complete, inspected_by, updated_by, 
		       when_updated, deleted, created_at
		FROM store.received 
		WHERE id = $1 AND deleted = false
	`

	var item models.ReceivedItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer, &item.Joints,
		&item.Rack, &item.SizeID, &item.Size, &item.Weight, &item.Grade, &item.Connection,
		&item.CTD, &item.WString, &item.Well, &item.Lease, &item.OrderedBy, &item.Notes,
		&item.CustomerPO, &item.DateReceived, &item.Background, &item.Norm, &item.Services,
		&item.BillToID, &item.EnteredBy, &item.WhenEntered, &item.Trucking, &item.Trailer,
		&item.InProduction, &item.InspectedDate, &item.ThreadingDate, &item.StraightenRequired,
		&item.ExcessMaterial, &item.Complete, &item.InspectedBy, &item.UpdatedBy,
		&item.WhenUpdated, &item.Deleted, &item.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("received item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get received item: %w", err)
	}

	return &item, nil
}

func (r *receivedRepository) Create(ctx context.Context, item *models.ReceivedItem) error {
	query := `
		INSERT INTO store.received (
			work_order, customer_id, customer, joints, rack, size_id, size, weight, 
			grade, connection, ctd, w_string, well, lease, ordered_by, notes, customer_po, 
			date_received, background, norm, services, bill_to_id, entered_by, when_entered, 
			trucking, trailer, in_production, inspected_date, threading_date, 
			straighten_required, excess_material, complete, inspected_by, updated_by, 
			when_updated, deleted, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, false, NOW()
		) RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		item.WorkOrder, item.CustomerID, item.Customer, item.Joints, item.Rack,
		item.SizeID, item.Size, item.Weight, item.Grade, item.Connection, item.CTD,
		item.WString, item.Well, item.Lease, item.OrderedBy, item.Notes, item.CustomerPO,
		item.DateReceived, item.Background, item.Norm, item.Services, item.BillToID,
		item.EnteredBy, item.WhenEntered, item.Trucking, item.Trailer, item.InProduction,
		item.InspectedDate, item.ThreadingDate, item.StraightenRequired, item.ExcessMaterial,
		item.Complete, item.InspectedBy, item.UpdatedBy, item.WhenUpdated,
	).Scan(&item.ID, &item.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create received item: %w", err)
	}

	return nil
}

func (r *receivedRepository) Update(ctx context.Context, item *models.ReceivedItem) error {
	query := `
		UPDATE store.received SET
			work_order = $2, customer_id = $3, customer = $4, joints = $5, rack = $6,
			size_id = $7, size = $8, weight = $9, grade = $10, connection = $11,
			ctd = $12, w_string = $13, well = $14, lease = $15, ordered_by = $16,
			notes = $17, customer_po = $18, date_received = $19, background = $20,
			norm = $21, services = $22, bill_to_id = $23, entered_by = $24,
			when_entered = $25, trucking = $26, trailer = $27, in_production = $28,
			inspected_date = $29, threading_date = $30, straighten_required = $31,
			excess_material = $32, complete = $33, inspected_by = $34,
			updated_by = $35, when_updated = $36
		WHERE id = $1 AND deleted = false
	`

	result, err := r.db.Exec(ctx, query, item.ID,
		item.WorkOrder, item.CustomerID, item.Customer, item.Joints, item.Rack,
		item.SizeID, item.Size, item.Weight, item.Grade, item.Connection, item.CTD,
		item.WString, item.Well, item.Lease, item.OrderedBy, item.Notes, item.CustomerPO,
		item.DateReceived, item.Background, item.Norm, item.Services, item.BillToID,
		item.EnteredBy, item.WhenEntered, item.Trucking, item.Trailer, item.InProduction,
		item.InspectedDate, item.ThreadingDate, item.StraightenRequired, item.ExcessMaterial,
		item.Complete, item.InspectedBy, item.UpdatedBy, item.WhenUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to update received item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("received item not found or already deleted")
	}

	return nil
}

func (r *receivedRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE store.received SET deleted = true WHERE id = $1 AND deleted = false`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete received item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("received item not found")
	}

	return nil
}

func (r *receivedRepository) GetFiltered(ctx context.Context, filters ReceivedFilters) ([]models.ReceivedItem, *models.Pagination, error) {
	filters.NormalizePagination()
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition
	conditions = append(conditions, "deleted = false")

	// Apply filters
	if filters.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, *filters.CustomerID)
		argIndex++
	}

	if filters.Status != "" {
		switch strings.ToLower(filters.Status) {
		case "pending":
			conditions = append(conditions, "in_production IS NULL")
		case "in_production":
			conditions = append(conditions, "in_production IS NOT NULL AND inspected_date IS NULL")
		case "completed":
			conditions = append(conditions, "complete = true")
		}
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("date_received >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("date_received <= $%d", argIndex))
		args = append(args, filters.DateTo)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM store.received %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count received items: %w", err)
	}

	// Calculate pagination
	totalPages := (total + filters.PerPage - 1) / filters.PerPage
	offset := (filters.Page - 1) * filters.PerPage

	pagination := &models.Pagination{
		Page:       filters.Page,
		PerPage:    filters.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}

	// Get filtered records
	query := fmt.Sprintf(`
		SELECT id, work_order, customer_id, customer, joints, rack, size_id, size, weight, 
		       grade, connection, ctd, w_string, well, lease, ordered_by, notes, customer_po, 
		       date_received, background, norm, services, bill_to_id, entered_by, when_entered, 
		       trucking, trailer, in_production, inspected_date, threading_date, 
		       straighten_required, excess_material, complete, inspected_by, updated_by, 
		       when_updated, deleted, created_at
		FROM store.received %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, filters.OrderBy, filters.OrderDir, argIndex, argIndex+1)

	args = append(args, filters.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query received items: %w", err)
	}
	defer rows.Close()

	var items []models.ReceivedItem
	for rows.Next() {
		var item models.ReceivedItem
		err := rows.Scan(
			&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer, &item.Joints,
			&item.Rack, &item.SizeID, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.CTD, &item.WString, &item.Well, &item.Lease, &item.OrderedBy, &item.Notes,
			&item.CustomerPO, &item.DateReceived, &item.Background, &item.Norm, &item.Services,
			&item.BillToID, &item.EnteredBy, &item.WhenEntered, &item.Trucking, &item.Trailer,
			&item.InProduction, &item.InspectedDate, &item.ThreadingDate, &item.StraightenRequired,
			&item.ExcessMaterial, &item.Complete, &item.InspectedBy, &item.UpdatedBy,
			&item.WhenUpdated, &item.Deleted, &item.CreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan received item: %w", err)
		}
		items = append(items, item)
	}

	return items, pagination, rows.Err()
}

func (r *receivedRepository) UpdateStatus(ctx context.Context, id int, status string, notes string) error {
	var query string
	var args []interface{}
	
	switch strings.ToLower(status) {
	case "in_production":
		query = `UPDATE store.received SET in_production = NOW(), notes = $2 WHERE id = $1`
		args = []interface{}{id, notes}
	case "inspected":
		query = `UPDATE store.received SET inspected_date = NOW(), notes = $2 WHERE id = $1`
		args = []interface{}{id, notes}
	case "completed":
		query = `UPDATE store.received SET complete = true, notes = $2 WHERE id = $1`
		args = []interface{}{id, notes}
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("received item not found")
	}

	return nil
}

func (r *receivedRepository) CanDelete(ctx context.Context, id int) (bool, string, error) {
	// Check if item has been moved to production
	var inProduction bool
	var hasInventory bool
	
	err := r.db.QueryRow(ctx, 
		`SELECT in_production IS NOT NULL FROM store.received WHERE id = $1`, 
		id).Scan(&inProduction)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, "received item not found", nil
		}
		return false, "", fmt.Errorf("failed to check production status: %w", err)
	}
	
	if inProduction {
		return false, "cannot delete items that are in production", nil
	}
	
	// Check if there's related inventory
	err = r.db.QueryRow(ctx, 
		`SELECT EXISTS(SELECT 1 FROM store.inventory i 
		 JOIN store.received r ON i.work_order = r.work_order 
		 WHERE r.id = $1 AND i.deleted = false)`, 
		id).Scan(&hasInventory)
	if err != nil {
		return false, "", fmt.Errorf("failed to check inventory: %w", err)
	}
	
	if hasInventory {
		return false, "cannot delete items with associated inventory", nil
	}
	
	return true, "", nil
}

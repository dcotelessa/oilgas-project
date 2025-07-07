package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

type InspectedRepository interface {
	Create(ctx context.Context, item *models.InspectionItem) error
	GetByID(ctx context.Context, id int) (*models.InspectionItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) (*models.InspectionItem, error)
	GetFiltered(ctx context.Context, filters InspectedFilters) ([]*models.InspectionItem, int, error)
	Update(ctx context.Context, item *models.InspectionItem) error
	Delete(ctx context.Context, id int) error
}

type InspectedFilters struct {
	CustomerID *int
	WorkOrder  *string
	Inspector  *string
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int
}

type inspectedRepository struct {
	db *pgxpool.Pool
}

func NewInspectedRepository(db *pgxpool.Pool) InspectedRepository {
	return &inspectedRepository{db: db}
}

func (r *inspectedRepository) Create(ctx context.Context, item *models.InspectionItem) error {
	query := `
		INSERT INTO store.inspected (
			work_order, customer_id, customer, joints, size, weight, grade, connection,
			passed_joints, failed_joints, inspection_date, inspector, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		item.WorkOrder, item.CustomerID, item.Customer, item.Joints,
		item.Size, item.Weight, item.Grade, item.Connection,
		item.PassedJoints, item.FailedJoints, item.InspectionDate,
		item.Inspector, item.Notes,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create inspection item: %w", err)
	}
	
	return nil
}

func (r *inspectedRepository) GetFiltered(ctx context.Context, filters InspectedFilters) ([]*models.InspectionItem, int, error) {
	// Basic implementation for testing
	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight, grade, connection,
		       passed_joints, failed_joints, inspection_date, inspector, notes, created_at, updated_at
		FROM store.inspected 
		WHERE 1=1
	`
	
	args := []interface{}{}
	argIndex := 1
	
	if filters.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, *filters.CustomerID)
		argIndex++
	}
	
	if filters.WorkOrder != nil {
		query += fmt.Sprintf(" AND work_order = $%d", argIndex)
		args = append(args, *filters.WorkOrder)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if filters.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filters.PerPage, (filters.Page-1)*filters.PerPage)
	}
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query inspected items: %w", err)
	}
	defer rows.Close()
	
	var items []*models.InspectionItem
	for rows.Next() {
		item := &models.InspectionItem{}
		err := rows.Scan(
			&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Size, &item.Weight, &item.Grade, &item.Connection,
			&item.PassedJoints, &item.FailedJoints, &item.InspectionDate,
			&item.Inspector, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan inspection item: %w", err)
		}
		items = append(items, item)
	}
	
	// Count total (simplified)
	total := len(items)
	
	return items, total, nil
}

// Other required methods...
func (r *inspectedRepository) GetByID(ctx context.Context, id int) (*models.InspectionItem, error) {
	// Implementation...
	return nil, fmt.Errorf("not implemented")
}

func (r *inspectedRepository) GetByWorkOrder(ctx context.Context, workOrder string) (*models.InspectionItem, error) {
	// Implementation...
	return nil, fmt.Errorf("not implemented")
}

func (r *inspectedRepository) Update(ctx context.Context, item *models.InspectionItem) error {
	// Implementation...
	return fmt.Errorf("not implemented")
}

func (r *inspectedRepository) Delete(ctx context.Context, id int) error {
	// Implementation...
	return fmt.Errorf("not implemented")
}

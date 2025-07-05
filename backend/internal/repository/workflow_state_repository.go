// backend/internal/repository/workflow_state_repository.go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

// WorkflowStateRepository handles workflow state transitions
type WorkflowStateRepository interface {
	AdvanceToProduction(ctx context.Context, workOrder string, username string) error
	AdvanceToInspection(ctx context.Context, workOrder string, inspectedBy string) error
	AdvanceToInventory(ctx context.Context, workOrder string) error
	MarkAsComplete(ctx context.Context, workOrder string) error
	
	GetWorkflowStatus(ctx context.Context, workOrder string) (*models.WorkflowStatus, error)
	GetJobsByState(ctx context.Context, state string, limit, offset int) ([]models.ReceivedItem, int, error)
}

type workflowStateRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowStateRepository(db *pgxpool.Pool) WorkflowStateRepository {
	return &workflowStateRepository{db: db}
}

func (r *workflowStateRepository) AdvanceToProduction(ctx context.Context, workOrder string, username string) error {
	// Check current state first
	status, err := r.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	if status.CurrentState != "received" {
		return fmt.Errorf("cannot advance to production from state: %s", status.CurrentState)
	}

	query := `
		UPDATE store.received 
		SET in_production = NOW(), updated_by = $2, when_updated = NOW()
		WHERE work_order = $1 AND deleted = false AND in_production IS NULL
	`

	result, err := r.db.Exec(ctx, query, workOrder, username)
	if err != nil {
		return fmt.Errorf("failed to advance to production: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("work order not found or already in production: %s", workOrder)
	}

	return nil
}

func (r *workflowStateRepository) AdvanceToInspection(ctx context.Context, workOrder string, inspectedBy string) error {
	// Check current state
	status, err := r.GetWorkflowStatus(ctx, workOrder)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	if status.CurrentState != "in_production" {
		return fmt.Errorf("cannot advance to inspection from state: %s", status.CurrentState)
	}

	query := `
		UPDATE store.received 
		SET inspected_date = NOW(), inspected_by = $2, updated_by = $2, when_updated = NOW()
		WHERE work_order = $1 AND deleted = false AND in_production IS NOT NULL AND inspected_date IS NULL
	`

	result, err := r.db.Exec(ctx, query, workOrder, inspectedBy)
	if err != nil {
		return fmt.Errorf("failed to advance to inspection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("work order not ready for inspection: %s", workOrder)
	}

	return nil
}

func (r *workflowStateRepository) AdvanceToInventory(ctx context.Context, workOrder string) error {
	// This creates inventory records from inspected received items
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if already in inventory
	var existsInInventory bool
	err = tx.QueryRow(ctx, 
		`SELECT EXISTS(SELECT 1 FROM store.inventory WHERE work_order = $1 AND deleted = false)`,
		workOrder).Scan(&existsInInventory)
	if err != nil {
		return fmt.Errorf("failed to check inventory status: %w", err)
	}

	if existsInInventory {
		return fmt.Errorf("work order already in inventory: %s", workOrder)
	}

	// Get received item details
	var receivedItem models.ReceivedItem
	query := `
		SELECT id, customer_id, customer, joints, size, weight, grade, connection, 
		       ctd, w_string, well, lease, customer_po, trucking, trailer
		FROM store.received 
		WHERE work_order = $1 AND deleted = false AND inspected_date IS NOT NULL
	`

	err = tx.QueryRow(ctx, query, workOrder).Scan(
		&receivedItem.ID, &receivedItem.CustomerID, &receivedItem.Customer,
		&receivedItem.Joints, &receivedItem.Size, &receivedItem.Weight,
		&receivedItem.Grade, &receivedItem.Connection, &receivedItem.CTD,
		&receivedItem.WString, &receivedItem.Well, &receivedItem.Lease,
		&receivedItem.CustomerPO, &receivedItem.Trucking, &receivedItem.Trailer,
	)
	if err != nil {
		return fmt.Errorf("work order not ready for inventory: %w", err)
	}

	// Create inventory record
	insertQuery := `
		INSERT INTO store.inventory (
			work_order, customer_id, customer, joints, size, weight, grade, connection,
			ctd, w_string, well_in, lease_in, customer_po, trucking, trailer,
			date_in, deleted, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			NOW(), false, NOW()
		)
	`

	_, err = tx.Exec(ctx, insertQuery,
		workOrder, receivedItem.CustomerID, receivedItem.Customer,
		receivedItem.Joints, receivedItem.Size, receivedItem.Weight,
		receivedItem.Grade, receivedItem.Connection, receivedItem.CTD,
		receivedItem.WString, receivedItem.Well, receivedItem.Lease,
		receivedItem.CustomerPO, receivedItem.Trucking, receivedItem.Trailer,
	)
	if err != nil {
		return fmt.Errorf("failed to create inventory record: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *workflowStateRepository) MarkAsComplete(ctx context.Context, workOrder string) error {
	query := `
		UPDATE store.received 
		SET complete = true, updated_by = 'system', when_updated = NOW()
		WHERE work_order = $1 AND deleted = false
	`

	result, err := r.db.Exec(ctx, query, workOrder)
	if err != nil {
		return fmt.Errorf("failed to mark as complete: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("work order not found: %s", workOrder)
	}

	return nil
}

func (r *workflowStateRepository) GetWorkflowStatus(ctx context.Context, workOrder string) (*models.WorkflowStatus, error) {
	query := `
		SELECT 
			r.work_order,
			r.customer,
			r.joints,
			r.grade,
			r.size,
			r.date_received,
			r.in_production,
			r.inspected_date,
			r.complete,
			CASE 
				WHEN i.date_in IS NOT NULL THEN 'inventory'
				WHEN r.complete = true THEN 'completed'
				WHEN r.inspected_date IS NOT NULL THEN 'inspected'
				WHEN r.in_production IS NOT NULL THEN 'in_production'
				ELSE 'received'
			END as current_state,
			i.date_in as inventory_date
		FROM store.received r
		LEFT JOIN store.inventory i ON r.work_order = i.work_order AND i.deleted = false
		WHERE r.work_order = $1 AND r.deleted = false
	`

	var status models.WorkflowStatus
	err := r.db.QueryRow(ctx, query, workOrder).Scan(
		&status.WorkOrder, &status.Customer, &status.Joints,
		&status.Grade, &status.Size, &status.DateReceived,
		&status.InProduction, &status.InspectedDate, &status.Complete,
		&status.CurrentState, &status.InventoryDate,
	)

	if err != nil {
		return nil, fmt.Errorf("work order not found: %w", err)
	}

	// Calculate days in current state
	var stateStartTime time.Time
	switch status.CurrentState {
	case "received":
		if status.DateReceived != nil {
			stateStartTime = *status.DateReceived
		}
	case "in_production":
		if status.InProduction != nil {
			stateStartTime = *status.InProduction
		}
	case "inspected":
		if status.InspectedDate != nil {
			stateStartTime = *status.InspectedDate
		}
	case "inventory":
		if status.InventoryDate != nil {
			stateStartTime = *status.InventoryDate
		}
	}

	if !stateStartTime.IsZero() {
		status.DaysInState = int(time.Since(stateStartTime).Hours() / 24)
	}

	return &status, nil
}

func (r *workflowStateRepository) GetJobsByState(ctx context.Context, state string, limit, offset int) ([]models.ReceivedItem, int, error) {
	var whereCondition string
	
	switch state {
	case "received":
		whereCondition = "r.in_production IS NULL"
	case "in_production":
		whereCondition = "r.in_production IS NOT NULL AND r.inspected_date IS NULL"
	case "inspected":
		whereCondition = "r.inspected_date IS NOT NULL AND r.complete = false"
	case "completed":
		whereCondition = "r.complete = true"
	default:
		return nil, 0, fmt.Errorf("invalid state: %s", state)
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM store.received r 
		WHERE r.deleted = false AND %s
	`, whereCondition)

	var total int
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Get jobs
	query := fmt.Sprintf(`
		SELECT 
			id, work_order, customer_id, customer, joints, rack, size_id, size, weight, 
			grade, connection, ctd, w_string, well, lease, ordered_by, notes, customer_po, 
			date_received, background, norm, services, bill_to_id, entered_by, when_entered, 
			trucking, trailer, in_production, inspected_date, threading_date, 
			straighten_required, excess_material, complete, inspected_by, updated_by, 
			when_updated, deleted, created_at
		FROM store.received r
		WHERE r.deleted = false AND %s
		ORDER BY r.date_received DESC
		LIMIT $1 OFFSET $2
	`, whereCondition)

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.ReceivedItem
	for rows.Next() {
		var job models.ReceivedItem
		err := rows.Scan(
			&job.ID, &job.WorkOrder, &job.CustomerID, &job.Customer, &job.Joints,
			&job.Rack, &job.SizeID, &job.Size, &job.Weight, &job.Grade, &job.Connection,
			&job.CTD, &job.WString, &job.Well, &job.Lease, &job.OrderedBy, &job.Notes,
			&job.CustomerPO, &job.DateReceived, &job.Background, &job.Norm, &job.Services,
			&job.BillToID, &job.EnteredBy, &job.WhenEntered, &job.Trucking, &job.Trailer,
			&job.InProduction, &job.InspectedDate, &job.ThreadingDate,
			&job.StraightenRequired, &job.ExcessMaterial, &job.Complete, &job.InspectedBy, 
			&job.UpdatedBy, &job.WhenUpdated, &job.Deleted, &job.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, total, rows.Err()
}

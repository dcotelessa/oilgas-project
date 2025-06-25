package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type WorkflowRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowRepository(db *pgxpool.Pool) repository.WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Dashboard operations with optimized queries
func (r *WorkflowRepository) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Get job summaries with single optimized query
	summaries, err := r.GetJobSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job summaries: %w", err)
	}
	stats.JobSummaries = summaries

	// Get recent activity
	recentActivity, err := r.GetRecentActivity(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	stats.RecentActivity = recentActivity

	// Get inventory by grade with conditional aggregation
	inventoryQuery := `
		SELECT 
			grade,
			SUM(joints) FILTER (WHERE joints > 0) as total_joints
		FROM store.inventory 
		WHERE deleted = false 
		GROUP BY grade 
		HAVING SUM(joints) FILTER (WHERE joints > 0) > 0
		ORDER BY total_joints DESC
	`
	
	rows, err := r.db.Query(ctx, inventoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory by grade: %w", err)
	}
	defer rows.Close()

	inventoryByGrade := make(map[string]int)
	for rows.Next() {
		var grade string
		var joints int
		if err := rows.Scan(&grade, &joints); err != nil {
			return nil, fmt.Errorf("failed to scan inventory row: %w", err)
		}
		inventoryByGrade[grade] = joints
	}
	stats.InventoryByGrade = inventoryByGrade

	// Get pending repairs count
	repairQuery := `
		SELECT COUNT(*) 
		FROM store.inspected 
		WHERE deleted = false 
		AND (pin > 0 OR cplg > 0 OR pc > 0)
		AND complete = false
	`
	err = r.db.QueryRow(ctx, repairQuery).Scan(&stats.PendingRepairs)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending repairs count: %w", err)
	}

	// Get total customers count
	customerQuery := `SELECT COUNT(*) FROM store.customers WHERE deleted = false`
	err = r.db.QueryRow(ctx, customerQuery).Scan(&stats.TotalCustomers)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers count: %w", err)
	}

	// Get top customers with active jobs
	topCustomersQuery := `
		SELECT 
			c.custid,
			c.customer,
			COUNT(DISTINCT CASE WHEN r.complete = false AND r.deleted = false THEN r.id END) as active_jobs,
			COALESCE(SUM(CASE WHEN i.joints > 0 AND i.deleted = false THEN i.joints END), 0) as total_joints
		FROM store.customers c
		LEFT JOIN store.received r ON c.custid = r.custid
		LEFT JOIN store.inventory i ON c.custid = i.custid
		WHERE c.deleted = false
		GROUP BY c.custid, c.customer
		HAVING COUNT(DISTINCT CASE WHEN r.complete = false AND r.deleted = false THEN r.id END) > 0
			OR COALESCE(SUM(CASE WHEN i.joints > 0 AND i.deleted = false THEN i.joints END), 0) > 0
		ORDER BY active_jobs DESC, total_joints DESC
		LIMIT 5
	`

	rows, err = r.db.Query(ctx, topCustomersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get top customers: %w", err)
	}
	defer rows.Close()

	var topCustomers []models.CustomerStats
	for rows.Next() {
		var customer models.CustomerStats
		if err := rows.Scan(&customer.CustomerID, &customer.CustomerName, &customer.ActiveJobs, &customer.TotalJoints); err != nil {
			return nil, fmt.Errorf("failed to scan customer stats: %w", err)
		}
		topCustomers = append(topCustomers, customer)
	}
	stats.TopCustomers = topCustomers

	return stats, nil
}

func (r *WorkflowRepository) GetJobSummaries(ctx context.Context) ([]models.JobSummary, error) {
	// Optimized single query with conditional aggregation
	query := `
		WITH job_states AS (
			SELECT 
				r.*,
				i.datein as inventory_date,
				i.joints as inventory_joints,
				CASE 
					WHEN i.datein IS NOT NULL AND i.joints < 0 THEN 'COMPLETED'
					WHEN i.datein IS NOT NULL AND i.joints > 0 THEN 'INVENTORY'
					WHEN r.inspected IS NOT NULL AND r.complete = true THEN 'INSPECTION'
					WHEN r.inproduction IS NOT NULL THEN 'PRODUCTION'
					ELSE 'RECEIVING'
				END as current_state,
				CASE 
					WHEN i.datein IS NOT NULL AND i.joints < 0 THEN 
						EXTRACT(DAYS FROM (CURRENT_DATE - r.daterecvd::date))
					WHEN i.datein IS NOT NULL AND i.joints > 0 THEN 
						EXTRACT(DAYS FROM (CURRENT_DATE - r.inspected::date))
					WHEN r.inspected IS NOT NULL AND r.complete = true THEN 
						EXTRACT(DAYS FROM (CURRENT_DATE - r.inproduction::date))
					WHEN r.inproduction IS NOT NULL THEN 
						EXTRACT(DAYS FROM (CURRENT_DATE - r.daterecvd::date))
					ELSE 
						EXTRACT(DAYS FROM (CURRENT_DATE - r.daterecvd::date))
				END as days_in_state
			FROM store.received r
			LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
			WHERE r.deleted = false
		)
		SELECT 
			current_state,
			COUNT(*) as job_count,
			COALESCE(SUM(joints), 0) as total_joints,
			ROUND(AVG(days_in_state), 1) as avg_days
		FROM job_states
		GROUP BY current_state
		ORDER BY 
			CASE current_state
				WHEN 'RECEIVING' THEN 1
				WHEN 'PRODUCTION' THEN 2
				WHEN 'INSPECTION' THEN 3
				WHEN 'INVENTORY' THEN 4
				WHEN 'SHIPPING' THEN 5
				WHEN 'COMPLETED' THEN 6
			END
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute job summaries query: %w", err)
	}
	defer rows.Close()

	var summaries []models.JobSummary
	for rows.Next() {
		var summary models.JobSummary
		var stateStr string
		if err := rows.Scan(&stateStr, &summary.Count, &summary.TotalJoints, &summary.AvgDays); err != nil {
			return nil, fmt.Errorf("failed to scan job summary: %w", err)
		}
		summary.State = models.WorkflowState(stateStr)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r *WorkflowRepository) GetRecentActivity(ctx context.Context, limit int) ([]models.Job, error) {
	query := `
		SELECT 
			r.id, r.wkorder, r.custid, r.customer, r.customerpo,
			r.size, r.weight, r.grade, r.connection, r.joints,
			r.ctd, r.wstring, COALESCE(r.well, '') as swgcc,
			COALESCE(r.well, '') as well, COALESCE(r.lease, '') as lease,
			COALESCE(r.rack, '') as rack, COALESCE(r.trucking, '') as trucking,
			r.daterecvd, r.inproduction, r.inspected, 
			i.datein, i.dateout,
			r.complete, r.deleted,
			COALESCE(r.orderedby, '') as orderedby,
			COALESCE(r.enteredby, '') as enteredby,
			COALESCE(r.inspectedby, '') as inspectedby,
			COALESCE(r.notes, '') as notes,
			COALESCE(r.services, '') as services,
			COALESCE(r.background, '') as background,
			GREATEST(r.daterecvd, r.inproduction, r.inspected, i.datein) as last_activity
		FROM store.received r
		LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
		WHERE r.deleted = false
		ORDER BY last_activity DESC NULLS LAST
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		var lastActivity sql.NullTime
		
		err := rows.Scan(
			&job.ID, &job.WorkOrder, &job.CustomerID, &job.Customer, &job.CustomerPO,
			&job.Size, &job.Weight, &job.Grade, &job.Connection, &job.Joints,
			&job.CTD, &job.WString, &job.SWGCC,
			&job.Well, &job.Lease, &job.Rack, &job.Trucking,
			&job.DateReceived, &job.InProduction, &job.Inspected,
			&job.DateIn, &job.DateOut,
			&job.Complete, &job.Deleted,
			&job.OrderedBy, &job.EnteredBy, &job.InspectedBy,
			&job.Notes, &job.Services, &job.Background,
			&lastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		job.CurrentState = job.GetCurrentState()
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Job operations
func (r *WorkflowRepository) GetJobs(ctx context.Context, filters repository.JobFilters) ([]models.Job, *models.Pagination, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition
	conditions = append(conditions, "r.deleted = false")

	// Apply filters
	if filters.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("r.custid = $%d", argIndex))
		args = append(args, *filters.CustomerID)
		argIndex++
	}

	if filters.WorkOrder != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.wkorder) LIKE UPPER($%d)", argIndex))
		args = append(args, "%"+filters.WorkOrder+"%")
		argIndex++
	}

	if filters.Grade != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.grade) = UPPER($%d)", argIndex))
		args = append(args, filters.Grade)
		argIndex++
	}

	if filters.Size != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.size) = UPPER($%d)", argIndex))
		args = append(args, filters.Size)
		argIndex++
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("r.daterecvd >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("r.daterecvd <= $%d", argIndex))
		args = append(args, filters.DateTo)
		argIndex++
	}

	if filters.State != nil {
		// Add state-specific conditions based on workflow logic
		switch *filters.State {
		case models.StateReceiving:
			conditions = append(conditions, "r.inproduction IS NULL")
		case models.StateProduction:
			conditions = append(conditions, "r.inproduction IS NOT NULL AND r.inspected IS NULL")
		case models.StateInspection:
			conditions = append(conditions, "r.inspected IS NOT NULL AND r.complete = true AND i.datein IS NULL")
		case models.StateInventory:
			conditions = append(conditions, "i.datein IS NOT NULL AND i.joints > 0")
		case models.StateCompleted:
			conditions = append(conditions, "i.dateout IS NOT NULL AND i.joints < 0")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT r.id)
		FROM store.received r
		LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count jobs: %w", err)
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

	// Get paginated results
	orderBy := "r.daterecvd"
	if filters.OrderBy != "" {
		orderBy = "r." + filters.OrderBy
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT
			r.id, r.wkorder, r.custid, r.customer, r.customerpo,
			r.size, r.weight, r.grade, r.connection, r.joints,
			r.ctd, r.wstring, COALESCE(r.well, '') as swgcc,
			COALESCE(r.well, '') as well, COALESCE(r.lease, '') as lease,
			COALESCE(r.rack, '') as rack, COALESCE(r.trucking, '') as trucking,
			r.daterecvd, r.inproduction, r.inspected,
			i.datein, i.dateout,
			r.complete, r.deleted,
			COALESCE(r.orderedby, '') as orderedby,
			COALESCE(r.enteredby, '') as enteredby,
			COALESCE(r.inspectedby, '') as inspectedby,
			COALESCE(r.notes, '') as notes,
			COALESCE(r.services, '') as services,
			COALESCE(r.background, '') as background
		FROM store.received r
		LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, filters.OrderDir, argIndex, argIndex+1)

	args = append(args, filters.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID, &job.WorkOrder, &job.CustomerID, &job.Customer, &job.CustomerPO,
			&job.Size, &job.Weight, &job.Grade, &job.Connection, &job.Joints,
			&job.CTD, &job.WString, &job.SWGCC,
			&job.Well, &job.Lease, &job.Rack, &job.Trucking,
			&job.DateReceived, &job.InProduction, &job.Inspected,
			&job.DateIn, &job.DateOut,
			&job.Complete, &job.Deleted,
			&job.OrderedBy, &job.EnteredBy, &job.InspectedBy,
			&job.Notes, &job.Services, &job.Background,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan job: %w", err)
		}

		job.CurrentState = job.GetCurrentState()
		jobs = append(jobs, job)
	}

	return jobs, pagination, nil
}

func (r *WorkflowRepository) GetJobByID(ctx context.Context, id int) (*models.Job, error) {
	query := `
		SELECT 
			r.id, r.wkorder, r.custid, r.customer, r.customerpo,
			r.size, r.weight, r.grade, r.connection, r.joints,
			r.ctd, r.wstring, COALESCE(r.well, '') as swgcc,
			COALESCE(r.well, '') as well, COALESCE(r.lease, '') as lease,
			COALESCE(r.rack, '') as rack, COALESCE(r.trucking, '') as trucking,
			r.daterecvd, r.inproduction, r.inspected,
			i.datein, i.dateout,
			r.complete, r.deleted,
			COALESCE(r.orderedby, '') as orderedby,
			COALESCE(r.enteredby, '') as enteredby,
			COALESCE(r.inspectedby, '') as inspectedby,
			COALESCE(r.notes, '') as notes,
			COALESCE(r.services, '') as services,
			COALESCE(r.background, '') as background
		FROM store.received r
		LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
		WHERE r.id = $1 AND r.deleted = false
	`

	var job models.Job
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.WorkOrder, &job.CustomerID, &job.Customer, &job.CustomerPO,
		&job.Size, &job.Weight, &job.Grade, &job.Connection, &job.Joints,
		&job.CTD, &job.WString, &job.SWGCC,
		&job.Well, &job.Lease, &job.Rack, &job.Trucking,
		&job.DateReceived, &job.InProduction, &job.Inspected,
		&job.DateIn, &job.DateOut,
		&job.Complete, &job.Deleted,
		&job.OrderedBy, &job.EnteredBy, &job.InspectedBy,
		&job.Notes, &job.Services, &job.Background,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job by ID: %w", err)
	}

	job.CurrentState = job.GetCurrentState()
	return &job, nil
}

func (r *WorkflowRepository) GetJobByWorkOrder(ctx context.Context, workOrder string) (*models.Job, error) {
	query := `
		SELECT 
			r.id, r.wkorder, r.custid, r.customer, r.customerpo,
			r.size, r.weight, r.grade, r.connection, r.joints,
			r.ctd, r.wstring, COALESCE(r.well, '') as swgcc,
			COALESCE(r.well, '') as well, COALESCE(r.lease, '') as lease,
			COALESCE(r.rack, '') as rack, COALESCE(r.trucking, '') as trucking,
			r.daterecvd, r.inproduction, r.inspected,
			i.datein, i.dateout,
			r.complete, r.deleted,
			COALESCE(r.orderedby, '') as orderedby,
			COALESCE(r.enteredby, '') as enteredby,
			COALESCE(r.inspectedby, '') as inspectedby,
			COALESCE(r.notes, '') as notes,
			COALESCE(r.services, '') as services,
			COALESCE(r.background, '') as background
		FROM store.received r
		LEFT JOIN store.inventory i ON r.wkorder = i.wkorder
		WHERE UPPER(r.wkorder) = UPPER($1) AND r.deleted = false
	`

	var job models.Job
	err := r.db.QueryRow(ctx, query, workOrder).Scan(
		&job.ID, &job.WorkOrder, &job.CustomerID, &job.Customer, &job.CustomerPO,
		&job.Size, &job.Weight, &job.Grade, &job.Connection, &job.Joints,
		&job.CTD, &job.WString, &job.SWGCC,
		&job.Well, &job.Lease, &job.Rack, &job.Trucking,
		&job.DateReceived, &job.InProduction, &job.Inspected,
		&job.DateIn, &job.DateOut,
		&job.Complete, &job.Deleted,
		&job.OrderedBy, &job.EnteredBy, &job.InspectedBy,
		&job.Notes, &job.Services, &job.Background,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job by work order: %w", err)
	}

	job.CurrentState = job.GetCurrentState()
	return &job, nil
}

func (r *WorkflowRepository) CreateJob(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO store.received (
			wkorder, custid, customer, lease, orderedby, well, daterecvd,
			billtoid, size, weight, grade, connection, ctd, wstring,
			joints, rack, background, norm, services, trucking,
			customerpo, notes, enteredby, when1, trailer
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25
		) RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		job.WorkOrder, job.CustomerID, job.Customer, job.Lease, job.OrderedBy,
		job.Well, job.DateReceived, "", job.Size, job.Weight, job.Grade,
		job.Connection, job.CTD, job.WString, job.Joints, job.Rack,
		job.Background, "", job.Services, job.Trucking, job.CustomerPO,
		job.Notes, job.EnteredBy, time.Now(), "",
	).Scan(&job.ID)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdateJob(ctx context.Context, job *models.Job) error {
	query := `
		UPDATE store.received SET
			customer = $2, lease = $3, orderedby = $4, well = $5,
			size = $6, weight = $7, grade = $8, connection = $9,
			ctd = $10, wstring = $11, joints = $12, rack = $13,
			background = $14, services = $15, trucking = $16,
			customerpo = $17, notes = $18, inspectedby = $19,
			inproduction = $20, inspected = $21, complete = $22,
			deleted = $23, when2 = $24
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		job.ID, job.Customer, job.Lease, job.OrderedBy, job.Well,
		job.Size, job.Weight, job.Grade, job.Connection,
		job.CTD, job.WString, job.Joints, job.Rack,
		job.Background, job.Services, job.Trucking,
		job.CustomerPO, job.Notes, job.InspectedBy,
		job.InProduction, job.Inspected, job.Complete,
		job.Deleted, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdateJobState(ctx context.Context, workOrder string, state models.WorkflowState) error {
	now := time.Now()
	
	switch state {
	case models.StateProduction:
		query := `UPDATE store.received SET inproduction = $1 WHERE wkorder = $2`
		_, err := r.db.Exec(ctx, query, now, workOrder)
		return err
	case models.StateInspection:
		query := `UPDATE store.received SET inspected = $1, complete = true WHERE wkorder = $2`
		_, err := r.db.Exec(ctx, query, now, workOrder)
		return err
	default:
		return fmt.Errorf("unsupported state transition: %s", state)
	}
}

func (r *WorkflowRepository) DeleteJob(ctx context.Context, id int) error {
	query := `UPDATE store.received SET deleted = true WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	return nil
}

// Inspection operations
func (r *WorkflowRepository) GetInspectionResults(ctx context.Context, workOrder string) ([]models.InspectionResult, error) {
	query := `
		SELECT 
			id, wkorder, color, cn, joints, accept, reject,
			pin, cplg, pc, COALESCE(rack, '') as rack,
			rep_pin, rep_cplg, rep_pc, complete
		FROM store.inspected 
		WHERE UPPER(wkorder) = UPPER($1) AND deleted = false
		ORDER BY cn, id
	`

	rows, err := r.db.Query(ctx, query, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspection results: %w", err)
	}
	defer rows.Close()

	var results []models.InspectionResult
	for rows.Next() {
		var result models.InspectionResult
		err := rows.Scan(
			&result.ID, &result.WorkOrder, &result.Color, &result.CN,
			&result.Joints, &result.Accept, &result.Reject,
			&result.Pin, &result.Coupling, &result.PC, &result.Rack,
			&result.RepairPin, &result.RepairCplg, &result.RepairPC,
			&result.Complete,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inspection result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *WorkflowRepository) CreateInspectionResult(ctx context.Context, result *models.InspectionResult) error {
	query := `
		INSERT INTO store.inspected (
			wkorder, color, cn, joints, accept, reject,
			pin, cplg, pc, rack, rep_pin, rep_cplg, rep_pc, complete
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		result.WorkOrder, result.Color, result.CN, result.Joints,
		result.Accept, result.Reject, result.Pin, result.Coupling,
		result.PC, result.Rack, result.RepairPin, result.RepairCplg,
		result.RepairPC, result.Complete,
	).Scan(&result.ID)

	if err != nil {
		return fmt.Errorf("failed to create inspection result: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdateInspectionResult(ctx context.Context, result *models.InspectionResult) error {
	query := `
		UPDATE store.inspected SET
			color = $2, cn = $3, joints = $4, accept = $5, reject = $6,
			pin = $7, cplg = $8, pc = $9, rack = $10,
			rep_pin = $11, rep_cplg = $12, rep_pc = $13, complete = $14
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		result.ID, result.Color, result.CN, result.Joints,
		result.Accept, result.Reject, result.Pin, result.Coupling,
		result.PC, result.Rack, result.RepairPin, result.RepairCplg,
		result.RepairPC, result.Complete,
	)

	if err != nil {
		return fmt.Errorf("failed to update inspection result: %w", err)
	}

	return nil
}

// Inventory operations
func (r *WorkflowRepository) GetInventory(ctx context.Context, filters repository.InventoryFilters) ([]models.InventoryItem, *models.Pagination, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition - exclude deleted and shipped (negative joints)
	if filters.IncludeShipped {
		conditions = append(conditions, "deleted = false")
	} else {
		conditions = append(conditions, "deleted = false AND joints > 0")
	}

	// Apply filters
	if filters.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("custid = $%d", argIndex))
		args = append(args, *filters.CustomerID)
		argIndex++
	}

	if filters.Grade != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(grade) = UPPER($%d)", argIndex))
		args = append(args, filters.Grade)
		argIndex++
	}

	if filters.Size != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(size) = UPPER($%d)", argIndex))
		args = append(args, filters.Size)
		argIndex++
	}

	if filters.Color != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(color) = UPPER($%d)", argIndex))
		args = append(args, filters.Color)
		argIndex++
	}

	if filters.CN != nil {
		conditions = append(conditions, fmt.Sprintf("cn = $%d", argIndex))
		args = append(args, *filters.CN)
		argIndex++
	}

	if filters.MinJoints != nil {
		conditions = append(conditions, fmt.Sprintf("joints >= $%d", argIndex))
		args = append(args, *filters.MinJoints)
		argIndex++
	}

	if filters.MaxJoints != nil {
		conditions = append(conditions, fmt.Sprintf("joints <= $%d", argIndex))
		args = append(args, *filters.MaxJoints)
		argIndex++
	}

	if filters.Rack != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(rack) LIKE UPPER($%d)", argIndex))
		args = append(args, "%"+filters.Rack+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM store.inventory %s`, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count inventory: %w", err)
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

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT 
			id, COALESCE(wkorder, '') as wkorder, rnumber, custid, customer,
			joints, COALESCE(rack, '') as rack, size, weight, grade,
			connection, ctd, wstring, swgcc, color, cn,
			COALESCE(customerpo, '') as customerpo, datein, dateout,
			COALESCE(location, '') as location
		FROM store.inventory
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, filters.OrderBy, filters.OrderDir, argIndex, argIndex+1)

	args = append(args, filters.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(
			&item.ID, &item.WorkOrder, &item.RNumber, &item.CustomerID,
			&item.Customer, &item.Joints, &item.Rack, &item.Size,
			&item.Weight, &item.Grade, &item.Connection, &item.CTD,
			&item.WString, &item.SWGCC, &item.Color, &item.CN,
			&item.CustomerPO, &item.DateIn, &item.DateOut, &item.Location,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	return items, pagination, nil
}

func (r *WorkflowRepository) GetInventoryByCustomer(ctx context.Context, customerID int) ([]models.InventoryItem, error) {
	query := `
		SELECT 
			id, COALESCE(wkorder, '') as wkorder, rnumber, custid, customer,
			joints, COALESCE(rack, '') as rack, size, weight, grade,
			connection, ctd, wstring, swgcc, color, cn,
			COALESCE(customerpo, '') as customerpo, datein, dateout,
			COALESCE(location, '') as location
		FROM store.inventory
		WHERE custid = $1 AND deleted = false AND joints > 0
		ORDER BY swgcc, cn, rack
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer inventory: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(
			&item.ID, &item.WorkOrder, &item.RNumber, &item.CustomerID,
			&item.Customer, &item.Joints, &item.Rack, &item.Size,
			&item.Weight, &item.Grade, &item.Connection, &item.CTD,
			&item.WString, &item.SWGCC, &item.Color, &item.CN,
			&item.CustomerPO, &item.DateIn, &item.DateOut, &item.Location,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *WorkflowRepository) CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	query := `
		INSERT INTO store.inventory (
			wkorder, rnumber, custid, customer, joints, rack,
			size, weight, grade, connection, ctd, wstring,
			swgcc, color, cn, customerpo, datein, location
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		item.WorkOrder, item.RNumber, item.CustomerID, item.Customer,
		item.Joints, item.Rack, item.Size, item.Weight, item.Grade,
		item.Connection, item.CTD, item.WString, item.SWGCC,
		item.Color, item.CN, item.CustomerPO, item.DateIn, item.Location,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	query := `
		UPDATE store.inventory SET
			joints = $2, rack = $3, size = $4, weight = $5, grade = $6,
			connection = $7, ctd = $8, wstring = $9, swgcc = $10,
			color = $11, cn = $12, customerpo = $13, location = $14,
			dateout = $15
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		item.ID, item.Joints, item.Rack, item.Size, item.Weight,
		item.Grade, item.Connection, item.CTD, item.WString,
		item.SWGCC, item.Color, item.CN, item.CustomerPO,
		item.Location, item.DateOut,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) ShipInventory(ctx context.Context, itemIDs []int, shipmentDetails map[string]interface{}) error {
	if len(itemIDs) == 0 {
		return fmt.Errorf("no items to ship")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update inventory items - set negative joints to indicate shipped
	query := `
		UPDATE store.inventory 
		SET joints = -ABS(joints), dateout = $1
		WHERE id = ANY($2) AND joints > 0
	`

	now := time.Now()
	_, err = tx.Exec(ctx, query, now, itemIDs)
	if err != nil {
		return fmt.Errorf("failed to ship inventory items: %w", err)
	}

	return tx.Commit(ctx)
}

// Customer operations
func (r *WorkflowRepository) GetCustomers(ctx context.Context, includeDeleted bool) ([]models.Customer, error) {
	query := `
		SELECT 
			custid, customer, COALESCE(billingaddress, '') as billingaddress,
			COALESCE(billingcity, '') as billingcity, 
			COALESCE(billingstate, '') as billingstate,
			COALESCE(billingzipcode, '') as billingzipcode,
			COALESCE(contact, '') as contact, COALESCE(phone, '') as phone,
			COALESCE(email, '') as email, deleted
		FROM store.customers
	`

	if !includeDeleted {
		query += " WHERE deleted = false"
	}
	query += " ORDER BY customer"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID, &customer.Name, &customer.BillingAddress,
			&customer.BillingCity, &customer.BillingState, &customer.BillingZip,
			&customer.Contact, &customer.Phone, &customer.Email, &customer.Deleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	return customers, nil
}

func (r *WorkflowRepository) GetCustomerByID(ctx context.Context, id int) (*models.Customer, error) {
	query := `
		SELECT 
			custid, customer, COALESCE(billingaddress, '') as billingaddress,
			COALESCE(billingcity, '') as billingcity, 
			COALESCE(billingstate, '') as billingstate,
			COALESCE(billingzipcode, '') as billingzipcode,
			COALESCE(contact, '') as contact, COALESCE(phone, '') as phone,
			COALESCE(email, '') as email, deleted
		FROM store.customers
		WHERE custid = $1
	`

	var customer models.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&customer.ID, &customer.Name, &customer.BillingAddress,
		&customer.BillingCity, &customer.BillingState, &customer.BillingZip,
		&customer.Contact, &customer.Phone, &customer.Email, &customer.Deleted,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *WorkflowRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO store.customers (
			customer, billingaddress, billingcity, billingstate,
			billingzipcode, contact, phone, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING custid
	`

	err := r.db.QueryRow(ctx, query,
		customer.Name, customer.BillingAddress, customer.BillingCity,
		customer.BillingState, customer.BillingZip, customer.Contact,
		customer.Phone, customer.Email,
	).Scan(&customer.ID)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	query := `
		UPDATE store.customers SET
			customer = $2, billingaddress = $3, billingcity = $4,
			billingstate = $5, billingzipcode = $6, contact = $7,
			phone = $8, email = $9, deleted = $10
		WHERE custid = $1
	`

	_, err := r.db.Exec(ctx, query,
		customer.ID, customer.Name, customer.BillingAddress,
		customer.BillingCity, customer.BillingState, customer.BillingZip,
		customer.Contact, customer.Phone, customer.Email, customer.Deleted,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) DeleteCustomer(ctx context.Context, id int) error {
	query := `UPDATE store.customers SET deleted = true WHERE custid = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	return nil
}

// Pipe size operations
func (r *WorkflowRepository) GetPipeSizes(ctx context.Context, customerID int) ([]models.PipeSize, error) {
	query := `
		SELECT sizeid, custid, size, weight, connection
		FROM store.swgc
		WHERE custid = $1
		ORDER BY size, weight, connection
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipe sizes: %w", err)
	}
	defer rows.Close()

	var sizes []models.PipeSize
	for rows.Next() {
		var size models.PipeSize
		err := rows.Scan(&size.ID, &size.CustomerID, &size.Size, &size.Weight, &size.Connection)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pipe size: %w", err)
		}
		sizes = append(sizes, size)
	}

	return sizes, nil
}

func (r *WorkflowRepository) CreatePipeSize(ctx context.Context, size *models.PipeSize) error {
	query := `
		INSERT INTO store.swgc (custid, size, weight, connection)
		VALUES ($1, $2, $3, $4)
		RETURNING sizeid
	`

	err := r.db.QueryRow(ctx, query,
		size.CustomerID, size.Size, size.Weight, size.Connection,
	).Scan(&size.ID)

	if err != nil {
		return fmt.Errorf("failed to create pipe size: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) UpdatePipeSize(ctx context.Context, size *models.PipeSize) error {
	query := `
		UPDATE store.swgc SET
			size = $2, weight = $3, connection = $4
		WHERE sizeid = $1
	`

	_, err := r.db.Exec(ctx, query, size.ID, size.Size, size.Weight, size.Connection)
	if err != nil {
		return fmt.Errorf("failed to update pipe size: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) DeletePipeSize(ctx context.Context, id int) error {
	query := `DELETE FROM store.swgc WHERE sizeid = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pipe size: %w", err)
	}
	return nil
}

// Grades
func (r *WorkflowRepository) GetGrades(ctx context.Context) ([]string, error) {
	query := `SELECT grade FROM store.grade ORDER BY grade`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []string
	for rows.Next() {
		var grade string
		if err := rows.Scan(&grade); err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, grade)
	}

	return grades, nil
}

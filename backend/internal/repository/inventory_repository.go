// backend/internal/repository/inventory_repository.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/internal/models"
)

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

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("inventory item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return &item, nil
}

func (r *inventoryRepository) GetByWorkOrder(ctx context.Context, workOrder string) ([]models.InventoryItem, error) {
	query := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory 
		WHERE work_order = $1 AND deleted = false
		ORDER BY cn, id
	`

	rows, err := r.db.Query(ctx, query, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory by work order: %w", err)
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
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *inventoryRepository) Create(ctx context.Context, item *models.InventoryItem) error {
	query := `
		INSERT INTO store.inventory (
			username, work_order, r_number, customer_id, customer, joints, rack,
			size, weight, grade, connection, ctd, w_string, swgcc, color, 
			customer_po, fletcher, date_in, date_out, well_in, lease_in, 
			well_out, lease_out, trucking, trailer, location, notes, pcode, 
			cn, ordered_by, deleted, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, false, NOW()
		) RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		item.Username, item.WorkOrder, item.RNumber, item.CustomerID, item.Customer,
		item.Joints, item.Rack, item.Size, item.Weight, item.Grade, item.Connection,
		item.CTD, item.WString, item.SWGCC, item.Color, item.CustomerPO, item.Fletcher,
		item.DateIn, item.DateOut, item.WellIn, item.LeaseIn, item.WellOut, item.LeaseOut,
		item.Trucking, item.Trailer, item.Location, item.Notes, item.PCode, item.CN,
		item.OrderedBy,
	).Scan(&item.ID, &item.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

func (r *inventoryRepository) Update(ctx context.Context, item *models.InventoryItem) error {
	query := `
		UPDATE store.inventory SET
			username = $2, work_order = $3, r_number = $4, customer_id = $5, customer = $6,
			joints = $7, rack = $8, size = $9, weight = $10, grade = $11, connection = $12,
			ctd = $13, w_string = $14, swgcc = $15, color = $16, customer_po = $17,
			fletcher = $18, date_in = $19, date_out = $20, well_in = $21, lease_in = $22,
			well_out = $23, lease_out = $24, trucking = $25, trailer = $26, location = $27,
			notes = $28, pcode = $29, cn = $30, ordered_by = $31
		WHERE id = $1 AND deleted = false
	`

	result, err := r.db.Exec(ctx, query, item.ID,
		item.Username, item.WorkOrder, item.RNumber, item.CustomerID, item.Customer,
		item.Joints, item.Rack, item.Size, item.Weight, item.Grade, item.Connection,
		item.CTD, item.WString, item.SWGCC, item.Color, item.CustomerPO, item.Fletcher,
		item.DateIn, item.DateOut, item.WellIn, item.LeaseIn, item.WellOut, item.LeaseOut,
		item.Trucking, item.Trailer, item.Location, item.Notes, item.PCode, item.CN,
		item.OrderedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory item not found or already deleted")
	}

	return nil
}

func (r *inventoryRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE store.inventory SET deleted = true WHERE id = $1 AND deleted = false`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory item not found")
	}

	return nil
}

func (r *inventoryRepository) GetFiltered(ctx context.Context, filters InventoryFilters) ([]models.InventoryItem, *models.Pagination, error) {
	filters.NormalizePagination()
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition
	if filters.IncludeDeleted {
		conditions = append(conditions, "deleted = false OR deleted = true")
	} else {
		conditions = append(conditions, "deleted = false")
	}

	// Apply filters
	if filters.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, *filters.CustomerID)
		argIndex++
	}

	if filters.Grade != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(grade) = UPPER($%d)", argIndex))
		args = append(args, filters.Grade)
		argIndex++
	}

	if filters.Size != "" {
		conditions = append(conditions, fmt.Sprintf("size = $%d", argIndex))
		args = append(args, filters.Size)
		argIndex++
	}

	if filters.Color != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(color) = UPPER($%d)", argIndex))
		args = append(args, filters.Color)
		argIndex++
	}

	if filters.Location != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(location) LIKE UPPER($%d)", argIndex))
		args = append(args, "%"+filters.Location+"%")
		argIndex++
	}

	if filters.Rack != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(rack) LIKE UPPER($%d)", argIndex))
		args = append(args, "%"+filters.Rack+"%")
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

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("date_in >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("date_in <= $%d", argIndex))
		args = append(args, filters.DateTo)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM store.inventory %s", whereClause)
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

	// Get filtered records
	query := fmt.Sprintf(`
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, filters.OrderBy, filters.OrderDir, argIndex, argIndex+1)

	args = append(args, filters.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query inventory: %w", err)
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
			return nil, nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	return items, pagination, rows.Err()
}

func (r *inventoryRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error) {
	searchTerm := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	
	whereClause := `
		WHERE deleted = false AND (
			LOWER(customer) LIKE $1 OR
			LOWER(grade) LIKE $1 OR  
			LOWER(size) LIKE $1 OR
			LOWER(location) LIKE $1 OR
			LOWER(rack) LIKE $1 OR
			LOWER(notes) LIKE $1 OR
			LOWER(customer_po) LIKE $1 OR
			LOWER(work_order) LIKE $1
		)
	`

	// Count total
	countQuery := "SELECT COUNT(*) FROM store.inventory " + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get results with relevance scoring
	searchQuery := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory ` + whereClause + `
		ORDER BY 
			CASE WHEN LOWER(customer) LIKE $1 THEN 1
			     WHEN LOWER(grade) LIKE $1 THEN 2
			     WHEN LOWER(customer_po) LIKE $1 THEN 3
			     ELSE 4 END,
			date_in DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
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

	return items, total, rows.Err()
}

func (r *inventoryRepository) GetSummary(ctx context.Context) (*models.InventorySummary, error) {
	summary := &models.InventorySummary{
		ItemsByGrade:    make(map[string]int),
		ItemsByCustomer: make(map[string]int),
		ItemsByLocation: make(map[string]int),
		LastUpdated:     time.Now(),
	}

	// Get totals
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(joints), 0) 
		FROM store.inventory 
		WHERE deleted = false AND joints > 0
	`).Scan(&summary.TotalItems, &summary.TotalJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Items by grade
	rows, err := r.db.Query(ctx, `
		SELECT grade, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND grade IS NOT NULL AND joints > 0
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

	// Items by customer (top 10)
	rows, err = r.db.Query(ctx, `
		SELECT customer, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND customer IS NOT NULL AND joints > 0
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

	// Items by location (top 10)
	rows, err = r.db.Query(ctx, `
		SELECT location, COUNT(*) 
		FROM store.inventory 
		WHERE deleted = false AND location IS NOT NULL AND joints > 0
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
		SELECT id, customer, joints, size, grade, location, date_in, customer_id, created_at
		FROM store.inventory 
		WHERE deleted = false AND joints > 0
		ORDER BY date_in DESC NULLS LAST, created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.ID, &item.Customer, &item.Joints, &item.Size, 
			&item.Grade, &item.Location, &item.DateIn, &item.CustomerID, &item.CreatedAt)
		if err == nil {
			summary.RecentActivity = append(summary.RecentActivity, item)
		}
	}

	return summary, nil
}

func (r *inventoryRepository) GetByCustomer(ctx context.Context, customerID int) ([]models.InventoryItem, error) {
	query := `
		SELECT id, username, work_order, r_number, customer_id, customer, joints, 
		       rack, size, weight, grade, connection, ctd, w_string, swgcc, color, 
		       customer_po, fletcher, date_in, date_out, well_in, lease_in, 
		       well_out, lease_out, trucking, trailer, location, notes, pcode, 
		       cn, ordered_by, deleted, created_at
		FROM store.inventory 
		WHERE customer_id = $1 AND deleted = false AND joints > 0
		ORDER BY grade, size, date_in DESC
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
			&item.ID, &item.Username, &item.WorkOrder, &item.RNumber, &item.CustomerID,
			&item.Customer, &item.Joints, &item.Rack, &item.Size, &item.Weight,
			&item.Grade, &item.Connection, &item.CTD, &item.WString, &item.SWGCC,
			&item.Color, &item.CustomerPO, &item.Fletcher, &item.DateIn, &item.DateOut,
			&item.WellIn, &item.LeaseIn, &item.WellOut, &item.LeaseOut, &item.Trucking,
			&item.Trailer, &item.Location, &item.Notes, &item.PCode, &item.CN,
			&item.OrderedBy, &item.Deleted, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *inventoryRepository) GetRecentActivity(ctx context.Context, days int) ([]models.InventoryItem, error) {
	query := `
		SELECT id, customer, joints, size, grade, location, date_in, customer_id, created_at
		FROM store.inventory 
		WHERE deleted = false 
		AND (date_in >= NOW() - INTERVAL '%d days' OR created_at >= NOW() - INTERVAL '%d days')
		ORDER BY COALESCE(date_in, created_at) DESC
		LIMIT 50
	`

	rows, err := r.db.Query(ctx, fmt.Sprintf(query, days, days))
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.ID, &item.Customer, &item.Joints, &item.Size, 
			&item.Grade, &item.Location, &item.DateIn, &item.CustomerID, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent activity: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

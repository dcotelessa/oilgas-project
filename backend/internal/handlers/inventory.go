// backend/internal/handlers/inventory.go
// Inventory API endpoints with tenant isolation and filtering
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/database"
)

type InventoryItem struct {
	ID         int     `json:"id" db:"id"`
	WorkOrder  string  `json:"work_order" db:"work_order"`
	RNumber    string  `json:"r_number" db:"r_number"`
	CustomerID int     `json:"customer_id" db:"customer_id"`
	Customer   string  `json:"customer" db:"customer"`
	Joints     int     `json:"joints" db:"joints"`
	Rack       string  `json:"rack" db:"rack"`
	Size       string  `json:"size" db:"size"`
	Weight     float64 `json:"weight" db:"weight"`
	Grade      string  `json:"grade" db:"grade"`
	Connection string  `json:"connection" db:"connection"`
	CTD        string  `json:"ctd" db:"ctd"`
	WString    string  `json:"w_string" db:"w_string"`
	Color      string  `json:"color" db:"color"`
	DateIn     string  `json:"date_in" db:"date_in"`
	DateOut    string  `json:"date_out,omitempty" db:"date_out"`
	WellIn     string  `json:"well_in" db:"well_in"`
	LeaseIn    string  `json:"lease_in" db:"lease_in"`
	WellOut    string  `json:"well_out,omitempty" db:"well_out"`
	LeaseOut   string  `json:"lease_out,omitempty" db:"lease_out"`
	Location   string  `json:"location" db:"location"`
	Notes      string  `json:"notes" db:"notes"`
	TenantID   string  `json:"tenant_id" db:"tenant_id"`
}

// GetInventory returns inventory items for the specified tenant
func GetInventory(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	search := strings.TrimSpace(c.Query("search"))
	customerIDStr := c.Query("customer_id")
	workOrder := strings.TrimSpace(c.Query("work_order"))
	size := strings.TrimSpace(c.Query("size"))
	grade := strings.TrimSpace(c.Query("grade"))
	location := strings.TrimSpace(c.Query("location"))
	dateFrom := strings.TrimSpace(c.Query("date_from")) // YYYY-MM-DD format
	dateTo := strings.TrimSpace(c.Query("date_to"))     // YYYY-MM-DD format
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	// Convert and validate pagination parameters
	limitInt, offsetInt := validatePagination(limit, offset)

	db, err := database.GetTenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
			"details": err.Error(),
		})
		return
	}

	// Build query with filters
	query := `
		SELECT id, COALESCE(work_order, '') as work_order, COALESCE(r_number, '') as r_number,
		       customer_id, customer, COALESCE(joints, 0) as joints, COALESCE(rack, '') as rack,
		       COALESCE(size, '') as size, COALESCE(weight, 0) as weight, COALESCE(grade, '') as grade,
		       COALESCE(connection, '') as connection, COALESCE(ctd, '') as ctd, 
		       COALESCE(w_string, '') as w_string, COALESCE(color, '') as color,
		       COALESCE(date_in::text, '') as date_in, COALESCE(date_out::text, '') as date_out,
		       COALESCE(well_in, '') as well_in, COALESCE(lease_in, '') as lease_in,
		       COALESCE(well_out, '') as well_out, COALESCE(lease_out, '') as lease_out,
		       COALESCE(location, '') as location, COALESCE(notes, '') as notes, tenant_id
		FROM store.inventory 
		WHERE tenant_id = $1 AND NOT deleted
	`
	args := []interface{}{tenantID}

	// Add customer filter
	if customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			query += ` AND customer_id = $` + strconv.Itoa(len(args)+1)
			args = append(args, customerID)
		}
	}

	// Add work order filter
	if workOrder != "" {
		query += ` AND work_order ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+workOrder+"%")
	}

	// Add size filter
	if size != "" {
		query += ` AND size ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+size+"%")
	}

	// Add grade filter
	if grade != "" {
		query += ` AND grade ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+grade+"%")
	}

	// Add location filter
	if location != "" {
		query += ` AND location ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+location+"%")
	}

	// Add date range filter
	if dateFrom != "" {
		query += ` AND date_in >= $` + strconv.Itoa(len(args)+1)
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		query += ` AND date_in <= $` + strconv.Itoa(len(args)+1)
		args = append(args, dateTo)
	}

	// Add general search filter
	if search != "" {
		search = sanitizeSearchTerm(search)
		query += ` AND (
			customer ILIKE $` + strconv.Itoa(len(args)+1) + ` OR 
			work_order ILIKE $` + strconv.Itoa(len(args)+1) + ` OR 
			size ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			grade ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			notes ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			location ILIKE $` + strconv.Itoa(len(args)+1) + `
		)`
		args = append(args, "%"+search+"%")
	}

	query += ` ORDER BY date_in DESC, id DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, limitInt, offsetInt)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Query failed",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var items []InventoryItem
	for rows.Next() {
		var item InventoryItem
		err := rows.Scan(
			&item.ID, &item.WorkOrder, &item.RNumber, &item.CustomerID, &item.Customer,
			&item.Joints, &item.Rack, &item.Size, &item.Weight, &item.Grade,
			&item.Connection, &item.CTD, &item.WString, &item.Color,
			&item.DateIn, &item.DateOut, &item.WellIn, &item.LeaseIn,
			&item.WellOut, &item.LeaseOut, &item.Location, &item.Notes, &item.TenantID,
		)
		if err != nil {
			continue // Skip problematic rows
		}
		items = append(items, item)
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted`
	countArgs := []interface{}{tenantID}
	
	// Apply same filters to count query
	if customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			countQuery += ` AND customer_id = $` + strconv.Itoa(len(countArgs)+1)
			countArgs = append(countArgs, customerID)
		}
	}
	if workOrder != "" {
		countQuery += ` AND work_order ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+workOrder+"%")
	}
	if size != "" {
		countQuery += ` AND size ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+size+"%")
	}
	if grade != "" {
		countQuery += ` AND grade ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+grade+"%")
	}
	if location != "" {
		countQuery += ` AND location ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+location+"%")
	}
	if dateFrom != "" {
		countQuery += ` AND date_in >= $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, dateFrom)
	}
	if dateTo != "" {
		countQuery += ` AND date_in <= $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, dateTo)
	}
	if search != "" {
		countQuery += ` AND (customer ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR work_order ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR size ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR grade ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR notes ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR location ILIKE $` + strconv.Itoa(len(countArgs)+1) + `)`
		countArgs = append(countArgs, "%"+search+"%")
	}

	var totalCount int
	db.QueryRow(countQuery, countArgs...).Scan(&totalCount)

	// Get summary statistics
	summaryStats := getInventorySummary(db, tenantID, countArgs[1:]) // Exclude tenant_id from countArgs

	// Calculate pagination info
	hasMore := (offsetInt + limitInt) < totalCount
	page := (offsetInt / limitInt) + 1

	response := gin.H{
		"tenant":    tenantID,
		"inventory": items,
		"count":     len(items),
		"total":     totalCount,
		"limit":     limitInt,
		"offset":    offsetInt,
		"page":      page,
		"has_more":  hasMore,
		"search":    search,
		"filters": gin.H{
			"customer_id": customerIDStr,
			"work_order":  workOrder,
			"size":        size,
			"grade":       grade,
			"location":    location,
			"date_from":   dateFrom,
			"date_to":     dateTo,
		},
	}

	if summaryStats != nil {
		response["summary"] = summaryStats
	}

	c.JSON(http.StatusOK, response)
}

// getInventorySummary returns summary statistics for the inventory query
func getInventorySummary(db *sql.DB, tenantID string, additionalFilters []interface{}) map[string]interface{} {
	// Build base query for summary
	baseQuery := `FROM store.inventory WHERE tenant_id = $1 AND NOT deleted`
	args := []interface{}{tenantID}
	args = append(args, additionalFilters...)

	// Note: In a real implementation, you'd need to rebuild the filter conditions
	// For now, we'll just get basic totals for the tenant
	
	summary := make(map[string]interface{})

	// Get total joints and weight
	var totalJoints sql.NullInt64
	var totalWeight sql.NullFloat64
	err := db.QueryRow(
		`SELECT SUM(joints), SUM(weight) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted`,
		tenantID,
	).Scan(&totalJoints, &totalWeight)
	
	if err == nil {
		if totalJoints.Valid {
			summary["total_joints"] = totalJoints.Int64
		}
		if totalWeight.Valid {
			summary["total_weight"] = totalWeight.Float64
		}
	}

	// Get unique counts
	var uniqueCustomers, uniqueWorkOrders, uniqueSizes, uniqueGrades int
	
	db.QueryRow(`SELECT COUNT(DISTINCT customer_id) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted`, tenantID).Scan(&uniqueCustomers)
	db.QueryRow(`SELECT COUNT(DISTINCT work_order) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted AND work_order IS NOT NULL AND work_order != ''`, tenantID).Scan(&uniqueWorkOrders)
	db.QueryRow(`SELECT COUNT(DISTINCT size) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted AND size IS NOT NULL AND size != ''`, tenantID).Scan(&uniqueSizes)
	db.QueryRow(`SELECT COUNT(DISTINCT grade) FROM store.inventory WHERE tenant_id = $1 AND NOT deleted AND grade IS NOT NULL AND grade != ''`, tenantID).Scan(&uniqueGrades)

	summary["unique_customers"] = uniqueCustomers
	summary["unique_work_orders"] = uniqueWorkOrders
	summary["unique_sizes"] = uniqueSizes
	summary["unique_grades"] = uniqueGrades

	return summary
}

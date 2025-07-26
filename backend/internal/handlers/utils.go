// backend/internal/handlers/utils.go
// Additional handler functions for work orders and inventory items
package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"your-project/internal/database"
)

// GetInventoryItem returns a specific inventory item by ID
func GetInventoryItem(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	itemIDStr := c.Param("id")

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory item ID",
			"message": "ID must be a number",
		})
		return
	}

	db, err := database.GetTenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
			"details": err.Error(),
		})
		return
	}

	query := `
		SELECT id, work_order, customer_id, customer, joints, size, weight,
		       grade, connection, date_in, well_in, lease_in, location, notes, tenant_id
		FROM store.inventory 
		WHERE id = $1 AND tenant_id = $2 AND NOT deleted
	`

	var item InventoryItem
	err = db.QueryRow(query, itemID, tenantID).Scan(
		&item.ID, &item.WorkOrder, &item.CustomerID, &item.Customer,
		&item.Joints, &item.Size, &item.Weight, &item.Grade,
		&item.Connection, &item.DateIn, &item.WellIn, &item.LeaseIn,
		&item.Location, &item.Notes, &item.TenantID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Inventory item not found",
				"message": fmt.Sprintf("No inventory item with ID %d found for tenant %s", itemID, tenantID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Query failed",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant": tenantID,
		"item":   item,
	})
}

// GetWorkOrders returns work orders for the tenant
func GetWorkOrders(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	search := c.Query("search")
	customerIDStr := c.Query("customer_id")
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Validate limits
	if limitInt > 1000 {
		limitInt = 1000
	}
	if limitInt < 1 {
		limitInt = 50
	}

	db, err := database.GetTenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
			"details": err.Error(),
		})
		return
	}

	// Aggregate work orders from inventory table
	query := `
		SELECT work_order, customer, customer_id, COUNT(*) as item_count, 
		       SUM(joints) as total_joints, MIN(date_in) as start_date,
		       MAX(date_in) as latest_date, 
		       STRING_AGG(DISTINCT location, ', ' ORDER BY location) as locations,
		       STRING_AGG(DISTINCT size, ', ' ORDER BY size) as sizes,
		       STRING_AGG(DISTINCT grade, ', ' ORDER BY grade) as grades
		FROM store.inventory 
		WHERE tenant_id = $1 AND NOT deleted AND work_order IS NOT NULL AND work_order != ''
	`
	args := []interface{}{tenantID}

	// Add customer filter if specified
	if customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			query += ` AND customer_id = $` + strconv.Itoa(len(args)+1)
			args = append(args, customerID)
		}
	}

	// Add search filter if specified
	if search != "" {
		query += ` AND (work_order ILIKE $` + strconv.Itoa(len(args)+1) + ` OR customer ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, "%"+search+"%")
	}

	query += ` GROUP BY work_order, customer, customer_id ORDER BY latest_date DESC LIMIT $` + 
		strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
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

	type WorkOrder struct {
		WorkOrder   string `json:"work_order"`
		Customer    string `json:"customer"`
		CustomerID  int    `json:"customer_id"`
		ItemCount   int    `json:"item_count"`
		TotalJoints int    `json:"total_joints"`
		StartDate   string `json:"start_date"`
		LatestDate  string `json:"latest_date"`
		Locations   string `json:"locations"`
		Sizes       string `json:"sizes"`
		Grades      string `json:"grades"`
	}

	var workOrders []WorkOrder
	for rows.Next() {
		var wo WorkOrder
		err := rows.Scan(&wo.WorkOrder, &wo.Customer, &wo.CustomerID, &wo.ItemCount, 
			&wo.TotalJoints, &wo.StartDate, &wo.LatestDate, &wo.Locations, &wo.Sizes, &wo.Grades)
		if err != nil {
			continue // Skip problematic rows
		}
		workOrders = append(workOrders, wo)
	}

	// Get total count for pagination
	countQuery := `
		SELECT COUNT(DISTINCT work_order)
		FROM store.inventory 
		WHERE tenant_id = $1 AND NOT deleted AND work_order IS NOT NULL AND work_order != ''
	`
	countArgs := []interface{}{tenantID}
	
	if customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			countQuery += ` AND customer_id = $` + strconv.Itoa(len(countArgs)+1)
			countArgs = append(countArgs, customerID)
		}
	}
	
	if search != "" {
		countQuery += ` AND (work_order ILIKE $` + strconv.Itoa(len(countArgs)+1) + ` OR customer ILIKE $` + strconv.Itoa(len(countArgs)+1) + `)`
		countArgs = append(countArgs, "%"+search+"%")
	}

	var totalCount int
	db.QueryRow(countQuery, countArgs...).Scan(&totalCount)

	c.JSON(http.StatusOK, gin.H{
		"tenant":      tenantID,
		"work_orders": workOrders,
		"count":       len(workOrders),
		"total":       totalCount,
		"limit":       limitInt,
		"offset":      offsetInt,
		"search":      search,
		"customer_id": customerIDStr,
	})
}

// GetWorkOrder returns details for a specific work order
func GetWorkOrder(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	workOrderID := c.Param("id")

	if workOrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Work order ID is required",
		})
		return
	}

	db, err := database.GetTenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
			"details": err.Error(),
		})
		return
	}

	// Get work order summary
	summaryQuery := `
		SELECT work_order, customer, customer_id, COUNT(*) as item_count, 
		       SUM(joints) as total_joints, MIN(date_in) as start_date,
		       MAX(date_in) as latest_date,
		       STRING_AGG(DISTINCT location, ', ' ORDER BY location) as locations
		FROM store.inventory 
		WHERE work_order = $1 AND tenant_id = $2 AND NOT deleted
		GROUP BY work_order, customer, customer_id
	`

	type WorkOrderSummary struct {
		WorkOrder   string `json:"work_order"`
		Customer    string `json:"customer"`
		CustomerID  int    `json:"customer_id"`
		ItemCount   int    `json:"item_count"`
		TotalJoints int    `json:"total_joints"`
		StartDate   string `json:"start_date"`
		LatestDate  string `json:"latest_date"`
		Locations   string `json:"locations"`
	}

	var summary WorkOrderSummary
	err = db.QueryRow(summaryQuery, workOrderID, tenantID).Scan(
		&summary.WorkOrder, &summary.Customer, &summary.CustomerID, &summary.ItemCount,
		&summary.TotalJoints, &summary.StartDate, &summary.LatestDate, &summary.Locations,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Work order not found",
				"message": fmt.Sprintf("No work order '%s' found for tenant %s", workOrderID, tenantID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Query failed",
				"details": err.Error(),
			})
		}
		return
	}

	// Get inventory items for this work order
	itemsQuery := `
		SELECT id, joints, size, weight, grade, connection, 
		       date_in, well_in, lease_in, location, notes
		FROM store.inventory 
		WHERE work_order = $1 AND tenant_id = $2 AND NOT deleted
		ORDER BY date_in, id
	`

	rows, err := db.Query(itemsQuery, workOrderID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get work order items",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, joints int
		var size, grade, connection, dateIn, wellIn, leaseIn, location, notes string
		var weight float64

		err := rows.Scan(&id, &joints, &size, &weight, &grade, &connection,
			&dateIn, &wellIn, &leaseIn, &location, &notes)
		if err != nil {
			continue // Skip problematic rows
		}

		items = append(items, map[string]interface{}{
			"id":         id,
			"joints":     joints,
			"size":       size,
			"weight":     weight,
			"grade":      grade,
			"connection": connection,
			"date_in":    dateIn,
			"well_in":    wellIn,
			"lease_in":   leaseIn,
			"location":   location,
			"notes":      notes,
		})
	}

	// Get related work orders for context
	relatedQuery := `
		SELECT DISTINCT work_order
		FROM store.inventory
		WHERE customer_id = $1 AND tenant_id = $2 AND work_order != $3 AND NOT deleted
		ORDER BY work_order
		LIMIT 10
	`

	relatedRows, err := db.Query(relatedQuery, summary.CustomerID, tenantID, workOrderID)
	if err == nil {
		defer relatedRows.Close()
		var relatedWorkOrders []string
		for relatedRows.Next() {
			var relatedWO string
			if err := relatedRows.Scan(&relatedWO); err == nil {
				relatedWorkOrders = append(relatedWorkOrders, relatedWO)
			}
		}
		
		c.JSON(http.StatusOK, gin.H{
			"tenant":           tenantID,
			"summary":          summary,
			"items":            items,
			"related_work_orders": relatedWorkOrders,
		})
	} else {
		// Return without related work orders if query fails
		c.JSON(http.StatusOK, gin.H{
			"tenant":  tenantID,
			"summary": summary,
			"items":   items,
		})
	}
}

// Helper function to validate and sanitize search terms
func sanitizeSearchTerm(term string) string {
	// Remove potentially problematic characters
	// This is a basic implementation - you might want more sophisticated sanitization
	if len(term) > 100 {
		term = term[:100]
	}
	return term
}

// Helper function to validate pagination parameters
func validatePagination(limitStr, offsetStr string) (int, int) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

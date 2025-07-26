// backend/internal/handlers/search.go
// Global search API endpoint for cross-table search within tenant
package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/database"
)

type SearchResult struct {
	Type   string      `json:"type"`
	ID     interface{} `json:"id"`
	Title  string      `json:"title"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"`
	Score  float64     `json:"score,omitempty"` // For relevance ranking
}

type SearchResponse struct {
	TenantID string         `json:"tenant"`
	Query    string         `json:"query"`
	Results  []SearchResult `json:"results"`
	Count    int            `json:"count"`
	Summary  SearchSummary  `json:"summary"`
}

type SearchSummary struct {
	Customers  int `json:"customers"`
	Inventory  int `json:"inventory"`
	WorkOrders int `json:"work_orders"`
	Total      int `json:"total"`
}

// GlobalSearch performs cross-table search within tenant
func GlobalSearch(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	query := strings.TrimSpace(c.Query("q"))
	searchType := c.Query("type") // Optional: filter by type (customers, inventory, work_orders)
	limit := c.DefaultQuery("limit", "50")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query 'q' parameter required",
			"message": "Provide a search term in the 'q' parameter",
			"example": "/api/v1/search?q=oil",
		})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query too short",
			"message": "Search query must be at least 2 characters",
		})
		return
	}

	limitInt, _ := strconv.Atoi(limit)
	if limitInt > 200 {
		limitInt = 200 // Cap at 200 results
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

	var results []SearchResult
	summary := SearchSummary{}

	// Sanitize search query
	query = sanitizeSearchTerm(query)

	// Search different entity types based on searchType filter
	switch searchType {
	case "customers":
		customerResults, count := searchCustomers(db, tenantID, query, limitInt)
		results = append(results, customerResults...)
		summary.Customers = count
	case "inventory":
		inventoryResults, count := searchInventory(db, tenantID, query, limitInt)
		results = append(results, inventoryResults...)
		summary.Inventory = count
	case "work_orders":
		workOrderResults, count := searchWorkOrders(db, tenantID, query, limitInt)
		results = append(results, workOrderResults...)
		summary.WorkOrders = count
	default:
		// Search all types (default behavior)
		customerResults, customerCount := searchCustomers(db, tenantID, query, limitInt/3)
		inventoryResults, inventoryCount := searchInventory(db, tenantID, query, limitInt/3)
		workOrderResults, workOrderCount := searchWorkOrders(db, tenantID, query, limitInt/3)

		results = append(results, customerResults...)
		results = append(results, inventoryResults...)
		results = append(results, workOrderResults...)

		summary.Customers = customerCount
		summary.Inventory = inventoryCount
		summary.WorkOrders = workOrderCount
	}

	summary.Total = len(results)

	// Sort results by relevance score if available
	if len(results) > 1 {
		sortResultsByRelevance(results, query)
	}

	c.JSON(http.StatusOK, SearchResponse{
		TenantID: tenantID,
		Query:    query,
		Results:  results,
		Count:    len(results),
		Summary:  summary,
	})
}

// searchCustomers searches within customers table
func searchCustomers(db *sql.DB, tenantID, query string, limit int) ([]SearchResult, int) {
	sqlQuery := `
		SELECT customer_id, customer, COALESCE(contact, '') as contact, 
		       COALESCE(billing_city, '') as billing_city, COALESCE(billing_state, '') as billing_state,
		       COALESCE(phone, '') as phone, COALESCE(email, '') as email
		FROM store.customers 
		WHERE tenant_id = $1 AND NOT deleted 
		AND (customer ILIKE $2 OR contact ILIKE $2 OR billing_city ILIKE $2 OR phone ILIKE $2 OR email ILIKE $2)
		ORDER BY 
			CASE 
				WHEN customer ILIKE $3 THEN 1
				WHEN contact ILIKE $3 THEN 2
				ELSE 3
			END,
			customer 
		LIMIT $4
	`

	rows, err := db.Query(sqlQuery, tenantID, "%"+query+"%", query+"%", limit)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var customerID int
		var customer, contact, city, state, phone, email string
		
		if err := rows.Scan(&customerID, &customer, &contact, &city, &state, &phone, &email); err != nil {
			continue
		}

		// Build detail string
		var details []string
		if contact != "" {
			details = append(details, "Contact: "+contact)
		}
		if city != "" && state != "" {
			details = append(details, fmt.Sprintf("Location: %s, %s", city, state))
		}
		if phone != "" {
			details = append(details, "Phone: "+phone)
		}

		detail := strings.Join(details, " • ")
		if detail == "" {
			detail = "Oil & Gas Customer"
		}

		// Calculate relevance score
		score := calculateRelevanceScore(customer, query) + calculateRelevanceScore(contact, query)*0.8

		results = append(results, SearchResult{
			Type:   "customer",
			ID:     customerID,
			Title:  customer,
			Detail: detail,
			Score:  score,
			Data: map[string]interface{}{
				"customer_id": customerID,
				"customer":    customer,
				"contact":     contact,
				"city":        city,
				"state":       state,
				"phone":       phone,
				"email":       email,
			},
		})
	}

	return results, len(results)
}

// searchInventory searches within inventory table
func searchInventory(db *sql.DB, tenantID, query string, limit int) ([]SearchResult, int) {
	sqlQuery := `
		SELECT id, COALESCE(work_order, '') as work_order, customer, 
		       COALESCE(size, '') as size, COALESCE(grade, '') as grade, 
		       COALESCE(joints, 0) as joints, COALESCE(location, '') as location,
		       COALESCE(date_in::text, '') as date_in
		FROM store.inventory 
		WHERE tenant_id = $1 AND NOT deleted 
		AND (work_order ILIKE $2 OR customer ILIKE $2 OR size ILIKE $2 OR 
		     grade ILIKE $2 OR notes ILIKE $2 OR location ILIKE $2)
		ORDER BY 
			CASE 
				WHEN work_order ILIKE $3 THEN 1
				WHEN customer ILIKE $3 THEN 2
				WHEN size ILIKE $3 OR grade ILIKE $3 THEN 3
				ELSE 4
			END,
			date_in DESC 
		LIMIT $4
	`

	rows, err := db.Query(sqlQuery, tenantID, "%"+query+"%", query+"%", limit)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, joints int
		var workOrder, customer, size, grade, location, dateIn string
		
		if err := rows.Scan(&id, &workOrder, &customer, &size, &grade, &joints, &location, &dateIn); err != nil {
			continue
		}

		// Build title and detail
		title := fmt.Sprintf("Inventory #%d", id)
		if workOrder != "" {
			title = fmt.Sprintf("WO: %s", workOrder)
		}

		var details []string
		details = append(details, customer)
		if joints > 0 {
			details = append(details, fmt.Sprintf("%d joints", joints))
		}
		if size != "" && grade != "" {
			details = append(details, fmt.Sprintf("%s %s", size, grade))
		}
		if location != "" {
			details = append(details, fmt.Sprintf("@ %s", location))
		}

		detail := strings.Join(details, " • ")

		// Calculate relevance score
		score := calculateRelevanceScore(workOrder, query) + 
				calculateRelevanceScore(customer, query)*0.8 + 
				calculateRelevanceScore(size+" "+grade, query)*0.6

		results = append(results, SearchResult{
			Type:   "inventory",
			ID:     id,
			Title:  title,
			Detail: detail,
			Score:  score,
			Data: map[string]interface{}{
				"id":         id,
				"work_order": workOrder,
				"customer":   customer,
				"size":       size,
				"grade":      grade,
				"joints":     joints,
				"location":   location,
				"date_in":    dateIn,
			},
		})
	}

	return results, len(results)
}

// searchWorkOrders searches for work orders (aggregated from inventory)
func searchWorkOrders(db *sql.DB, tenantID, query string, limit int) ([]SearchResult, int) {
	sqlQuery := `
		SELECT work_order, customer, COUNT(*) as item_count, 
		       SUM(joints) as total_joints, MAX(date_in) as latest_date,
		       STRING_AGG(DISTINCT location, ', ' ORDER BY location) as locations
		FROM store.inventory 
		WHERE tenant_id = $1 AND NOT deleted AND work_order IS NOT NULL AND work_order != ''
		AND (work_order ILIKE $2 OR customer ILIKE $2)
		GROUP BY work_order, customer
		ORDER BY 
			CASE 
				WHEN work_order ILIKE $3 THEN 1
				WHEN customer ILIKE $3 THEN 2
				ELSE 3
			END,
			latest_date DESC
		LIMIT $4
	`

	rows, err := db.Query(sqlQuery, tenantID, "%"+query+"%", query+"%", limit)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var itemCount, totalJoints int
		var workOrder, customer, latestDate, locations string
		
		if err := rows.Scan(&workOrder, &customer, &itemCount, &totalJoints, &latestDate, &locations); err != nil {
			continue
		}

		// Build detail string
		var details []string
		details = append(details, customer)
		details = append(details, fmt.Sprintf("%d items", itemCount))
		if totalJoints > 0 {
			details = append(details, fmt.Sprintf("%d joints", totalJoints))
		}
		if latestDate != "" {
			details = append(details, fmt.Sprintf("Latest: %s", latestDate))
		}

		detail := strings.Join(details, " • ")

		// Calculate relevance score
		score := calculateRelevanceScore(workOrder, query) + calculateRelevanceScore(customer, query)*0.7

		results = append(results, SearchResult{
			Type:   "work_order",
			ID:     workOrder,
			Title:  fmt.Sprintf("WO: %s", workOrder),
			Detail: detail,
			Score:  score,
			Data: map[string]interface{}{
				"work_order":    workOrder,
				"customer":      customer,
				"item_count":    itemCount,
				"total_joints":  totalJoints,
				"latest_date":   latestDate,
				"locations":     locations,
			},
		})
	}

	return results, len(results)
}

// calculateRelevanceScore calculates a simple relevance score for search ranking
func calculateRelevanceScore(text, query string) float64 {
	if text == "" || query == "" {
		return 0.0
	}

	text = strings.ToLower(text)
	query = strings.ToLower(query)

	score := 0.0

	// Exact match gets highest score
	if text == query {
		score += 10.0
	}

	// Starts with query gets high score
	if strings.HasPrefix(text, query) {
		score += 5.0
	}

	// Contains query gets medium score
	if strings.Contains(text, query) {
		score += 2.0
	}

	// Word boundary match gets bonus
	if strings.Contains(" "+text+" ", " "+query+" ") {
		score += 3.0
	}

	return score
}

// sortResultsByRelevance sorts search results by relevance score
func sortResultsByRelevance(results []SearchResult, query string) {
	// Simple bubble sort by score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// Helper function to validate and sanitize search terms (imported from utils)
func sanitizeSearchTerm(term string) string {
	// Remove potentially problematic characters
	if len(term) > 100 {
		term = term[:100]
	}
	
	// Remove or escape dangerous SQL characters
	term = strings.ReplaceAll(term, "'", "''")  // Escape single quotes
	term = strings.ReplaceAll(term, ";", "")    // Remove semicolons
	term = strings.ReplaceAll(term, "--", "")   // Remove SQL comments
	term = strings.ReplaceAll(term, "/*", "")   // Remove SQL block comments
	term = strings.ReplaceAll(term, "*/", "")   // Remove SQL block comments
	
	return strings.TrimSpace(term)
}

// Helper function to validate pagination parameters (imported from utils)
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

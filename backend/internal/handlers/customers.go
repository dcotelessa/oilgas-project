// backend/internal/handlers/customers.go
// Customer API endpoints with tenant isolation
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/database"
)

type Customer struct {
	CustomerID      int    `json:"customer_id" db:"customer_id"`
	Customer        string `json:"customer" db:"customer"`
	BillingAddress  string `json:"billing_address" db:"billing_address"`
	BillingCity     string `json:"billing_city" db:"billing_city"`
	BillingState    string `json:"billing_state" db:"billing_state"`
	BillingZipcode  string `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         string `json:"contact" db:"contact"`
	Phone           string `json:"phone" db:"phone"`
	Fax             string `json:"fax" db:"fax"`
	Email           string `json:"email" db:"email"`
	TenantID        string `json:"tenant_id" db:"tenant_id"`
	ImportedAt      string `json:"imported_at,omitempty" db:"imported_at"`
	CreatedAt       string `json:"created_at" db:"created_at"`
}

// GetCustomers returns customers for the specified tenant
func GetCustomers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	search := strings.TrimSpace(c.Query("search"))
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")
	state := strings.TrimSpace(c.Query("state"))
	city := strings.TrimSpace(c.Query("city"))

	// Convert and validate pagination parameters
	limitInt, offsetInt := validatePagination(limit, offset)

	// Get tenant database connection
	db, err := database.GetTenantDB(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
			"details": err.Error(),
		})
		return
	}

	// Build query with search and filters
	query := `
		SELECT customer_id, customer, COALESCE(billing_address, '') as billing_address, 
		       COALESCE(billing_city, '') as billing_city, COALESCE(billing_state, '') as billing_state, 
		       COALESCE(billing_zipcode, '') as billing_zipcode, COALESCE(contact, '') as contact, 
		       COALESCE(phone, '') as phone, COALESCE(fax, '') as fax, 
		       COALESCE(email, '') as email, tenant_id, 
		       COALESCE(imported_at::text, '') as imported_at,
		       created_at::text as created_at
		FROM store.customers 
		WHERE tenant_id = $1 AND NOT deleted
	`
	args := []interface{}{tenantID}

	// Add search filter
	if search != "" {
		search = sanitizeSearchTerm(search)
		query += ` AND (
			customer ILIKE $` + strconv.Itoa(len(args)+1) + ` OR 
			contact ILIKE $` + strconv.Itoa(len(args)+1) + ` OR 
			billing_city ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			billing_state ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			phone ILIKE $` + strconv.Itoa(len(args)+1) + ` OR
			email ILIKE $` + strconv.Itoa(len(args)+1) + `
		)`
		args = append(args, "%"+search+"%")
	}

	// Add state filter
	if state != "" {
		query += ` AND billing_state ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+state+"%")
	}

	// Add city filter
	if city != "" {
		query += ` AND billing_city ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+city+"%")
	}

	query += ` ORDER BY customer LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
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

	var customers []Customer
	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.CustomerID, &customer.Customer, &customer.BillingAddress,
			&customer.BillingCity, &customer.BillingState, &customer.BillingZipcode,
			&customer.Contact, &customer.Phone, &customer.Fax, &customer.Email,
			&customer.TenantID, &customer.ImportedAt, &customer.CreatedAt,
		)
		if err != nil {
			continue // Skip problematic rows
		}
		customers = append(customers, customer)
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM store.customers WHERE tenant_id = $1 AND NOT deleted`
	countArgs := []interface{}{tenantID}
	
	if search != "" {
		countQuery += ` AND (customer ILIKE $2 OR contact ILIKE $2 OR billing_city ILIKE $2 OR billing_state ILIKE $2 OR phone ILIKE $2 OR email ILIKE $2)`
		countArgs = append(countArgs, "%"+search+"%")
	}
	
	if state != "" {
		countQuery += ` AND billing_state ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+state+"%")
	}
	
	if city != "" {
		countQuery += ` AND billing_city ILIKE $` + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, "%"+city+"%")
	}

	var totalCount int
	db.QueryRow(countQuery, countArgs...).Scan(&totalCount)

	// Calculate pagination info
	hasMore := (offsetInt + limitInt) < totalCount
	page := (offsetInt / limitInt) + 1

	c.JSON(http.StatusOK, gin.H{
		"tenant":     tenantID,
		"customers":  customers,
		"count":      len(customers),
		"total":      totalCount,
		"limit":      limitInt,
		"offset":     offsetInt,
		"page":       page,
		"has_more":   hasMore,
		"search":     search,
		"filters": gin.H{
			"state": state,
			"city":  city,
		},
	})
}

// GetCustomer returns a specific customer by ID
func GetCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	customerIDStr := c.Param("id")

	customerID, err := strconv.Atoi(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
			"message": "Customer ID must be a number",
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
		SELECT customer_id, customer, COALESCE(billing_address, '') as billing_address, 
		       COALESCE(billing_city, '') as billing_city, COALESCE(billing_state, '') as billing_state, 
		       COALESCE(billing_zipcode, '') as billing_zipcode, COALESCE(contact, '') as contact, 
		       COALESCE(phone, '') as phone, COALESCE(fax, '') as fax, 
		       COALESCE(email, '') as email, tenant_id,
		       COALESCE(imported_at::text, '') as imported_at,
		       created_at::text as created_at
		FROM store.customers 
		WHERE customer_id = $1 AND tenant_id = $2 AND NOT deleted
	`

	var customer Customer
	err = db.QueryRow(query, customerID, tenantID).Scan(
		&customer.CustomerID, &customer.Customer, &customer.BillingAddress,
		&customer.BillingCity, &customer.BillingState, &customer.BillingZipcode,
		&customer.Contact, &customer.Phone, &customer.Fax, &customer.Email,
		&customer.TenantID, &customer.ImportedAt, &customer.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
				"message": "No customer found with the specified ID for this tenant",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Query failed",
				"details": err.Error(),
			})
		}
		return
	}

	// Get related data for this customer
	relatedData := getCustomerRelatedData(db, customerID, tenantID)

	response := gin.H{
		"tenant":   tenantID,
		"customer": customer,
	}

	// Add related data if available
	if relatedData != nil {
		response["related"] = relatedData
	}

	c.JSON(http.StatusOK, response)
}

// getCustomerRelatedData gets inventory and work order counts for a customer
func getCustomerRelatedData(db *sql.DB, customerID int, tenantID string) map[string]interface{} {
	related := make(map[string]interface{})

	// Get inventory count
	var inventoryCount int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM store.inventory WHERE customer_id = $1 AND tenant_id = $2 AND NOT deleted",
		customerID, tenantID,
	).Scan(&inventoryCount)
	if err == nil {
		related["inventory_count"] = inventoryCount
	}

	// Get work order count
	var workOrderCount int
	err = db.QueryRow(
		"SELECT COUNT(DISTINCT work_order) FROM store.inventory WHERE customer_id = $1 AND tenant_id = $2 AND NOT deleted AND work_order IS NOT NULL AND work_order != ''",
		customerID, tenantID,
	).Scan(&workOrderCount)
	if err == nil {
		related["work_order_count"] = workOrderCount
	}

	// Get recent work orders
	workOrderQuery := `
		SELECT DISTINCT work_order, MAX(date_in) as latest_date
		FROM store.inventory 
		WHERE customer_id = $1 AND tenant_id = $2 AND NOT deleted 
		AND work_order IS NOT NULL AND work_order != ''
		GROUP BY work_order
		ORDER BY latest_date DESC
		LIMIT 5
	`
	
	rows, err := db.Query(workOrderQuery, customerID, tenantID)
	if err == nil {
		defer rows.Close()
		var recentWorkOrders []map[string]interface{}
		for rows.Next() {
			var workOrder, latestDate string
			if err := rows.Scan(&workOrder, &latestDate); err == nil {
				recentWorkOrders = append(recentWorkOrders, map[string]interface{}{
					"work_order":  workOrder,
					"latest_date": latestDate,
				})
			}
		}
		if len(recentWorkOrders) > 0 {
			related["recent_work_orders"] = recentWorkOrders
		}
	}

	if len(related) == 0 {
		return nil
	}

	return related
}

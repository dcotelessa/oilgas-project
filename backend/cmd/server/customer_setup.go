// backend/cmd/server/customer_setup.go
// Example integration of the enhanced customer domain

package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	
	"oilgas-backend/internal/customer"
)

// SetupCustomerDomain initializes the customer domain with all dependencies
func SetupCustomerDomain(db *sqlx.DB, router *gin.RouterGroup) {
	// Initialize repository
	customerRepo := customer.NewRepository(db)
	
	// Initialize service with business logic
	customerService := customer.NewService(customerRepo)
	
	// Initialize handlers
	customerHandler := customer.NewHandler(customerService)
	
	// Register routes
	customerHandler.RegisterRoutes(router)
	
	log.Println("Customer domain initialized successfully")
}

// SetupEnterpriseCustomerRoutes sets up admin-only routes
func SetupEnterpriseCustomerRoutes(db *sqlx.DB, router *gin.RouterGroup) {
	customerRepo := customer.NewRepository(db)
	customerService := customer.NewService(customerRepo)
	customerHandler := customer.NewHandler(customerService)
	
	// Register enterprise routes (admin only)
	customerHandler.RegisterEnterpriseRoutes(router)
}

// Example usage in main server setup
func ExampleServerSetup() {
	// Database connection (example)
	db, err := sqlx.Connect("postgres", "connection_string_here")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Gin router setup
	router := gin.Default()
	
	// API versioning
	v1 := router.Group("/api/v1")
	
	// Apply tenant middleware (example - implement based on your auth system)
	v1.Use(TenantMiddleware())
	
	// Setup customer domain
	SetupCustomerDomain(db, v1)
	
	// Admin routes (enterprise)
	adminRoutes := router.Group("/api/v1/admin")
	adminRoutes.Use(AdminMiddleware()) // Implement admin auth
	SetupEnterpriseCustomerRoutes(db, adminRoutes)
	
	// Start server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Example middleware functions (implement based on your auth system)
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant from JWT token, header, or session
		tenantID := extractTenantFromRequest(c)
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
			c.Abort()
			return
		}
		
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user has admin role
		if !isAdminUser(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Helper functions (implement based on your auth system)
func extractTenantFromRequest(c *gin.Context) string {
	// Implementation depends on your authentication system
	// Could be from JWT claims, header, session, etc.
	return c.GetHeader("X-Tenant-ID") // Example
}

func isAdminUser(c *gin.Context) bool {
	// Check user role from token/session
	// Implementation depends on your auth system
	return c.GetHeader("X-User-Role") == "admin" // Example
}

// Example API usage documentation
/*
Customer Domain API Endpoints:

## Basic Operations
GET    /api/v1/customers                    - List customers with filtering
POST   /api/v1/customers                    - Create new customer
GET    /api/v1/customers/:id                - Get customer details
PUT    /api/v1/customers/:id                - Update customer
DELETE /api/v1/customers/:id                - Delete customer (soft delete)

## Search & Analytics
GET    /api/v1/customers/search?q=term      - Search customers
GET    /api/v1/customers/:id/analytics      - Get customer analytics

## Enhanced Features
PUT    /api/v1/customers/:id/contacts       - Update customer contacts

## Enterprise (Admin Only)
GET    /api/v1/admin/enterprise/customers/summary     - Customer summary by tenant
POST   /api/v1/admin/enterprise/customers/batch       - Get customers by IDs

## Query Parameters for Filtering:
- q: Search query (name, contact, email)
- state: Filter by billing state
- has_orders: Filter customers with/without orders
- active: Filter active/inactive customers
- date_from/date_to: Filter by creation date
- sort_by: Sort field (name, created_at, work_orders, revenue, last_activity)
- sort_order: Sort direction (asc, desc)
- limit: Page size (default 20, max 100)
- offset: Pagination offset

## Example Requests:

### Create Customer
POST /api/v1/customers
{
  "customer": "Acme Oil Company",
  "billing_address": "123 Oil Field Road",
  "billing_city": "Houston",
  "billing_state": "TX",
  "billing_zipcode": "77001",
  "contact": "John Smith",
  "phone": "555-123-4567",
  "email": "john@acmeoil.com",
  "preferred_payment_terms": "Net 30",
  "default_po_required": true,
  "credit_limit": 50000.00
}

### Search Customers
GET /api/v1/customers/search?q=Acme

### Get Customers with Filtering
GET /api/v1/customers?state=TX&has_orders=true&sort_by=revenue&sort_order=desc&limit=10

### Update Customer Contacts
PUT /api/v1/customers/123/contacts
{
  "primary_contact": {
    "name": "John Smith",
    "title": "Operations Manager",
    "phone": "555-123-4567",
    "email": "john@acmeoil.com",
    "preferred": true
  },
  "billing_contact": {
    "name": "Jane Doe",
    "title": "Accounting Manager", 
    "phone": "555-123-4568",
    "email": "jane@acmeoil.com"
  }
}

### Get Customer Analytics
GET /api/v1/customers/123/analytics
Response:
{
  "analytics": {
    "customer_id": 123,
    "total_work_orders": 45,
    "completed_orders": 40,
    "pending_orders": 5,
    "total_revenue": 125000.00,
    "average_job_value": 2777.78,
    "last_order_date": "2025-07-15T10:30:00Z",
    "recent_work_orders": [...],
    "monthly_revenue": [...],
    "service_breakdown": [...]
  }
}
*/

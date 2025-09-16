// backend/cmd/bakersfield/main.go - Bakersfield location service
package main

import (
	"log"
	"os"
	"time"
	
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	
	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/customer"
	"oilgas-backend/internal/shared/database"
)

func main() {
	// Database configuration for Bakersfield location
	dbConfig := &database.Config{
		CentralDBURL: getDBURL("CENTRAL_AUTH_DB_URL", "DEV_CENTRAL_AUTH_DB_URL"),
		TenantDBs: map[string]string{
			"bakersfield": getDBURL("BAKERSFIELD_DB_URL", "DEV_BAKERSFIELD_DB_URL"),
		},
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  time.Hour,
	}
	
	// Initialize database manager
	dbManager, err := database.NewDatabaseManager(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database manager:", err)
	}
	defer dbManager.Close()
	
	// Initialize services
	authRepo := auth.NewRepository(dbManager.GetCentralDB())
	authSvc := auth.NewService(dbManager, authRepo)
	customerRepo := customer.NewRepository(dbManager)
	customerCache := customer.NewInMemoryCache(time.Hour)
	customerSvc := customer.NewService(customerRepo, authSvc, customerCache)
	customerHandlers := customer.NewHandlers(customerSvc)
	
	// Setup router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	// Apply middleware
	api := router.Group("/api/v1")
	api.Use(tenantMiddleware("bakersfield")) // Set Bakersfield as default tenant
	api.Use(authMiddleware(authSvc))         // Auth validation
	
	// Register routes
	customerHandlers.RegisterRoutes(api, authMiddleware(authSvc))
	
	log.Println("Bakersfield location service starting on :8081")
	log.Fatal(router.Run(":8081"))
}

func tenantMiddleware(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// For Bakersfield service, always use bakersfield tenant
		c.Set("tenant_id", defaultTenant)
		c.Next()
	}
}

func authMiddleware(authSvc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user from session/token
		userID := extractUserFromRequest(c) // Your implementation
		tenantID := c.GetString("tenant_id")
		
		// Validate user has access to this tenant
		hasAccess, err := authSvc.ValidateUserTenantAccess(c.Request.Context(), userID, tenantID)
		if err != nil || !hasAccess {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		
		// Set user context
		c.Set("user_id", userID)
		
		// Check if user is a customer contact and set customer filter
		customerID, err := authSvc.GetUserCustomerContext(c.Request.Context(), userID, tenantID)
		if err == nil && customerID != nil {
			c.Set("customer_filter", *customerID)
			c.Set("user_role", "customer_contact")
		}
		
		c.Next()
	}
}

func extractUserFromRequest(c *gin.Context) int {
	// Extract user ID from JWT token or session
	// This is your existing auth implementation
	return 1 // Placeholder
}

// getDBURL returns dev URL if available, otherwise production URL
func getDBURL(prodKey, devKey string) string {
	if devURL := os.Getenv(devKey); devURL != "" {
		return devURL
	}
	return os.Getenv(prodKey)
}
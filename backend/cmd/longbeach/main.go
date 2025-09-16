// backend/cmd/longbeach/main.go - Long Beach location service
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
	// Database configuration for Long Beach location
	dbConfig := &database.Config{
		CentralDBURL: os.Getenv("CENTRAL_AUTH_DB_URL"),
		TenantDBs: map[string]string{
			"longbeach": os.Getenv("LONGBEACH_DB_URL"),
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
	authSvc := auth.NewAuthService(dbManager)
	customerRepo := customer.NewRepository(dbManager)
	customerCache := customer.NewInMemoryCache(time.Hour)
	customerSvc := customer.NewService(customerRepo, authSvc, customerCache)
	customerHandlers := customer.NewHandlers(customerSvc)
	
	// Setup router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	// Apply middleware
	api := router.Group("/api/v1")
	api.Use(tenantMiddleware("longbeach")) // Set Long Beach as default tenant
	api.Use(authMiddleware(authSvc))       // Auth validation
	
	// Register routes
	customerHandlers.RegisterRoutes(api, authMiddleware(authSvc))
	
	log.Println("Long Beach location service starting on :8080")
	log.Fatal(router.Run(":8080"))
}

func tenantMiddleware(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// For Long Beach service, always use longbeach tenant
		c.Set("tenant_id", defaultTenant)
		c.Next()
	}
}

func authMiddleware(authSvc *auth.AuthService) gin.HandlerFunc {
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

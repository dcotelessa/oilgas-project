// backend/cmd/server/main.go
// Enhanced API server with tenant-aware routing and handlers
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
	// Import your new internal packages
	"oilgas-backend/internal/middleware"
	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router with middleware
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware for development
	router.Use(corsMiddleware())

	// Health check endpoint (no tenant required)
	router.GET("/health", healthCheckHandler)

	// Root API info endpoint (no tenant required)
	router.GET("/api", apiInfoHandler)

	// API v1 routes with tenant middleware
	setupAPIRoutes(router)

	// Admin endpoints (for tenant management)
	setupAdminRoutes(router)

	// Graceful shutdown handler
	defer database.CloseTenantConnections()

	// Start server
	startServer(router)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Tenant")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "oil-gas-inventory-api",
		"version":   "1.0.0",
		"tenant_support": true,
	})
}

func apiInfoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Oil & Gas Inventory API",
		"version": "1.0.0",
		"docs":    "Add X-Tenant header to access tenant-specific endpoints",
		"examples": gin.H{
			"customers": "curl -H 'X-Tenant: longbeach' '/api/v1/customers'",
			"search":    "curl -H 'X-Tenant: longbeach' '/api/v1/search?q=oil'",
			"inventory": "curl -H 'X-Tenant: longbeach' '/api/v1/inventory?customer_id=123'",
		},
		"tenant_commands": gin.H{
			"create": "migrator tenant-create <tenant_id>",
			"list":   "migrator tenant-list",
			"status": "migrator tenant-status <tenant_id>",
		},
	})
}

func setupAPIRoutes(router *gin.Engine) {
	// All v1 routes require tenant middleware
	v1 := router.Group("/api/v1")
	v1.Use(middleware.TenantMiddleware())
	{
		// Customer endpoints
		v1.GET("/customers", handlers.GetCustomers)
		v1.GET("/customers/:id", handlers.GetCustomer)

		// Inventory endpoints
		v1.GET("/inventory", handlers.GetInventory)
		v1.GET("/inventory/:id", handlers.GetInventoryItem)

		// Work order endpoints
		v1.GET("/work-orders", handlers.GetWorkOrders)
		v1.GET("/work-orders/:id", handlers.GetWorkOrder)

		// Search endpoints
		v1.GET("/search", handlers.GlobalSearch)

		// Tenant status endpoint
		v1.GET("/status", func(c *gin.Context) {
			tenantID := c.GetString("tenant_id")
			c.JSON(http.StatusOK, gin.H{
				"tenant":    tenantID,
				"database":  fmt.Sprintf("oilgas_%s", tenantID),
				"status":    "active",
				"timestamp": time.Now().Unix(),
				"endpoints": gin.H{
					"customers": fmt.Sprintf("/api/v1/customers"),
					"inventory": fmt.Sprintf("/api/v1/inventory"),
					"search":    fmt.Sprintf("/api/v1/search?q=<term>"),
					"work_orders": fmt.Sprintf("/api/v1/work-orders"),
				},
			})
		})
	}
}

func setupAdminRoutes(router *gin.Engine) {
	// Admin endpoints (no tenant required, but could add auth later)
	admin := router.Group("/admin")
	{
		// Tenant management
		admin.GET("/tenants", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Tenant management endpoint",
				"note":    "Use migrator commands for tenant operations",
				"commands": gin.H{
					"list":   "migrator tenant-list",
					"create": "migrator tenant-create <id>",
					"status": "migrator tenant-status <id>",
				},
			})
		})

		// System health
		admin.GET("/health", func(c *gin.Context) {
			// Get connection pool stats
			poolStats := database.GetConnectionPoolManager().GetAllConnectionStats()
			
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"timestamp": time.Now().Unix(),
				"connection_pools": poolStats,
				"active_tenants": len(poolStats),
			})
		})

		// Connection monitoring
		admin.GET("/connections", func(c *gin.Context) {
			database.GetConnectionPoolManager().MonitorConnectionPools()
			
			c.JSON(http.StatusOK, gin.H{
				"message": "Connection monitoring logged",
				"timestamp": time.Now().Unix(),
			})
		})
	}
}

func startServer(router *gin.Engine) {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("🚀 Starting Oil & Gas Inventory API server on port %s\n", port)
	fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔌 API info: http://localhost:%s/api\n", port)
	fmt.Printf("🏢 Tenant API: http://localhost:%s/api/v1/customers (requires X-Tenant header)\n", port)
	fmt.Printf("🔧 Admin panel: http://localhost:%s/admin/health\n", port)
	fmt.Println()
	fmt.Println("📖 Example tenant API calls:")
	fmt.Printf("  curl -H 'X-Tenant: longbeach' 'http://localhost:%s/api/v1/customers'\n", port)
	fmt.Printf("  curl -H 'X-Tenant: longbeach' 'http://localhost:%s/api/v1/search?q=oil'\n", port)
	fmt.Printf("  curl -H 'X-Tenant: longbeach' 'http://localhost:%s/api/v1/inventory?customer_id=123'\n", port)
	fmt.Println()
	fmt.Println("🏗️  Create tenants with:")
	fmt.Println("  go run migrator.go tenant-create <tenant_id>")
	fmt.Println("  go run migrator.go tenant-list")

	log.Fatal(http.ListenAndServe(":"+port, router))
}

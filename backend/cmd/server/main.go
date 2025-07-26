// backend/cmd/server/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	
	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/middleware"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/tenant"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Set Gin mode
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize auth database
	authDB, err := initAuthDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize auth database: %v", err)
	}
	defer authDB.Close()

	// Initialize tenant database manager
	baseConnStr := getBaseConnectionString()
	dbManager := tenant.NewDatabaseManager(baseConnStr)
	defer dbManager.CloseAll()

	// Initialize repositories
	authRepo := repository.NewAuthRepository(authDB)

	// Initialize services
	authService := services.NewAuthService(authRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	customerHandler := handlers.NewCustomerHandler()
	inventoryHandler := handlers.NewInventoryHandler()
	receivedHandler := handlers.NewReceivedHandler()

	// Initialize middleware
	tenantMiddleware := middleware.NewTenantMiddleware(authRepo, dbManager)

	// Create Gin router
	router := gin.New()
	
	// Apply global middleware
	router.Use(middleware.Logging())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "oil-gas-inventory-api",
			"version":   "1.0.0",
		})
	})

	// Auth endpoints (no tenant required)
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/tenants", authHandler.GetUserTenants)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// API v1 routes (tenant required)
	v1 := router.Group("/api/v1")
	v1.Use(tenantMiddleware.RequireAuth())
	{
		// Status endpoint
		v1.GET("/status", func(c *gin.Context) {
			tenant := middleware.GetTenant(c)
			c.JSON(http.StatusOK, gin.H{
				"message": "Oil & Gas Inventory API",
				"status":  "running",
				"tenant":  tenant.Code,
				"env":     os.Getenv("APP_ENV"),
			})
		})

		// Customer endpoints
		customers := v1.Group("/customers")
		{
			customers.GET("", customerHandler.GetCustomers)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.GET("/search", customerHandler.SearchCustomers)
		}

		// Inventory endpoints  
		inventory := v1.Group("/inventory")
		{
			inventory.GET("", inventoryHandler.GetInventory)
			inventory.GET("/:id", inventoryHandler.GetInventoryItem)
			inventory.GET("/search", inventoryHandler.SearchInventory)
		}

		// Received endpoints
		received := v1.Group("/received")
		{
			received.GET("", receivedHandler.GetReceived)
			received.GET("/:id", receivedHandler.GetReceivedItem)
		}

		// Reference data endpoints
		v1.GET("/grades", handleGetGrades)
		v1.GET("/sizes", handleGetSizes)
	}

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("🚀 Starting Oil & Gas Inventory API server on port %s\n", port)
	fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔌 API base: http://localhost:%s/api/v1\n", port)
	fmt.Printf("🔐 Auth endpoints: http://localhost:%s/api/v1/auth\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}

func initAuthDatabase() (*sql.DB, error) {
	authDatabaseURL := os.Getenv("AUTH_DATABASE_URL")
	if authDatabaseURL == "" {
		authDatabaseURL = "postgres://user:password@localhost/oilgas_auth?sslmode=disable"
	}

	db, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping auth database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	return db, nil
}

func getBaseConnectionString() string {
	baseConn := os.Getenv("TENANT_DATABASE_BASE_URL")
	if baseConn == "" {
		baseConn = "postgres://user:password@localhost"
	}
	return baseConn
}

// Reference data handlers (will be moved to proper handlers later)
func handleGetGrades(c *gin.Context) {
	tenantDB := middleware.GetTenantDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	rows, err := tenantDB.Query("SELECT grade, description FROM store.grade ORDER BY grade")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch grades"})
		return
	}
	defer rows.Close()

	var grades []map[string]string
	for rows.Next() {
		var grade, description string
		if err := rows.Scan(&grade, &description); err != nil {
			continue
		}
		grades = append(grades, map[string]string{
			"grade":       grade,
			"description": description,
		})
	}

	c.JSON(http.StatusOK, grades)
}

func handleGetSizes(c *gin.Context) {
	tenantDB := middleware.GetTenantDB(c)
	if tenantDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	rows, err := tenantDB.Query("SELECT size_id, size, description FROM store.sizes ORDER BY size")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sizes"})
		return
	}
	defer rows.Close()

	var sizes []map[string]interface{}
	for rows.Next() {
		var sizeID int
		var size, description string
		if err := rows.Scan(&sizeID, &size, &description); err != nil {
			continue
		}
		sizes = append(sizes, map[string]interface{}{
			"size_id":     sizeID,
			"size":        size,
			"description": description,
		})
	}

	c.JSON(http.StatusOK, sizes)
}

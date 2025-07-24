// cmd/server/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/repository/postgres"
)

type ServerDependencies struct {
	CustomerRepo  repository.CustomerRepository
	InventoryRepo repository.InventoryRepository
}

type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Load environment variables from project root
	envFiles := []string{
		"../.env.local",  // Project root .env.local
		"../.env",        // Project root .env  
		".env.local",     // Backend .env.local
		".env",           // Backend .env
	}
	
	var envLoaded bool
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			log.Printf("Loaded environment from: %s", envFile)
			envLoaded = true
			break
		}
	}
	
	if !envLoaded {
		log.Printf("Warning: No .env file found, using system environment variables")
	}

	// Database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}
	defer db.Close()

	// Repository dependencies - use postgres package constructors
	customerRepo := postgres.NewCustomerRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	
	deps := &ServerDependencies{
		CustomerRepo:  customerRepo,
		InventoryRepo: inventoryRepo,
	}

	// Gin setup
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Routes
	setupRoutes(router, deps)

	// Start server
	port := getPort()
	fmt.Printf("🚀 Oil & Gas Inventory API server starting on port %s\n", port)
	fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔌 API base: http://localhost:%s/api/v1\n", port)
	fmt.Printf("📊 Sample endpoints:\n")
	fmt.Printf("   GET /api/v1/customers\n")
	fmt.Printf("   GET /api/v1/customers/search?q=oil\n")
	fmt.Printf("   GET /api/v1/inventory\n")

	log.Fatal(http.ListenAndServe(":"+port, router))
}

func setupDatabase() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Try to construct from individual components if DATABASE_URL not set
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "5433")
		user := getEnvOrDefault("DB_USER", "postgres")
		password := getEnvOrDefault("DB_PASSWORD", "postgres123")  // Match your .env.local
		dbname := getEnvOrDefault("DB_NAME", "oilgas_inventory_local")
		sslmode := getEnvOrDefault("DB_SSLMODE", "disable")
		
		databaseURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
		
		log.Printf("Constructed DATABASE_URL for: %s", dbname)
	} else {
		log.Printf("Using DATABASE_URL: %s", maskPassword(databaseURL))
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connection successful")

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Helper to mask password in logs
func maskPassword(dbURL string) string {
	// Simple password masking for logs
	parts := strings.Split(dbURL, "@")
	if len(parts) >= 2 {
		userPart := strings.Split(parts[0], ":")
		if len(userPart) >= 3 {
			userPart[len(userPart)-1] = "***"
			parts[0] = strings.Join(userPart, ":")
		}
		return strings.Join(parts, "@")
	}
	return dbURL
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func setupRoutes(router *gin.Engine, deps *ServerDependencies) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "oil-gas-inventory-api",
			"version":   "1.0.0",
			"database":  "connected",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/status", handleStatus)
		
		// Customer routes
		v1.GET("/customers", handleGetCustomers(deps))
		v1.GET("/customers/:id", handleGetCustomer(deps))
		v1.GET("/customers/search", handleSearchCustomers(deps))
		v1.POST("/customers", handleCreateCustomer(deps))
		
		// Inventory routes
		v1.GET("/inventory", handleGetInventory(deps))
		v1.GET("/inventory/:id", handleGetInventoryItem(deps))
		v1.GET("/inventory/search", handleSearchInventory(deps))
		
		// Simple reference data (hardcoded for now)
		v1.GET("/grades", handleGetGrades)
		v1.GET("/sizes", handleGetSizes)
	}
}

func handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Oil & Gas Inventory API",
		"status":  "running",
		"env":     os.Getenv("APP_ENV"),
		"time":    time.Now().Format(time.RFC3339),
	})
}

func handleGetCustomers(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		customers, err := deps.CustomerRepo.GetAll(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Error:   "database_error",
				Code:    500,
				Message: "Failed to retrieve customers",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"customers": customers,
			"count":     len(customers),
		})
	}
}

func handleGetCustomer(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "invalid_id",
				Code:    400,
				Message: "Customer ID must be a number",
			})
			return
		}

		customer, err := deps.CustomerRepo.GetByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, APIError{
				Error:   "customer_not_found",
				Code:    404,
				Message: "Customer not found",
			})
			return
		}

		c.JSON(http.StatusOK, customer)
	}
}

func handleSearchCustomers(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "missing_query",
				Code:    400,
				Message: "Query parameter 'q' is required",
			})
			return
		}

		customers, err := deps.CustomerRepo.Search(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Error:   "search_error",
				Code:    500,
				Message: "Failed to search customers",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"customers": customers,
			"count":     len(customers),
			"query":     query,
		})
	}
}

func handleCreateCustomer(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var customer repository.Customer
		if err := c.ShouldBindJSON(&customer); err != nil {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "invalid_json",
				Code:    400,
				Message: "Invalid customer data: " + err.Error(),
			})
			return
		}

		// Basic validation - use the correct field name
		if customer.Customer == "" {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "missing_name",
				Code:    400,
				Message: "Customer name is required",
			})
			return
		}

		err := deps.CustomerRepo.Create(ctx, &customer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Error:   "create_failed",
				Code:    500,
				Message: "Failed to create customer",
			})
			return
		}

		c.JSON(http.StatusCreated, customer)
	}
}

func handleGetInventory(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// Use empty filters for GetAll
		filters := repository.InventoryFilters{}
		inventory, err := deps.InventoryRepo.GetAll(ctx, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Error:   "database_error",
				Code:    500,
				Message: "Failed to retrieve inventory",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"inventory": inventory,
			"count":     len(inventory),
		})
	}
}

func handleGetInventoryItem(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "invalid_id",
				Code:    400,
				Message: "Inventory ID must be a number",
			})
			return
		}

		item, err := deps.InventoryRepo.GetByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusNotFound, APIError{
				Error:   "item_not_found",
				Code:    404,
				Message: "Inventory item not found",
			})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

func handleSearchInventory(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, APIError{
				Error:   "missing_query",
				Code:    400,
				Message: "Query parameter 'q' is required",
			})
			return
		}

		inventory, err := deps.InventoryRepo.Search(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Error:   "search_error",
				Code:    500,
				Message: "Failed to search inventory",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"inventory": inventory,
			"count":     len(inventory),
			"query":     query,
		})
	}
}

// Simple hardcoded reference data handlers (since we don't have Grade/Size repos yet)
func handleGetGrades(c *gin.Context) {
	grades := []map[string]string{
		{"grade": "J55", "description": "Standard grade steel casing"},
		{"grade": "L80", "description": "Higher strength grade"},
		{"grade": "N80", "description": "Medium strength grade"},
		{"grade": "P110", "description": "Premium performance grade"},
	}
	c.JSON(http.StatusOK, gin.H{
		"grades": grades,
		"count":  len(grades),
	})
}

func handleGetSizes(c *gin.Context) {
	sizes := []map[string]string{
		{"size": "5 1/2\"", "description": "5.5 inch diameter"},
		{"size": "7\"", "description": "7 inch diameter"},
		{"size": "9 5/8\"", "description": "9.625 inch diameter"},
		{"size": "13 3/8\"", "description": "13.375 inch diameter"},
	}
	c.JSON(http.StatusOK, gin.H{
		"sizes": sizes,
		"count": len(sizes),
	})
}

func getPort() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}
	return port
}

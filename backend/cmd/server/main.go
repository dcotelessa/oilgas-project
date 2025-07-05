// backend/cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/utils"
)

func main() {
	// Load environment variables - try root .env.local first
	if err := godotenv.Load("../.env.local"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load(".env.local"); err != nil {
				if err := godotenv.Load(".env"); err != nil {
					log.Println("No .env file found")
				}
			}
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := utils.NewDBConnection(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize cache
	cacheConfig := cache.Config{
		TTL:             parseEnvDuration("CACHE_TTL", "5m"),
		CleanupInterval: parseEnvDuration("CACHE_CLEANUP_INTERVAL", "10m"),
		MaxSize:         parseEnvInt("CACHE_MAX_SIZE", 1000),
	}
	cacheService := cache.New(cacheConfig)

	// Initialize repositories
	repos := repository.New(db)

	// Initialize services using your existing structure
	servicesContainer := services.New(repos, cacheService)

	// Initialize handlers using your existing structure
	handlersContainer := handlers.New(servicesContainer)

	// Initialize Gin router
	r := gin.Default()

	// Configure trusted proxies
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// CORS middleware
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"service":     "oil-gas-inventory",
			"version":     "1.0.0",
			"environment": os.Getenv("APP_ENV"),
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Dashboard routes
		api.GET("/dashboard", handlersContainer.GetDashboard)

		// Customer routes
		customers := api.Group("/customers")
		{
			customers.GET("", handlersContainer.GetCustomers)
			customers.GET("/:customerID", handlersContainer.GetCustomer)
			customers.POST("", handlersContainer.CreateCustomer)
			customers.PUT("/:customerID", handlersContainer.UpdateCustomer)
			customers.DELETE("/:customerID", handlersContainer.DeleteCustomer)
			// Customer workflow as a sub-route
			customers.GET("/:customerID/workflow", handlersContainer.GetCustomerWorkflow)
		}

		// Inventory routes
		inventory := api.Group("/inventory")
		{
			inventory.GET("", handlersContainer.GetInventory)
			inventory.GET("/:id", handlersContainer.GetInventoryItem)
			inventory.POST("", handlersContainer.CreateInventoryItem)
			inventory.PUT("/:id", handlersContainer.UpdateInventoryItem)
			inventory.DELETE("/:id", handlersContainer.DeleteInventoryItem)
			inventory.GET("/summary", handlersContainer.GetInventorySummary)
		}

		// Search routes
		search := api.Group("/search")
		{
			search.GET("/inventory", handlersContainer.SearchInventory)
		}

		// Analytics routes
		analytics := api.Group("/analytics")
		{
			analytics.GET("/customers/activity", handlersContainer.GetCustomerActivity)
			analytics.GET("/customers/top", handlersContainer.GetTopCustomers)
			analytics.GET("/grades", handlersContainer.GetGradeAnalytics)
		}

		// Reference data
		api.GET("/grades", handlersContainer.GetGrades)

		// System routes
		system := api.Group("/system")
		{
			system.GET("/cache/stats", handlersContainer.GetCacheStats)
			system.POST("/cache/clear", handlersContainer.ClearCache)
		}
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("APP_ENV"))
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("📊 API Base: http://localhost:%s/api/v1", port)
	log.Fatal(r.Run(":" + port))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func parseEnvDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func parseEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

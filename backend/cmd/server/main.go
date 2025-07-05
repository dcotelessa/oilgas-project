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
	// Load environment variables from project root or local
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
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

	// Initialize repositories using your existing structure
	repos := repository.New(db)

	// Initialize services
	workflowService := services.NewWorkflowService(repos, cacheService)

	// Initialize handlers
	inventoryHandler := handlers.NewInventoryHandler(workflowService)

	// Initialize Gin router
	r := gin.Default()

	// Configure trusted proxies (fix the warning)
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
		// Dashboard and workflow endpoints
		api.GET("/dashboard", inventoryHandler.GetDashboard)
		api.GET("/customers/:customerID/workflow", inventoryHandler.GetCustomerWorkflow)
		api.GET("/search", inventoryHandler.SearchInventory)

		// Legacy endpoints for compatibility
		api.GET("/grades", getGrades)
		api.GET("/customers", getCustomers)
		api.GET("/inventory", getInventory)
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("APP_ENV"))
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("📊 Dashboard: http://localhost:%s/api/v1/dashboard", port)
	log.Printf("🔍 Search: http://localhost:%s/api/v1/search?q=<query>", port)
	log.Fatal(r.Run(":" + port))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// In development, allow localhost origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		
		// Check if origin is allowed
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

// Utility functions for environment parsing
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

// Legacy placeholder handlers for backward compatibility
func getGrades(c *gin.Context) {
	grades := []string{"J55", "JZ55", "L80", "N80", "P105", "P110"}
	utils.SuccessResponse(c, http.StatusOK, gin.H{"grades": grades})
}

func getCustomers(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, gin.H{"customers": []interface{}{}})
}

func getInventory(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, gin.H{"inventory": []interface{}{}})
}

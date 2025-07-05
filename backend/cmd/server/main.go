// backend/cmd/server/main.go - Updated to use actual repository structure
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
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	// Set Gin mode
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

	// Initialize cache with configuration
	cacheConfig := cache.Config{
		TTL:             parseEnvDuration("CACHE_TTL", "5m"),
		CleanupInterval: parseEnvDuration("CACHE_CLEANUP_INTERVAL", "10m"),
		MaxSize:         parseEnvInt("CACHE_MAX_SIZE", 1000),
	}
	appCache := cache.New(cacheConfig)

	repos := repository.New(db)

	services := services.New(repos, appCache)

	handlers := handlers.New(services)

	r := gin.Default()

	// Configure trusted proxies
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// CORS middleware
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		// Optionally check repository health
		healthStatus := gin.H{
			"status":      "healthy",
			"service":     "oil-gas-inventory",
			"version":     "1.0.0",
			"environment": os.Getenv("APP_ENV"),
			"timestamp":   time.Now().UTC(),
		}

		// Check database connectivity
		if err := repos.HealthCheck(c.Request.Context()); err != nil {
			healthStatus["status"] = "unhealthy"
			healthStatus["database_error"] = err.Error()
			c.JSON(http.StatusServiceUnavailable, healthStatus)
			return
		}

		// Check cache stats
		cacheStats := appCache.GetStats()
		healthStatus["cache"] = gin.H{
			"items":     cacheStats.Items,
			"hit_ratio": appCache.GetHitRatio(),
		}

		c.JSON(http.StatusOK, healthStatus)
	})

	// API routes using updated handler structure
	api := r.Group("/api/v1")
	{
		// Register all routes through the handlers
		handlers.RegisterRoutes(api)

		// Legacy compatibility endpoints
		api.GET("/dashboard", getLegacyDashboard(services))
		api.GET("/search", getLegacySearch(services))
	}

	// Start server with graceful shutdown support
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("APP_ENV"))
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("📊 API: http://localhost:%s/api/v1", port)
	
	// Log available repositories
	log.Printf("📦 Loaded repositories: %v", getRepositoryNames(repos))
	
	log.Fatal(r.Run(":" + port))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow specific origins in development
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

// Legacy compatibility functions
func getLegacyDashboard(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := services.Analytics.GetDashboardStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get dashboard stats",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"dashboard": stats,
			"message":   "Use /api/v1/analytics/dashboard for new endpoint",
		})
	}
}

func getLegacySearch(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Search query 'q' is required",
			})
			return
		}

		results, err := services.Search.GlobalSearch(c.Request.Context(), query, map[string]interface{}{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Search failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"message": "Use /api/v1/search for new endpoint",
		})
	}
}

// Utility functions
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

func getRepositoryNames(repos *repository.Repositories) []string {
	repoMap := repos.GetAll()
	names := make([]string, 0, len(repoMap))
	for name := range repoMap {
		names = append(names, name)
	}
	return names
}

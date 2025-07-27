// backend/cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/middleware"
	"oilgas-backend/pkg/cache"
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

	// Initialize database
	pool, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	// Initialize in-memory cache (no Redis dependency)
	memCache := cache.NewWithDefaultExpiration(10*time.Minute, 5*time.Minute)

	// Initialize components
	sessionManager := auth.NewTenantSessionManager(pool, memCache)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(sessionManager)
	customerHandler := handlers.NewCustomerHandler()
	inventoryHandler := handlers.NewInventoryHandler()

	// Setup router
	router := setupRouter(authHandler, customerHandler, inventoryHandler)

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("🚀 Oil & Gas Inventory API Server Starting\\n")
	fmt.Printf("📋 Health check: http://localhost:%s/health\\n", port)
	fmt.Printf("🔌 API base: http://localhost:%s/api/v1\\n", port)
	fmt.Printf("🔐 Auth: Session-based with tenant isolation\\n")
	fmt.Printf("💾 Cache: In-memory (%d items)\\n", memCache.Count())

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeDatabase() (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("Warning: DATABASE_URL not set, some features will be limited")
		return nil, nil // Allow server to start without DB for testing
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Optimized connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connected with %d max connections", config.MaxConns)
	return pool, nil
}

func setupRouter(authHandler *handlers.AuthHandler, customerHandler *handlers.CustomerHandler, inventoryHandler *handlers.InventoryHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Health check (no tenant required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"timestamp":  time.Now().Unix(),
			"service":    "oil-gas-inventory-api",
			"version":    "1.0.0",
			"features": []string{
				"tenant-isolation",
				"session-auth",
				"in-memory-cache",
			},
		})
	})

	// API info (no tenant required)
	router.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Oil & Gas Inventory API",
			"version": "1.0.0",
			"docs":    "Add X-Tenant header to access tenant-specific endpoints",
			"examples": gin.H{
				"customers": "curl -H 'X-Tenant: demo' '/api/v1/customers'",
				"inventory": "curl -H 'X-Tenant: demo' '/api/v1/inventory'",
			},
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	
	// Authentication endpoints (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected endpoints (require tenant)
	protected := v1.Group("")
	protected.Use(middleware.TenantMiddleware())
	{
		// Auth endpoints
		protected.GET("/auth/me", authHandler.Me)
		
		// Business endpoints
		protected.GET("/customers", customerHandler.GetCustomers)
		protected.GET("/customers/:id", customerHandler.GetCustomer)
		protected.GET("/inventory", inventoryHandler.GetInventory)
		protected.GET("/inventory/:id", inventoryHandler.GetInventoryItem)
		
		// Status endpoint
		protected.GET("/status", func(c *gin.Context) {
			tenantID := c.GetString("tenant_id")
			c.JSON(http.StatusOK, gin.H{
				"tenant":    tenantID,
				"database":  fmt.Sprintf("oilgas_%s", tenantID),
				"status":    "active",
				"timestamp": time.Now().Unix(),
				"endpoints": gin.H{
					"customers": "/api/v1/customers",
					"inventory": "/api/v1/inventory",
				},
			})
		})
	}

	return router
}

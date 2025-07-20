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

	"github.com/dcotelessa/oilgas-project/internal/auth"
	"github.com/dcotelessa/oilgas-project/internal/handlers"
	"github.com/dcotelessa/oilgas-project/pkg/cache"
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

	// Initialize database with proven settings
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

	// Setup router
	router := setupRouter(sessionManager, authHandler)

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("🚀 Starting Oil & Gas Inventory API server on port %s\n", port)
	fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔌 API base: http://localhost:%s/api/v1\n", port)
	fmt.Printf("🔐 Authentication: Session-based with tenant isolation\n")

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeDatabase() (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Proven connection pool settings for tenant isolation
	config.MaxConns = 25
	config.MinConns = 10
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

func setupRouter(sessionManager *auth.TenantSessionManager, authHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "oil-gas-inventory-api",
			"version":   "1.0.0-phase3",
			"features": []string{
				"tenant-isolation",
				"session-auth",
				"row-level-security",
				"in-memory-cache",
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

	// Protected endpoints
	protected := v1.Group("")
	protected.Use(sessionManager.RequireAuth())
	{
		protected.GET("/auth/me", authHandler.Me)
		
		// Placeholder for Phase 4 endpoints
		protected.GET("/customers", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Customers endpoint - Phase 4 implementation",
				"tenant_id": c.GetString("tenant_id"),
				"user_role": c.GetString("user_role"),
			})
		})
		
		protected.GET("/inventory", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Inventory endpoint - Phase 4 implementation",
				"tenant_id": c.GetString("tenant_id"),
				"user_role": c.GetString("user_role"),
			})
		})
	}

	return router
}

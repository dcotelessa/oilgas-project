package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"oilgas-backend/internal/handlers"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/cache"
)

type Server struct {
	router   *gin.Engine
	db       *pgxpool.Pool
	cache    *cache.Cache
	handlers *handlers.Handlers
}

var startTime = time.Now()

func main() {
	log.Println("🚀 Starting Oil & Gas Inventory System...")
	
	// Load environment variables
	loadEnvFiles()

	// Initialize server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer server.Cleanup()

	// Start server with graceful shutdown
	server.Start()
}

func loadEnvFiles() {
	// Try multiple env file locations
	envFiles := []string{
		".env",
		".env.local", 
		"../.env",
		"../.env.local",
	}

	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Printf("📄 Loaded environment from %s", file)
			return
		}
	}
	log.Println("⚠️  No .env file found, using system environment")
}

func NewServer() (*Server, error) {
	// Database connection
	db, err := initDatabase()
	if err != nil {
		return nil, fmt.Errorf("database init: %w", err)
	}

	// Cache initialization
	cacheConfig := cache.Config{
		TTL:             getDurationEnv("CACHE_TTL", 5*time.Minute),
		CleanupInterval: getDurationEnv("CACHE_CLEANUP_INTERVAL", 10*time.Minute),
		MaxSize:         getIntEnv("CACHE_MAX_SIZE", 1000),
	}
	appCache := cache.New(cacheConfig)
	log.Printf("💾 Cache initialized (TTL: %v, Max: %d)", cacheConfig.TTL, cacheConfig.MaxSize)

	// Repository layer
	repos := repository.New(db)

	// Service layer
	services := services.New(repos, appCache)

	// Handler layer
	handlers := handlers.New(services)

	// Router setup
	router := setupRouter(handlers)

	return &Server{
		router:   router,
		db:       db,
		cache:    appCache,
		handlers: handlers,
	}, nil
}

func initDatabase() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Connection pool configuration
	config.MaxConns = 30
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Printf("🐘 Database connected successfully (pool: %d-%d)", config.MinConns, config.MaxConns)
	return db, nil
}

func setupRouter(h *handlers.Handlers) *gin.Engine {
	// Set Gin mode
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Configure trusted proxies
	r.SetTrustedProxies([]string{
		"127.0.0.1", "::1",
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	})

	// Health check
	r.GET("/health", healthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		// Customer routes
		customers := api.Group("/customers")
		{
			customers.GET("", h.GetCustomers)
			customers.GET("/:id", h.GetCustomer)
			customers.POST("", h.CreateCustomer)
			customers.PUT("/:id", h.UpdateCustomer)
			customers.DELETE("/:id", h.DeleteCustomer)
		}

		// Inventory routes
		inventory := api.Group("/inventory")
		{
			inventory.GET("", h.GetInventory)
			inventory.GET("/:id", h.GetInventoryItem)
			inventory.POST("", h.CreateInventoryItem)
			inventory.PUT("/:id", h.UpdateInventoryItem)
			inventory.DELETE("/:id", h.DeleteInventoryItem)
		}

		// Grade routes
		grades := api.Group("/grades")
		{
			grades.GET("", h.GetGrades)
			grades.POST("", h.CreateGrade)
			grades.DELETE("/:grade", h.DeleteGrade)
		}

		// Received routes (incoming inventory)
		received := api.Group("/received")
		{
			received.GET("", h.GetReceived)
			received.GET("/:id", h.GetReceivedItem)
			received.POST("", h.CreateReceivedItem)
			received.PUT("/:id", h.UpdateReceivedItem)
			received.DELETE("/:id", h.DeleteReceivedItem)
		}

		// Fletcher routes (threading/inspection)
		fletcher := api.Group("/fletcher")
		{
			fletcher.GET("", h.GetFletcherItems)
			fletcher.GET("/:id", h.GetFletcherItem)
			fletcher.POST("", h.CreateFletcherItem)
			fletcher.PUT("/:id", h.UpdateFletcherItem)
			fletcher.DELETE("/:id", h.DeleteFletcherItem)
		}

		// Bakeout routes
		bakeout := api.Group("/bakeout")
		{
			bakeout.GET("", h.GetBakeoutItems)
			bakeout.GET("/:id", h.GetBakeoutItem)
			bakeout.POST("", h.CreateBakeoutItem)
			bakeout.PUT("/:id", h.UpdateBakeoutItem)
			bakeout.DELETE("/:id", h.DeleteBakeoutItem)
		}

		// Search and reporting routes
		search := api.Group("/search")
		{
			search.GET("/customers", h.SearchCustomers)
			search.GET("/inventory", h.SearchInventory)
			search.GET("/global", h.GlobalSearch)
		}

		// Analytics routes
		analytics := api.Group("/analytics")
		{
			analytics.GET("/dashboard", h.GetDashboardStats)
			analytics.GET("/inventory-summary", h.GetInventorySummary)
			analytics.GET("/customer-activity", h.GetCustomerActivity)
		}

		// System routes
		system := api.Group("/system")
		{
			system.GET("/cache/stats", h.GetCacheStats)
			system.POST("/cache/clear", h.ClearCache)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		
		// In production, get from env
		if os.Getenv("APP_ENV") == "production" {
			if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
				allowedOrigins = strings.Split(corsOrigins, ",")
				// Trim whitespace
				for i := range allowedOrigins {
					allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
				}
			}
		}
		
		// Check if origin is allowed
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
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

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"service":     "oil-gas-inventory",
		"version":     "1.0.0",
		"environment": os.Getenv("APP_ENV"),
		"timestamp":   time.Now().Format(time.RFC3339),
		"uptime":      time.Since(startTime).String(),
	})
}

func (s *Server) Start() {
	port := getEnv("APP_PORT", "8000")
	
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		log.Printf("🌍 Environment: %s", getEnv("APP_ENV", "development"))
		log.Printf("🔗 Health check: http://localhost:%s/health", port)
		log.Printf("📚 API base: http://localhost:%s/api/v1", port)
		log.Printf("📊 Cache stats: http://localhost:%s/api/v1/system/cache/stats", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

func (s *Server) Cleanup() {
	log.Println("🧹 Cleaning up resources...")
	
	if s.cache != nil {
		s.cache.Close()
		log.Println("💾 Cache closed")
	}
	
	if s.db != nil {
		s.db.Close()
		log.Println("🐘 Database connection closed")
	}
}

// Utility functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

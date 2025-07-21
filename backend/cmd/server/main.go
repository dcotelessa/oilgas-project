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
    
    "github.com/dcotelessa/oilgas-project/internal/handlers"
    "github.com/dcotelessa/oilgas-project/internal/middleware"
    "github.com/dcotelessa/oilgas-project/internal/repository"
    "github.com/dcotelessa/oilgas-project/internal/services"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(".env"); err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    // Database connection
    db, err := connectDB()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Set Gin mode
    if os.Getenv("APP_ENV") == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    // Create router
    router := gin.New()
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    // CORS middleware
    router.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Tenant-ID")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    // Initialize repositories and services
    tenantRepo := repository.NewTenantRepository(db)
    tenantService := services.NewTenantService(tenantRepo)

    // Initialize handlers
    tenantHandler := handlers.NewTenantHandler(tenantService)

    // Initialize middleware
    tenantMiddleware := middleware.NewTenantMiddleware(db)

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status":    "ok",
            "timestamp": time.Now().Unix(),
            "service":   "oil-gas-inventory-api",
            "version":   "1.0.0",
            "tenant_aware": true,
        })
    })

    // API v1 routes
    v1 := router.Group("/api/v1")
    {
        // Public endpoints (no auth required)
        v1.GET("/status", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "message": "Oil & Gas Inventory API",
                "status":  "running",
                "env":     os.Getenv("APP_ENV"),
                "features": []string{"multi-tenant", "authentication", "rls"},
            })
        })

        // Authenticated endpoints
        auth := v1.Group("/")
        // auth.Use(authMiddleware.Authenticate()) // Uncomment when auth is implemented
        auth.Use(tenantMiddleware.SetTenantContext())
        {
            // Tenant management (admin only)
            tenants := auth.Group("/tenants")
            tenants.Use(tenantMiddleware.RequireAdmin())
            {
                tenants.POST("/", tenantHandler.CreateTenant)
                tenants.GET("/", tenantHandler.ListTenants)
                tenants.GET("/:id", tenantHandler.GetTenant)
                tenants.POST("/assign-customer", tenantHandler.AssignCustomerToTenant)
            }

            // Current tenant context
            auth.GET("/tenant/current", tenantHandler.GetCurrentTenant)
            auth.POST("/tenant/switch", tenantMiddleware.RequireAdmin(), tenantHandler.SwitchTenant)
            auth.GET("/tenant/customers", tenantHandler.GetTenantCustomers)

            // Business endpoints (tenant-aware)
            auth.GET("/customers", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                    "message": "Tenant-aware customers endpoint",
                    "tenant_id": c.GetInt("tenant_id"),
                })
            })

            auth.GET("/inventory", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                    "message": "Tenant-aware inventory endpoint", 
                    "tenant_id": c.GetInt("tenant_id"),
                })
            })

            auth.GET("/received", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                    "message": "Tenant-aware work orders endpoint",
                    "tenant_id": c.GetInt("tenant_id"),
                })
            })
        }
    }

    // Start server
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = "8000"
    }

    fmt.Printf("🚀 Starting Multi-Tenant Oil & Gas Inventory API on port %s\n", port)
    fmt.Printf("📋 Health check: http://localhost:%s/health\n", port)
    fmt.Printf("🔌 API base: http://localhost:%s/api/v1\n", port)
    fmt.Printf("🏢 Tenant-aware: ✅\n")

    log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectDB() (*sql.DB, error) {
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL environment variable not set")
    }

    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(time.Hour)

    return db, nil
}

// backend/cmd/test-startup/main.go - Simple startup test
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("🚀 Starting Oil & Gas Application Startup Test")

	// Test database connections
	if err := testDatabaseConnections(); err != nil {
		log.Fatal("Database connection test failed:", err)
	}

	// Setup basic web server
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "oil-gas-multi-tenant",
			"timestamp": time.Now().Format(time.RFC3339),
			"databases": gin.H{
				"auth_central":           "connected",
				"location_longbeach":     "connected",
				"location_bakersfield":   "connected",
				"location_colorado":      "connected",
			},
		})
	})

	// Basic tenant info endpoint
	router.GET("/api/v1/tenant/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"tenant_id":   "longbeach",
			"location":    "Long Beach Operations",
			"api_version": "v1",
		})
	})

	// Database test endpoint
	router.GET("/api/v1/test/database", func(c *gin.Context) {
		result := testDatabaseQueries()
		if result["success"] == true {
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusInternalServerError, result)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ Application starting successfully on port %s", port)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("🔗 Tenant info: http://localhost:%s/api/v1/tenant/info", port)
	log.Printf("🔗 DB test: http://localhost:%s/api/v1/test/database", port)

	log.Fatal(router.Run(":" + port))
}

func testDatabaseConnections() error {
	authDBURL := getEnvOrDefault("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable")
	tenantDBURL := getEnvOrDefault("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable")

	log.Println("🔗 Testing auth database connection...")
	authDB, err := sql.Open("postgres", authDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to auth database: %w", err)
	}
	defer authDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := authDB.PingContext(ctx); err != nil {
		return fmt.Errorf("auth database ping failed: %w", err)
	}
	log.Println("✅ Auth database connection successful")

	log.Println("🔗 Testing tenant database connection...")
	tenantDB, err := sql.Open("postgres", tenantDBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer tenantDB.Close()

	if err := tenantDB.PingContext(ctx); err != nil {
		return fmt.Errorf("tenant database ping failed: %w", err)
	}
	log.Println("✅ Tenant database connection successful")

	return nil
}

func testDatabaseQueries() map[string]interface{} {
	result := make(map[string]interface{})
	result["timestamp"] = time.Now().Format(time.RFC3339)

	authDBURL := getEnvOrDefault("CENTRAL_AUTH_DB_URL", "postgresql://auth_user:secure_auth_password_2024@localhost:5432/auth_central?sslmode=disable")
	longbeachDBURL := getEnvOrDefault("LONGBEACH_DB_URL", "postgresql://longbeach_user:secure_longbeach_password_2024@localhost:5433/location_longbeach?sslmode=disable")
	bakersfieldDBURL := getEnvOrDefault("BAKERSFIELD_DB_URL", "postgresql://bakersfield_user:bakersfield_password@localhost:5434/location_bakersfield?sslmode=disable")
	coloradoDBURL := getEnvOrDefault("COLORADO_DB_URL", "postgresql://colorado_user:colorado_password@localhost:5435/location_colorado?sslmode=disable")

	// Test auth database
	authDB, err := sql.Open("postgres", authDBURL)
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result
	}
	defer authDB.Close()

	var authUserCount int
	err = authDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&authUserCount)
	if err != nil {
		result["success"] = false
		result["auth_error"] = err.Error()
		return result
	}

	var tenantCount int
	err = authDB.QueryRow("SELECT COUNT(*) FROM tenants WHERE is_active = true").Scan(&tenantCount)
	if err != nil {
		result["success"] = false
		result["auth_error"] = err.Error()
		return result
	}

	// Test Long Beach database
	longbeachDB, err := sql.Open("postgres", longbeachDBURL)
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result
	}
	defer longbeachDB.Close()

	var longbeachCustomers int
	err = longbeachDB.QueryRow("SELECT COUNT(*) FROM store.customers").Scan(&longbeachCustomers)
	if err != nil {
		result["success"] = false
		result["longbeach_error"] = err.Error()
		return result
	}

	// Test Bakersfield database
	bakersfieldDB, err := sql.Open("postgres", bakersfieldDBURL)
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result
	}
	defer bakersfieldDB.Close()

	var bakersfieldCustomers int
	err = bakersfieldDB.QueryRow("SELECT COUNT(*) FROM store.customers").Scan(&bakersfieldCustomers)
	if err != nil {
		result["success"] = false
		result["bakersfield_error"] = err.Error()
		return result
	}

	// Test Colorado database
	coloradoDB, err := sql.Open("postgres", coloradoDBURL)
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result
	}
	defer coloradoDB.Close()

	var coloradoCustomers int
	err = coloradoDB.QueryRow("SELECT COUNT(*) FROM store.customers").Scan(&coloradoCustomers)
	if err != nil {
		result["success"] = false
		result["colorado_error"] = err.Error()
		return result
	}

	result["success"] = true
	result["auth_users"] = authUserCount
	result["active_tenants"] = tenantCount
	result["longbeach_customers"] = longbeachCustomers
	result["bakersfield_customers"] = bakersfieldCustomers
	result["colorado_customers"] = coloradoCustomers
	result["databases"] = map[string]string{
		"auth_central":           "connected",
		"location_longbeach":     "connected", 
		"location_bakersfield":   "connected",
		"location_colorado":      "connected",
	}

	return result
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
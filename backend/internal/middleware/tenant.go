// backend/internal/middleware/tenant.go
package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// TenantMiddleware adds tenant routing to API endpoints
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Tenant header required",
				"message": "Please provide tenant ID in X-Tenant header",
			})
			c.Abort()
			return
		}

		// Validate tenant ID format
		if !isValidTenantID(tenantID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant ID format",
				"message": "Tenant ID must contain only lowercase letters, numbers, and underscores",
			})
			c.Abort()
			return
		}

		// Verify tenant database exists
		if !tenantDatabaseExists(tenantID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Tenant not found",
				"message": fmt.Sprintf("Tenant database 'oilgas_%s' does not exist", tenantID),
			})
			c.Abort()
			return
		}

		// Store tenant info in context
		c.Set("tenant_id", tenantID)
		c.Set("tenant_db", fmt.Sprintf("oilgas_%s", tenantID))
		c.Next()
	}
}

func isValidTenantID(tenantID string) bool {
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return false
	}

	for _, char := range tenantID {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

func tenantDatabaseExists(tenantID string) bool {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return false
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return false
	}
	defer db.Close()

	var exists bool
	dbName := fmt.Sprintf("oilgas_%s", tenantID)
	query := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	err = db.QueryRow(query, dbName).Scan(&exists)
	
	return err == nil && exists
}

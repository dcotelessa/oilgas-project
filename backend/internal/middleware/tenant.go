// backend/internal/middleware/tenant.go
package middleware

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/tenant"
)

type TenantMiddleware struct {
	authRepo  *repository.AuthRepository
	dbManager *tenant.DatabaseManager
}

func NewTenantMiddleware(authRepo *repository.AuthRepository, dbManager *tenant.DatabaseManager) *TenantMiddleware {
	return &TenantMiddleware{
		authRepo:  authRepo,
		dbManager: dbManager,
	}
}

func (tm *TenantMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from header
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
			c.Abort()
			return
		}
		
		// Get tenant code from header
		tenantCode := c.GetHeader("X-Tenant")
		if tenantCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant required"})
			c.Abort()
			return
		}
		
		// Validate tenant
		tenant, err := tm.authRepo.GetTenantByCode(tenantCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant"})
			c.Abort()
			return
		}
		
		// Get tenant database connection
		tenantDB, err := tm.dbManager.GetConnection(tenant.DatabaseName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
			c.Abort()
			return
		}
		
		// Store in context for handlers
		c.Set("tenant", tenant)
		c.Set("tenantDB", tenantDB)
		c.Next()
	}
}

// Helper functions to get from context
func GetTenantDB(c *gin.Context) *sql.DB {
	if db, exists := c.Get("tenantDB"); exists {
		return db.(*sql.DB)
	}
	return nil
}

func GetTenant(c *gin.Context) *models.Tenant {
	if tenant, exists := c.Get("tenant"); exists {
		return tenant.(*models.Tenant)
	}
	return nil
}

// internal/auth/middleware.go
package auth

import (
	"net/http"
	"sql"
	
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/tenant"
)

type TenantMiddleware struct {
	authRepo  *AuthRepository
	dbManager *tenant.DatabaseManager
}

func NewTenantMiddleware(authRepo *AuthRepository, dbManager *tenant.DatabaseManager) *TenantMiddleware {
	return &TenantMiddleware{
		authRepo:  authRepo,
		dbManager: dbManager,
	}
}

func (tm *TenantMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple session-based auth for now
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
			c.Abort()
			return
		}
		
		// TODO: Validate session and get user/tenant info
		// For now, extract tenant from header
		tenantCode := c.GetHeader("X-Tenant")
		if tenantCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant required"})
			c.Abort()
			return
		}
		
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

// Helper to get tenant database from context
func GetTenantDB(c *gin.Context) *sql.DB {
	if db, exists := c.Get("tenantDB"); exists {
		return db.(*sql.DB)
	}
	return nil
}

func GetTenant(c *gin.Context) *Tenant {
	if tenant, exists := c.Get("tenant"); exists {
		return tenant.(*Tenant)
	}
	return nil
}


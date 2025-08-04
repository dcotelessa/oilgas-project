// backend/internal/auth/middleware.go
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantSlug := c.GetHeader("X-Tenant-ID")
		if tenantSlug == "" {
			tenantSlug = c.Query("tenant")
		}
		
		if tenantSlug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
			c.Abort()
			return
		}
		
		tenantSlug = strings.ToLower(strings.TrimSpace(tenantSlug))
		
		tenant := &models.Tenant{
			Slug: tenantSlug,
			Name: tenantSlug,
		}
		
		c.Set("tenant", tenant)
		c.Set("tenant_id", tenantSlug)
		c.Set("tenant_slug", tenantSlug)
		c.Next()
	}
}

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("Authorization")
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}
		
		sessionID = strings.TrimPrefix(sessionID, "Bearer ")
		
		user, err := authService.ValidateSession(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}
		
		c.Set("user", user)
		c.Next()
	}
}

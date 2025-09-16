// backend/internal/auth/middleware.go
package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	authService Service
}

func NewMiddleware(authService Service) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			m.unauthorizedResponse(c, "Authorization token required")
			return
		}

		user, session, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.unauthorizedResponse(c, "Invalid or expired token")
			return
		}

		if !user.IsActive {
			m.unauthorizedResponse(c, "Account is inactive")
			return
		}

		m.setUserContext(c, user, session)
		m.setTenantContext(c, user, session)
		m.setCustomerContext(c, user)

		c.Next()
	}
}

func (m *Middleware) RequireRole(requiredRoles ...UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		for _, role := range requiredRoles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		m.forbiddenResponse(c, "Insufficient role permissions")
	}
}

func (m *Middleware) RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		tenantID := m.getTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
			c.Abort()
			return
		}

		if !m.hasPermissionInTenant(user, tenantID, permission) {
			m.forbiddenResponse(c, fmt.Sprintf("Permission %s denied in tenant %s", permission, tenantID))
			return
		}

		c.Next()
	}
}

func (m *Middleware) RequireTenantAccess(allowedTenants ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		tenantID := m.extractTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
			c.Abort()
			return
		}

		if len(allowedTenants) > 0 {
			allowed := false
			for _, allowed_tenant := range allowedTenants {
				if tenantID == allowed_tenant {
					allowed = true
					break
				}
			}
			if !allowed {
				m.forbiddenResponse(c, fmt.Sprintf("Access denied to tenant %s", tenantID))
				return
			}
		}

		if !user.CanAccessTenant(tenantID) {
			m.forbiddenResponse(c, fmt.Sprintf("User does not have access to tenant %s", tenantID))
			return
		}

		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

func (m *Middleware) RequireCustomerAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		if !user.IsCustomerContact() {
			m.forbiddenResponse(c, "Customer contact access required")
			return
		}

		if err := m.validateCustomerAccess(c, user); err != nil {
			m.forbiddenResponse(c, err.Error())
			return
		}

		c.Next()
	}
}

func (m *Middleware) RequireEnterpriseAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		if !user.CanPerformCrossTenantOperation() {
			m.forbiddenResponse(c, "Enterprise access required")
			return
		}

		c.Next()
	}
}

func (m *Middleware) RequireYardAccess(yardLocation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := m.getUser(c)
		if user == nil {
			m.unauthorizedResponse(c, "Authentication required")
			return
		}

		tenantID := m.getTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
			c.Abort()
			return
		}

		if !user.HasAccessToYard(tenantID, yardLocation) {
			m.forbiddenResponse(c, fmt.Sprintf("Access denied to yard %s in tenant %s", yardLocation, tenantID))
			return
		}

		c.Set("yard_location", yardLocation)
		c.Next()
	}
}

func (m *Middleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return ""
	}

	return token
}

func (m *Middleware) extractTenantID(c *gin.Context) string {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID != "" {
		return tenantID
	}

	tenantID = c.Query("tenant_id")
	if tenantID != "" {
		return tenantID
	}

	tenantID = c.Param("tenant_id")
	if tenantID != "" {
		return tenantID
	}

	user := m.getUser(c)
	if user != nil && user.PrimaryTenantID != "" {
		return user.PrimaryTenantID
	}

	return ""
}

func (m *Middleware) setUserContext(c *gin.Context, user *User, session *Session) {
	c.Set("user", user)
	c.Set("session", session)
	c.Set("user_id", user.ID)
	c.Set("user_role", string(user.Role))
	c.Set("is_enterprise_user", user.IsEnterpriseUser)
}

func (m *Middleware) setTenantContext(c *gin.Context, user *User, session *Session) {
	if session.TenantContext != nil {
		c.Set("tenant_context", session.TenantContext)
		c.Set("tenant_id", session.TenantContext.TenantID)
	} else if user.PrimaryTenantID != "" {
		c.Set("primary_tenant_id", user.PrimaryTenantID)
	}

	if len(user.TenantAccess) > 0 {
		c.Set("tenant_access_list", user.TenantAccess)
	}
}

func (m *Middleware) setCustomerContext(c *gin.Context, user *User) {
	if user.IsCustomerContact() {
		c.Set("customer_id", *user.CustomerID)
		c.Set("contact_type", string(user.ContactType))
		c.Set("is_customer_contact", true)

		context := user.GetCustomerAccessContext()
		if context != nil {
			c.Set("customer_context", context)
			c.Set("accessible_yards", context.AccessibleYards)
		}
	}
}

func (m *Middleware) getUser(c *gin.Context) *User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*User)
}

func (m *Middleware) getTenantID(c *gin.Context) string {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return ""
	}
	return tenantID.(string)
}

func (m *Middleware) hasPermissionInTenant(user *User, tenantID string, permission Permission) bool {
	if user.Role == RoleSystemAdmin {
		return true
	}

	if user.IsEnterpriseUser && (user.Role == RoleEnterpriseAdmin || user.Role == RoleAdmin) {
		return true
	}

	return user.HasPermissionInTenant(tenantID, permission)
}

func (m *Middleware) validateCustomerAccess(c *gin.Context, user *User) error {
	customerIDParam := c.Param("customer_id")
	if customerIDParam == "" {
		return nil
	}

	requestedCustomerID, err := strconv.Atoi(customerIDParam)
	if err != nil {
		return fmt.Errorf("invalid customer ID format")
	}

	if *user.CustomerID != requestedCustomerID {
		return fmt.Errorf("access denied to customer %d (user belongs to customer %d)", requestedCustomerID, *user.CustomerID)
	}

	return nil
}

func (m *Middleware) unauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error":     "Unauthorized",
		"message":   message,
		"timestamp": time.Now().UTC(),
	})
	c.Abort()
}

func (m *Middleware) forbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":     "Forbidden",
		"message":   message,
		"timestamp": time.Now().UTC(),
	})
	c.Abort()
}

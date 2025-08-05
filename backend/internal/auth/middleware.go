// backend/internal/auth/middleware.go
package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	service Service
}

func NewMiddleware(service Service) *Middleware {
	return &Middleware{service: service}
}

// RequireAuth validates JWT token and sets user context
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}
		
		user, tenantContext, err := m.service.ValidateToken(c.Request.Context(), token)
		if err != nil {
			switch err {
			case ErrTokenExpired:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			case ErrInvalidCredentials:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			}
			c.Abort()
			return
		}
		
		// Set user context for downstream handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("tenant_context", tenantContext)
		if tenantContext != nil {
			c.Set("tenant_id", tenantContext.TenantID)
		}
		c.Set("user_role", user.Role)
		
		c.Next()
	}
}

// RequireRole validates that user has one of the specified roles
func (m *Middleware) RequireRole(roles ...UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		userObj := user.(*User)
		
		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if userObj.Role == role {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role permissions"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequirePermission validates that user has specific permission in current tenant
func (m *Middleware) RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context not found"})
			c.Abort()
			return
		}
		
		check := UserPermissionCheck{
			UserID:     userID.(int),
			TenantID:   tenantID.(string),
			Permission: permission,
		}
		
		if err := m.service.CheckPermission(c.Request.Context(), check); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireYardAccess validates that user has access to specific yard
func (m *Middleware) RequireYardAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context not found"})
			c.Abort()
			return
		}
		
		// Get yard location from URL parameter or query string
		yardLocation := c.Param("yardLocation")
		if yardLocation == "" {
			yardLocation = c.Query("yard_location")
		}
		
		if yardLocation == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Yard location not specified"})
			c.Abort()
			return
		}
		
		if err := m.service.CheckYardAccess(c.Request.Context(), userID.(int), tenantID.(string), yardLocation); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Yard access denied"})
			c.Abort()
			return
		}
		
		// Set yard context for downstream handlers
		c.Set("yard_location", yardLocation)
		
		c.Next()
	}
}

// RequireCustomerContact validates that user is a customer contact
func (m *Middleware) RequireCustomerContact() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		userObj := user.(*User)
		if !userObj.IsCustomerContact() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Customer contact access required"})
			c.Abort()
			return
		}
		
		// Set customer context
		c.Set("customer_id", *userObj.CustomerID)
		c.Set("contact_type", userObj.ContactType)
		
		c.Next()
	}
}

// RequireEnterpriseAccess validates that user has enterprise-level access
func (m *Middleware) RequireEnterpriseAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		userObj := user.(*User)
		if !userObj.CanPerformCrossTenantOperation() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Enterprise access required"})
			c.Abort()
			return
		}
		
		// Get enterprise context
		enterpriseContext, err := m.service.GetEnterpriseContext(c.Request.Context(), userObj.ID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Failed to get enterprise context"})
			c.Abort()
			return
		}
		
		c.Set("enterprise_context", enterpriseContext)
		
		c.Next()
	}
}

// ValidateTenantAccess ensures user has access to specified tenant
func (m *Middleware) ValidateTenantAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		
		// Get tenant ID from URL parameter
		requestedTenantID := c.Param("tenantId")
		if requestedTenantID == "" {
			requestedTenantID = c.Query("tenant_id")
		}
		
		if requestedTenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not specified"})
			c.Abort()
			return
		}
		
		userObj := user.(*User)
		
		// Check if user has access to this tenant
		if !userObj.CanAccessTenant(requestedTenantID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant access denied"})
			c.Abort()
			return
		}
		
		// Override tenant context with requested tenant
		c.Set("tenant_id", requestedTenantID)
		
		c.Next()
	}
}

// SetTenantContext sets tenant context from various sources
func (m *Middleware) SetTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get tenant ID from various sources
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}
		
		// If no tenant specified, use user's primary tenant
		if tenantID == "" {
			if user, exists := c.Get("user"); exists {
				userObj := user.(*User)
				tenantID = userObj.PrimaryTenantID
			}
		}
		
		if tenantID != "" {
			c.Set("tenant_id", tenantID)
		}
		
		c.Next()
	}
}

// CustomerContactFilter filters data to only show customer's own data
func (m *Middleware) CustomerContactFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}
		
		userObj := user.(*User)
		
		// If user is a customer contact, set customer filter
		if userObj.IsCustomerContact() && userObj.CustomerID != nil {
			c.Set("customer_filter", *userObj.CustomerID)
		}
		
		c.Next()
	}
}

// LogUserActivity logs user activity for audit purposes
func (m *Middleware) LogUserActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log before processing
		if user, exists := c.Get("user"); exists {
			userObj := user.(*User)
			c.Header("X-User-ID", fmt.Sprintf("%d", userObj.ID))
			c.Header("X-User-Role", string(userObj.Role))
			
			if tenantID, exists := c.Get("tenant_id"); exists {
				c.Header("X-Tenant-ID", tenantID.(string))
			}
		}
		
		c.Next()
		
		// Log after processing (could log to audit system)
		// Implementation depends on your logging requirements
	}
}

// CORS middleware for auth endpoints
func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
		c.Header("Access-Control-Expose-Headers", "X-User-ID, X-User-Role, X-Tenant-ID")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// Helper methods
func (m *Middleware) extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Remove "Bearer " prefix
		if len(authHeader) > 7 && strings.ToLower(authHeader[:7]) == "bearer " {
			return authHeader[7:]
		}
	}
	
	// Try query parameter as fallback
	return c.Query("token")
}

// Utility functions for getting context values
func GetUserFromContext(c *gin.Context) (*User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userObj, ok := user.(*User)
	return userObj, ok
}

func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int)
	return id, ok
}

func GetTenantIDFromContext(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	id, ok := tenantID.(string)
	return id, ok
}

func GetTenantContextFromContext(c *gin.Context) (*TenantAccess, bool) {
	tenantContext, exists := c.Get("tenant_context")
	if !exists {
		return nil, false
	}
	context, ok := tenantContext.(*TenantAccess)
	return context, ok
}

func GetCustomerIDFromContext(c *gin.Context) (int, bool) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		return 0, false
	}
	id, ok := customerID.(int)
	return id, ok
}

func GetYardLocationFromContext(c *gin.Context) (string, bool) {
	yardLocation, exists := c.Get("yard_location")
	if !exists {
		return "", false
	}
	location, ok := yardLocation.(string)
	return location, ok
}

func GetEnterpriseContextFromContext(c *gin.Context) (*EnterpriseContext, bool) {
	enterpriseContext, exists := c.Get("enterprise_context")
	if !exists {
		return nil, false
	}
	context, ok := enterpriseContext.(*EnterpriseContext)
	return context, ok
}

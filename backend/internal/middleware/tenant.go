package middleware

import (
    "database/sql"
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
)

type TenantMiddleware struct {
    db *sql.DB
}

func NewTenantMiddleware(db *sql.DB) *TenantMiddleware {
    return &TenantMiddleware{db: db}
}

// SetTenantContext sets the tenant context for the current request
func (tm *TenantMiddleware) SetTenantContext() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user info from auth middleware (assumed to be set)
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            c.Abort()
            return
        }

        // Get tenant from header or query param for admin users
        requestedTenant := c.GetHeader("X-Tenant-ID")
        if requestedTenant == "" {
            requestedTenant = c.Query("tenant_id")
        }

        var tenantID int
        var tenantSlug string
        var userRole string
        var isAdmin bool

        // Query user's tenant and role
        query := `
            SELECT 
                u.tenant_id,
                t.tenant_slug,
                COALESCE(utr.role, 'viewer') as role,
                CASE WHEN utr.role = 'admin' THEN true ELSE false END as is_admin
            FROM store.users u
            LEFT JOIN store.tenants t ON u.tenant_id = t.tenant_id
            LEFT JOIN store.user_tenant_roles utr ON u.user_id = utr.user_id AND u.tenant_id = utr.tenant_id
            WHERE u.user_id = $1 AND u.active = true
        `
        
        err := tm.db.QueryRow(query, userID).Scan(&tenantID, &tenantSlug, &userRole, &isAdmin)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant context"})
            c.Abort()
            return
        }

        // Allow admin users to switch tenant context
        if isAdmin && requestedTenant != "" {
            if requestedTenantID, err := strconv.Atoi(requestedTenant); err == nil {
                // Verify tenant exists
                var exists bool
                err = tm.db.QueryRow("SELECT EXISTS(SELECT 1 FROM store.tenants WHERE tenant_id = $1 AND active = true)", 
                    requestedTenantID).Scan(&exists)
                if err == nil && exists {
                    tenantID = requestedTenantID
                    // Get the slug for the switched tenant
                    tm.db.QueryRow("SELECT tenant_slug FROM store.tenants WHERE tenant_id = $1", 
                        requestedTenantID).Scan(&tenantSlug)
                }
            }
        }

        // Set database session variables for RLS
        _, err = tm.db.Exec("SET app.current_tenant_id = $1", tenantID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set tenant context"})
            c.Abort()
            return
        }

        _, err = tm.db.Exec("SET app.current_user_id = $1", userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set user context"})
            c.Abort()
            return
        }

        _, err = tm.db.Exec("SET app.user_role = $1", userRole)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set user role"})
            c.Abort()
            return
        }

        // Set context for use in handlers
        c.Set("tenant_id", tenantID)
        c.Set("tenant_slug", tenantSlug)
        c.Set("user_role", userRole)
        c.Set("is_admin", isAdmin)

        c.Next()
    }
}

// RequireAdmin ensures the user has admin privileges
func (tm *TenantMiddleware) RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        isAdmin, exists := c.Get("is_admin")
        if !exists || !isAdmin.(bool) {
            c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// RequireRole ensures the user has the specified role or higher
func (tm *TenantMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
    roleHierarchy := map[string]int{
        "viewer":   1,
        "operator": 2,
        "manager":  3,
        "admin":    4,
    }

    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
            c.Abort()
            return
        }

        userLevel := roleHierarchy[userRole.(string)]
        requiredLevel := roleHierarchy[requiredRole]

        if userLevel < requiredLevel {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Insufficient privileges",
                "required_role": requiredRole,
                "user_role": userRole,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

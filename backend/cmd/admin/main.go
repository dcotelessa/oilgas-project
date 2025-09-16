// backend/cmd/admin/main.go - Multi-tenant admin application
package main

import (
	"log"
	"net/http"
	"os"
	"time"
	
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	
	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/customer"
	"oilgas-backend/internal/shared/database"
)

func main() {
	// Database configuration for admin (all tenants initialized and ready)
	dbConfig := &database.Config{
		CentralDBURL: getDBURL("CENTRAL_AUTH_DB_URL", "DEV_CENTRAL_AUTH_DB_URL"),
		TenantDBs: map[string]string{
			"longbeach":   getDBURL("LONGBEACH_DB_URL", "DEV_LONGBEACH_DB_URL"),
			"bakersfield": getDBURL("BAKERSFIELD_DB_URL", "DEV_BAKERSFIELD_DB_URL"),
			"colorado":    getDBURL("COLORADO_DB_URL", "DEV_COLORADO_DB_URL"),
		},
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  time.Hour,
	}
	
	// Initialize database manager
	dbManager, err := database.NewDatabaseManager(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database manager:", err)
	}
	defer dbManager.Close()
	
	// Initialize services
	authRepo := auth.NewRepository(dbManager.GetCentralDB())
	authSvc := auth.NewService(dbManager, authRepo)
	customerRepo := customer.NewRepository(dbManager)
	customerCache := customer.NewInMemoryCache(time.Hour)
	customerSvc := customer.NewService(customerRepo, authSvc, customerCache)
	
	// Initialize handlers
	authHandlers := auth.NewAuthHandler(authSvc)
	customerHandlers := customer.NewHandlers(customerSvc)
	adminHandlers := NewAdminHandlers(authSvc, customerSvc)
	
	// Setup router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	// Public routes (no auth required)
	public := router.Group("/api/v1")
	public.POST("/login", authHandlers.Login)
	public.POST("/logout", authHandlers.Logout)
	public.GET("/health", healthCheck(dbManager))
	
	// Admin routes (auth required)
	admin := router.Group("/api/v1/admin")
	admin.Use(authMiddleware(authSvc))
	admin.Use(adminOnlyMiddleware())
	
	// Admin user management
	admin.POST("/users", adminHandlers.CreateUser)
	admin.GET("/users", adminHandlers.ListUsers)
	admin.PUT("/users/:id", adminHandlers.UpdateUser)
	admin.DELETE("/users/:id", adminHandlers.DeleteUser)
	
	// Tenant management
	admin.GET("/tenants", adminHandlers.ListTenants)
	admin.POST("/tenants/:tenant_id/switch", adminHandlers.SwitchTenant)
	
	// Multi-tenant customer routes (dynamic tenant switching)
	tenantRoutes := router.Group("/api/v1/:tenant_id")
	tenantRoutes.Use(authMiddleware(authSvc))
	tenantRoutes.Use(tenantAccessMiddleware(authSvc))
	tenantRoutes.Use(dynamicTenantMiddleware())
	
	// Register customer routes with dynamic tenant support
	customerHandlers.RegisterRoutes(tenantRoutes, authMiddleware(authSvc))
	
	log.Println("Multi-tenant admin application starting on :8080")
	log.Fatal(router.Run(":8080"))
}

// AdminHandlers handles admin-specific operations
type AdminHandlers struct {
	authSvc     auth.Service
	customerSvc customer.Service
}

func NewAdminHandlers(authSvc auth.Service, customerSvc customer.Service) *AdminHandlers {
	return &AdminHandlers{
		authSvc:     authSvc,
		customerSvc: customerSvc,
	}
}

func (h *AdminHandlers) CreateUser(c *gin.Context) {
	// Implementation for creating users
	c.JSON(http.StatusNotImplemented, gin.H{"message": "User creation not yet implemented"})
}

func (h *AdminHandlers) ListUsers(c *gin.Context) {
	// Implementation for listing users
	c.JSON(http.StatusNotImplemented, gin.H{"message": "User listing not yet implemented"})
}

func (h *AdminHandlers) UpdateUser(c *gin.Context) {
	// Implementation for updating users
	c.JSON(http.StatusNotImplemented, gin.H{"message": "User updates not yet implemented"})
}

func (h *AdminHandlers) DeleteUser(c *gin.Context) {
	// Implementation for deleting users
	c.JSON(http.StatusNotImplemented, gin.H{"message": "User deletion not yet implemented"})
}

func (h *AdminHandlers) ListTenants(c *gin.Context) {
	// Return available tenants for the authenticated user
	tenants := []gin.H{
		{"id": "longbeach", "name": "Long Beach Operations", "active": true, "customer_count": "6+"},
		{"id": "bakersfield", "name": "Bakersfield Operations", "active": true, "customer_count": "0 (ready for data)"},
		{"id": "colorado", "name": "Colorado Operations", "active": true, "customer_count": "0 (ready for data)"},
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (h *AdminHandlers) SwitchTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	userID := c.GetInt("user_id")
	
	// Validate user has access to the tenant
	hasAccess, err := h.authSvc.ValidateUserTenantAccess(c.Request.Context(), userID, tenantID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to tenant: " + tenantID})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "Switched to tenant: " + tenantID,
		"tenant_id": tenantID,
	})
}

// Middleware functions

func authMiddleware(authSvc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Remove "Bearer " prefix
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		
		// Validate token
		user, session, err := authSvc.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// Set user context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", string(user.Role))
		c.Set("session", session)
		
		c.Next()
	}
}

func adminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		
		// Check if user has admin privileges
		if userRole != "ADMIN" && userRole != "ENTERPRISE_ADMIN" && userRole != "SYSTEM_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func tenantAccessMiddleware(authSvc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		tenantID := c.Param("tenant_id")
		
		// Validate user has access to this tenant
		hasAccess, err := authSvc.ValidateUserTenantAccess(c.Request.Context(), userID, tenantID)
		if err != nil || !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to tenant: " + tenantID})
			c.Abort()
			return
		}
		
		// Check if user is a customer contact and set customer filter
		customerID, err := authSvc.GetUserCustomerContext(c.Request.Context(), userID, tenantID)
		if err == nil && customerID != nil {
			c.Set("customer_filter", *customerID)
			c.Set("user_role", "customer_contact")
		}
		
		c.Next()
	}
}

func dynamicTenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Param("tenant_id")
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

func healthCheck(dbManager *database.DatabaseManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"databases": gin.H{},
		}
		
		// Check central auth database
		if err := dbManager.GetCentralDB().Ping(); err != nil {
			status["status"] = "unhealthy"
			status["databases"].(gin.H)["auth"] = "error: " + err.Error()
		} else {
			status["databases"].(gin.H)["auth"] = "healthy"
		}
		
		// Check all tenant databases
		tenants := []string{"longbeach", "bakersfield", "colorado"}
		for _, tenant := range tenants {
			if db, err := dbManager.GetTenantDB(tenant); err != nil {
				status["status"] = "unhealthy"
				status["databases"].(gin.H)[tenant] = "error: " + err.Error()
			} else if err := db.Ping(); err != nil {
				status["status"] = "unhealthy"
				status["databases"].(gin.H)[tenant] = "error: " + err.Error()
			} else {
				status["databases"].(gin.H)[tenant] = "healthy"
			}
		}
		
		httpStatus := http.StatusOK
		if status["status"] == "unhealthy" {
			httpStatus = http.StatusServiceUnavailable
		}
		
		c.JSON(httpStatus, status)
	}
}

// getDBURL returns dev URL if available, otherwise production URL
func getDBURL(prodKey, devKey string) string {
	if devURL := os.Getenv(devKey); devURL != "" {
		return devURL
	}
	return os.Getenv(prodKey)
}
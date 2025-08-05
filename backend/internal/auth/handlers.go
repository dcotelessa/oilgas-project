// backend/internal/auth/handlers.go
package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	service Service
}

func NewHandlers(service Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes sets up all auth-related routes
func (h *Handlers) RegisterRoutes(router *gin.RouterGroup, middleware *Middleware) {
	// Public auth endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)  
		auth.POST("/refresh", h.RefreshToken)
	}

	// Protected user management endpoints
	users := router.Group("/users")
	users.Use(middleware.RequireAuth())
	{
		users.POST("/customer-contacts", middleware.RequirePermission(PermissionUserManagement), h.CreateCustomerContact)
		users.POST("/enterprise-users", middleware.RequirePermission(PermissionUserManagement), h.CreateEnterpriseUser)
		users.GET("/search", middleware.RequirePermission(PermissionUserManagement), h.SearchUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission(PermissionUserManagement), h.DeactivateUser)
		users.PUT("/:id/tenant-access", middleware.RequireEnterpriseAccess(), h.UpdateUserTenantAccess)
		users.PUT("/:id/yard-access", middleware.RequirePermission(PermissionUserManagement), h.UpdateUserYardAccess)
	}

	// Permission checking endpoints
	permissions := router.Group("/permissions")
	permissions.Use(middleware.RequireAuth())
	{
		permissions.POST("/check", h.CheckPermission)
		permissions.GET("/user/:id", h.GetUserPermissions)
	}

	// Admin endpoints (enterprise access required)
	admin := router.Group("/admin")
	admin.Use(middleware.RequireAuth())
	admin.Use(middleware.RequireEnterpriseAccess())
	{
		admin.GET("/users", h.ListAllUsers)
		admin.GET("/stats", h.GetUserStats)
		admin.GET("/customer-contacts/:customerId", h.GetCustomerContacts)
	}
}

// ============================================================================
// AUTHENTICATION ENDPOINTS
// ============================================================================

// LoginRequest for authentication
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID string `json:"tenant_id"`
}

// LoginResponse for authentication
type LoginResponse struct {
	Token           string           `json:"token"`
	User            UserResponse     `json:"user"`
	TenantContext   *TenantAccess    `json:"tenant_context,omitempty"`
	ExpiresAt       time.Time        `json:"expires_at"`
	RefreshToken    string           `json:"refresh_token"`
}

// RefreshTokenRequest for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login authenticates a user and returns JWT token
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login request",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username and password are required",
		})
		return
	}

	response, err := h.service.Authenticate(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
		case ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User account is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout invalidates the current session
func (h *Handlers) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No token provided",
		})
		return
	}

	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// RefreshToken generates a new access token using refresh token
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid refresh request",
		})
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case ErrInvalidCredentials, ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired refresh token",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Token refresh failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// USER MANAGEMENT ENDPOINTS
// ============================================================================

// CreateCustomerContact creates a new customer contact user
func (h *Handlers) CreateCustomerContact(c *gin.Context) {
	var req CreateCustomerContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer contact request",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if err := h.validateCustomerContactRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.CreateCustomerContact(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this email or username already exists",
			})
		case ErrInvalidTenant:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create customer contact",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Customer contact created successfully",
		"user": user.ToResponse(),
	})
}

// CreateEnterpriseUser creates a new enterprise user
func (h *Handlers) CreateEnterpriseUser(c *gin.Context) {
	var req CreateEnterpriseUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise user request",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if err := h.validateEnterpriseUserRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.CreateEnterpriseUser(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this email or username already exists",
			})
		case ErrInvalidTenant:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create enterprise user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Enterprise user created successfully",
		"user": user.ToResponse(),
	})
}

// GetUser retrieves user details (with proper access control)
func (h *Handlers) GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Check if user can access this user record
	currentUser, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User context not found",
		})
		return
	}

	// Users can always access their own record
	// Admin/enterprise users can access others
	if currentUser.ID != userID && !currentUser.CanManageOtherUsers() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}

// UpdateUser updates user information
func (h *Handlers) UpdateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var updates UserUpdates
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid update request",
			"details": err.Error(),
		})
		return
	}

	// Check access permissions
	currentUser, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User context not found",
		})
		return
	}

	// Users can update their own basic info, admins can update others
	if currentUser.ID != userID && !currentUser.CanManageOtherUsers() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// Restrict what non-admin users can update about themselves
	if currentUser.ID == userID && !currentUser.CanManageOtherUsers() {
		updates = h.filterUserSelfUpdates(updates)
	}

	user, err := h.service.UpdateUser(c.Request.Context(), userID, updates)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email or username already in use",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user": user.ToResponse(),
	})
}

// DeactivateUser deactivates a user account
func (h *Handlers) DeactivateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Users cannot deactivate themselves
	currentUser, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User context not found",
		})
		return
	}

	if currentUser.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot deactivate your own account",
		})
		return
	}

	if err := h.service.DeactivateUser(c.Request.Context(), userID); err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to deactivate user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
	})
}

// SearchUsers searches for users with filters
func (h *Handlers) SearchUsers(c *gin.Context) {
	// Get search parameters
	query := c.Query("q")
	tenantID := c.Query("tenant_id")
	role := c.Query("role")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Build search filters
	filters := UserSearchFilters{
		Query:    query,
		TenantID: tenantID,
		Role:     UserRole(role),
		Limit:    limit,
		Offset:   offset,
	}

	// Get current user context for filtering
	currentUser, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User context not found",
		})
		return
	}

	// Apply user-specific filtering
	filters = h.applyUserSearchFiltering(currentUser, filters)

	users, total, err := h.service.SearchUsers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// ============================================================================
// TENANT ACCESS MANAGEMENT
// ============================================================================

// UpdateUserTenantAccess updates user's tenant access (enterprise admin only)
func (h *Handlers) UpdateUserTenantAccess(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req UpdateUserTenantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant access request",
			"details": err.Error(),
		})
		return
	}

	req.UserID = userID

	if err := h.service.UpdateUserTenantAccess(c.Request.Context(), req); err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		case ErrInvalidTenant:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update tenant access",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant access updated successfully",
	})
}

// UpdateUserYardAccess updates user's yard-level access
func (h *Handlers) UpdateUserYardAccess(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req UpdateUserYardAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid yard access request",
			"details": err.Error(),
		})
		return
	}

	req.UserID = userID

	if err := h.service.UpdateUserYardAccess(c.Request.Context(), req); err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		case ErrInvalidTenant:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant or yard specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update yard access",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Yard access updated successfully",
	})
}

// ============================================================================
// PERMISSION CHECKING ENDPOINTS
// ============================================================================

// CheckPermission checks if user has specific permission
func (h *Handlers) CheckPermission(c *gin.Context) {
	var req UserPermissionCheck
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid permission check request",
			"details": err.Error(),
		})
		return
	}

	// Use current user if not specified
	if req.UserID == 0 {
		if userID, exists := GetUserIDFromContext(c); exists {
			req.UserID = userID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User ID required",
			})
			return
		}
	}

	// Use current tenant if not specified
	if req.TenantID == "" {
		if tenantID, exists := GetTenantIDFromContext(c); exists {
			req.TenantID = tenantID
		}
	}

	err := h.service.CheckPermission(c.Request.Context(), req)
	hasPermission := err == nil

	response := gin.H{
		"has_permission": hasPermission,
		"user_id":       req.UserID,
		"tenant_id":     req.TenantID,
		"permission":    req.Permission,
	}

	if req.YardLocation != nil {
		response["yard_location"] = *req.YardLocation
	}

	if !hasPermission && err != ErrPermissionDenied {
		response["error"] = err.Error()
	}

	c.JSON(http.StatusOK, response)
}

// GetUserPermissions retrieves all permissions for a user in a tenant
func (h *Handlers) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		if tid, exists := GetTenantIDFromContext(c); exists {
			tenantID = tid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tenant ID required",
			})
			return
		}
	}

	permissions, err := h.service.GetUserPermissions(c.Request.Context(), userID, tenantID)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user permissions",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"permissions": permissions,
	})
}

// ============================================================================
// ADMIN ENDPOINTS (Enterprise Access Required)
// ============================================================================

// ListAllUsers lists all users across tenants (enterprise admin only)
func (h *Handlers) ListAllUsers(c *gin.Context) {
	// Get search parameters
	tenantID := c.Query("tenant_id")
	role := c.Query("role")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Validate limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	// Get enterprise context to determine accessible tenants
	enterpriseContext, exists := GetEnterpriseContextFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Enterprise context required",
		})
		return
	}

	// Build filters
	filters := AdminUserFilters{
		TenantID:          tenantID,
		Role:              UserRole(role),
		Limit:             limit,
		Offset:            offset,
		AccessibleTenants: extractAccessibleTenantIDs(enterpriseContext.AccessibleTenants),
	}

	users, total, err := h.service.ListAllUsers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list users",
		})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  userResponses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUserStats provides user statistics (enterprise admin only)
func (h *Handlers) GetUserStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	// Get enterprise context for filtering
	enterpriseContext, exists := GetEnterpriseContextFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Enterprise context required",
		})
		return
	}

	stats, err := h.service.GetUserStats(c.Request.Context(), UserStatsRequest{
		TenantID:          tenantID,
		AccessibleTenants: extractAccessibleTenantIDs(enterpriseContext.AccessibleTenants),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCustomerContacts retrieves all contacts for a specific customer
func (h *Handlers) GetCustomerContacts(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
		})
		return
	}

	users, err := h.service.GetCustomerContacts(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get customer contacts",
		})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id": customerID,
		"contacts":    userResponses,
	})
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// extractToken extracts JWT token from request
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if len(authHeader) > 7 && strings.ToLower(authHeader[:7]) == "bearer " {
			return authHeader[7:]
		}
	}
	
	// Try query parameter as fallback
	return c.Query("token")
}

// validateCustomerContactRequest validates customer contact creation request
func (h *Handlers) validateCustomerContactRequest(req CreateCustomerContactRequest) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.ContactType == "" {
		req.ContactType = ContactPrimary
	}
	return nil
}

// validateEnterpriseUserRequest validates enterprise user creation request
func (h *Handlers) validateEnterpriseUserRequest(req CreateEnterpriseUserRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.Role == "" {
		return fmt.Errorf("role is required")
	}
	if req.PrimaryTenantID == "" && !req.IsEnterpriseUser {
		return fmt.Errorf("primary tenant ID is required for non-enterprise users")
	}
	if len(req.TenantAccess) == 0 {
		return fmt.Errorf("at least one tenant access is required")
	}
	return nil
}

// filterUserSelfUpdates restricts what users can update about themselves
func (h *Handlers) filterUserSelfUpdates(updates UserUpdates) UserUpdates {
	// Users can only update basic profile information
	filtered := UserUpdates{
		FullName: updates.FullName,
		Email:    updates.Email,
	}
	
	// Password changes allowed if both old and new provided
	if updates.CurrentPassword != nil && updates.NewPassword != nil {
		filtered.CurrentPassword = updates.CurrentPassword
		filtered.NewPassword = updates.NewPassword
	}
	
	return filtered
}

// applyUserSearchFiltering applies user-specific search filtering
func (h *Handlers) applyUserSearchFiltering(currentUser *User, filters UserSearchFilters) UserSearchFilters {
	// Customer contacts can only see other contacts from same customer
	if currentUser.IsCustomerContact() && currentUser.CustomerID != nil {
		filters.CustomerID = currentUser.CustomerID
		filters.OnlyCustomerContacts = true
	}
	
	// Non-enterprise users are limited to their tenant access
	if !currentUser.IsEnterpriseUser {
		accessibleTenants := make([]string, 0, len(currentUser.TenantAccess))
		for _, access := range currentUser.TenantAccess {
			accessibleTenants = append(accessibleTenants, access.TenantID)
		}
		filters.AccessibleTenants = accessibleTenants
	}
	
	return filters
}

// extractAccessibleTenantIDs extracts tenant IDs from tenant access list
func extractAccessibleTenantIDs(tenantAccess []TenantAccess) []string {
	tenantIDs := make([]string, 0, len(tenantAccess))
	for _, access := range tenantAccess {
		tenantIDs = append(tenantIDs, access.TenantID)
	}
	return tenantIDs
}

// ============================================================================
// REQUEST/RESPONSE MODELS FOR SEARCH AND ADMIN
// ============================================================================

type UserSearchFilters struct {
	Query                string     `json:"query"`
	TenantID            string     `json:"tenant_id"`
	Role                UserRole   `json:"role"`
	CustomerID          *int       `json:"customer_id"`
	OnlyCustomerContacts bool       `json:"only_customer_contacts"`
	AccessibleTenants   []string   `json:"accessible_tenants"`
	Limit               int        `json:"limit"`
	Offset              int        `json:"offset"`
}

type AdminUserFilters struct {
	TenantID          string   `json:"tenant_id"`
	Role              UserRole `json:"role"`
	AccessibleTenants []string `json:"accessible_tenants"`
	Limit             int      `json:"limit"`
	Offset            int      `json:"offset"`
}

type UserStatsRequest struct {
	TenantID          string   `json:"tenant_id"`
	AccessibleTenants []string `json:"accessible_tenants"`
}

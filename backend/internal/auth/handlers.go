// backend/internal/auth/handlers.go
package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	service Service
}

func NewHandlers(service Service) *Handlers {
	return &Handlers{service: service}
}

// Authentication endpoints
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}
	
	response, err := h.service.Authenticate(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is inactive"})
		case ErrTenantAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to tenant"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		}
		return
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *Handlers) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}
	
	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *Handlers) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	response, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// User management endpoints
func (h *Handlers) CreateCustomerContact(c *gin.Context) {
	var req CreateCustomerContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}
	
	user, err := h.service.CreateCustomerContact(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrCustomerNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not found"})
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		case ErrInvalidYardAccess:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid yard access configuration"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer contact"})
		}
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"user": user.ToResponse()})
}

func (h *Handlers) CreateEnterpriseUser(c *gin.Context) {
	var req CreateEnterpriseUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}
	
	user, err := h.service.CreateEnterpriseUser(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		case ErrInvalidUserRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role"})
		case ErrInvalidYardAccess:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid yard access configuration"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create enterprise user"})
		}
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"user": user.ToResponse()})
}

func (h *Handlers) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}

func (h *Handlers) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var updates UserUpdates
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	user, err := h.service.UpdateUser(c.Request.Context(), userID, updates)
	if err != nil {
		switch err {
		case ErrInvalidUserRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user role"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}

func (h *Handlers) UpdateUserTenantAccess(c *gin.Context) {
	var req UpdateUserTenantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	if err := h.service.UpdateUserTenantAccess(c.Request.Context(), req); err != nil {
		switch err {
		case ErrInvalidYardAccess:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid yard access configuration"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant access"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Tenant access updated successfully"})
}

func (h *Handlers) UpdateUserYardAccess(c *gin.Context) {
	var req UpdateUserYardAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	if err := h.service.UpdateUserYardAccess(c.Request.Context(), req); err != nil {
		switch err {
		case ErrInvalidYardAccess:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid yard access configuration"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update yard access"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Yard access updated successfully"})
}

func (h *Handlers) DeactivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	if err := h.service.DeactivateUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// User query endpoints
func (h *Handlers) GetUsersByTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}
	
	users, err := h.service.GetUsersByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}
	
	// Convert to response format
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}
	
	c.JSON(http.StatusOK, gin.H{"users": userResponses})
}

func (h *Handlers) GetCustomerContacts(c *gin.Context) {
	customerIDStr := c.Param("customerId")
	customerID, err := strconv.Atoi(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	
	users, err := h.service.GetCustomerContacts(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer contacts"})
		return
	}
	
	// Convert to response format
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}
	
	c.JSON(http.StatusOK, gin.H{"contacts": userResponses})
}

func (h *Handlers) SearchUsers(c *gin.Context) {
	var filter UserSearchFilter
	
	// Extract query parameters
	filter.Query = c.Query("q")
	filter.SortBy = c.DefaultQuery("sort_by", "full_name")
	filter.SortOrder = c.DefaultQuery("sort_order", "asc")
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}
	
	if roleStr := c.Query("role"); roleStr != "" {
		role := UserRole(roleStr)
		filter.Role = &role
	}
	
	if tenantID := c.Query("tenant_id"); tenantID != "" {
		filter.TenantID = &tenantID
	}
	
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			filter.CustomerID = &customerID
		}
	}
	
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}
	
	if isEnterpriseStr := c.Query("is_enterprise_user"); isEnterpriseStr != "" {
		if isEnterprise, err := strconv.ParseBool(isEnterpriseStr); err == nil {
			filter.IsEnterpriseUser = &isEnterprise
		}
	}
	
	users, err := h.service.SearchUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}
	
	// Convert to response format
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}
	
	c.JSON(http.StatusOK, gin.H{
		"users":  userResponses,
		"count":  len(userResponses),
		"filter": filter,
	})
}

func (h *Handlers) GetUserStats(c *gin.Context) {
	stats, err := h.service.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user statistics"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// Enterprise context endpoints
func (h *Handlers) GetEnterpriseContext(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	context, err := h.service.GetEnterpriseContext(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case ErrEnterpriseAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "Enterprise access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get enterprise context"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"context": context})
}

func (h *Handlers) GetCustomerAccessContext(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	context, err := h.service.GetCustomerAccessContext(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case ErrNotCustomerContact:
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not a customer contact"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer access context"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"context": context})
}

// Permission checking endpoints
func (h *Handlers) CheckUserPermission(c *gin.Context) {
	var check UserPermissionCheck
	if err := c.ShouldBindJSON(&check); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	if err := h.service.CheckPermission(c.Request.Context(), check); err != nil {
		switch err {
		case ErrPermissionDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied", "has_permission": false})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"has_permission": true})
}

func (h *Handlers) CheckYardAccess(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	tenantID := c.Query("tenant_id")
	yardLocation := c.Query("yard_location")
	
	if tenantID == "" || yardLocation == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and yard_location are required"})
		return
	}
	
	if err := h.service.CheckYardAccess(c.Request.Context(), userID, tenantID, yardLocation); err != nil {
		switch err {
		case ErrYardAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "Yard access denied", "has_access": false})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check yard access"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"has_access": true})
}

func (h *Handlers) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}
	
	permissions, err := h.service.GetUserPermissions(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// Session management endpoints
func (h *Handlers) InvalidateUserSessions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	if err := h.service.InvalidateUserSessions(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate user sessions"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User sessions invalidated successfully"})
}

func (h *Handlers) CleanupExpiredSessions(c *gin.Context) {
	if err := h.service.CleanupExpiredSessions(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup expired sessions"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Expired sessions cleaned up successfully"})
}

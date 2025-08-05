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
	return &Handlers{
		service: service,
	}
}

// Fix #1: SearchUsers method signature and UserSearchFilters type
func (h *Handlers) SearchUsers(c *gin.Context) {
	var filters UserSearchFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid search parameters",
			"details": err.Error(),
		})
		return
	}

	// Get current user from context for access control
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	user := currentUser.(*User)
	
	// Apply user-specific filtering
	filters = h.applyUserSearchFiltering(user, filters)

	// Fixed: Correct method call with proper variable assignment (2 return values)
	users, total, err := h.service.SearchUsers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"filters": filters,
	})
}

// Fix #2: Remove the undefined ListAllUsers method call and replace with SearchUsers
func (h *Handlers) ListAllUsers(c *gin.Context) {
	var filters UserSearchFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	user := currentUser.(*User)
	
	// Enterprise admins can see all users, others are restricted
	if !user.CanPerformCrossTenantOperation() {
		filters = h.applyUserSearchFiltering(user, filters)
	}

	// Use SearchUsers instead of the non-existent ListAllUsers
	users, total, err := h.service.SearchUsers(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
	})
}

// Fix #3: GetUserStats method signature (no parameters needed)
func (h *Handlers) GetUserStats(c *gin.Context) {
	// Fixed: Remove the UserStatsRequest parameter
	stats, err := h.service.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper method for user search filtering
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

// Update user profile (self-service)
func (h *Handlers) UpdateProfile(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	user := currentUser.(*User)
	
	var updates UserUpdates
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid update request",
			"details": err.Error(),
		})
		return
	}

	// Filter updates to only allow profile changes for regular users
	filtered := h.filterProfileUpdates(&updates)
	
	updatedUser, err := h.service.UpdateUser(c.Request.Context(), user.ID, filtered)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Current password is incorrect",
			})
		case ErrWeakPassword:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "New password does not meet security requirements",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profile",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": updatedUser.ToResponse(),
		"message": "Profile updated successfully",
	})
}

// Admin user management
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

	updatedUser, err := h.service.UpdateUser(c.Request.Context(), userID, updates)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		case ErrInvalidUserRole:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user role specified",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": updatedUser.ToResponse(),
		"message": "User updated successfully",
	})
}

// Filter profile updates to prevent privilege escalation
func (h *Handlers) filterProfileUpdates(updates *UserUpdates) UserUpdates {
	// Only allow basic profile information
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

// Authentication endpoints
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login request",
			"details": err.Error(),
		})
		return
	}

	response, err := h.service.Authenticate(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
		case ErrTenantAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to requested tenant",
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

func (h *Handlers) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No token provided",
		})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.service.Logout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// backend/internal/auth/handlers.go
package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService Service
}

func NewAuthHandler(authService Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}


type UserInfo struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Capture security info for logging
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Authenticate using the new auth service
	response, err := h.authService.Authenticate(c.Request.Context(), req.Email, req.Password)

	if err != nil {
		// Log security event (failed login attempt)
		log.Printf("LOGIN_FAILED: email=%s ip=%s ua=%s error=%v", 
			req.Email, clientIP, userAgent, err)

		// Map auth service errors to HTTP responses
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case ErrTenantAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to specified tenant"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	// Log successful login for security audit
	tenantID := ""
	if response.TenantContext != nil {
		tenantID = response.TenantContext.TenantID
	}
	log.Printf("LOGIN_SUCCESS: user=%s tenant=%s ip=%s", 
		response.User.Email, tenantID, clientIP)

	// Return clean response
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization header"})
		return
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
		return
	}

	// Log security event
	log.Printf("LOGOUT: ip=%s", c.ClientIP())

	// Logout using auth service
	err := h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		log.Printf("LOGOUT_ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// Get user info from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userInterface.(*User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	tenantID := c.GetString("tenant_id")

	c.JSON(http.StatusOK, gin.H{
		"user": UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
			TenantID: tenantID,
		},
	})
}

func (h *AuthHandler) RegisterCustomerContact(c *gin.Context) {
	var req struct {
		Email       string   `json:"email" binding:"required,email"`
		FullName    string   `json:"full_name" binding:"required"`
		Password    string   `json:"password" binding:"required"`
		TenantID    string   `json:"tenant_id" binding:"required"`
		CustomerID  int      `json:"customer_id" binding:"required"`
		YardAccess  []string `json:"yard_access"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.CreateCustomerContact(c.Request.Context(), &CreateCustomerContactRequest{
		Email:      req.Email,
		FullName:   req.FullName,
		Password:   req.Password,
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		YardAccess: []YardAccess{}, // Convert []string to []YardAccess if needed
	})

	if err != nil {
		switch err {
		case ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Customer contact created successfully",
		"user": UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
			TenantID: req.TenantID,
		},
	})
}

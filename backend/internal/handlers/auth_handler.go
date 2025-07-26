// backend/internal/handlers/auth_handler.go
package handlers

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, tenants, sessionID, err := h.authService.AuthenticateUser(loginRequest.Email, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"tenants":    tenants,
		"session_id": sessionID,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Invalidate session
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetUserTenants(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session required"})
		return
	}

	// TODO: Get user from session and return their tenants
	c.JSON(http.StatusOK, gin.H{"message": "User tenants endpoint"})
}

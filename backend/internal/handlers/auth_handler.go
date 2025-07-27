// backend/internal/handlers/auth_handler.go
package handlers

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/models"
)

type AuthHandler struct {
	authService    *auth.Service
	sessionManager *auth.TenantSessionManager
}

func NewAuthHandler(sessionManager *auth.TenantSessionManager) *AuthHandler {
	return &AuthHandler{
		authService:    auth.NewService(),
		sessionManager: sessionManager,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := c.GetHeader("Authorization")
	if sessionID != "" {
		h.sessionManager.DeleteSession(c.Request.Context(), sessionID)
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// Get user from context (set by middleware)
	if user, exists := c.Get("user"); exists {
		c.JSON(http.StatusOK, gin.H{"user": user})
		return
	}
	
	c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
}

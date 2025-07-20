package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/dcotelessa/oilgas-project/internal/auth"
)

type AuthHandler struct {
	sessionManager *auth.TenantSessionManager
}

func NewAuthHandler(sessionManager *auth.TenantSessionManager) *AuthHandler {
	return &AuthHandler{sessionManager: sessionManager}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required"`
		TenantSlug string `json:"tenant_slug,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Default tenant if not specified
	if req.TenantSlug == "" {
		req.TenantSlug = "default"
	}

	// Extract IP and User Agent
	ipAddr := auth.ExtractIP(c)
	userAgent := auth.ExtractUserAgent(c)

	session, err := h.sessionManager.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
		req.TenantSlug,
		ipAddr,
		userAgent,
	)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			c.JSON(401, gin.H{"error": "Invalid email or password"})
		case auth.ErrAccountLocked:
			c.JSON(423, gin.H{"error": "Account is locked due to too many failed attempts"})
		case auth.ErrNoTenantAccess:
			c.JSON(403, gin.H{"error": "No access to specified tenant"})
		default:
			c.JSON(500, gin.H{"error": "Login failed"})
		}
		return
	}

	// Set secure cookie
	c.SetCookie("session_id", session.ID, int(auth.SessionExpiration.Seconds()),
		"/", "", false, true)

	c.JSON(200, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":         session.UserID,
			"email":      session.Email,
			"role":       session.Role,
			"company":    session.Company,
			"tenant":     session.TenantSlug,
		},
		"session_expires": session.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := auth.ExtractSessionID(c)
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "No active session"})
		return
	}

	err := h.sessionManager.RevokeSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to logout"})
		return
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	sessionInterface, exists := c.Get("session")
	if !exists {
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	session := sessionInterface.(*auth.TenantSession)
	c.JSON(200, gin.H{
		"user": gin.H{
			"id":         session.UserID,
			"email":      session.Email,
			"role":       session.Role,
			"company":    session.Company,
			"tenant":     session.TenantSlug,
		},
		"session": gin.H{
			"expires_at": session.ExpiresAt,
			"ip_address": session.IPAddress,
		},
	})
}

// backend/internal/handlers/auth_handler.go
package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/auth"
	"oilgas-backend/internal/models"
)

type AuthHandler struct {
	authService    *auth.Service         // Use EXISTING auth service
	sessionManager *auth.SessionManager  // Use EXISTING session manager
}

func NewAuthHandler(authService *auth.Service, sessionManager *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Default tenant if not specified
	if req.TenantID == "" {
		req.TenantID = "default"
	}

	// Capture HTTP-layer security info (separate from business logic)
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Business logic: Pure authentication (no HTTP concerns)
	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		// Log security event (failed login attempt)
		log.Printf("LOGIN_FAILED: email=%s ip=%s ua=%s error=%v", 
			req.Email, clientIP, userAgent, err)

		switch err {
		case auth.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case auth.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case auth.ErrTenantNotFound:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to specified tenant"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	// Log successful login for security audit
	log.Printf("LOGIN_SUCCESS: user=%s tenant=%s ip=%s ua=%s", 
		resp.User.Email, resp.Tenant.Slug, clientIP, userAgent)

	// HTTP-layer concerns: Set secure cookie
	c.SetCookie("session_id", resp.SessionID, int(24*time.Hour.Seconds()),
		"/", "", false, true)

	// Return clean response
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":       resp.User.ID,
			"email":    resp.User.Email,
			"role":     resp.User.Role,
			"company":  resp.User.Company,
			"tenant":   resp.Tenant.Slug,
		},
		"session_expires": resp.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active session"})
		return
	}

	// Log security event
	log.Printf("LOGOUT: session=%s ip=%s", sessionID, c.ClientIP())

	// Use EXISTING session manager
	h.sessionManager.DeleteSession(sessionID)

	// Clear cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// Get user from context (set by existing auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"role":     user.Role,
			"company":  user.Company,
			"tenant":   c.GetString("tenant_id"),
		},
	})
}

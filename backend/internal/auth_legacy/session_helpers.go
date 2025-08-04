package auth

import (
	"strings"
	"github.com/gin-gonic/gin"
)

func ExtractSessionID(c *gin.Context) string {
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		return cookie
	}

	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

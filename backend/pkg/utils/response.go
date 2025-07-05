// backend/pkg/utils/response.go
package utils

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, err error) {
	c.JSON(status, APIResponse{
		Success: false,
		Error:   err.Error(),
	})
}

func MessageResponse(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
	})
}

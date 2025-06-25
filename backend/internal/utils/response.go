package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool                `json:"success"`
	Message    string              `json:"message,omitempty"`
	Data       interface{}         `json:"data"`
	Pagination *models.Pagination  `json:"pagination"`
	Error      interface{}         `json:"error,omitempty"`
	Timestamp  time.Time           `json:"timestamp"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, data interface{}) {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

// CreatedResponse sends a created response (201)
func CreatedResponse(c *gin.Context, data interface{}) {
	response := StandardResponse{
		Success:   true,
		Message:   "Resource created successfully",
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusCreated, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Include error details in development mode
	if gin.Mode() == gin.DebugMode && err != nil {
		response.Error = gin.H{
			"details": err.Error(),
		}
	}

	c.JSON(statusCode, response)
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, data interface{}, pagination *models.Pagination) {
	response := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, errors map[string]string) {
	response := StandardResponse{
		Success:   false,
		Message:   "Validation failed",
		Error:     errors,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusBadRequest, response)
}

// NotFoundResponse sends a not found response
func NotFoundResponse(c *gin.Context, resource string) {
	response := StandardResponse{
		Success:   false,
		Message:   resource + " not found",
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusNotFound, response)
}

// UnauthorizedResponse sends an unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusUnauthorized, response)
}

// ForbiddenResponse sends a forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusForbidden, response)
}

// ConflictResponse sends a conflict response
func ConflictResponse(c *gin.Context, message string) {
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusConflict, response)
}

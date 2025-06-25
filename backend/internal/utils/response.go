// backend/internal/utils/response.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	APIResponse
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Pages  int `json:"pages"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	APIResponse
	ValidationErrors []ValidationError `json:"validation_errors"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessResponse is an alias for Success (for backward compatibility)
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	Success(c, data, message)
}

// SuccessWithPagination sends a successful paginated response
func SuccessWithPagination(c *gin.Context, data interface{}, total, limit, offset int, message string) {
	pages := (total + limit - 1) / limit // Calculate total pages
	
	c.JSON(http.StatusOK, PaginatedResponse{
		APIResponse: APIResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Pages:  pages,
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Success: false,
		Message: message,
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	c.JSON(statusCode, response)
}

// ErrorResponse is an alias for Error (for backward compatibility)
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	Error(c, statusCode, message, err)
}

// ValidationError sends a validation error response
func ValidationErrors(c *gin.Context, errors []ValidationError, message string) {
	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		APIResponse: APIResponse{
			Success: false,
			Message: message,
		},
		ValidationErrors: errors,
	})
}

// NotFound sends a 404 response
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, resource+" not found", nil)
}

// BadRequest sends a 400 response
func BadRequest(c *gin.Context, message string, err error) {
	Error(c, http.StatusBadRequest, message, err)
}

// InternalServerError sends a 500 response
func InternalServerError(c *gin.Context, message string, err error) {
	Error(c, http.StatusInternalServerError, message, err)
}

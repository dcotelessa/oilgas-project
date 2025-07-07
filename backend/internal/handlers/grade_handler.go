// backend/internal/handlers/grade_handler.go
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
	"oilgas-backend/pkg/validation"
)

type GradeHandler struct {
	gradeService services.GradeService
}

func NewGradeHandler(gradeService services.GradeService) *GradeHandler {
	return &GradeHandler{
		gradeService: gradeService,
	}
}

func (h *GradeHandler) GetGrades(c *gin.Context) {
	grades, err := h.gradeService.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve grades", err)
		return
	}

	utils.SuccessResponse(c, grades, "Grades retrieved successfully")
}

func (h *GradeHandler) CreateGrade(c *gin.Context) {
	var req validation.GradeValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate grade data
	if err := req.Validate(); err != nil {
		utils.BadRequest(c, "Validation failed", err)
		return
	}

	// Create grade through service
	grade, err := h.gradeService.Create(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Grade already exists",
			})
			return
		}
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create grade", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Grade created successfully",
		"grade":   grade,
	})
}

func (h *GradeHandler) DeleteGrade(c *gin.Context) {
	gradeName := c.Param("grade")
	if gradeName == "" {
		utils.BadRequest(c, "Grade name is required", nil)
		return
	}

	// Delete grade through service
	err := h.gradeService.Delete(c.Request.Context(), gradeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Grade")
			return
		}
		if strings.Contains(err.Error(), "in use") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Cannot delete grade that is currently in use",
			})
			return
		}
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete grade", err)
		return
	}

	utils.SuccessResponse(c, nil, "Grade deleted successfully")
}

func (h *GradeHandler) GetGradeUsage(c *gin.Context) {
	gradeName := c.Param("grade")
	if gradeName == "" {
		utils.BadRequest(c, "Grade name is required", nil)
		return
	}

	// Fixed: Use GetUsageStats instead of GetUsage
	usage, err := h.gradeService.GetUsageStats(c.Request.Context(), gradeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Grade")
			return
		}
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get grade usage", err)
		return
	}

	utils.SuccessResponse(c, usage, "Grade usage retrieved successfully")
}

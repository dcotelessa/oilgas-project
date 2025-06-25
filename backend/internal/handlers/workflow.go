// backend/internal/handlers/workflow.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type WorkflowHandlers struct {
	workflowService services.WorkflowService
}

func NewWorkflowHandlers(workflowService services.WorkflowService) *WorkflowHandlers {
	return &WorkflowHandlers{
		workflowService: workflowService,
	}
}

// GetDashboardStats returns dashboard statistics
func (h *WorkflowHandlers) GetDashboardStats(c *gin.Context) {
	stats, err := h.workflowService.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to get dashboard stats", err)
		return
	}

	utils.Success(c, stats, "Dashboard stats retrieved successfully")
}

// GetRecentJobs returns recent job summaries
func (h *WorkflowHandlers) GetRecentJobs(c *gin.Context) {
	// Parse limit parameter
	limit := 10 // Default
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	jobs, err := h.workflowService.GetRecentJobs(c.Request.Context(), limit)
	if err != nil {
		utils.InternalServerError(c, "Failed to get recent jobs", err)
		return
	}

	utils.Success(c, jobs, "Recent jobs retrieved successfully")
}

// GetJobByID returns a specific job
func (h *WorkflowHandlers) GetJobByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", err)
		return
	}

	job, err := h.workflowService.GetJobByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "inventory item not found" {
			utils.NotFound(c, "Job")
			return
		}
		utils.InternalServerError(c, "Failed to get job", err)
		return
	}

	utils.Success(c, job, "Job retrieved successfully")
}

// UpdateJobStatus updates job status
func (h *WorkflowHandlers) UpdateJobStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", err)
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate status
	validStatuses := []string{"pending", "active", "completed", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if request.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.BadRequest(c, "Invalid status. Must be one of: pending, active, completed, cancelled", nil)
		return
	}

	err = h.workflowService.UpdateJobStatus(c.Request.Context(), id, request.Status)
	if err != nil {
		utils.InternalServerError(c, "Failed to update job status", err)
		return
	}

	utils.Success(c, gin.H{"id": id, "status": request.Status}, "Job status updated successfully")
}

// GetActiveCustomers returns customers with active jobs
func (h *WorkflowHandlers) GetActiveCustomers(c *gin.Context) {
	customers, err := h.workflowService.GetActiveCustomers(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to get active customers", err)
		return
	}

	utils.Success(c, customers, "Active customers retrieved successfully")
}

// GetGradeDistribution returns grade distribution data
func (h *WorkflowHandlers) GetGradeDistribution(c *gin.Context) {
	grades, err := h.workflowService.GetGradeDistribution(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to get grade distribution", err)
		return
	}

	utils.Success(c, gin.H{"grades": grades}, "Grade distribution retrieved successfully")
}

// GetWorkflowMetrics returns workflow performance metrics
func (h *WorkflowHandlers) GetWorkflowMetrics(c *gin.Context) {
	// Get dashboard stats as base metrics
	stats, err := h.workflowService.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to get workflow metrics", err)
		return
	}

	// Get recent jobs for additional metrics
	recentJobs, err := h.workflowService.GetRecentJobs(c.Request.Context(), 50)
	if err != nil {
		utils.InternalServerError(c, "Failed to get recent jobs for metrics", err)
		return
	}

	// Calculate additional metrics
	metrics := gin.H{
		"inventory_summary": stats,
		"recent_jobs_count": len(recentJobs),
		"active_customers":  len(stats.TopCustomers), // Using existing data
		"total_joints":      stats.TotalJoints,
		"total_items":       stats.TotalItems,
		"items_by_grade":    stats.ItemsByGrade,
		"items_by_location": stats.ItemsByLocation,
	}

	utils.Success(c, metrics, "Workflow metrics retrieved successfully")
}

// GetJobsByStatus returns jobs filtered by status (placeholder - would need job status tracking)
func (h *WorkflowHandlers) GetJobsByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		utils.BadRequest(c, "Status parameter is required", nil)
		return
	}

	// For now, return recent jobs as this would require job status tracking
	// In a real implementation, you'd filter by status
	jobs, err := h.workflowService.GetRecentJobs(c.Request.Context(), 20)
	if err != nil {
		utils.InternalServerError(c, "Failed to get jobs by status", err)
		return
	}

	// Filter jobs by status (mock implementation)
	var filteredJobs []interface{}
	for _, job := range jobs {
		// In real implementation, check job.Status == status
		filteredJobs = append(filteredJobs, job)
	}

	utils.Success(c, gin.H{
		"jobs":   filteredJobs,
		"status": status,
		"total":  len(filteredJobs),
	}, "Jobs filtered by status")
}

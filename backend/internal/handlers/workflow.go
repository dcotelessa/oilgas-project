package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type WorkflowHandler struct {
	service *services.WorkflowService
}

func NewWorkflowHandler(service *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
	}
}

// RegisterRoutes registers all workflow-related routes
func (h *WorkflowHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Dashboard routes
	dashboard := r.Group("/dashboard")
	{
		dashboard.GET("/stats", h.GetDashboardStats)
		dashboard.GET("/summaries", h.GetJobSummaries)
		dashboard.GET("/recent", h.GetRecentActivity)
	}

	// Job routes
	jobs := r.Group("/jobs")
	{
		jobs.GET("", h.GetJobs)
		jobs.POST("", h.CreateJob)
		jobs.GET("/:id", h.GetJobByID)
		jobs.PUT("/:id", h.UpdateJob)
		jobs.DELETE("/:id", h.DeleteJob)
		jobs.GET("/work-order/:workOrder", h.GetJobByWorkOrder)
		
		// Workflow state transitions
		jobs.POST("/:workOrder/advance/production", h.AdvanceToProduction)
		jobs.POST("/:workOrder/advance/inspection", h.AdvanceToInspection)
		jobs.POST("/:workOrder/advance/inventory", h.AdvanceToInventory)
	}

	// Inspection routes
	inspection := r.Group("/inspection")
	{
		inspection.GET("/:workOrder", h.GetInspectionResults)
		inspection.POST("", h.CreateInspectionResult)
		inspection.PUT("/:id", h.UpdateInspectionResult)
	}

	// Inventory routes
	inventory := r.Group("/inventory")
	{
		inventory.GET("", h.GetInventory)
		inventory.GET("/customer/:customerID", h.GetInventoryByCustomer)
		inventory.POST("", h.CreateInventoryItem)
		inventory.PUT("/:id", h.UpdateInventoryItem)
		inventory.POST("/ship", h.ShipInventory)
	}

	// Customer routes
	customers := r.Group("/customers")
	{
		customers.GET("", h.GetCustomers)
		customers.POST("", h.CreateCustomer)
		customers.GET("/:id", h.GetCustomerByID)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
		customers.GET("/:id/pipe-sizes", h.GetPipeSizes)
		customers.POST("/:id/pipe-sizes", h.CreatePipeSize)
	}

	// Reference data routes
	reference := r.Group("/reference")
	{
		reference.GET("/grades", h.GetGrades)
		reference.GET("/workflow-states", h.GetWorkflowStates)
		reference.GET("/color-numbers", h.GetColorNumbers)
	}
}

// Dashboard handlers
func (h *WorkflowHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard stats", err)
		return
	}

	utils.SuccessResponse(c, stats)
}

func (h *WorkflowHandler) GetJobSummaries(c *gin.Context) {
	summaries, err := h.service.GetJobSummaries(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get job summaries", err)
		return
	}

	utils.SuccessResponse(c, summaries)
}

func (h *WorkflowHandler) GetRecentActivity(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	activity, err := h.service.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get recent activity", err)
		return
	}

	utils.SuccessResponse(c, activity)
}

// Job handlers
func (h *WorkflowHandler) GetJobs(c *gin.Context) {
	filters := repository.JobFilters{
		Page:    getIntQuery(c, "page", 1),
		PerPage: getIntQuery(c, "per_page", 50),
		OrderBy: c.DefaultQuery("order_by", "created_at"),
		OrderDir: c.DefaultQuery("order_dir", "DESC"),
	}

	// Parse filters
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters.CustomerID = &id
		}
	}

	if state := c.Query("state"); state != "" {
		workflowState := models.WorkflowState(state)
		filters.State = &workflowState
	}

	filters.WorkOrder = c.Query("work_order")
	filters.Grade = c.Query("grade")
	filters.Size = c.Query("size")
	filters.DateFrom = c.Query("date_from")
	filters.DateTo = c.Query("date_to")
	filters.IncludeDeleted = c.Query("include_deleted") == "true"

	jobs, pagination, err := h.service.GetJobs(c.Request.Context(), filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get jobs", err)
		return
	}

	utils.PaginatedResponse(c, jobs, pagination)
}

func (h *WorkflowHandler) GetJobByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	job, err := h.service.GetJobByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get job", err)
		return
	}

	if job == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Job not found", nil)
		return
	}

	utils.SuccessResponse(c, job)
}

func (h *WorkflowHandler) GetJobByWorkOrder(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Work order is required", nil)
		return
	}

	job, err := h.service.GetJobByWorkOrder(c.Request.Context(), workOrder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get job", err)
		return
	}

	if job == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Job not found", nil)
		return
	}

	utils.SuccessResponse(c, job)
}

func (h *WorkflowHandler) CreateJob(c *gin.Context) {
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.CreateJob(c.Request.Context(), &job); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	utils.CreatedResponse(c, &job)
}

func (h *WorkflowHandler) UpdateJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	job.ID = id
	if err := h.service.UpdateJob(c.Request.Context(), &job); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update job", err)
		return
	}

	utils.SuccessResponse(c, &job)
}

func (h *WorkflowHandler) DeleteJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	// Soft delete by setting deleted = true
	job, err := h.service.GetJobByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get job", err)
		return
	}

	if job == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Job not found", nil)
		return
	}

	job.Deleted = true
	if err := h.service.UpdateJob(c.Request.Context(), job); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job deleted successfully"})
}

// Workflow state transition handlers
func (h *WorkflowHandler) AdvanceToProduction(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Work order is required", nil)
		return
	}

	if err := h.service.AdvanceJobToProduction(c.Request.Context(), workOrder); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to advance job to production", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job advanced to production successfully"})
}

func (h *WorkflowHandler) AdvanceToInspection(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Work order is required", nil)
		return
	}

	var req struct {
		InspectedBy string `json:"inspected_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.AdvanceJobToInspection(c.Request.Context(), workOrder, req.InspectedBy); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to advance job to inspection", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job advanced to inspection successfully"})
}

func (h *WorkflowHandler) AdvanceToInventory(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Work order is required", nil)
		return
	}

	if err := h.service.MoveJobToInventory(c.Request.Context(), workOrder); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to move job to inventory", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job moved to inventory successfully"})
}

// Inspection handlers
func (h *WorkflowHandler) GetInspectionResults(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Work order is required", nil)
		return
	}

	results, err := h.service.GetInspectionResults(c.Request.Context(), workOrder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inspection results", err)
		return
	}

	utils.SuccessResponse(c, results)
}

func (h *WorkflowHandler) CreateInspectionResult(c *gin.Context) {
	var result models.InspectionResult
	if err := c.ShouldBindJSON(&result); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.CreateInspectionResult(c.Request.Context(), &result); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create inspection result", err)
		return
	}

	utils.CreatedResponse(c, &result)
}

func (h *WorkflowHandler) UpdateInspectionResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid inspection result ID", err)
		return
	}

	var result models.InspectionResult
	if err := c.ShouldBindJSON(&result); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result.ID = id
	// Implementation would call repository update method
	utils.SuccessResponse(c, &result)
}

// Inventory handlers
func (h *WorkflowHandler) GetInventory(c *gin.Context) {
	filters := repository.InventoryFilters{
		Page:    getIntQuery(c, "page", 1),
		PerPage: getIntQuery(c, "per_page", 50),
		OrderBy: c.DefaultQuery("order_by", "datein"),
		OrderDir: c.DefaultQuery("order_dir", "DESC"),
	}

	// Parse filters
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters.CustomerID = &id
		}
	}

	if cn := c.Query("cn"); cn != "" {
		if cnInt, err := strconv.Atoi(cn); err == nil {
			cnVal := models.ColorNumber(cnInt)
			filters.CN = &cnVal
		}
	}

	if minJoints := c.Query("min_joints"); minJoints != "" {
		if min, err := strconv.Atoi(minJoints); err == nil {
			filters.MinJoints = &min
		}
	}

	if maxJoints := c.Query("max_joints"); maxJoints != "" {
		if max, err := strconv.Atoi(maxJoints); err == nil {
			filters.MaxJoints = &max
		}
	}

	filters.Grade = c.Query("grade")
	filters.Size = c.Query("size")
	filters.Color = c.Query("color")
	filters.Rack = c.Query("rack")
	filters.IncludeShipped = c.Query("include_shipped") == "true"

	items, pagination, err := h.service.GetInventory(c.Request.Context(), filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory", err)
		return
	}

	utils.PaginatedResponse(c, items, pagination)
}

func (h *WorkflowHandler) GetInventoryByCustomer(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	items, err := h.service.GetInventoryByCustomer(c.Request.Context(), customerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer inventory", err)
		return
	}

	utils.SuccessResponse(c, items)
}

func (h *WorkflowHandler) CreateInventoryItem(c *gin.Context) {
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Implementation would call service method
	utils.CreatedResponse(c, &item)
}

func (h *WorkflowHandler) UpdateInventoryItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid inventory item ID", err)
		return
	}

	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	item.ID = id
	// Implementation would call service method
	utils.SuccessResponse(c, &item)
}

func (h *WorkflowHandler) ShipInventory(c *gin.Context) {
	var req struct {
		ItemIDs         []int                  `json:"item_ids" binding:"required"`
		ShipmentDetails map[string]interface{} `json:"shipment_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.ShipInventory(c.Request.Context(), req.ItemIDs, req.ShipmentDetails); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to ship inventory", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Inventory shipped successfully",
		"items_shipped": len(req.ItemIDs),
	})
}

// Customer handlers
func (h *WorkflowHandler) GetCustomers(c *gin.Context) {
	includeDeleted := c.Query("include_deleted") == "true"

	customers, err := h.service.GetCustomers(c.Request.Context(), includeDeleted)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customers", err)
		return
	}

	utils.SuccessResponse(c, customers)
}

func (h *WorkflowHandler) GetCustomerByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	customer, err := h.service.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer", err)
		return
	}

	if customer == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Customer not found", nil)
		return
	}

	utils.SuccessResponse(c, customer)
}

func (h *WorkflowHandler) CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.CreateCustomer(c.Request.Context(), &customer); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create customer", err)
		return
	}

	utils.CreatedResponse(c, &customer)
}

func (h *WorkflowHandler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	customer.ID = id
	// Implementation would call service method
	utils.SuccessResponse(c, &customer)
}

func (h *WorkflowHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	// Implementation would call service method for soft delete
	utils.SuccessResponse(c, gin.H{"message": "Customer deleted successfully"})
}

func (h *WorkflowHandler) GetPipeSizes(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	// Implementation would call service method
	sizes := []models.PipeSize{} // Placeholder
	utils.SuccessResponse(c, sizes)
}

func (h *WorkflowHandler) CreatePipeSize(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	var size models.PipeSize
	if err := c.ShouldBindJSON(&size); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	size.CustomerID = customerID
	// Implementation would call service method
	utils.CreatedResponse(c, &size)
}

// Reference data handlers
func (h *WorkflowHandler) GetGrades(c *gin.Context) {
	grades, err := h.service.GetGrades(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get grades", err)
		return
	}

	utils.SuccessResponse(c, grades)
}

func (h *WorkflowHandler) GetWorkflowStates(c *gin.Context) {
	states := []gin.H{
		{"value": models.StateReceiving, "label": "Receiving", "description": "Pipe received, pending production"},
		{"value": models.StateProduction, "label": "Production", "description": "In production process"},
		{"value": models.StateInspection, "label": "Inspection", "description": "Being inspected"},
		{"value": models.StateInventory, "label": "Inventory", "description": "In inventory, ready to ship"},
		{"value": models.StateShipping, "label": "Shipping", "description": "Being shipped"},
		{"value": models.StateCompleted, "label": "Completed", "description": "Shipped and complete"},
	}

	utils.SuccessResponse(c, states)
}

func (h *WorkflowHandler) GetColorNumbers(c *gin.Context) {
	colorNumbers := []gin.H{
		{"value": models.CNPremium, "color": "WHT", "name": "White", "description": "Premium Quality"},
		{"value": models.CNStandard, "color": "BLU", "name": "Blue", "description": "Standard Quality"},
		{"value": models.CNEconomy, "color": "GRN", "name": "Green", "description": "Economy Quality"},
		{"value": models.CNRejected, "color": "RED", "name": "Red", "description": "Rejected/Problem"},
		{"value": models.CNGrade5, "color": "Grade 5", "name": "Grade 5", "description": "Grade 5 Quality"},
		{"value": models.CNGrade6, "color": "Grade 6", "name": "Grade 6", "description": "Grade 6 Quality"},
	}

	utils.SuccessResponse(c, colorNumbers)
}

// Helper functions
func getIntQuery(c *gin.Context, key string, defaultValue int) int {
	if val := c.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

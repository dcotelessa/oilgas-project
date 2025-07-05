// backend/internal/handlers/handlers.go - Updated to match actual services
package handlers

import (
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
)

type Handlers struct {
	Analytics     *AnalyticsHandler
	Customer      *CustomerHandler
	Grade         *GradeHandler
	Inventory     *InventoryHandler
	Received      *ReceivedHandler
	WorkflowState *WorkflowStateHandler
	Search        *SearchHandler
	System        *SystemHandler
}

func New(services *services.Services) *Handlers {
	return &Handlers{
		Analytics:     NewAnalyticsHandler(services.Analytics),
		Customer:      NewCustomerHandler(services.Customer),
		Grade:         NewGradeHandler(services.Grade),
		Inventory:     NewInventoryHandler(services.Inventory),
		Received:      NewReceivedHandler(services.Received),
		WorkflowState: NewWorkflowStateHandler(services.WorkflowState),
		Search:        NewSearchHandler(services.Search),
		System:        NewSystemHandler(services),
	}
}

// RegisterRoutes sets up all API routes using your actual repository structure
func (h *Handlers) RegisterRoutes(r *gin.RouterGroup) {
	// Analytics routes - dashboard, reporting, metrics
	analytics := r.Group("/analytics")
	{
		analytics.GET("/dashboard", h.Analytics.GetDashboardStats)
		analytics.GET("/customer-activity", h.Analytics.GetCustomerActivity)
		analytics.GET("/inventory-analytics", h.Analytics.GetInventoryAnalytics)
		analytics.GET("/grade-distribution", h.Analytics.GetGradeDistribution)
		analytics.GET("/location-utilization", h.Analytics.GetLocationUtilization)
	}

	// Customer routes - existing structure
	customers := r.Group("/customers")
	{
		customers.GET("", h.Customer.GetCustomers)
		customers.POST("", h.Customer.CreateCustomer)
		customers.GET("/:id", h.Customer.GetCustomer)
		customers.PUT("/:id", h.Customer.UpdateCustomer)
		customers.DELETE("/:id", h.Customer.DeleteCustomer)
		customers.GET("/search", h.Customer.SearchCustomers)
	}

	// Grade routes - existing structure  
	grades := r.Group("/grades")
	{
		grades.GET("", h.Grade.GetGrades)
		grades.POST("", h.Grade.CreateGrade)
		grades.DELETE("/:grade", h.Grade.DeleteGrade)
		grades.GET("/:grade/usage", h.Grade.GetGradeUsage)
	}

	// Inventory routes - existing structure
	inventory := r.Group("/inventory")
	{
		inventory.GET("", h.Inventory.GetInventory)
		inventory.POST("", h.Inventory.CreateInventoryItem)
		inventory.GET("/:id", h.Inventory.GetInventoryItem)
		inventory.PUT("/:id", h.Inventory.UpdateInventoryItem)
		inventory.DELETE("/:id", h.Inventory.DeleteInventoryItem)
		inventory.GET("/summary", h.Inventory.GetInventorySummary)
		inventory.GET("/search", h.Inventory.SearchInventory)
	}

	// Received routes - new based on your repository
	received := r.Group("/received")
	{
		received.GET("", h.Received.GetReceived)
		received.POST("", h.Received.CreateReceived)
		received.GET("/:id", h.Received.GetReceivedItem)
		received.PUT("/:id", h.Received.UpdateReceived)
		received.DELETE("/:id", h.Received.DeleteReceived)
		received.GET("/work-order/:workOrder", h.Received.GetByWorkOrder)
		received.PUT("/:id/status", h.Received.UpdateStatus)
		received.GET("/pending-inspection", h.Received.GetPendingInspection)
	}

	// Workflow state routes - new based on your repository
	workflow := r.Group("/workflow")
	{
		workflow.GET("/:workOrder/state", h.WorkflowState.GetCurrentState)
		workflow.POST("/:workOrder/transition", h.WorkflowState.TransitionTo)
		workflow.GET("/:workOrder/history", h.WorkflowState.GetStateHistory)
		workflow.GET("/state/:state/items", h.WorkflowState.GetItemsByState)
		workflow.POST("/validate-transition", h.WorkflowState.ValidateTransition)
	}

	// Search routes - consolidated search functionality
	search := r.Group("/search")
	{
		search.GET("/global", h.Search.GlobalSearch)
		search.GET("/customers", h.Search.SearchCustomers)
		search.GET("/inventory", h.Search.SearchInventory)
		search.GET("/received", h.Search.SearchReceived)
		search.GET("/suggestions", h.Search.GetSuggestions)
	}

	// System routes - health, cache, metrics
	system := r.Group("/system")
	{
		system.GET("/cache/stats", h.System.GetCacheStats)
		system.POST("/cache/clear", h.System.ClearCache)
		system.GET("/health", h.System.GetSystemHealth)
		system.GET("/metrics", h.System.GetMetrics)
	}
}

// Handler constructor functions - these need to be implemented in separate files

func NewAnalyticsHandler(service services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func NewReceivedHandler(service services.ReceivedService) *ReceivedHandler {
	return &ReceivedHandler{service: service}
}

func NewWorkflowStateHandler(service services.WorkflowStateService) *WorkflowStateHandler {
	return &WorkflowStateHandler{service: service}
}

func NewSearchHandler(service services.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func NewSystemHandler(services *services.Services) *SystemHandler {
	return &SystemHandler{services: services}
}

// Handler structs - these would be in separate files
type AnalyticsHandler struct {
	service services.AnalyticsService
}

type ReceivedHandler struct {
	service services.ReceivedService
}

type WorkflowStateHandler struct {
	service services.WorkflowStateService
}

type SearchHandler struct {
	service services.SearchService
}

type SystemHandler struct {
	services *services.Services
}

// Example implementation for ReceivedHandler methods
func (h *ReceivedHandler) GetReceived(c *gin.Context) {
	// Parse filters and pagination
	filters := parseReceivedFilters(c)
	limit := getIntQuery(c, "limit", 50)
	offset := getIntQuery(c, "offset", 0)

	items, total, err := h.service.GetAll(c.Request.Context(), filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get received items", err)
		return
	}

	utils.SuccessWithPagination(c, items, total, limit, offset, "Retrieved received items")
}

func (h *ReceivedHandler) CreateReceived(c *gin.Context) {
	var item models.ReceivedItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	err := h.service.Create(c.Request.Context(), &item)
	if err != nil {
		utils.InternalServerError(c, "Failed to create received item", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Received item created successfully",
		"item":    item,
	})
}

func (h *ReceivedHandler) GetReceivedItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerError(c, "Failed to get received item", err)
		return
	}

	if item == nil {
		utils.NotFound(c, "Received item")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"item": item})
}

func (h *ReceivedHandler) UpdateReceived(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	var item models.ReceivedItem
	if err := c.ShouldBindJSON(&item); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	item.ID = id
	err = h.service.Update(c.Request.Context(), &item)
	if err != nil {
		utils.InternalServerError(c, "Failed to update received item", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Received item updated successfully",
		"item":    item,
	})
}

func (h *ReceivedHandler) DeleteReceived(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerError(c, "Failed to delete received item", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Received item deleted successfully",
	})
}

func (h *ReceivedHandler) GetByWorkOrder(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	item, err := h.service.GetByWorkOrder(c.Request.Context(), workOrder)
	if err != nil {
		utils.InternalServerError(c, "Failed to get received item", err)
		return
	}

	if item == nil {
		utils.NotFound(c, "Received item")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"item": item})
}

func (h *ReceivedHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	err = h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		utils.InternalServerError(c, "Failed to update status", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

func (h *ReceivedHandler) GetPendingInspection(c *gin.Context) {
	items, err := h.service.GetPendingInspection(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "Failed to get pending inspection items", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

// Helper functions for parsing requests
func parseReceivedFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})
	
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters["customer_id"] = id
		}
	}
	
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	
	if grade := c.Query("grade"); grade != "" {
		filters["grade"] = grade
	}
	
	if size := c.Query("size"); size != "" {
		filters["size"] = size
	}
	
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}
	
	return filters
}

func getIntQuery(c *gin.Context, key string, defaultValue int) int {
	if val := c.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// backend/internal/handlers/handlers.go
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

// RegisterRoutes sets up all API routes
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

	// Customer routes
	customers := r.Group("/customers")
	{
		customers.GET("", h.Customer.GetCustomers)
		customers.POST("", h.Customer.CreateCustomer)
		customers.GET("/:id", h.Customer.GetCustomer)
		customers.PUT("/:id", h.Customer.UpdateCustomer)
		customers.DELETE("/:id", h.Customer.DeleteCustomer)
		// Note: SearchCustomers method needs to be implemented in customer_handler.go
		// customers.GET("/search", h.Customer.SearchCustomers)
	}

	// Grade routes  
	grades := r.Group("/grades")
	{
		grades.GET("", h.Grade.GetGrades)
		grades.POST("", h.Grade.CreateGrade)
		grades.DELETE("/:grade", h.Grade.DeleteGrade)
		grades.GET("/:grade/usage", h.Grade.GetGradeUsage)
	}

	// Inventory routes
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

	// Received routes
	received := r.Group("/received")
	{
		received.GET("", h.Received.GetReceived)
		received.POST("", h.Received.CreateReceived)
		received.GET("/:id", h.Received.GetReceivedItem)
		received.PUT("/:id", h.Received.UpdateReceived)
		received.DELETE("/:id", h.Received.DeleteReceived)
		received.GET("/work-order/:workOrder", h.Received.GetByWorkOrder)
		received.PUT("/:id/status", h.Received.UpdateStatus)
		received.POST("/:id/transition", h.Received.TransitionStatus)
		received.GET("/pending-inspection", h.Received.GetPendingInspection)
	}

	// Workflow state routes
	workflow := r.Group("/workflow")
	{
	    workflow.GET("/:workOrder/state", h.WorkflowState.GetCurrentState)
	    workflow.POST("/:workOrder/transition", h.WorkflowState.TransitionTo)
	    workflow.GET("/:workOrder/history", h.WorkflowState.GetStateHistory)
	    workflow.GET("/state/:state/items", h.WorkflowState.GetItemsByState)
	    workflow.GET("/state/:state/jobs", h.WorkflowState.GetJobsByState)
	    workflow.POST("/validate-transition", h.WorkflowState.ValidateTransition)
	}

	// Search routes
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

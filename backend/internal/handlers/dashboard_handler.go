package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type DashboardHandler struct {
	inventoryService services.InventoryService
	customerService  services.CustomerService
}

func NewDashboardHandler(inventoryService services.InventoryService, customerService services.CustomerService) *DashboardHandler {
	return &DashboardHandler{
		inventoryService: inventoryService,
		customerService:  customerService,
	}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// Get inventory summary
	inventorySummary, err := h.inventoryService.GetSummary(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory summary", err)
		return
	}

	// Get customers count - using a simple approach since we don't have a Count method
	customers, err := h.customerService.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customers", err)
		return
	}

	// Build dashboard response
	dashboard := gin.H{
		"inventory": inventorySummary,
		"customers": gin.H{
			"total": len(customers),
		},
		"summary": gin.H{
			"total_items":     inventorySummary.TotalItems,
			"total_joints":    inventorySummary.TotalJoints,
			"total_customers": len(customers),
			"active_grades":   len(inventorySummary.ItemsByGrade),
		},
	}

	utils.SuccessResponse(c, dashboard, "Dashboard data retrieved successfully")
}

func (h *DashboardHandler) GetCustomerWorkflow(c *gin.Context) {
	customerIDParam := c.Param("customerID")
	customerID, err := strconv.Atoi(customerIDParam)
	if err != nil {
		utils.BadRequest(c, "Invalid customer ID", err)
		return
	}

	// Get customer details
	customer, err := h.customerService.GetByID(c.Request.Context(), customerIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer", err)
		return
	}

	// Get customer's inventory using filters
	filters := map[string]interface{}{
		"customer_id": customerID,
	}
	
	items, total, err := h.inventoryService.GetFiltered(c.Request.Context(), filters, 100, 0)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer inventory", err)
		return
	}

	// Build workflow summary
	summary := gin.H{
		"customer":        customer,
		"total_items":     total,
		"items":           items,
		"grade_breakdown": buildGradeBreakdown(items),
	}

	utils.SuccessResponse(c, summary, "Customer workflow retrieved successfully")
}

// Helper function to build grade breakdown
func buildGradeBreakdown(items []models.InventoryItem) map[string]int {
	breakdown := make(map[string]int)
	
	for _, item := range items {
		if item.Grade != "" {
			breakdown[item.Grade]++
		}
	}
	
	return breakdown
}

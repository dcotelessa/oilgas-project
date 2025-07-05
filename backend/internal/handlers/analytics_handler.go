package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type AnalyticsHandler struct {
	inventoryService services.InventoryService
	customerService  services.CustomerService
}

func NewAnalyticsHandler(inventoryService services.InventoryService, customerService services.CustomerService) *AnalyticsHandler {
	return &AnalyticsHandler{
		inventoryService: inventoryService,
		customerService:  customerService,
	}
}

func (h *AnalyticsHandler) GetInventorySummary(c *gin.Context) {
	summary, err := h.inventoryService.GetSummary(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory summary", err)
		return
	}

	utils.SuccessResponse(c, summary, "Inventory summary retrieved successfully")
}

func (h *AnalyticsHandler) GetCustomerActivity(c *gin.Context) {
	// Parse date range
	days := 30 // default
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Parse customer ID filter (optional)
	var customerIDFilter *int
	if custParam := c.Query("customer_id"); custParam != "" {
		if id, err := strconv.Atoi(custParam); err == nil {
			customerIDFilter = &id
		}
	}

	// Get all customers for now - in a real implementation you'd have a more efficient method
	customers, err := h.customerService.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customers", err)
		return
	}

	// Build activity summary
	activity := gin.H{
		"period_days":    days,
		"customer_count": len(customers),
		"generated_at":   time.Now().UTC(),
	}

	if customerIDFilter != nil {
		activity["customer_id"] = *customerIDFilter
		// Add customer-specific activity here
	}

	utils.SuccessResponse(c, activity, "Customer activity retrieved successfully")
}

func (h *AnalyticsHandler) GetTopCustomers(c *gin.Context) {
	// Parse limit
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Parse time period
	days := 30
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Get customers - simplified implementation
	customers, err := h.customerService.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customers", err)
		return
	}

	// Take only the requested limit
	if len(customers) > limit {
		customers = customers[:limit]
	}

	result := gin.H{
		"top_customers": customers,
		"limit":         limit,
		"period_days":   days,
	}

	utils.SuccessResponse(c, result, "Top customers retrieved successfully")
}

func (h *AnalyticsHandler) GetGradeAnalytics(c *gin.Context) {
	// Get inventory summary which contains grade distribution
	summary, err := h.inventoryService.GetSummary(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory summary", err)
		return
	}

	analytics := gin.H{
		"grade_distribution": summary.ItemsByGrade,
		"total_grades":       len(summary.ItemsByGrade),
		"generated_at":       time.Now().UTC(),
	}

	utils.SuccessResponse(c, analytics, "Grade analytics retrieved successfully")
}

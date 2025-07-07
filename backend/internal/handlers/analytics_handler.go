// backend/internal/handlers/analytics_handler.go
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
	service services.AnalyticsService
}

func NewAnalyticsHandler(service services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard stats", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"dashboard":    stats,
		"generated_at": time.Now().UTC(),
	}, "Dashboard stats retrieved successfully")
}

func (h *AnalyticsHandler) GetCustomerActivity(c *gin.Context) {
	// Parse parameters
	days := 30 // default
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	var customerID *int
	if custParam := c.Query("customer_id"); custParam != "" {
		if id, err := strconv.Atoi(custParam); err == nil {
			customerID = &id
		}
	}

	activity, err := h.service.GetCustomerActivity(c.Request.Context(), days, customerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer activity", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"activity":    activity,
		"period_days": days,
		"customer_id": customerID,
	}, "Customer activity retrieved successfully")
}

func (h *AnalyticsHandler) GetInventoryAnalytics(c *gin.Context) {
	analytics, err := h.service.GetInventoryAnalytics(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory analytics", err)
		return
	}

	utils.SuccessResponse(c, analytics, "Inventory analytics retrieved successfully")
}

func (h *AnalyticsHandler) GetGradeDistribution(c *gin.Context) {
	distribution, err := h.service.GetGradeDistribution(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get grade distribution", err)
		return
	}

	utils.SuccessResponse(c, distribution, "Grade distribution retrieved successfully")
}

func (h *AnalyticsHandler) GetLocationUtilization(c *gin.Context) {
	utilization, err := h.service.GetLocationUtilization(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get location utilization", err)
		return
	}

	utils.SuccessResponse(c, utilization, "Location utilization retrieved successfully")
}

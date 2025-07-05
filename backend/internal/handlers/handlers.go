package handlers

import (
	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type Handlers struct {
	Customer  *CustomerHandler
	Inventory *InventoryHandler
	Dashboard *DashboardHandler
	Analytics *AnalyticsHandler
}

func New(services *services.Services) *Handlers {
	return &Handlers{
		Customer:  NewCustomerHandler(services.Customer),
		Inventory: NewInventoryHandler(services.Inventory),
		Dashboard: NewDashboardHandler(services.Inventory, services.Customer),
		Analytics: NewAnalyticsHandler(services.Inventory, services.Customer),
	}
}

// Convenience methods that delegate to specific handlers
func (h *Handlers) GetCustomers(c *gin.Context) {
	h.Customer.GetCustomers(c)
}

func (h *Handlers) GetCustomer(c *gin.Context) {
	h.Customer.GetCustomer(c)
}

func (h *Handlers) CreateCustomer(c *gin.Context) {
	h.Customer.CreateCustomer(c)
}

func (h *Handlers) UpdateCustomer(c *gin.Context) {
	h.Customer.UpdateCustomer(c)
}

func (h *Handlers) DeleteCustomer(c *gin.Context) {
	h.Customer.DeleteCustomer(c)
}

func (h *Handlers) GetInventory(c *gin.Context) {
	h.Inventory.GetInventory(c)
}

func (h *Handlers) GetInventoryItem(c *gin.Context) {
	h.Inventory.GetInventoryItem(c)
}

func (h *Handlers) CreateInventoryItem(c *gin.Context) {
	h.Inventory.CreateInventoryItem(c)
}

func (h *Handlers) UpdateInventoryItem(c *gin.Context) {
	h.Inventory.UpdateInventoryItem(c)
}

func (h *Handlers) DeleteInventoryItem(c *gin.Context) {
	h.Inventory.DeleteInventoryItem(c)
}

func (h *Handlers) SearchInventory(c *gin.Context) {
	h.Inventory.SearchInventory(c)
}

func (h *Handlers) GetInventorySummary(c *gin.Context) {
	h.Inventory.GetInventorySummary(c)
}

// Dashboard handlers
func (h *Handlers) GetDashboard(c *gin.Context) {
	h.Dashboard.GetDashboard(c)
}

func (h *Handlers) GetCustomerWorkflow(c *gin.Context) {
	h.Dashboard.GetCustomerWorkflow(c)
}

// Analytics handlers
func (h *Handlers) GetCustomerActivity(c *gin.Context) {
	h.Analytics.GetCustomerActivity(c)
}

func (h *Handlers) GetTopCustomers(c *gin.Context) {
	h.Analytics.GetTopCustomers(c)
}

func (h *Handlers) GetGradeAnalytics(c *gin.Context) {
	h.Analytics.GetGradeAnalytics(c)
}

// Simple handlers for basic endpoints
func (h *Handlers) GetGrades(c *gin.Context) {
	grades := []string{"J55", "JZ55", "K55", "L80", "N80", "P105", "P110", "Q125", "T95", "C90", "C95", "S135"}
	utils.SuccessResponse(c, grades, "Grades retrieved successfully")
}

func (h *Handlers) GetCacheStats(c *gin.Context) {
	// TODO: Implement cache stats from services
	stats := map[string]interface{}{
		"hits":   0,
		"misses": 0,
		"items":  0,
	}
	utils.SuccessResponse(c, stats, "Cache stats retrieved")
}

func (h *Handlers) ClearCache(c *gin.Context) {
	// TODO: Implement cache clearing
	utils.SuccessResponse(c, nil, "Cache cleared successfully")
}

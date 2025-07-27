// internal/handlers/inventory_handler.go
package handlers

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	// TODO: Will integrate with repository layer later
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{}
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	
	// Mock inventory data
	inventory := []map[string]interface{}{
		{
			"id":          1,
			"customer_id": 1,
			"work_order":  "WO-2024-001",
			"size":        "5 1/2\"",
			"grade":       "J55",
			"joints":      100,
			"tenant_id":   tenantID,
		},
		{
			"id":          2,
			"customer_id": 2,
			"work_order":  "WO-2024-002", 
			"size":        "7\"",
			"grade":       "L80",
			"joints":      75,
			"tenant_id":   tenantID,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"inventory": inventory,
		"count":     len(inventory),
		"tenant":    tenantID,
	})
}

func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory ID"})
		return
	}
	
	tenantID := c.GetString("tenant_id")
	
	// Mock inventory item
	item := map[string]interface{}{
		"id":          id,
		"customer_id": 1,
		"work_order":  "WO-2024-001",
		"size":        "5 1/2\"",
		"grade":       "J55",
		"joints":      100,
		"tenant_id":   tenantID,
		"location":    "Yard A, Section 3",
		"date_in":     "2024-01-15",
	}
	
	c.JSON(http.StatusOK, gin.H{"item": item})
}

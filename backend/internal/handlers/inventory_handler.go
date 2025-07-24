// internal/handlers/inventory_handler.go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
)

type InventoryHandler struct {
	service *services.InventoryService
}

func NewInventoryHandler(service *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	filters := repository.InventoryFilters{}

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			filters.CustomerID = &customerID
		}
	}

	if grade := c.Query("grade"); grade != "" {
		filters.Grade = &grade
	}

	if size := c.Query("size"); size != "" {
		filters.Size = &size
	}

	if location := c.Query("location"); location != "" {
		filters.Location = &location
	}

	if availableStr := c.Query("available"); availableStr != "" {
		if available, err := strconv.ParseBool(availableStr); err == nil {
			filters.Available = &available
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	inventory, err := h.service.GetInventory(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"inventory": inventory})
}

func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory ID"})
		return
	}

	item, err := h.service.GetInventoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *InventoryHandler) GetInventoryByWorkOrder(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "work order is required"})
		return
	}

	items, err := h.service.GetInventoryByWorkOrder(c.Request.Context(), workOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *InventoryHandler) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	items, err := h.service.SearchInventory(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *InventoryHandler) GetAvailableInventory(c *gin.Context) {
	items, err := h.service.GetAvailableInventory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

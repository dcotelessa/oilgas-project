package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
	"oilgas-backend/pkg/validation"
)

type InventoryHandler struct {
	inventoryService services.InventoryService
}

func NewInventoryHandler(inventoryService services.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

func (h *InventoryHandler) GetInventory(c *gin.Context) {
	// Parse query parameters for filtering
	filters := map[string]interface{}{}
	
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters["customer_id"] = id
		}
	}
	
	if grade := c.Query("grade"); grade != "" {
		filters["grade"] = grade
	}
	
	if size := c.Query("size"); size != "" {
		filters["size"] = size
	}
	
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	
	if rack := c.Query("rack"); rack != "" {
		filters["rack"] = rack
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get inventory through service
	inventory, total, err := h.inventoryService.GetFiltered(c.Request.Context(), filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve inventory", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"inventory": inventory,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"filters":   filters,
	})
}

func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid inventory item ID", err)
		return
	}

	// Get inventory item through service
	item, err := h.inventoryService.GetByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Inventory item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve inventory item", err)
		return
	}

	utils.SuccessResponse(c, item, "Inventory item retrieved successfully")
}

func (h *InventoryHandler) CreateInventoryItem(c *gin.Context) {
	var req validation.InventoryValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate the inventory data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":           false,
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Create inventory item through service
	item, err := h.inventoryService.Create(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create inventory item", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Inventory item created successfully",
		"item":    item,
	})
}

func (h *InventoryHandler) UpdateInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid inventory item ID", err)
		return
	}

	var req validation.InventoryValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate the inventory data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":           false,
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Update inventory item through service
	item, err := h.inventoryService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Inventory item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update inventory item", err)
		return
	}

	utils.SuccessResponse(c, item, "Inventory item updated successfully")
}

func (h *InventoryHandler) DeleteInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid inventory item ID", err)
		return
	}

	// Delete inventory item through service
	err = h.inventoryService.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Inventory item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete inventory item", err)
		return
	}

	utils.SuccessResponse(c, nil, "Inventory item deleted successfully")
}

func (h *InventoryHandler) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query parameter 'q' is required", nil)
		return
	}

	// Parse pagination
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Search through service
	results, total, err := h.inventoryService.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Search failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
		"total":   total,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *InventoryHandler) GetInventorySummary(c *gin.Context) {
	summary, err := h.inventoryService.GetSummary(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get inventory summary", err)
		return
	}

	utils.SuccessResponse(c, summary, "Inventory summary retrieved successfully")
}

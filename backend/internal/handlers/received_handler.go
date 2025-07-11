// backend/internal/handlers/received_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
	"oilgas-backend/pkg/validation"
)

type ReceivedHandler struct {
	receivedService services.ReceivedService
}

func NewReceivedHandler(receivedService services.ReceivedService) *ReceivedHandler {
	return &ReceivedHandler{
		receivedService: receivedService,
	}
}

func (h *ReceivedHandler) GetReceived(c *gin.Context) {
	// Parse filters and pagination
	filters := parseReceivedFilters(c)
	limit := getIntQuery(c, "limit", 50)
	offset := getIntQuery(c, "offset", 0)

	items, total, err := h.receivedService.GetAll(c.Request.Context(), filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get received items", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *ReceivedHandler) CreateReceived(c *gin.Context) {
	var req validation.ReceivedValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate received item data
	if err := req.Validate(); err != nil {
		utils.BadRequest(c, "Validation failed", err)
		return
	}

	// Create received item through service
	item, err := h.receivedService.Create(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Work order already exists",
			})
			return
		}
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create received item", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Received item created successfully",
		"data":    item,
	})
}

func (h *ReceivedHandler) GetReceivedItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	item, err := h.receivedService.GetByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Received item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get received item", err)
		return
	}

	utils.SuccessResponse(c, item, "Received item retrieved successfully")
}

func (h *ReceivedHandler) UpdateReceived(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	var req validation.ReceivedValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate received item data
	if err := req.Validate(); err != nil {
		utils.BadRequest(c, "Validation failed", err)
		return
	}

	// Update received item through service
	item, err := h.receivedService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Received item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update received item", err)
		return
	}

	utils.SuccessResponse(c, item, "Received item updated successfully")
}

func (h *ReceivedHandler) DeleteReceived(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	// Delete received item through service
	err = h.receivedService.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Received item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete received item", err)
		return
	}

	utils.SuccessResponse(c, nil, "Received item deleted successfully")
}

func (h *ReceivedHandler) GetByWorkOrder(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	item, err := h.receivedService.GetByWorkOrder(c.Request.Context(), workOrder)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Received item")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get received item", err)
		return
	}

	utils.SuccessResponse(c, item, "Received item retrieved successfully")
}

func (h *ReceivedHandler) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
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

	// Use the centralized function from models
	status, err := models.StringToWorkflowState(req.Status)
	if err != nil {
		utils.BadRequest(c, "Invalid status", err)
		return
	}

	// Use the correct field name: receivedService instead of service
	err = h.receivedService.UpdateStatus(c.Request.Context(), id, status, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "invalid transition") || 
		   strings.Contains(err.Error(), "validation failed") {
			utils.BadRequest(c, "Invalid status transition", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", err)
		return
	}

	// Get updated item to return in response
	updated, err := h.receivedService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve updated item", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"item":       updated,
		"new_status": req.Status,
		"notes":      req.Notes,
	}, "Status updated successfully")
}

func (h *ReceivedHandler) GetPendingInspection(c *gin.Context) {
	items, err := h.receivedService.GetPendingInspection(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pending inspection items", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   len(items),
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
	
	if workOrder := c.Query("work_order"); workOrder != "" {
		filters["work_order"] = workOrder
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

func (h *ReceivedHandler) TransitionStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.BadRequest(c, "Invalid ID", err)
		return
	}

	var req struct {
		ToState string `json:"to_state" binding:"required"`
		Notes   string `json:"notes,omitempty"`
		Reason  string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	// Convert string to WorkflowState
	targetState, err := models.StringToWorkflowState(req.ToState)
	if err != nil {
		utils.BadRequest(c, "Invalid target state", err)
		return
	}

	// Get current item to show transition
	current, err := h.receivedService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Item not found", err)
		return
	}

	currentState := current.GetCurrentState()
	
	// Perform transition
	notes := req.Notes
	if req.Reason != "" {
		notes = fmt.Sprintf("%s (Reason: %s)", notes, req.Reason)
	}

	err = h.receivedService.UpdateStatus(c.Request.Context(), id, targetState, notes)
	if err != nil {
		if strings.Contains(err.Error(), "invalid transition") || 
		   strings.Contains(err.Error(), "validation failed") {
			utils.BadRequest(c, "Invalid status transition", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to transition status", err)
		return
	}

	// Get updated item
	updated, err := h.receivedService.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve updated item", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"item":          updated,
		"transition": gin.H{
			"from":   currentState.String(),
			"to":     targetState.String(),
			"notes":  notes,
		},
	}, "Status transition completed successfully")
}

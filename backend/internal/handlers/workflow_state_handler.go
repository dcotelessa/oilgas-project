// backend/internal/handlers/workflow_state_handler.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/models"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type WorkflowStateHandler struct {
	service services.WorkflowStateService
}

func NewWorkflowStateHandler(service services.WorkflowStateService) *WorkflowStateHandler {
	return &WorkflowStateHandler{service: service}
}

func (h *WorkflowStateHandler) GetCurrentState(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	state, err := h.service.GetCurrentState(c.Request.Context(), workOrder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get current state", err)
		return
	}

	if state == "" {
		utils.NotFound(c, "Work order")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"work_order": workOrder,
		"state":      state.String(), // Convert to string for JSON
	}, "Current state retrieved successfully")
}

func (h *WorkflowStateHandler) TransitionTo(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	var req struct {
		NewState string `json:"new_state" binding:"required"`
		Notes    string `json:"notes,omitempty"`
		User     string `json:"user,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	// SIMPLIFIED: Use the unified TransitionTo method
	err := h.service.TransitionTo(c.Request.Context(), workOrder, req.NewState, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid state transition", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to transition state", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":    "State transition successful",
		"work_order": workOrder,
		"new_state":  req.NewState,
	}, "State transition completed")
}

func (h *WorkflowStateHandler) GetStateHistory(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	history, err := h.service.GetStateHistory(c.Request.Context(), workOrder)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get state history", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"work_order": workOrder,
		"history":    history,
		"total":      len(history),
	}, "State history retrieved successfully")
}

func (h *WorkflowStateHandler) GetItemsByState(c *gin.Context) {
	stateParam := c.Param("state")
	if stateParam == "" {
		utils.BadRequest(c, "State is required", nil)
		return
	}

	// Use the centralized function from models
	state, err := models.StringToWorkflowState(stateParam)
	if err != nil {
		utils.BadRequest(c, "Invalid state", err)
		return
	}

	items, err := h.service.GetItemsByState(c.Request.Context(), state)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get items by state", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"state": stateParam,
		"items": items,
		"total": len(items),
	}, "Items by state retrieved successfully")
}

func (h *WorkflowStateHandler) ValidateTransition(c *gin.Context) {
	var req struct {
		WorkOrder string `json:"work_order" binding:"required"`
		ToState   string `json:"to_state" binding:"required"`
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

	err = h.service.ValidateTransition(c.Request.Context(), req.WorkOrder, targetState)
	if err != nil {
		utils.SuccessResponse(c, gin.H{
			"valid": false,
			"error": err.Error(),
		}, "Transition validation completed")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"valid":      true,
		"work_order": req.WorkOrder,
		"to_state":   req.ToState,
	}, "Transition is valid")
}

func (h *WorkflowStateHandler) GetJobsByState(c *gin.Context) {
	stateParam := c.Param("state")
	if stateParam == "" {
		utils.BadRequest(c, "State is required", nil)
		return
	}

	// Convert string to WorkflowState
	state, err := models.StringToWorkflowState(stateParam)
	if err != nil {
		utils.BadRequest(c, "Invalid state", err)
		return
	}

	// Get pagination parameters
	limit := 50 // default
	offset := 0 // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	jobs, total, err := h.service.GetJobsByState(c.Request.Context(), state, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get jobs by state", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"state":  stateParam,
		"jobs":   jobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}, "Jobs by state retrieved successfully")
}

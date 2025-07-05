// backend/internal/handlers/workflow_state_handler.go
package handlers

import (
	"net/http"

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

	if state == nil {
		utils.NotFound(c, "Work order")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"work_order": workOrder,
		"state": state,
	})
}

func (h *WorkflowStateHandler) TransitionTo(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		utils.BadRequest(c, "Work order is required", nil)
		return
	}

	var req struct {
		NewState models.WorkflowState `json:"new_state" binding:"required"`
		Notes    string              `json:"notes,omitempty"`
		User     string              `json:"user,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	err := h.service.TransitionTo(c.Request.Context(), workOrder, req.NewState, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "invalid transition") {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid state transition", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to transition state", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "State transition successful",
		"work_order": workOrder,
		"new_state": req.NewState,
	})
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

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"work_order": workOrder,
		"history": history,
		"total": len(history),
	})
}

func (h *WorkflowStateHandler) GetItemsByState(c *gin.Context) {
	stateParam := c.Param("state")
	if stateParam == "" {
		utils.BadRequest(c, "State is required", nil)
		return
	}

	state := models.WorkflowState(stateParam)
	
	items, err := h.service.GetItemsByState(c.Request.Context(), state)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get items by state", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"state": state,
		"items": items,
		"total": len(items),
	})
}

func (h *WorkflowStateHandler) ValidateTransition(c *gin.Context) {
	var req struct {
		From models.WorkflowState `json:"from" binding:"required"`
		To   models.WorkflowState `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err)
		return
	}

	err := h.service.ValidateTransition(c.Request.Context(), req.From, req.To)
	if err != nil {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"valid": true,
		"from": req.From,
		"to": req.To,
	})
}

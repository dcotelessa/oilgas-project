// backend/internal/handlers/search_handler.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
)

type SearchHandler struct {
	service services.SearchService
}

func NewSearchHandler(service services.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	if len(strings.TrimSpace(query)) < 2 {
		utils.BadRequest(c, "Search query must be at least 2 characters", nil)
		return
	}

	// Parse filters
	filters := make(map[string]interface{})
	if scope := c.Query("scope"); scope != "" {
		filters["scope"] = scope
	}
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters["customer_id"] = id
		}
	}
	if grade := c.Query("grade"); grade != "" {
		filters["grade"] = grade
	}

	results, err := h.service.GlobalSearch(c.Request.Context(), query, filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Global search failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"results": results,
		"query": query,
		"filters": filters,
	})
}

func (h *SearchHandler) SearchCustomers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getIntQuery(c, "limit", 50)
	offset := getIntQuery(c, "offset", 0)

	customers, total, err := h.service.SearchCustomers(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Customer search failed", err)
		return
	}

	utils.SuccessWithPagination(c, customers, total, limit, offset, "Customer search results")
}

func (h *SearchHandler) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getIntQuery(c, "limit", 50)
	offset := getIntQuery(c, "offset", 0)

	items, total, err := h.service.SearchInventory(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Inventory search failed", err)
		return
	}

	utils.SuccessWithPagination(c, items, total, limit, offset, "Inventory search results")
}

func (h *SearchHandler) SearchReceived(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getIntQuery(c, "limit", 50)
	offset := getIntQuery(c, "offset", 0)

	items, total, err := h.service.SearchReceived(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Received items search failed", err)
		return
	}

	utils.SuccessWithPagination(c, items, total, limit, offset, "Received items search results")
}

func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" || len(strings.TrimSpace(query)) < 2 {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"suggestions": []interface{}{},
		})
		return
	}

	limit := getIntQuery(c, "limit", 10)
	if limit > 20 {
		limit = 20 // Cap suggestions for performance
	}

	suggestions, err := h.service.GetSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get suggestions", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"suggestions": suggestions,
		"query": query,
		"total": len(suggestions),
	})
}

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

	// Parse pagination
	limit := getSearchIntQuery(c, "limit", 50)
	offset := getSearchIntQuery(c, "offset", 0)

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

	// Call GlobalSearch with correct signature - returns (*SearchResults, error)
	results, err := h.service.GlobalSearch(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Global search failed", err)
		return
	}

	// Fix: Access the correct fields from SearchResults
	utils.SuccessResponse(c, gin.H{
		"results":    results.Results,    // Fixed: Access Results field
		"total":      results.TotalHits,  // Fixed: Access TotalHits field
		"query":      query,
		"filters":    filters,
		"limit":      limit,
		"offset":     offset,
		"search_time": results.SearchTime,
	}, "Global search completed")
}

func (h *SearchHandler) SearchCustomers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getSearchIntQuery(c, "limit", 50)
	offset := getSearchIntQuery(c, "offset", 0)

	customers, total, err := h.service.SearchCustomers(c.Request.Context(), query, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Customer search failed", err)
		return
	}

	utils.SuccessWithPagination(c, customers, total, limit, offset, "Customer search completed")
}

func (h *SearchHandler) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getSearchIntQuery(c, "limit", 50)
	offset := getSearchIntQuery(c, "offset", 0)

	// Parse additional filters
	filters := make(map[string]interface{})
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

	items, total, err := h.service.SearchInventory(c.Request.Context(), query, filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Inventory search failed", err)
		return
	}

	utils.SuccessWithPagination(c, items, total, limit, offset, "Inventory search completed")
}

func (h *SearchHandler) SearchReceived(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	limit := getSearchIntQuery(c, "limit", 50)
	offset := getSearchIntQuery(c, "offset", 0)

	// Parse additional filters
	filters := make(map[string]interface{})
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters["customer_id"] = id
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if workOrder := c.Query("work_order"); workOrder != "" {
		filters["work_order"] = workOrder
	}

	items, total, err := h.service.SearchReceived(c.Request.Context(), query, filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Received items search failed", err)
		return
	}

	utils.SuccessWithPagination(c, items, total, limit, offset, "Received items search completed")
}

func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequest(c, "Search query 'q' is required", nil)
		return
	}

	if len(strings.TrimSpace(query)) < 2 {
		utils.SuccessResponse(c, gin.H{
			"suggestions": []interface{}{},
			"query":       query,
			"total":       0,
		}, "Empty suggestions for short query")
		return
	}

	// Parse entity type filter
	entityType := c.Query("type")
	if entityType == "" {
		entityType = "all" // Default to all entity types
	}

	// Limit suggestions to a reasonable number
	limit := getSearchIntQuery(c, "limit", 10)
	if limit > 20 {
		limit = 20 // Cap suggestions for performance
	}

	suggestions, err := h.service.GetSuggestions(c.Request.Context(), query, entityType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get suggestions", err)
		return
	}

	// Limit to requested number
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	utils.SuccessResponse(c, gin.H{
		"suggestions": suggestions,
		"query":       query,
		"total":       len(suggestions),
		"type":        entityType,
	}, "Search suggestions retrieved")
}

// Helper function - renamed to avoid conflict with received_handler.go
func getSearchIntQuery(c *gin.Context, key string, defaultValue int) int {
	if val := c.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

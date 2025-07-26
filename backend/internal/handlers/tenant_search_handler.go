// backend/internal/handlers/tenant_search_handler.go
// New cross-table search handler (tenant-only feature)
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
)

type TenantSearchHandler struct {
	searchService *services.TenantSearchService
}

func NewTenantSearchHandler(searchService *services.TenantSearchService) *TenantSearchHandler {
	return &TenantSearchHandler{
		searchService: searchService,
	}
}

func (h *TenantSearchHandler) GlobalSearch(c *gin.Context) {
	// This is a tenant-only feature
	tenantID, hasTenant := c.Get("tenant_id")
	if !hasTenant || tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant required for global search",
			"message": "Add X-Tenant header to search across tenant data",
		})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search query 'q' parameter required",
			"example": "/api/v1/search?q=oil",
		})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search query must be at least 2 characters",
		})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	results, err := h.searchService.GlobalSearchForTenant(c.Request.Context(), tenantID.(string), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

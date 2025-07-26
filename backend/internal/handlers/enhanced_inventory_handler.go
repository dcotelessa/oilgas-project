// backend/internal/handlers/enhanced_inventory_handler.go
// Enhanced inventory handler with tenant isolation
package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
)

type EnhancedInventoryHandler struct {
	service       *services.InventoryService       // Your existing service
	tenantService *services.TenantInventoryService // New tenant-aware service
}

func NewEnhancedInventoryHandler(service *services.InventoryService, tenantService *services.TenantInventoryService) *EnhancedInventoryHandler {
	return &EnhancedInventoryHandler{
		service:       service,
		tenantService: tenantService,
	}
}

// Enhanced GetInventory with tenant isolation and advanced filtering
func (h *EnhancedInventoryHandler) GetInventory(c *gin.Context) {
	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service with enhanced filtering
		h.getInventoryForTenant(c, tenantID.(string))
	} else {
		// Fall back to your existing implementation
		h.getInventoryLegacy(c)
	}
}

func (h *EnhancedInventoryHandler) getInventoryForTenant(c *gin.Context, tenantID string) {
	// Build enhanced filters from query parameters
	filters := repository.InventoryFilters{
		Search:    strings.TrimSpace(c.Query("search")),
		WorkOrder: strings.TrimSpace(c.Query("work_order")),
		Size:      strings.TrimSpace(c.Query("size")),
		Grade:     strings.TrimSpace(c.Query("grade")),
		Location:  strings.TrimSpace(c.Query("location")),
		Limit:     50, // Default
		Offset:    0,  // Default
	}

	// Parse customer ID filter
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			filters.CustomerID = &customerID
		}
	}

	// Parse date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters.DateFrom = &parsedDate
		}
	}
	
	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters.DateTo = &parsedDate
		}
	}

	// Parse available filter
	if availableStr := c.Query("available"); availableStr != "" {
		if available, err := strconv.ParseBool(availableStr); err == nil {
			filters.Available = &available
		}
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}
	
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	// Get inventory with filters, total count, and summary
	inventory, total, err := h.tenantService.GetInventoryForTenant(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get summary statistics
	summary, err := h.tenantService.GetInventorySummaryForTenant(c.Request.Context(), tenantID, filters)
	if err != nil {
		// Continue without summary if there's an error
		summary = nil
	}

	// Calculate pagination info
	hasMore := (filters.Offset + filters.Limit) < total
	page := (filters.Offset / filters.Limit) + 1

	response := gin.H{
		"tenant":    tenantID,
		"inventory": inventory,
		"count":     len(inventory),
		"total":     total,
		"limit":     filters.Limit,
		"offset":    filters.Offset,
		"page":      page,
		"has_more":  hasMore,
		"filters": gin.H{
			"customer_id": c.Query("customer_id"),
			"work_order":  filters.WorkOrder,
			"size":        filters.Size,
			"grade":       filters.Grade,
			"location":    filters.Location,
			"date_from":   c.Query("date_from"),
			"date_to":     c.Query("date_to"),
			"available":   c.Query("available"),
			"search":      filters.Search,
		},
	}

	if summary != nil {
		response["summary"] = summary
	}

	c.JSON(http.StatusOK, response)
}

func (h *EnhancedInventoryHandler) getInventoryLegacy(c *gin.Context) {
	// Your existing implementation
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

// Enhanced GetInventoryItem with tenant isolation
func (h *EnhancedInventoryHandler) GetInventoryItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory ID"})
		return
	}

	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		item, err := h.tenantService.GetInventoryByIDForTenant(c.Request.Context(), tenantID.(string), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if item == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant": tenantID,
			"item":   item,
		})
	} else {
		// Fall back to your existing implementation
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
}

// Enhanced GetInventoryByWorkOrder with tenant isolation
func (h *EnhancedInventoryHandler) GetInventoryByWorkOrder(c *gin.Context) {
	workOrder := c.Param("workOrder")
	if workOrder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "work order is required"})
		return
	}

	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		items, err := h.tenantService.GetInventoryByWorkOrderForTenant(c.Request.Context(), tenantID.(string), workOrder)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant": tenantID,
			"items":  items,
			"count":  len(items),
		})
	} else {
		// Fall back to your existing implementation
		items, err := h.service.GetInventoryByWorkOrder(c.Request.Context(), workOrder)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

// Enhanced SearchInventory with tenant isolation
func (h *EnhancedInventoryHandler) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		items, err := h.tenantService.SearchInventoryForTenant(c.Request.Context(), tenantID.(string), query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant": tenantID,
			"items":  items,
			"count":  len(items),
			"search": query,
		})
	} else {
		// Fall back to your existing implementation
		items, err := h.service.SearchInventory(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

// Enhanced GetAvailableInventory with tenant isolation
func (h *EnhancedInventoryHandler) GetAvailableInventory(c *gin.Context) {
	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		items, err := h.tenantService.GetAvailableInventoryForTenant(c.Request.Context(), tenantID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant": tenantID,
			"items":  items,
			"count":  len(items),
		})
	} else {
		// Fall back to your existing implementation
		items, err := h.service.GetAvailableInventory(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

// New Work Orders endpoints (tenant-aware only)
func (h *EnhancedInventoryHandler) GetWorkOrders(c *gin.Context) {
	// This is a new tenant-only feature
	tenantID, hasTenant := c.Get("tenant_id")
	if !hasTenant || tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant required for work orders endpoint"})
		return
	}

	// Build filters
	filters := repository.WorkOrderFilters{
		Search: strings.TrimSpace(c.Query("search")),
		Limit:  50,
		Offset: 0,
	}

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil {
			filters.CustomerID = &customerID
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}
	
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	workOrders, total, err := h.tenantService.GetWorkOrdersForTenant(c.Request.Context(), tenantID.(string), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasMore := (filters.Offset + filters.Limit) < total
	page := (filters.Offset / filters.Limit) + 1

	c.JSON(http.StatusOK, gin.H{
		"tenant":      tenantID,
		"work_orders": workOrders,
		"count":       len(workOrders),
		"total":       total,
		"limit":       filters.Limit,
		"offset":      filters.Offset,
		"page":        page,
		"has_more":    hasMore,
		"filters": gin.H{
			"search":      filters.Search,
			"customer_id": c.Query("customer_id"),
		},
	})
}

func (h *EnhancedInventoryHandler) GetWorkOrder(c *gin.Context) {
	// This is a new tenant-only feature
	tenantID, hasTenant := c.Get("tenant_id")
	if !hasTenant || tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant required for work order details endpoint"})
		return
	}

	workOrderID := c.Param("id")
	if workOrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "work order ID is required"})
		return
	}

	details, err := h.tenantService.GetWorkOrderDetailsForTenant(c.Request.Context(), tenantID.(string), workOrderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if details == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant":  tenantID,
		"details": details,
	})
}

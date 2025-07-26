// backend/internal/handlers/enhanced_customer_handler.go
// Enhanced customer handler with tenant isolation
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/services"
)

type EnhancedCustomerHandler struct {
	service       *services.CustomerService       // Your existing service
	tenantService *services.TenantCustomerService // New tenant-aware service
}

func NewEnhancedCustomerHandler(service *services.CustomerService, tenantService *services.TenantCustomerService) *EnhancedCustomerHandler {
	return &EnhancedCustomerHandler{
		service:       service,
		tenantService: tenantService,
	}
}

// Enhanced GetCustomers with tenant isolation and advanced filtering
func (h *EnhancedCustomerHandler) GetCustomers(c *gin.Context) {
	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		h.getCustomersForTenant(c, tenantID.(string))
	} else {
		// Fall back to your existing implementation
		h.service.GetAllCustomers(c.Request.Context())
	}
}

func (h *EnhancedCustomerHandler) getCustomersForTenant(c *gin.Context, tenantID string) {
	// Build filters from query parameters
	filters := repository.CustomerFilters{
		Search: strings.TrimSpace(c.Query("search")),
		State:  strings.TrimSpace(c.Query("state")),
		City:   strings.TrimSpace(c.Query("city")),
		Limit:  50, // Default
		Offset: 0,  // Default
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

	// Get customers with filters and total count
	customers, total, err := h.tenantService.GetCustomersWithFiltersForTenant(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pagination info
	hasMore := (filters.Offset + filters.Limit) < total
	page := (filters.Offset / filters.Limit) + 1

	c.JSON(http.StatusOK, gin.H{
		"tenant":    tenantID,
		"customers": customers,
		"count":     len(customers),
		"total":     total,
		"limit":     filters.Limit,
		"offset":    filters.Offset,
		"page":      page,
		"has_more":  hasMore,
		"filters": gin.H{
			"search": filters.Search,
			"state":  filters.State,
			"city":   filters.City,
		},
	})
}

// Enhanced GetCustomer with tenant isolation and related data
func (h *EnhancedCustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		h.getCustomerForTenant(c, tenantID.(string), id)
	} else {
		// Fall back to your existing implementation
		customer, err := h.service.GetCustomerByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if customer == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"customer": customer})
	}
}

func (h *EnhancedCustomerHandler) getCustomerForTenant(c *gin.Context, tenantID string, customerID int) {
	// Get customer
	customer, err := h.tenantService.GetCustomerByIDForTenant(c.Request.Context(), tenantID, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if customer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	// Get related data (inventory count, work orders, etc.)
	relatedData, err := h.tenantService.GetCustomerRelatedDataForTenant(c.Request.Context(), tenantID, customerID)
	if err != nil {
		// Continue without related data if there's an error
		c.JSON(http.StatusOK, gin.H{
			"tenant":   tenantID,
			"customer": customer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant":   tenantID,
		"customer": customer,
		"related":  relatedData,
	})
}

// Enhanced SearchCustomers with tenant isolation
func (h *EnhancedCustomerHandler) SearchCustomers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	// Check if this is a tenant-aware request
	tenantID, hasTenant := c.Get("tenant_id")
	
	if hasTenant && tenantID != "" {
		// Use tenant-aware service
		customers, err := h.tenantService.SearchCustomersForTenant(c.Request.Context(), tenantID.(string), query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant":    tenantID,
			"customers": customers,
			"count":     len(customers),
			"search":    query,
		})
	} else {
		// Fall back to your existing implementation
		customers, err := h.service.SearchCustomers(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"customers": customers})
	}
}

// backend/internal/customer/handlers.go
package customer

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes sets up the customer endpoints
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	customers := router.Group("/customers")
	{
		customers.GET("", h.GetCustomers)
		customers.POST("", h.CreateCustomer)
		customers.GET("/search", h.SearchCustomers)
		customers.GET("/:id", h.GetCustomer)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
		customers.GET("/:id/analytics", h.GetCustomerAnalytics)
		customers.PUT("/:id/contacts", h.UpdateCustomerContacts)
	}
}

// GetCustomers retrieves customers with filtering and pagination
func (h *Handler) GetCustomers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	filters := CustomerFilters{
		Query:     c.Query("q"),
		State:     c.Query("state"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Parse optional parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	// Parse boolean filters
	if hasOrdersStr := c.Query("has_orders"); hasOrdersStr != "" {
		if hasOrders, err := strconv.ParseBool(hasOrdersStr); err == nil {
			filters.HasOrders = &hasOrders
		}
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.Active = &active
		}
	}

	// Parse date filters
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	response, err := h.service.GetCustomersForTenant(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCustomer retrieves a single customer with analytics
func (h *Handler) GetCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	customer, err := h.service.GetCustomerByIDForTenant(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer": customer})
}

// SearchCustomers provides enhanced search functionality
func (h *Handler) SearchCustomers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query parameter 'q' is required"})
		return
	}

	customers, err := h.service.SearchCustomersForTenant(c.Request.Context(), tenantID, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"count":     len(customers),
		"query":     query,
	})
}

// CreateCustomer creates a new customer
func (h *Handler) CreateCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.service.CreateCustomerForTenant(c.Request.Context(), tenantID, req)
	if err != nil {
		if customerErr, ok := err.(CustomerError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": customerErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"customer": customer})
}

// UpdateCustomer updates an existing customer
func (h *Handler) UpdateCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.service.UpdateCustomerForTenant(c.Request.Context(), tenantID, id, req)
	if err != nil {
		if err == ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		if customerErr, ok := err.(CustomerError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": customerErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer": customer})
}

// DeleteCustomer soft deletes a customer
func (h *Handler) DeleteCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	err = h.service.DeleteCustomerForTenant(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}

// GetCustomerAnalytics provides detailed customer analytics
func (h *Handler) GetCustomerAnalytics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	analytics, err := h.service.GetCustomerAnalyticsForTenant(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// UpdateCustomerContacts updates customer contact information
func (h *Handler) UpdateCustomerContacts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	var req UpdateContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateCustomerContactsForTenant(c.Request.Context(), tenantID, id, req)
	if err != nil {
		if err == ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contacts updated successfully"})
}

// Enterprise endpoints for cross-tenant operations (admin only)
func (h *Handler) RegisterEnterpriseRoutes(router *gin.RouterGroup) {
	enterprise := router.Group("/enterprise/customers")
	{
		enterprise.GET("/summary", h.GetCustomerSummary)
		enterprise.POST("/batch", h.GetCustomersBatch)
	}
}

// GetCustomerSummary provides tenant-wide customer counts (admin only)
func (h *Handler) GetCustomerSummary(c *gin.Context) {
	// TODO: Add admin role check
	summary, err := h.service.GetCustomerSummaryByTenant(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// GetCustomersBatch retrieves customers by IDs across tenants (admin only)
func (h *Handler) GetCustomersBatch(c *gin.Context) {
	// TODO: Add admin role check
	var req struct {
		CustomerIDs []int `json:"customer_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customers, err := h.service.GetCustomersByIDs(c.Request.Context(), req.CustomerIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"count":     len(customers),
	})
}

// backend/internal/customer/handlers.go
package customer

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	service Service
}

func NewHandlers(service Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	customers := router.Group("/customers")
	customers.Use(authMiddleware)
	
	customers.GET("", h.SearchCustomers)
	customers.POST("", h.CreateCustomer)
	customers.GET("/:id", h.GetCustomer)
	customers.PUT("/:id", h.UpdateCustomer)
	customers.DELETE("/:id", h.DeleteCustomer)
	
	customers.GET("/:id/contacts", h.GetCustomerContacts)
	customers.POST("/:id/contacts", h.RegisterCustomerContact)
	customers.DELETE("/:id/contacts/:userId", h.RemoveCustomerContact)
	
	customers.GET("/:id/analytics", h.GetCustomerAnalytics)
	customers.GET("/analytics", h.GetTenantAnalytics)
}

func (h *Handlers) GetCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	
	customer, err := h.service.GetCustomer(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	
	c.JSON(http.StatusOK, customer)
}

func (h *Handlers) SearchCustomers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	
	filters := SearchFilters{
		Name:        c.Query("name"),
		CompanyCode: c.Query("company_code"),
		TaxID:       c.Query("tax_id"),
	}
	
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}
	
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}
	
	customers, total, err := h.service.SearchCustomers(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search customers"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":  customers,
		"total": total,
	})
}

type CreateCustomerRequest struct {
	Name        string      `json:"name" binding:"required"`
	CompanyCode string      `json:"company_code" binding:"required"`
	BillingInfo BillingInfo `json:"billing_info" binding:"required"`
}

func (h *Handlers) CreateCustomer(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	customer := &Customer{
		TenantID:    tenantID,
		Name:        req.Name,
		CompanyCode: req.CompanyCode,
		BillingInfo: req.BillingInfo,
		Status:      StatusActive,
	}
	
	err := h.service.CreateCustomer(c.Request.Context(), customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, customer)
}

type RegisterContactRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

func (h *Handlers) RegisterCustomerContact(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	customerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	
	var req RegisterContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err = h.service.RegisterCustomerContact(
		c.Request.Context(), tenantID, customerID,
		req.Email, req.FirstName, req.LastName, req.Role,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Customer contact registered successfully"})
}

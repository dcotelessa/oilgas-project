package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/dcotelessa/oil-gas-inventory/internal/models"
    "github.com/dcotelessa/oil-gas-inventory/internal/services"
)

type TenantHandler struct {
    tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
    return &TenantHandler{
        tenantService: tenantService,
    }
}

// CreateTenant creates a new tenant (admin only)
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    var tenant models.Tenant
    if err := c.ShouldBindJSON(&tenant); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.tenantService.CreateTenant(&tenant); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, tenant)
}

// GetTenant gets tenant by ID
func (h *TenantHandler) GetTenant(c *gin.Context) {
    idStr := c.Param("id")
    tenantID, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
        return
    }

    tenant, err := h.tenantService.GetTenant(tenantID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
        return
    }

    c.JSON(http.StatusOK, tenant)
}

// ListTenants lists all tenants (admin only)
func (h *TenantHandler) ListTenants(c *gin.Context) {
    tenants, err := h.tenantService.ListTenants()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "tenants": tenants,
        "count":   len(tenants),
    })
}

// GetCurrentTenant returns the current user's tenant context
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
    tenantID, exists := c.Get("tenant_id")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant context not found"})
        return
    }

    tenant, err := h.tenantService.GetTenant(tenantID.(int))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Current tenant not found"})
        return
    }

    // Include user's role and permissions
    c.JSON(http.StatusOK, gin.H{
        "tenant":   tenant,
        "role":     c.GetString("user_role"),
        "is_admin": c.GetBool("is_admin"),
    })
}

// SwitchTenant allows admin users to switch tenant context
func (h *TenantHandler) SwitchTenant(c *gin.Context) {
    var request struct {
        TenantID int `json:"tenant_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify tenant exists
    tenant, err := h.tenantService.GetTenant(request.TenantID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
        return
    }

    if !tenant.Active {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant is not active"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Tenant context switched",
        "tenant":  tenant,
    })
}

// AssignCustomerToTenant assigns a customer to a tenant
func (h *TenantHandler) AssignCustomerToTenant(c *gin.Context) {
    var request struct {
        CustomerID       int    `json:"customer_id" binding:"required"`
        TenantID         int    `json:"tenant_id" binding:"required"`
        RelationshipType string `json:"relationship_type"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Default relationship type
    if request.RelationshipType == "" {
        request.RelationshipType = "primary"
    }

    // Get current user ID for audit
    userID, _ := c.Get("user_id")

    err := h.tenantService.AssignCustomerToTenant(
        request.CustomerID,
        request.TenantID,
        request.RelationshipType,
        userID.(int),
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Customer assigned to tenant successfully"})
}

// GetTenantCustomers gets all customers assigned to a tenant
func (h *TenantHandler) GetTenantCustomers(c *gin.Context) {
    tenantID, exists := c.Get("tenant_id")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant context not found"})
        return
    }

    assignments, err := h.tenantService.GetTenantCustomers(tenantID.(int))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "assignments": assignments,
        "count":       len(assignments),
    })
}

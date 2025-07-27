// internal/handlers/customer_handler.go
package handlers

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	// Will integrate with repository layer later
}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{}
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	
	// Mock data for now - will integrate with repository
	customers := []map[string]interface{}{
		{
			"id":       1,
			"name":     "Sample Oil Company",
			"email":    "contact@sampleoil.com",
			"tenant_id": tenantID,
		},
		{
			"id":       2,
			"name":     "Demo Gas Corp",
			"email":    "info@demogas.com", 
			"tenant_id": tenantID,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"count":     len(customers),
		"tenant":    tenantID,
	})
}

func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}
	
	tenantID := c.GetString("tenant_id")
	
	// Mock customer data
	customer := map[string]interface{}{
		"id":       id,
		"name":     "Sample Oil Company",
		"email":    "contact@sampleoil.com",
		"tenant_id": tenantID,
		"address":  "123 Oil Field Road",
		"phone":    "(555) 123-4567",
	}
	
	c.JSON(http.StatusOK, gin.H{"customer": customer})
}

package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/internal/utils"
	"oilgas-backend/pkg/validation"
)

type CustomerHandler struct {
	customerService services.CustomerService
}

func NewCustomerHandler(customerService services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	customers, err := h.customerService.GetAll(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customers", err)
		return
	}

	utils.SuccessResponse(c, customers, "Customers retrieved successfully")
}

func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	customerID := c.Param("customerID")
	customer, err := h.customerService.GetByID(c.Request.Context(), customerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid") {
			utils.NotFound(c, "Customer")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customer", err)
		return
	}

	utils.SuccessResponse(c, customer, "Customer retrieved successfully")
}

func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req validation.CustomerValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate customer data
	if err := req.Validate(); err != nil {
		utils.BadRequest(c, "Validation failed", err)
		return
	}

	// Create customer through service
	customer, err := h.customerService.Create(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Customer with this name already exists",
			})
			return
		}
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create customer", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"message":  "Customer created successfully",
		"customer": customer,
	})
}

func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	customerIDParam := c.Param("customerID")
	
	var req validation.CustomerValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid JSON", err)
		return
	}

	// Validate customer data
	if err := req.Validate(); err != nil {
		utils.BadRequest(c, "Validation failed", err)
		return
	}

	// Update customer through service
	customer, err := h.customerService.Update(c.Request.Context(), customerIDParam, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Customer")
			return
		}
		
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Customer name already exists",
			})
			return
		}

		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update customer", err)
		return
	}

	utils.SuccessResponse(c, customer, "Customer updated successfully")
}

func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	customerIDParam := c.Param("customerID")

	// Soft delete customer through service
	err := h.customerService.Delete(c.Request.Context(), customerIDParam)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFound(c, "Customer")
			return
		}

		if strings.Contains(err.Error(), "active inventory") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Cannot delete customer with active inventory items",
				"note":    "Move or complete all inventory items first",
			})
			return
		}

		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete customer", err)
		return
	}

	utils.SuccessResponse(c, nil, "Customer deleted successfully")
}

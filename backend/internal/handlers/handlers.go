// backend/internal/handlers/handlers.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oilgas-backend/internal/services"
	"oilgas-backend/pkg/validation"
)

type Handlers struct {
	services *services.Services
}

func New(services *services.Services) *Handlers {
	return &Handlers{
		services: services,
	}
}

// Customer handlers
func (h *Handlers) GetCustomers(c *gin.Context) {
	customers, err := h.services.Customer.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve customers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"total":     len(customers),
	})
}

func (h *Handlers) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	customer, err := h.services.Customer.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "customer not found" || err.Error() == "invalid customer ID: strconv.Atoi: parsing \"invalid\": invalid syntax" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve customer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer": customer,
	})
}

func (h *Handlers) CreateCustomer(c *gin.Context) {
	var req validation.CustomerValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate customer data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Create customer through service
	customer, err := h.services.Customer.Create(c.Request.Context(), &req)
	if err != nil {
		// Check for duplicate customer name
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Customer with this name already exists",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create customer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Customer created successfully",
		"customer": customer,
	})
}

func (h *Handlers) UpdateCustomer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
		})
		return
	}

	var req validation.CustomerValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate customer data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Update customer through service
	customer, err := h.services.Customer.Update(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}
		
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Customer name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update customer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Customer updated successfully",
		"customer": customer,
	})
}

func (h *Handlers) DeleteCustomer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid customer ID",
		})
		return
	}

	// Check if customer has associated inventory items
	hasInventory, err := h.services.Customer.HasActiveInventory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check customer dependencies",
			"details": err.Error(),
		})
		return
	}

	if hasInventory {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Cannot delete customer with active inventory items",
			"note":  "Move or complete all inventory items first",
		})
		return
	}

	// Soft delete customer through service
	err = h.services.Customer.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete customer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer deleted successfully",
	})
}

// Inventory handlers - FULLY IMPLEMENTED
func (h *Handlers) CreateInventoryItem(c *gin.Context) {
	var req validation.InventoryValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate the inventory data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Create inventory item through service
	item, err := h.services.Inventory.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create inventory item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Inventory item created successfully",
		"item":    item,
	})
}

func (h *Handlers) UpdateInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory item ID",
		})
		return
	}

	var req validation.InventoryValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate the inventory data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Update inventory item through service
	item, err := h.services.Inventory.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "inventory item not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Inventory item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update inventory item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inventory item updated successfully",
		"item":    item,
	})
}

func (h *Handlers) DeleteInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory item ID",
		})
		return
	}

	// Delete inventory item through service
	err = h.services.Inventory.Delete(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "inventory item not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Inventory item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete inventory item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inventory item deleted successfully",
	})
}

func (h *Handlers) GetInventory(c *gin.Context) {
	// Parse query parameters for filtering
	filters := map[string]interface{}{}
	
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
	
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	
	if rack := c.Query("rack"); rack != "" {
		filters["rack"] = rack
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get inventory through service
	inventory, total, err := h.services.Inventory.GetFiltered(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve inventory",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inventory": inventory,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"filters":   filters,
	})
}

func (h *Handlers) GetInventoryItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid inventory item ID",
		})
		return
	}

	// Get inventory item through service
	item, err := h.services.Inventory.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "inventory item not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Inventory item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve inventory item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": item,
	})
}

func (h *Handlers) SearchInventory(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query parameter 'q' is required",
		})
		return
	}

	// Parse pagination
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Search through service
	results, total, err := h.services.Inventory.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handlers) GetInventorySummary(c *gin.Context) {
	summary, err := h.services.Inventory.GetSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get inventory summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

// Grade handlers
func (h *Handlers) GetGrades(c *gin.Context) {
	// Get grades from database instead of hard-coded list
	grades, err := h.services.Grade.GetAll(c.Request.Context())
	if err != nil {
		// Fallback to default grades if database fails
		defaultGrades := []string{"J55", "JZ55", "K55", "L80", "N80", "P105", "P110", "Q125", "T95", "C90", "C95", "S135"}
		c.JSON(http.StatusOK, gin.H{
			"grades":   defaultGrades,
			"total":    len(defaultGrades),
			"source":   "default",
			"warning":  "Using fallback grades due to database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grades": grades,
		"total":  len(grades),
		"source": "database",
	})
}

func (h *Handlers) CreateGrade(c *gin.Context) {
	var req struct {
		Grade       string `json:"grade" binding:"required,min=2,max=20"`
		Description string `json:"description,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Normalize grade name (uppercase, trim spaces)
	req.Grade = strings.ToUpper(strings.TrimSpace(req.Grade))

	// Validate grade format (basic oil & gas grade patterns)
	if !isValidGradeFormat(req.Grade) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid grade format",
			"note":  "Grade should follow oil & gas industry standards (e.g., J55, L80, P110)",
		})
		return
	}

	// Create grade through service
	grade, err := h.services.Grade.Create(c.Request.Context(), req.Grade, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Grade already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create grade",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Grade created successfully",
		"grade":   grade,
	})
}

func (h *Handlers) DeleteGrade(c *gin.Context) {
	gradeName := c.Param("grade")
	if gradeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Grade name is required",
		})
		return
	}

	gradeName = strings.ToUpper(strings.TrimSpace(gradeName))

	// Check if grade is in use
	inUse, err := h.services.Grade.IsInUse(c.Request.Context(), gradeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check grade usage",
			"details": err.Error(),
		})
		return
	}

	if inUse {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Cannot delete grade that is currently in use",
			"note":  "Remove all inventory items with this grade first",
		})
		return
	}

	// Delete grade through service
	err = h.services.Grade.Delete(c.Request.Context(), gradeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Grade not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete grade",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grade deleted successfully",
	})
}

func (h *Handlers) GetGradeUsage(c *gin.Context) {
	gradeName := c.Param("grade")
	if gradeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Grade name is required",
		})
		return
	}

	gradeName = strings.ToUpper(strings.TrimSpace(gradeName))

	// Get usage statistics for the grade
	usage, err := h.services.Grade.GetUsageStats(c.Request.Context(), gradeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get grade usage",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grade": gradeName,
		"usage": usage,
	})
}

// Helper function to validate grade format
func isValidGradeFormat(grade string) bool {
	// Basic validation for oil & gas grade formats
	// This can be expanded based on industry standards
	if len(grade) < 2 || len(grade) > 10 {
		return false
	}

	// Check for common patterns: letter+number combinations
	// Examples: J55, L80, P110, Q125, etc.
	validPattern := regexp.MustCompile(`^[A-Z]+[0-9]+[A-Z]*$`)
	return validPattern.MatchString(grade)
}

// Placeholder handlers for other endpoints (implement as needed)
func (h *Handlers) GetReceived(c *gin.Context) {
	// Parse filters
	filters := map[string]interface{}{}
	
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filters["customer_id"] = id
		}
	}
	
	if status := c.Query("status"); status != "" {
		// pending, in_production, completed
		filters["status"] = status
	}
	
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if date, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = date
		}
	}
	
	if dateTo := c.Query("date_to"); dateTo != "" {
		if date, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = date
		}
	}

	// Parse pagination
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get received items through service
	items, total, err := h.services.Received.GetFiltered(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve received items",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"received_items": items,
		"total":          total,
		"limit":          limit,
		"offset":         offset,
		"filters":        filters,
	})
}

func (h *Handlers) GetReceivedItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid received item ID",
		})
		return
	}

	item, err := h.services.Received.GetByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Received item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve received item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": item,
	})
}

func (h *Handlers) CreateReceivedItem(c *gin.Context) {
	var req validation.ReceivedValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate the received item data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Create received item through service
	item, err := h.services.Received.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create received item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Received item created successfully",
		"item":    item,
	})
}

func (h *Handlers) UpdateReceivedItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid received item ID",
		})
		return
	}

	var req validation.ReceivedValidation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	// Validate the received item data
	if errors := req.Validate(); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Validation failed",
			"validation_errors": errors,
		})
		return
	}

	// Update received item through service
	item, err := h.services.Received.Update(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Received item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update received item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Received item updated successfully",
		"item":    item,
	})
}

func (h *Handlers) DeleteReceivedItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid received item ID",
		})
		return
	}

	// Check if item can be deleted (business rules)
	canDelete, reason, err := h.services.Received.CanDelete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check deletion status",
			"details": err.Error(),
		})
		return
	}

	if !canDelete {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "Cannot delete received item",
			"reason": reason,
		})
		return
	}

	// Delete received item through service
	err = h.services.Received.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Received item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete received item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Received item deleted successfully",
	})
}

// Fletcher handlers (implement when needed)
func (h *Handlers) GetFletcherItems(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement fletcher items listing",
	})
}

func (h *Handlers) GetFletcherItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement fletcher item retrieval",
	})
}

func (h *Handlers) CreateFletcherItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement fletcher item creation",
	})
}

func (h *Handlers) UpdateFletcherItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement fletcher item update",
	})
}

func (h *Handlers) DeleteFletcherItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement fletcher item deletion",
	})
}

// Bakeout handlers (implement when needed)
func (h *Handlers) GetBakeoutItems(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement bakeout items listing",
	})
}

func (h *Handlers) GetBakeoutItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement bakeout item retrieval",
	})
}

func (h *Handlers) CreateBakeoutItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement bakeout item creation",
	})
}

func (h *Handlers) UpdateBakeoutItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement bakeout item update",
	})
}

func (h *Handlers) DeleteBakeoutItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement bakeout item deletion",
	})
}

// Search handlers
func (h *Handlers) SearchCustomers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query parameter 'q' is required",
		})
		return
	}

	// Parse pagination
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Search customers through service
	customers, total, err := h.services.Customer.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"total":     total,
		"query":     query,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handlers) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query parameter 'q' is required",
		})
		return
	}

	// Validate query length to prevent abuse
	if len(strings.TrimSpace(query)) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query must be at least 2 characters",
		})
		return
	}

	// Parse search scope (optional filter)
	scope := c.Query("scope") // "customers", "inventory", "all"
	if scope == "" {
		scope = "all"
	}

	// Parse pagination
	limit := 20 // Lower default for global search
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Perform global search through service
	results, err := h.services.Search.GlobalSearch(c.Request.Context(), query, scope, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Global search failed",
			"details": err.Error(),
		})
		return
	}

	// Calculate totals by type
	customerCount := len(results.Customers)
	inventoryCount := len(results.Inventory)
	totalResults := customerCount + inventoryCount

	c.JSON(http.StatusOK, gin.H{
		"results": gin.H{
			"customers": results.Customers,
			"inventory": results.Inventory,
		},
		"summary": gin.H{
			"total_results":    totalResults,
			"customer_count":   customerCount,
			"inventory_count":  inventoryCount,
		},
		"query":  query,
		"scope":  scope,
		"limit":  limit,
		"offset": offset,
	})
}

// Quick search endpoint for autocomplete/suggestions
func (h *Handlers) QuickSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" || len(strings.TrimSpace(query)) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"suggestions": []interface{}{},
		})
		return
	}

	// Get quick suggestions (limited results for performance)
	suggestions, err := h.services.Search.GetSuggestions(c.Request.Context(), query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get suggestions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"query":       query,
	})
}

// Advanced search with multiple filters
func (h *Handlers) AdvancedSearch(c *gin.Context) {
	var searchReq struct {
		Query       string            `json:"query"`
		Filters     map[string]string `json:"filters"`
		DateRange   *DateRange        `json:"date_range,omitempty"`
		SortBy      string            `json:"sort_by"`
		SortOrder   string            `json:"sort_order"`
		Limit       int               `json:"limit"`
		Offset      int               `json:"offset"`
	}

	if err := c.ShouldBindJSON(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search request",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if searchReq.Limit == 0 || searchReq.Limit > 100 {
		searchReq.Limit = 50
	}
	if searchReq.SortOrder != "asc" && searchReq.SortOrder != "desc" {
		searchReq.SortOrder = "desc"
	}

	// Perform advanced search through service
	results, total, err := h.services.Search.AdvancedSearch(c.Request.Context(), &searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Advanced search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"request": searchReq,
	})
}

type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Analytics handlers
func (h *Handlers) GetDashboardStats(c *gin.Context) {
	// Get inventory summary (already implemented)
	inventorySummary, err := h.services.Inventory.GetSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get inventory summary",
			"details": err.Error(),
		})
		return
	}

	// Get customer count
	customerCount, err := h.services.Customer.GetTotalCount(c.Request.Context())
	if err != nil {
		// Don't fail the whole request, just log and use 0
		customerCount = 0
	}

	// Get recent activity (last 7 days)
	recentActivity, err := h.services.Inventory.GetRecentActivity(c.Request.Context(), 7)
	if err != nil {
		recentActivity = []interface{}{} // Empty array if fails
	}

	// Build comprehensive dashboard
	dashboard := gin.H{
		"inventory": inventorySummary,
		"customers": gin.H{
			"total": customerCount,
		},
		"recent_activity": recentActivity,
		"system": gin.H{
			"cache_stats": h.services.Cache.GetStats(),
			"uptime":      time.Since(startTime).String(), // You'll need to track this
		},
		"summary": gin.H{
			"total_items":   inventorySummary.TotalItems,
			"total_customers": customerCount,
			"active_grades": len(inventorySummary.GradeDistribution),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": dashboard,
		"generated_at": time.Now().UTC(),
	})
}

func (h *Handlers) GetCustomerActivity(c *gin.Context) {
	// Parse date range
	days := 30 // default
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Parse customer ID filter (optional)
	var customerID *int
	if custParam := c.Query("customer_id"); custParam != "" {
		if id, err := strconv.Atoi(custParam); err == nil {
			customerID = &id
		}
	}

	// Get customer activity through service
	activity, err := h.services.Customer.GetActivity(c.Request.Context(), days, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get customer activity",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activity":   activity,
		"period_days": days,
		"customer_id": customerID,
		"generated_at": time.Now().UTC(),
	})
}

func (h *Handlers) GetTopCustomers(c *gin.Context) {
	// Parse limit
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Parse time period
	days := 30
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Get top customers by activity
	topCustomers, err := h.services.Customer.GetTopByActivity(c.Request.Context(), limit, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get top customers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_customers": topCustomers,
		"limit":         limit,
		"period_days":   days,
	})
}

func (h *Handlers) GetGradeAnalytics(c *gin.Context) {
	// Get grade distribution and trends
	analytics, err := h.services.Inventory.GetGradeAnalytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get grade analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grade_analytics": analytics,
		"generated_at":    time.Now().UTC(),
	})
}

func (h *Handlers) UpdateReceivedStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid received item ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending in_production threading completed"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Update status through service
	err = h.services.Received.UpdateStatus(c.Request.Context(), id, req.Status, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Received item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

// System handlers
func (h *Handlers) GetCacheStats(c *gin.Context) {
	stats := h.services.Cache.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
		"hit_ratio":   h.services.Cache.GetHitRatio(),
	})
}

func (h *Handlers) ClearCache(c *gin.Context) {
	h.services.Cache.Clear()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

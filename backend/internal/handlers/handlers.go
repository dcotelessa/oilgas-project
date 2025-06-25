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
	// TODO: Implement customer creation with validation
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement customer creation",
		"note":    "Requires CustomerValidation and repository Create method",
	})
}

func (h *Handlers) UpdateCustomer(c *gin.Context) {
	// TODO: Implement customer update
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement customer update",
		"note":    "Requires repository Update method",
	})
}

func (h *Handlers) DeleteCustomer(c *gin.Context) {
	// TODO: Implement customer deletion
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement customer deletion",
		"note":    "Requires repository Delete method",
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
	// For now, return static grades - could be made dynamic later
	grades := []string{"J55", "JZ55", "K55", "L80", "N80", "P105", "P110", "Q125", "T95", "C90", "C95", "S135"}
	c.JSON(http.StatusOK, gin.H{
		"grades": grades,
		"total":  len(grades),
	})
}

func (h *Handlers) CreateGrade(c *gin.Context) {
	// TODO: Implement grade creation
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement grade creation",
		"note":    "Add to grade repository when needed",
	})
}

func (h *Handlers) DeleteGrade(c *gin.Context) {
	// TODO: Implement grade deletion
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement grade deletion",
		"note":    "Add to grade repository when needed",
	})
}

// Placeholder handlers for other endpoints (implement as needed)
func (h *Handlers) GetReceived(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement received items listing",
		"note":    "Add ReceivedRepository when needed by frontend",
	})
}

func (h *Handlers) GetReceivedItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement received item retrieval",
	})
}

func (h *Handlers) CreateReceivedItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement received item creation",
	})
}

func (h *Handlers) UpdateReceivedItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement received item update",
	})
}

func (h *Handlers) DeleteReceivedItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement received item deletion",
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
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement customer search",
		"note":    "Add to customer repository when needed",
	})
}

func (h *Handlers) GlobalSearch(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement global search",
		"note":    "Search across customers, inventory, received items",
	})
}

// Analytics handlers
func (h *Handlers) GetDashboardStats(c *gin.Context) {
	// For now, delegate to inventory summary
	summary, err := h.services.Inventory.GetSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dashboard stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": summary,
	})
}

func (h *Handlers) GetCustomerActivity(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "TODO: Implement customer activity",
		"note":    "Add analytics queries when needed",
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

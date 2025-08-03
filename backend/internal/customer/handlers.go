// backend/internal/customer/handlers.go
// Clean HTTP handlers using service interface
package customer

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for customer operations
type Handler struct {
	service CustomerService // Use interface instead of concrete type
}

// NewHandler creates a new customer handler
func NewHandler(service CustomerService) *Handler {
	return &Handler{service: service}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse creates an error response
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// SuccessResponse creates a success response
func SuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RegisterRoutes registers all customer routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Customer CRUD operations
	router.HandleFunc("/customers", h.GetCustomers).Methods("GET")
	router.HandleFunc("/customers", h.CreateCustomer).Methods("POST")
	router.HandleFunc("/customers/{id:[0-9]+}", h.GetCustomerByID).Methods("GET")
	router.HandleFunc("/customers/{id:[0-9]+}", h.UpdateCustomer).Methods("PUT")
	router.HandleFunc("/customers/{id:[0-9]+}", h.DeleteCustomer).Methods("DELETE")
	
	// Search and analytics
	router.HandleFunc("/customers/search", h.SearchCustomers).Methods("GET")
	router.HandleFunc("/customers/stats", h.GetCustomerStats).Methods("GET")
	
	// Tenant management
	router.HandleFunc("/tenant/context", h.SetTenantContext).Methods("POST")
	router.HandleFunc("/tenant/context", h.GetTenantContext).Methods("GET")
}

// GetCustomers retrieves customers for a tenant
func (h *Handler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}
	
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	customers, err := h.service.GetCustomersByTenant(tenantID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, customers, "Customers retrieved successfully")
}

// GetCustomerByID retrieves a specific customer
func (h *Handler) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}
	
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	customer, err := h.service.GetCustomerByID(tenantID, customerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SuccessResponse(w, customer, "Customer retrieved successfully")
}

// SearchCustomers performs filtered customer search
func (h *Handler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}
	
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	// Parse query parameters
	filter := CustomerFilter{
		TenantID: tenantID,
		State:    r.URL.Query().Get("state"),
		Search:   r.URL.Query().Get("search"),
	}

	// Parse boolean parameters
	if includeDeleted := r.URL.Query().Get("include_deleted"); includeDeleted == "true" {
		filter.IncludeDeleted = true
	}

	// Parse pagination parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	customers, err := h.service.SearchCustomers(filter)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, customers, "Search completed successfully")
}

// CreateCustomer creates a new customer
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required in header")
		return
	}

	var customer Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// Set tenant ID from header
	customer.TenantID = tenantID

	if err := h.service.CreateCustomer(&customer); err != nil {
		if strings.Contains(err.Error(), "validation") {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SuccessResponse(w, customer, "Customer created successfully")
}

// UpdateCustomer updates an existing customer
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required in header")
		return
	}

	var customer Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// Set ID and tenant from URL and header
	customer.CustomerID = customerID
	customer.TenantID = tenantID

	if err := h.service.UpdateCustomer(&customer); err != nil {
		if strings.Contains(err.Error(), "validation") {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		} else if strings.Contains(err.Error(), "not found") {
			ErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SuccessResponse(w, customer, "Customer updated successfully")
}

// DeleteCustomer soft deletes a customer
func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required in header")
		return
	}

	if err := h.service.SoftDeleteCustomer(tenantID, customerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			ErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SuccessResponse(w, nil, "Customer deleted successfully")
}

// GetCustomerStats returns customer statistics
func (h *Handler) GetCustomerStats(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}
	
	if tenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	stats, err := h.service.GetCustomerStats(tenantID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, stats, "Statistics retrieved successfully")
}

// SetTenantContext sets the tenant context
func (h *Handler) SetTenantContext(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TenantID string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if request.TenantID == "" {
		ErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	if err := h.service.SetTenantContext(request.TenantID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, map[string]string{"tenant_id": request.TenantID}, "Tenant context set successfully")
}

// GetTenantContext returns the current tenant context
func (h *Handler) GetTenantContext(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.service.GetCurrentTenant()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, map[string]string{"tenant_id": tenantID}, "Current tenant context retrieved")
}

// Health check endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	SuccessResponse(w, map[string]string{
		"status":  "healthy",
		"service": "customer-service",
	}, "Service is healthy")
}

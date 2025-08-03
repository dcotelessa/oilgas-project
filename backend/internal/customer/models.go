// backend/internal/customer/models.go
package customer

import (
	"time"
)

// Customer represents an oil & gas company with enhanced billing and analytics
type Customer struct {
	CustomerID      int        `json:"customer_id" db:"customer_id"`
	Customer        string     `json:"customer" db:"customer" binding:"required"`
	BillingAddress  *string    `json:"billing_address" db:"billing_address"`
	BillingCity     *string    `json:"billing_city" db:"billing_city"`
	BillingState    *string    `json:"billing_state" db:"billing_state"`
	BillingZipcode  *string    `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         *string    `json:"contact" db:"contact"`
	Phone           *string    `json:"phone" db:"phone"`
	Fax             *string    `json:"fax" db:"fax"`
	Email           *string    `json:"email" db:"email"`
	TenantID        string     `json:"tenant_id" db:"tenant_id"`
	
	// Enhanced contact management
	PrimaryContact  *Contact   `json:"primary_contact,omitempty"`
	BillingContact  *Contact   `json:"billing_contact,omitempty"`
	
	// Customer preferences
	PreferredPaymentTerms *string `json:"preferred_payment_terms" db:"preferred_payment_terms"`
	PreferredShippingMethod *string `json:"preferred_shipping_method" db:"preferred_shipping_method"`
	DefaultPORequired    bool     `json:"default_po_required" db:"default_po_required"`
	CreditLimit          *float64 `json:"credit_limit" db:"credit_limit"`
	
	// System fields
	ImportedAt      *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted         bool       `json:"deleted" db:"deleted"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	
	// Analytics (computed fields - not stored)
	WorkOrderCount  int        `json:"work_order_count,omitempty" db:"-"`
	TotalRevenue    float64    `json:"total_revenue,omitempty" db:"-"`
	LastActivity    *time.Time `json:"last_activity,omitempty" db:"-"`
	ActiveJobs      int        `json:"active_jobs,omitempty" db:"-"`
}

// Contact represents enhanced contact information
type Contact struct {
	Name     string  `json:"name" db:"name"`
	Title    *string `json:"title" db:"title"`
	Phone    *string `json:"phone" db:"phone"`
	Email    *string `json:"email" db:"email"`
	Mobile   *string `json:"mobile" db:"mobile"`
	Preferred bool   `json:"preferred" db:"preferred"`
}

// CustomerAnalytics provides detailed customer performance metrics
type CustomerAnalytics struct {
	CustomerID       int                 `json:"customer_id"`
	TotalWorkOrders  int                 `json:"total_work_orders"`
	CompletedOrders  int                 `json:"completed_orders"`
	PendingOrders    int                 `json:"pending_orders"`
	TotalRevenue     float64             `json:"total_revenue"`
	AverageJobValue  float64             `json:"average_job_value"`
	LastOrderDate    *time.Time          `json:"last_order_date"`
	RecentWorkOrders []RecentWorkOrder   `json:"recent_work_orders"`
	MonthlyRevenue   []MonthlyRevenue    `json:"monthly_revenue"`
	ServiceBreakdown []ServiceBreakdown  `json:"service_breakdown"`
}

type RecentWorkOrder struct {
	WorkOrder    string     `json:"work_order"`
	ServiceDate  *time.Time `json:"service_date"`
	Status       string     `json:"status"`
	TotalValue   float64    `json:"total_value"`
	Description  string     `json:"description"`
}

type MonthlyRevenue struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type ServiceBreakdown struct {
	Service    string  `json:"service"`
	Count      int     `json:"count"`
	Revenue    float64 `json:"revenue"`
	Percentage float64 `json:"percentage"`
}

// CustomerFilters for search and filtering
type CustomerFilters struct {
	Query        string    `json:"query" form:"q"`
	State        string    `json:"state" form:"state"`
	Active       *bool     `json:"active" form:"active"`
	HasOrders    *bool     `json:"has_orders" form:"has_orders"`
	CreditStatus string    `json:"credit_status" form:"credit_status"`
	DateFrom     *time.Time `json:"date_from" form:"date_from"`
	DateTo       *time.Time `json:"date_to" form:"date_to"`
	SortBy       string    `json:"sort_by" form:"sort_by"`
	SortOrder    string    `json:"sort_order" form:"sort_order"`
	Limit        int       `json:"limit" form:"limit"`
	Offset       int       `json:"offset" form:"offset"`
}

// Request/Response DTOs
type CreateCustomerRequest struct {
	Customer                string   `json:"customer" binding:"required,min=2,max=255"`
	BillingAddress          *string  `json:"billing_address" binding:"omitempty,max=500"`
	BillingCity             *string  `json:"billing_city" binding:"omitempty,max=100"`
	BillingState            *string  `json:"billing_state" binding:"omitempty,len=2"`
	BillingZipcode          *string  `json:"billing_zipcode" binding:"omitempty,max=20"`
	Contact                 *string  `json:"contact" binding:"omitempty,max=255"`
	Phone                   *string  `json:"phone" binding:"omitempty,max=50"`
	Fax                     *string  `json:"fax" binding:"omitempty,max=50"`
	Email                   *string  `json:"email" binding:"omitempty,email,max=255"`
	PreferredPaymentTerms   *string  `json:"preferred_payment_terms" binding:"omitempty,max=100"`
	PreferredShippingMethod *string  `json:"preferred_shipping_method" binding:"omitempty,max=100"`
	DefaultPORequired       bool     `json:"default_po_required"`
	CreditLimit             *float64 `json:"credit_limit" binding:"omitempty,min=0"`
}

type UpdateCustomerRequest struct {
	Customer                *string  `json:"customer" binding:"omitempty,min=2,max=255"`
	BillingAddress          *string  `json:"billing_address" binding:"omitempty,max=500"`
	BillingCity             *string  `json:"billing_city" binding:"omitempty,max=100"`
	BillingState            *string  `json:"billing_state" binding:"omitempty,len=2"`
	BillingZipcode          *string  `json:"billing_zipcode" binding:"omitempty,max=20"`
	Contact                 *string  `json:"contact" binding:"omitempty,max=255"`
	Phone                   *string  `json:"phone" binding:"omitempty,max=50"`
	Fax                     *string  `json:"fax" binding:"omitempty,max=50"`
	Email                   *string  `json:"email" binding:"omitempty,email,max=255"`
	PreferredPaymentTerms   *string  `json:"preferred_payment_terms" binding:"omitempty,max=100"`
	PreferredShippingMethod *string  `json:"preferred_shipping_method" binding:"omitempty,max=100"`
	DefaultPORequired       *bool    `json:"default_po_required"`
	CreditLimit             *float64 `json:"credit_limit" binding:"omitempty,min=0"`
}

type UpdateContactsRequest struct {
	PrimaryContact *Contact `json:"primary_contact"`
	BillingContact *Contact `json:"billing_contact"`
}

type CustomerSearchResponse struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	HasMore   bool       `json:"has_more"`
}

// Validation methods
func (c *Customer) Validate() error {
	if len(c.Customer) == 0 {
		return ErrCustomerNameRequired
	}
	if len(c.Customer) > 255 {
		return ErrCustomerNameTooLong
	}
	if c.Email != nil && *c.Email != "" {
		// Basic email validation could be enhanced
		if len(*c.Email) > 255 {
			return ErrEmailTooLong
		}
	}
	return nil
}

// Domain errors
type CustomerError string

func (e CustomerError) Error() string {
	return string(e)
}

const (
	ErrCustomerNotFound      CustomerError = "customer not found"
	ErrCustomerNameRequired  CustomerError = "customer name is required"
	ErrCustomerNameTooLong   CustomerError = "customer name too long (max 255 characters)"
	ErrEmailTooLong          CustomerError = "email too long (max 255 characters)"
	ErrInvalidCustomerID     CustomerError = "invalid customer ID"
	ErrCustomerExists        CustomerError = "customer already exists"
	ErrTenantMismatch        CustomerError = "customer belongs to different tenant"
)

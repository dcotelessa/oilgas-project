// backend/internal/models/customer.go
package models

import "time"

// Customer represents an oil & gas company that purchases services
type Customer struct {
	CustomerID      int        `json:"customer_id" db:"customer_id"`
	Customer        string     `json:"customer" db:"customer"`
	BillingAddress  *string    `json:"billing_address" db:"billing_address"`
	BillingCity     *string    `json:"billing_city" db:"billing_city"`
	BillingState    *string    `json:"billing_state" db:"billing_state"`
	BillingZipcode  *string    `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         *string    `json:"contact" db:"contact"`
	Phone           *string    `json:"phone" db:"phone"`
	Fax             *string    `json:"fax" db:"fax"`
	Email           *string    `json:"email" db:"email"`
	TenantID        string     `json:"tenant_id" db:"tenant_id"`
	ImportedAt      *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted         bool       `json:"deleted" db:"deleted"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// CustomerRelatedData aggregates related information for a customer
type CustomerRelatedData struct {
	InventoryCount   int               `json:"inventory_count"`
	WorkOrderCount   int               `json:"work_order_count"`
	LastActivity     *time.Time        `json:"last_activity"`
	RecentWorkOrders []RecentWorkOrder `json:"recent_work_orders,omitempty"`
}

type RecentWorkOrder struct {
	WorkOrder  string `json:"work_order"`
	LatestDate string `json:"latest_date"`
}

// CreateCustomerRequest for API endpoints
type CreateCustomerRequest struct {
	Customer       string  `json:"customer" binding:"required"`
	BillingAddress *string `json:"billing_address"`
	BillingCity    *string `json:"billing_city"`
	BillingState   *string `json:"billing_state"`
	BillingZipcode *string `json:"billing_zipcode"`
	Contact        *string `json:"contact"`
	Phone          *string `json:"phone"`
	Fax            *string `json:"fax"`
	Email          *string `json:"email"`
}

type UpdateCustomerRequest struct {
	Customer       *string `json:"customer"`
	BillingAddress *string `json:"billing_address"`
	BillingCity    *string `json:"billing_city"`
	BillingState   *string `json:"billing_state"`
	BillingZipcode *string `json:"billing_zipcode"`
	Contact        *string `json:"contact"`
	Phone          *string `json:"phone"`
	Fax            *string `json:"fax"`
	Email          *string `json:"email"`
}

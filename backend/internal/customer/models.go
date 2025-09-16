// backend/internal/customer/models.go
package customer

import "time"

type Customer struct {
	ID            int       `json:"id" db:"id"`
	TenantID      string    `json:"tenant_id" db:"tenant_id" validate:"required,max=100"`
	Name          string    `json:"name" db:"name" validate:"required,max=255"`
	CompanyCode   *string   `json:"company_code,omitempty" db:"company_code" validate:"omitempty,alphanum,min=2,max=50"`
	Status        Status    `json:"status" db:"status" validate:"required,oneof=active inactive suspended"`
	TaxID         *string   `json:"tax_id,omitempty" db:"tax_id" validate:"omitempty,max=50"`
	PaymentTerms  string    `json:"payment_terms" db:"payment_terms" validate:"max=50"`
	BillingStreet *string   `json:"billing_street,omitempty" db:"billing_street" validate:"omitempty,max=255"`
	BillingCity   *string   `json:"billing_city,omitempty" db:"billing_city" validate:"omitempty,max=100"`
	BillingState  *string   `json:"billing_state,omitempty" db:"billing_state" validate:"omitempty,max=10"`
	BillingZip    *string   `json:"billing_zip_code,omitempty" db:"billing_zip_code" validate:"omitempty,max=20"`
	BillingCountry string   `json:"billing_country" db:"billing_country" validate:"len=2"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive" 
	StatusSuspended Status = "suspended"
)

type ContactType string

const (
	ContactTypePrimary  ContactType = "PRIMARY"
	ContactTypeBilling  ContactType = "BILLING"
	ContactTypeShipping ContactType = "SHIPPING"
	ContactTypeApprover ContactType = "APPROVER"
)

type CustomerContact struct {
	ID           int         `json:"id" db:"id"`
	CustomerID   int         `json:"customer_id" db:"customer_id" validate:"required"`
	AuthUserID   int         `json:"auth_user_id" db:"auth_user_id" validate:"required"`
	ContactType  ContactType `json:"contact_type" db:"contact_type" validate:"required,oneof=PRIMARY BILLING SHIPPING APPROVER"`
	IsPrimary    bool        `json:"is_primary" db:"is_primary"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	FullName     *string     `json:"full_name,omitempty" db:"full_name" validate:"omitempty,max=255"`
	Email        *string     `json:"email,omitempty" db:"email" validate:"omitempty,email,max=255"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

type CustomerAnalytics struct {
	CustomerID      int        `json:"customer_id"`
	TotalWorkOrders int        `json:"total_work_orders"`
	ActiveOrders    int        `json:"active_orders"`
	TotalRevenue    float64    `json:"total_revenue"`
	AvgOrderValue   float64    `json:"avg_order_value"`
	LastOrderDate   *time.Time `json:"last_order_date,omitempty"`
}

type SearchFilters struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,max=255"`
	CompanyCode string   `json:"company_code,omitempty" validate:"omitempty,alphanum,max=50"`
	Status      []Status `json:"status,omitempty"`
	TaxID       string   `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	Limit       int      `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
	Offset      int      `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// CustomerFilter for service layer filtering
type CustomerFilter struct {
	TenantID    string   `json:"tenant_id,omitempty"`
	Name        string   `json:"name,omitempty"`
	CompanyCode string   `json:"company_code,omitempty"`
	Status      []Status `json:"status,omitempty"`
	TaxID       string   `json:"tax_id,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	Offset      int      `json:"offset,omitempty"`
}

// Address represents billing/shipping address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

// BillingInfo groups billing-related information
type BillingInfo struct {
	TaxID        string  `json:"tax_id"`
	PaymentTerms string  `json:"payment_terms"`
	Address      Address `json:"address"`
}

// ContactRegistrationResponse for contact registration workflows
type ContactRegistrationResponse struct {
	CustomerContact *CustomerContact `json:"customer_contact"`
	AuthUser        interface{}      `json:"auth_user"` // Will be auth.User when auth is fixed
	Message         string           `json:"message"`
}

// ContactInfo represents contact information for registration
type ContactInfo struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// BulkContactRegistrationRequest for bulk contact operations
type BulkContactRegistrationRequest struct {
	Contacts []ContactInfo `json:"contacts"`
}

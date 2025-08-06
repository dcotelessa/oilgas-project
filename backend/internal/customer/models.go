// backend/internal/customer/models.go
package customer

import (
    "time"
    "database/sql/driver"
    "encoding/json"
)

// Customer represents the main customer entity
type Customer struct {
    ID                   int       `json:"id" db:"id"`
    TenantID             string    `json:"tenant_id" db:"tenant_id"`
    Name                 string    `json:"name" db:"name"`
    CompanyCode          *string   `json:"company_code" db:"company_code"`
    Address              *string   `json:"address" db:"address"`
    Phone                *string   `json:"phone" db:"phone"`
    Email                *string   `json:"email" db:"email"`
    BillingContactName   *string   `json:"billing_contact_name" db:"billing_contact_name"`
    BillingContactEmail  *string   `json:"billing_contact_email" db:"billing_contact_email"`
    BillingContactPhone  *string   `json:"billing_contact_phone" db:"billing_contact_phone"`
    IsActive             bool      `json:"is_active" db:"is_active"`
    CreatedAt            time.Time `json:"created_at" db:"created_at"`
    UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
    
    // Joined data from relationships
    ContactCount         int       `json:"contact_count,omitempty"`
    ContactEmails        []string  `json:"contact_emails,omitempty"`
}

// CustomerAuthContact represents the junction table for customer-auth relationships
type CustomerAuthContact struct {
    ID              int      `json:"id" db:"id"`
    CustomerID      int      `json:"customer_id" db:"customer_id"`
    AuthUserID      int      `json:"auth_user_id" db:"auth_user_id"`
    ContactType     string   `json:"contact_type" db:"contact_type"`
    YardPermissions []string `json:"yard_permissions" db:"yard_permissions"`
    IsActive        bool     `json:"is_active" db:"is_active"`
    CreatedBy       *int     `json:"created_by" db:"created_by"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Implement sql.Scanner and driver.Valuer for YardPermissions slice
func (yp *[]string) Scan(value interface{}) error {
    if value == nil {
        *yp = []string{}
        return nil
    }
    // PostgreSQL array parsing logic here
    return json.Unmarshal(value.([]byte), yp)
}

func (yp []string) Value() (driver.Value, error) {
    return json.Marshal(yp)
}

// CustomerWithContacts represents customer with full contact information
type CustomerWithContacts struct {
    Customer
    Contacts []CustomerContact `json:"contacts"`
}

// CustomerContact represents a customer's auth contact with user details
type CustomerContact struct {
    UserID          int      `json:"user_id" db:"user_id"`
    Email           string   `json:"email" db:"email"`
    FullName        string   `json:"full_name" db:"full_name"`
    ContactType     string   `json:"contact_type" db:"contact_type"`
    YardPermissions []string `json:"yard_permissions" db:"yard_permissions"`
    IsActive        bool     `json:"is_active" db:"contact_active"`
    LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
}

// Request/Response structs
type CreateCustomerRequest struct {
    Name                string  `json:"name" binding:"required"`
    CompanyCode         *string `json:"company_code"`
    Address             *string `json:"address"`
    Phone               *string `json:"phone"`
    Email               *string `json:"email"`
    BillingContactName  *string `json:"billing_contact_name"`
    BillingContactEmail *string `json:"billing_contact_email"`
    BillingContactPhone *string `json:"billing_contact_phone"`
}

type UpdateCustomerRequest struct {
    Name                *string `json:"name"`
    CompanyCode         *string `json:"company_code"`
    Address             *string `json:"address"`
    Phone               *string `json:"phone"`
    Email               *string `json:"email"`
    BillingContactName  *string `json:"billing_contact_name"`
    BillingContactEmail *string `json:"billing_contact_email"`
    BillingContactPhone *string `json:"billing_contact_phone"`
    IsActive            *bool   `json:"is_active"`
}

type CustomerSearchFilters struct {
    Query       string `form:"q"`
    CompanyCode string `form:"company_code"`
    IsActive    *bool  `form:"is_active"`
    HasContacts *bool  `form:"has_contacts"`
    Page        int    `form:"page"`
    Limit       int    `form:"limit"`
}

// Admin contact registration request
type RegisterContactRequest struct {
    CustomerID      int      `json:"customer_id" binding:"required"`
    Email           string   `json:"email" binding:"required,email"`
    FullName        string   `json:"full_name" binding:"required"`
    ContactType     string   `json:"contact_type" binding:"required"`
    YardPermissions []string `json:"yard_permissions"`
    TemporaryPassword bool   `json:"temporary_password"`
}

type BulkContactRegistrationRequest struct {
    CustomerID int                    `json:"customer_id" binding:"required"`
    Contacts   []RegisterContactRequest `json:"contacts" binding:"required,max=10"`
}

type ContactRegistrationResponse struct {
    ContactID         int    `json:"contact_id"`
    UserID            int    `json:"user_id"`
    Email             string `json:"email"`
    TemporaryPassword string `json:"temporary_password,omitempty"`
    Success           bool   `json:"success"`
    Error             string `json:"error,omitempty"`
}

// Customer analytics
type CustomerAnalytics struct {
    TotalCustomers      int `json:"total_customers"`
    ActiveCustomers     int `json:"active_customers"`
    CustomersWithContacts int `json:"customers_with_contacts"`
    TotalContacts       int `json:"total_contacts"`
    ContactsByType      map[string]int `json:"contacts_by_type"`
}

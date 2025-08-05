// backend/internal/auth/models.go
package auth

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Enhanced User model with enterprise and yard-level access
type User struct {
	ID               int               `json:"id" db:"id"`
	Username         string            `json:"username" db:"username"`
	Email            string            `json:"email" db:"email"`
	FullName         string            `json:"full_name" db:"full_name"`
	PasswordHash     string            `json:"-" db:"password_hash"`
	
	// Role and access control (CFM compatibility)
	Role             UserRole          `json:"role" db:"role"`
	AccessLevel      int               `json:"access_level" db:"access_level"`
	IsEnterpriseUser bool              `json:"is_enterprise_user" db:"is_enterprise_user"`
	
	// Multi-tenant and yard access
	TenantAccess     TenantAccessList  `json:"tenant_access" db:"tenant_access"`
	PrimaryTenantID  string            `json:"primary_tenant_id" db:"primary_tenant_id"`
	
	// Customer relationship (contact-based login)
	CustomerID       *int              `json:"customer_id" db:"customer_id"`
	ContactType      ContactType       `json:"contact_type" db:"contact_type"`
	
	// Session management
	IsActive         bool              `json:"is_active" db:"is_active"`
	LastLoginAt      *time.Time        `json:"last_login_at" db:"last_login_at"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
}

// UserRole defines system-wide roles
type UserRole string

const (
	RoleCustomerContact  UserRole = "CUSTOMER_CONTACT"  // Customer contact login
	RoleOperator        UserRole = "OPERATOR"          // Yard operations
	RoleManager         UserRole = "MANAGER"           // Site management  
	RoleAdmin           UserRole = "ADMIN"             // Tenant admin
	RoleEnterpriseAdmin UserRole = "ENTERPRISE_ADMIN"  // Cross-tenant access
	RoleSystemAdmin     UserRole = "SYSTEM_ADMIN"      // Full system access
)

// ContactType for customer contact users
type ContactType string

const (
	ContactPrimary   ContactType = "PRIMARY"
	ContactBilling   ContactType = "BILLING"
	ContactShipping  ContactType = "SHIPPING"
	ContactApprover  ContactType = "APPROVER"
)

// TenantAccess defines user access to specific tenant with yard-level permissions
type TenantAccess struct {
	TenantID         string            `json:"tenant_id"`
	Role             UserRole          `json:"role"`
	Permissions      []Permission      `json:"permissions"`
	YardAccess       []YardAccess      `json:"yard_access"`
}

// YardAccess defines granular yard-level permissions
type YardAccess struct {
	YardLocation       string `json:"yard_location"`
	CanViewWorkOrders  bool   `json:"can_view_work_orders"`
	CanCreateWorkOrders bool  `json:"can_create_work_orders"`
	CanApproveOrders   bool   `json:"can_approve_orders"`
	CanViewInventory   bool   `json:"can_view_inventory"`
	CanManageTransport bool   `json:"can_manage_transport"`
	CanExportData      bool   `json:"can_export_data"`
}

// TenantAccessList for database storage
type TenantAccessList []TenantAccess

// Value implements driver.Valuer for database storage
func (tal TenantAccessList) Value() (driver.Value, error) {
	return json.Marshal(tal)
}

// Scan implements sql.Scanner for database retrieval
func (tal *TenantAccessList) Scan(value interface{}) error {
	if value == nil {
		*tal = TenantAccessList{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TenantAccessList", value)
	}
	
	return json.Unmarshal(bytes, tal)
}

// Session represents active user sessions
type Session struct {
	ID                string    `json:"id" db:"id"`
	UserID            int       `json:"user_id" db:"user_id"`
	Token             string    `json:"-" db:"token"`
	RefreshToken      string    `json:"-" db:"refresh_token"`
	TenantContext     *TenantAccess `json:"tenant_context" db:"tenant_context"`
	ExpiresAt         time.Time `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt  *time.Time `json:"refresh_expires_at" db:"refresh_expires_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	LastUsedAt        time.Time `json:"last_used_at" db:"last_used_at"`
	UserAgent         string    `json:"user_agent" db:"user_agent"`
	IPAddress         string    `json:"ip_address" db:"ip_address"`
}

// UserResponse for API responses (excludes sensitive fields)
type UserResponse struct {
	ID               int               `json:"id"`
	Username         string            `json:"username"`
	Email            string            `json:"email"`
	FullName         string            `json:"full_name"`
	Role             UserRole          `json:"role"`
	AccessLevel      int               `json:"access_level"`
	IsEnterpriseUser bool              `json:"is_enterprise_user"`
	TenantAccess     TenantAccessList  `json:"tenant_access"`
	PrimaryTenantID  string            `json:"primary_tenant_id"`
	CustomerID       *int              `json:"customer_id,omitempty"`
	ContactType      ContactType       `json:"contact_type,omitempty"`
	IsActive         bool              `json:"is_active"`
	LastLoginAt      *time.Time        `json:"last_login_at"`
	CreatedAt        time.Time         `json:"created_at"`
}

// Convert User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		FullName:         u.FullName,
		Role:             u.Role,
		AccessLevel:      u.AccessLevel,
		IsEnterpriseUser: u.IsEnterpriseUser,
		TenantAccess:     u.TenantAccess,
		PrimaryTenantID:  u.PrimaryTenantID,
		CustomerID:       u.CustomerID,
		ContactType:      u.ContactType,
		IsActive:         u.IsActive,
		LastLoginAt:      u.LastLoginAt,
		CreatedAt:        u.CreatedAt,
	}
}

// User permission check for middleware
type UserPermissionCheck struct {
	UserID       int        `json:"user_id"`
	TenantID     string     `json:"tenant_id"`
	YardLocation *string    `json:"yard_location,omitempty"`
	Permission   Permission `json:"permission"`
	ResourceID   *string    `json:"resource_id,omitempty"`
}

// Enterprise context for cross-tenant operations
type EnterpriseContext struct {
	UserID            int              `json:"user_id"`
	AccessibleTenants []TenantAccess   `json:"accessible_tenants"`
	IsEnterpriseAdmin bool             `json:"is_enterprise_admin"`
	CrossTenantPerms  []Permission     `json:"cross_tenant_permissions"`
}

// Customer access context for contact users
type CustomerAccessContext struct {
	CustomerID       int                     `json:"customer_id"`
	AccessibleYards  []YardAccess           `json:"accessible_yards"`
	TenantAccess     map[string]TenantAccess `json:"tenant_access"` // tenantID -> access
}

// Request/Response types for user management
type CreateCustomerContactRequest struct {
	CustomerID  int          `json:"customer_id" binding:"required"`
	TenantID    string       `json:"tenant_id" binding:"required"`
	Email       string       `json:"email" binding:"required,email"`
	FullName    string       `json:"full_name" binding:"required"`
	Password    string       `json:"password" binding:"required,min=8"`
	ContactType ContactType  `json:"contact_type"`
	YardAccess  []YardAccess `json:"yard_access"`
}

type CreateEnterpriseUserRequest struct {
	Username         string           `json:"username" binding:"required"`
	Email            string           `json:"email" binding:"required,email"`
	FullName         string           `json:"full_name" binding:"required"`
	Password         string           `json:"password" binding:"required,min=8"`
	Role             UserRole         `json:"role" binding:"required"`
	IsEnterpriseUser bool             `json:"is_enterprise_user"`
	PrimaryTenantID  string           `json:"primary_tenant_id"`
	TenantAccess     []TenantAccess   `json:"tenant_access" binding:"required"`
}

type UpdateUserTenantAccessRequest struct {
	UserID       int            `json:"user_id"`
	TenantAccess []TenantAccess `json:"tenant_access" binding:"required"`
}

type UpdateUserYardAccessRequest struct {
	UserID       int          `json:"user_id"`
	TenantID     string       `json:"tenant_id" binding:"required"`
	YardAccess   []YardAccess `json:"yard_access" binding:"required"`
}

type UserUpdates struct {
	FullName        *string    `json:"full_name,omitempty"`
	Email           *string    `json:"email,omitempty"`
	Role            *UserRole  `json:"role,omitempty"`
	IsActive        *bool      `json:"is_active,omitempty"`
	CurrentPassword *string    `json:"current_password,omitempty"`
	NewPassword     *string    `json:"new_password,omitempty"`
}

// Helper methods for User model
func (u *User) IsCustomerContact() bool {
	return u.Role == RoleCustomerContact && u.CustomerID != nil
}

func (u *User) CanManageOtherUsers() bool {
	return u.Role == RoleAdmin || u.Role == RoleEnterpriseAdmin || u.Role == RoleSystemAdmin
}

func (u *User) CanPerformCrossTenantOperation() bool {
	return u.IsEnterpriseUser && (u.Role == RoleEnterpriseAdmin || u.Role == RoleSystemAdmin)
}

func (u *User) CanAccessTenant(tenantID string) bool {
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	for _, access := range u.TenantAccess {
		if access.TenantID == tenantID {
			return true
		}
	}
	
	return false
}

func (u *User) HasAccessToYard(tenantID, yardLocation string) bool {
	for _, tenantAccess := range u.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			for _, yardAccess := range tenantAccess.YardAccess {
				if yardAccess.YardLocation == yardLocation {
					return true
				}
			}
		}
	}
	
	return false
}

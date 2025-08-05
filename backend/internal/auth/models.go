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
	CanRead          bool              `json:"can_read"`
	CanWrite         bool              `json:"can_write"`
	CanDelete        bool              `json:"can_delete"`
	CanApprove       bool              `json:"can_approve"`
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

// Implement driver.Valuer for database storage
func (tal TenantAccessList) Value() (driver.Value, error) {
	return json.Marshal(tal)
}

// Implement sql.Scanner for database retrieval
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

// Session represents active user sessions with proper fields
type Session struct {
	ID                string           `json:"id" db:"id"`
	UserID            int              `json:"user_id" db:"user_id"`
	TenantID          string           `json:"tenant_id" db:"tenant_id"`         // Added missing field
	Token             string           `json:"-" db:"token"`
	RefreshToken      string           `json:"-" db:"refresh_token"`
	TenantContext     *TenantAccess    `json:"tenant_context" db:"tenant_context"`
	IsActive          bool             `json:"is_active" db:"is_active"`         // Added missing field
	ExpiresAt         time.Time        `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt  *time.Time       `json:"refresh_expires_at" db:"refresh_expires_at"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	LastUsedAt        time.Time        `json:"last_used_at" db:"last_used_at"`
	UserAgent         string           `json:"user_agent" db:"user_agent"`
	IPAddress         string           `json:"ip_address" db:"ip_address"`
}

// Permission represents system permissions
type Permission string

const (
	PermissionViewInventory     Permission = "VIEW_INVENTORY"
	PermissionCreateWorkOrder   Permission = "CREATE_WORK_ORDER"
	PermissionApproveWorkOrder  Permission = "APPROVE_WORK_ORDER"
	PermissionManageTransport   Permission = "MANAGE_TRANSPORT"
	PermissionExportData       Permission = "EXPORT_DATA"
	PermissionUserManagement   Permission = "USER_MANAGEMENT"
	PermissionCrossTenantView  Permission = "CROSS_TENANT_VIEW"
)

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

// LoginRequest for authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TenantID string `json:"tenant_id,omitempty"`
}

// LoginResponse for authentication
type LoginResponse struct {
	Token         string          `json:"token"`
	User          UserResponse    `json:"user"`
	TenantContext *TenantAccess   `json:"tenant_context"`
	ExpiresAt     time.Time       `json:"expires_at"`
	RefreshToken  string          `json:"refresh_token"`
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

// UserUpdates for updating user information (moved from service.go to avoid duplicate)
type UserUpdates struct {
	FullName         *string    `json:"full_name,omitempty"`
	Email            *string    `json:"email,omitempty"`
	Role             *UserRole  `json:"role,omitempty"`
	IsActive         *bool      `json:"is_active,omitempty"`
	IsEnterpriseUser *bool      `json:"is_enterprise_user,omitempty"`
	PrimaryTenantID  *string    `json:"primary_tenant_id,omitempty"`
	CurrentPassword  *string    `json:"current_password,omitempty"`
	NewPassword      *string    `json:"new_password,omitempty"`
}

// UserSearchFilters for search functionality (fix singular/plural issue)
type UserSearchFilters struct {
	Query                string     `json:"query"`
	TenantID            string     `json:"tenant_id"`
	Role                UserRole   `json:"role"`
	CustomerID          *int       `json:"customer_id"`
	OnlyCustomerContacts bool       `json:"only_customer_contacts"`
	AccessibleTenants   []string   `json:"accessible_tenants"`
	Limit               int        `json:"limit"`
	Offset              int        `json:"offset"`
}

// AdminUserFilters for admin user management
type AdminUserFilters struct {
	TenantID          string   `json:"tenant_id"`
	Role              UserRole `json:"role"`
	AccessibleTenants []string `json:"accessible_tenants"`
	Limit             int      `json:"limit"`
	Offset            int      `json:"offset"`
}

// UserStatsRequest for user statistics (no parameters needed)
type UserStatsRequest struct {
	// Empty struct - stats calculated for current user's accessible tenants
}

// UserStats response
type UserStats struct {
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	CustomerContacts int `json:"customer_contacts"`
	EnterpriseUsers  int `json:"enterprise_users"`
}

// User permission check for middleware
type UserPermissionCheck struct {
	UserID       int        `json:"user_id"`
	TenantID     string     `json:"tenant_id"`
	YardLocation *string    `json:"yard_location,omitempty"`
	Permission   Permission `json:"permission"`
	ResourceID   *string    `json:"resource_id,omitempty"`
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

// HasPermissionInTenant checks if user has a specific permission in a tenant
func (u *User) HasPermissionInTenant(tenantID string, permission Permission) bool {
	// Enterprise users have all permissions
	if u.IsEnterpriseUser && (u.Role == RoleEnterpriseAdmin || u.Role == RoleSystemAdmin) {
		return true
	}
	
	// Check tenant-specific permissions
	for _, access := range u.TenantAccess {
		if access.TenantID == tenantID {
			for _, p := range access.Permissions {
				if p == permission {
					return true
				}
			}
		}
	}
	
	return false
}

// GetCustomerAccessContext returns customer access context for customer contacts
func (u *User) GetCustomerAccessContext() *CustomerAccessContext {
	if !u.IsCustomerContact() {
		return nil
	}
	
	// Build yard access from all tenant access
	allYards := []YardAccess{}
	tenantMap := make(map[string]TenantAccess)
	
	for _, access := range u.TenantAccess {
		allYards = append(allYards, access.YardAccess...)
		tenantMap[access.TenantID] = access
	}
	
	return &CustomerAccessContext{
		CustomerID:      *u.CustomerID,
		AccessibleYards: allYards,
		TenantAccess:    tenantMap,
	}
}

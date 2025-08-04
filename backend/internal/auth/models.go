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
	
	// Role and access control (CFM access levels 1-5+)
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
	
	// Yard-level access within tenant
	YardAccess       []YardAccess      `json:"yard_access"`
	
	// Tenant-level permissions
	CanRead          bool              `json:"can_read"`
	CanWrite         bool              `json:"can_write"`
	CanDelete        bool              `json:"can_delete"`
	CanApprove       bool              `json:"can_approve"`
}

// YardAccess defines permissions within specific yard
type YardAccess struct {
	YardLocation     string       `json:"yard_location"`
	CanViewWorkOrders    bool         `json:"can_view_work_orders"`
	CanViewInventory     bool         `json:"can_view_inventory"`
	CanCreateWorkOrders  bool         `json:"can_create_work_orders"`
	CanApproveOrders     bool         `json:"can_approve_orders"`
	CanManageTransport   bool         `json:"can_manage_transport"`
	CanExportData        bool         `json:"can_export_data"`
}

type TenantAccessList []TenantAccess

// Database serialization for TenantAccessList
func (tal TenantAccessList) Value() (driver.Value, error) {
	return json.Marshal(tal)
}

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

// Permission defines granular access control
type Permission string

const (
	// Customer domain permissions
	PermissionCustomerRead   Permission = "CUSTOMER_READ"
	PermissionCustomerWrite  Permission = "CUSTOMER_WRITE"
	PermissionCustomerDelete Permission = "CUSTOMER_DELETE"
	
	// Inventory domain permissions
	PermissionInventoryRead   Permission = "INVENTORY_READ"
	PermissionInventoryWrite  Permission = "INVENTORY_WRITE"
	PermissionInventoryDelete Permission = "INVENTORY_DELETE"
	PermissionInventoryExport Permission = "INVENTORY_EXPORT"
	
	// Work order domain permissions
	PermissionWorkOrderRead     Permission = "WORK_ORDER_READ"
	PermissionWorkOrderWrite    Permission = "WORK_ORDER_WRITE"
	PermissionWorkOrderDelete   Permission = "WORK_ORDER_DELETE"
	PermissionWorkOrderApprove  Permission = "WORK_ORDER_APPROVE"
	PermissionWorkOrderInvoice  Permission = "WORK_ORDER_INVOICE"
	
	// Transport permissions
	PermissionTransportRead  Permission = "TRANSPORT_READ"
	PermissionTransportWrite Permission = "TRANSPORT_WRITE"
	
	// Enterprise permissions
	PermissionCrossTenantView Permission = "CROSS_TENANT_VIEW"
	PermissionUserManagement  Permission = "USER_MANAGEMENT"
	PermissionSystemConfig    Permission = "SYSTEM_CONFIG"
)

// Session represents active user sessions
type Session struct {
	ID          string    `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Token       string    `json:"-" db:"token"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Authentication requests and responses
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	TenantID  string `json:"tenant_id"` // Optional - can be derived from user
}

type LoginResponse struct {
	Token           string           `json:"token"`
	User            UserResponse     `json:"user"`
	TenantContext   TenantAccess     `json:"tenant_context"`
	ExpiresAt       time.Time        `json:"expires_at"`
	RefreshToken    string           `json:"refresh_token"`
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
	UserID      int        `json:"user_id"`
	TenantID    string     `json:"tenant_id"`
	YardLocation *string   `json:"yard_location,omitempty"`
	Permission  Permission `json:"permission"`
	ResourceID  *string    `json:"resource_id,omitempty"`
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
	CustomerID       int                    `json:"customer_id"`
	AccessibleYards  []YardAccess          `json:"accessible_yards"`
	TenantAccess     map[string]TenantAccess `json:"tenant_access"` // tenantID -> access
}

// Helper methods for User model
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

func (u *User) HasAccessToTenant(tenantID string) bool {
	if u.IsEnterpriseUser {
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

func (u *User) HasPermissionInTenant(tenantID string, permission Permission) bool {
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	for _, access := range u.TenantAccess {
		if access.TenantID == tenantID {
			for _, perm := range access.Permissions {
				if perm == permission {
					return true
				}
			}
		}
	}
	
	return false
}

func (u *User) HasYardPermission(tenantID, yardLocation string, checkFunc func(YardAccess) bool) bool {
	for _, tenantAccess := range u.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			for _, yardAccess := range tenantAccess.YardAccess {
				if yardAccess.YardLocation == yardLocation {
					return checkFunc(yardAccess)
				}
			}
		}
	}
	
	return false
}

func (u *User) CanPerformCrossTenantOperation() bool {
	return u.IsEnterpriseUser || u.Role == RoleEnterpriseAdmin || u.Role == RoleSystemAdmin
}

func (u *User) IsCustomerContact() bool {
	return u.Role == RoleCustomerContact && u.CustomerID != nil
}

func (u *User) GetAccessibleYardsForTenant(tenantID string) []YardAccess {
	for _, tenantAccess := range u.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			return tenantAccess.YardAccess
		}
	}
	
	return []YardAccess{}
}

// Customer contact specific methods
func (u *User) GetCustomerAccessContext() *CustomerAccessContext {
	if !u.IsCustomerContact() {
		return nil
	}
	
	var allYards []YardAccess
	tenantAccessMap := make(map[string]TenantAccess)
	
	for _, tenantAccess := range u.TenantAccess {
		tenantAccessMap[tenantAccess.TenantID] = tenantAccess
		allYards = append(allYards, tenantAccess.YardAccess...)
	}
	
	return &CustomerAccessContext{
		CustomerID:      *u.CustomerID,
		AccessibleYards: allYards,
		TenantAccess:    tenantAccessMap,
	}
}

// Create/Update user requests
type CreateCustomerContactRequest struct {
	CustomerID   int         `json:"customer_id" binding:"required"`
	TenantID     string      `json:"tenant_id" binding:"required"`
	Email        string      `json:"email" binding:"required,email"`
	FullName     string      `json:"full_name" binding:"required"`
	Password     string      `json:"password" binding:"required,min=8"`
	ContactType  ContactType `json:"contact_type" binding:"required"`
	YardAccess   []YardAccess `json:"yard_access"`
}

type CreateEnterpriseUserRequest struct {
	Username         string           `json:"username" binding:"required"`
	Email            string           `json:"email" binding:"required,email"`
	FullName         string           `json:"full_name" binding:"required"`
	Password         string           `json:"password" binding:"required,min=8"`
	Role             UserRole         `json:"role" binding:"required"`
	IsEnterpriseUser bool             `json:"is_enterprise_user"`
	PrimaryTenantID  string           `json:"primary_tenant_id" binding:"required"`
	TenantAccess     TenantAccessList `json:"tenant_access"`
}

type UpdateUserTenantAccessRequest struct {
	UserID       int              `json:"user_id" binding:"required"`
	TenantAccess TenantAccessList `json:"tenant_access" binding:"required"`
}

type UpdateUserYardAccessRequest struct {
	UserID       int          `json:"user_id" binding:"required"`
	TenantID     string       `json:"tenant_id" binding:"required"`
	YardAccess   []YardAccess `json:"yard_access" binding:"required"`
}

// Custom errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e AuthError) Error() string {
	return e.Message
}

var (
	ErrInvalidCredentials      = AuthError{"INVALID_CREDENTIALS", "invalid credentials"}
	ErrUserInactive           = AuthError{"USER_INACTIVE", "user account is inactive"}
	ErrTenantAccessDenied     = AuthError{"TENANT_ACCESS_DENIED", "access denied to tenant"}
	ErrYardAccessDenied       = AuthError{"YARD_ACCESS_DENIED", "access denied to yard"}
	ErrPermissionDenied       = AuthError{"PERMISSION_DENIED", "permission denied"}
	ErrEnterpriseAccessDenied = AuthError{"ENTERPRISE_ACCESS_DENIED", "enterprise access denied"}
	ErrNotCustomerContact     = AuthError{"NOT_CUSTOMER_CONTACT", "user is not a customer contact"}
	ErrInvalidUserRole        = AuthError{"INVALID_USER_ROLE", "invalid user role"}
	ErrUserAlreadyExists      = AuthError{"USER_ALREADY_EXISTS", "user already exists"}
	ErrCustomerNotFound       = AuthError{"CUSTOMER_NOT_FOUND", "customer not found"}
	ErrInvalidYardAccess      = AuthError{"INVALID_YARD_ACCESS", "invalid yard access configuration"}
)

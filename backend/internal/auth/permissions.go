// backend/internal/auth/permissions.go
package auth

import (
	"fmt"
	"strings"
)

// Permission represents a specific permission in the system
type Permission string

// ============================================================================
// PERMISSION CONSTANTS
// ============================================================================

// Inventory permissions
const (
	PermissionInventoryRead   Permission = "INVENTORY_READ"
	PermissionInventoryWrite  Permission = "INVENTORY_WRITE"
	PermissionInventoryExport Permission = "INVENTORY_EXPORT"
	PermissionInventoryDelete Permission = "INVENTORY_DELETE"
)

// Work order permissions
const (
	PermissionWorkOrderView     Permission = "WORK_ORDER_VIEW"
	PermissionWorkOrderCreate   Permission = "WORK_ORDER_CREATE"
	PermissionWorkOrderUpdate   Permission = "WORK_ORDER_UPDATE"
	PermissionWorkOrderApprove  Permission = "WORK_ORDER_APPROVE"
	PermissionWorkOrderDelete   Permission = "WORK_ORDER_DELETE"
	PermissionWorkOrderInvoice  Permission = "WORK_ORDER_INVOICE"
)

// Customer permissions
const (
	PermissionCustomerView   Permission = "CUSTOMER_VIEW"
	PermissionCustomerCreate Permission = "CUSTOMER_CREATE"
	PermissionCustomerUpdate Permission = "CUSTOMER_UPDATE"
	PermissionCustomerDelete Permission = "CUSTOMER_DELETE"
	PermissionCustomerManage Permission = "CUSTOMER_MANAGE"
)

// Transport/Logistics permissions
const (
	PermissionTransportView   Permission = "TRANSPORT_VIEW"
	PermissionTransportManage Permission = "TRANSPORT_MANAGE"
	PermissionLogisticsRead   Permission = "LOGISTICS_READ"
	PermissionLogisticsWrite  Permission = "LOGISTICS_WRITE"
)

// User management permissions
const (
	PermissionUserView       Permission = "USER_VIEW"
	PermissionUserCreate     Permission = "USER_CREATE"
	PermissionUserUpdate     Permission = "USER_UPDATE"
	PermissionUserDelete     Permission = "USER_DELETE"
	PermissionUserManagement Permission = "USER_MANAGEMENT"
	PermissionRoleManagement Permission = "ROLE_MANAGEMENT"
)

// Admin and enterprise permissions
const (
	PermissionCrossTenantView    Permission = "CROSS_TENANT_VIEW"
	PermissionCrossTenantManage  Permission = "CROSS_TENANT_MANAGE"
	PermissionTenantManagement   Permission = "TENANT_MANAGEMENT"
	PermissionSystemAdmin        Permission = "SYSTEM_ADMIN"
	PermissionEnterpriseReports  Permission = "ENTERPRISE_REPORTS"
	PermissionAuditView          Permission = "AUDIT_VIEW"
)

// Data export permissions
const (
	PermissionDataExport     Permission = "DATA_EXPORT"
	PermissionReportGenerate Permission = "REPORT_GENERATE"
	PermissionAnalytics      Permission = "ANALYTICS"
)

// ============================================================================
// ROLE TO PERMISSION MAPPINGS
// ============================================================================

// RolePermissions defines the default permissions for each role
var RolePermissions = map[UserRole][]Permission{
	RoleCustomerContact: {
		// Customer contacts can view their own data
		PermissionWorkOrderView,
		PermissionInventoryRead,
		PermissionTransportView,
		PermissionCustomerView, // Own customer only
	},
	
	RoleOperator: {
		// Yard operators handle day-to-day operations
		PermissionInventoryRead,
		PermissionInventoryWrite,
		PermissionWorkOrderView,
		PermissionWorkOrderCreate,
		PermissionWorkOrderUpdate,
		PermissionTransportView,
		PermissionTransportManage,
		PermissionLogisticsRead,
		PermissionLogisticsWrite,
		PermissionCustomerView,
	},
	
	RoleManager: {
		// Site managers have broader operational control
		PermissionInventoryRead,
		PermissionInventoryWrite,
		PermissionInventoryExport,
		PermissionWorkOrderView,
		PermissionWorkOrderCreate,
		PermissionWorkOrderUpdate,
		PermissionWorkOrderApprove,
		PermissionWorkOrderInvoice,
		PermissionTransportView,
		PermissionTransportManage,
		PermissionLogisticsRead,
		PermissionLogisticsWrite,
		PermissionCustomerView,
		PermissionCustomerUpdate,
		PermissionDataExport,
		PermissionReportGenerate,
		PermissionAnalytics,
		PermissionUserView,
	},
	
	RoleAdmin: {
		// Tenant administrators have full control within their tenant
		PermissionInventoryRead,
		PermissionInventoryWrite,
		PermissionInventoryExport,
		PermissionInventoryDelete,
		PermissionWorkOrderView,
		PermissionWorkOrderCreate,
		PermissionWorkOrderUpdate,
		PermissionWorkOrderApprove,
		PermissionWorkOrderDelete,
		PermissionWorkOrderInvoice,
		PermissionTransportView,
		PermissionTransportManage,
		PermissionLogisticsRead,
		PermissionLogisticsWrite,
		PermissionCustomerView,
		PermissionCustomerCreate,
		PermissionCustomerUpdate,
		PermissionCustomerDelete,
		PermissionCustomerManage,
		PermissionDataExport,
		PermissionReportGenerate,
		PermissionAnalytics,
		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserManagement,
		PermissionAuditView,
	},
	
	RoleEnterpriseAdmin: {
		// Enterprise admins have cross-tenant capabilities
		PermissionCrossTenantView,
		PermissionCrossTenantManage,
		PermissionEnterpriseReports,
		PermissionUserManagement,
		PermissionRoleManagement,
		PermissionAuditView,
		// Plus all admin permissions
		PermissionInventoryRead,
		PermissionInventoryWrite,
		PermissionInventoryExport,
		PermissionInventoryDelete,
		PermissionWorkOrderView,
		PermissionWorkOrderCreate,
		PermissionWorkOrderUpdate,
		PermissionWorkOrderApprove,
		PermissionWorkOrderDelete,
		PermissionWorkOrderInvoice,
		PermissionTransportView,
		PermissionTransportManage,
		PermissionLogisticsRead,
		PermissionLogisticsWrite,
		PermissionCustomerView,
		PermissionCustomerCreate,
		PermissionCustomerUpdate,
		PermissionCustomerDelete,
		PermissionCustomerManage,
		PermissionDataExport,
		PermissionReportGenerate,
		PermissionAnalytics,
		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,
	},
	
	RoleSystemAdmin: {
		// System admins have unrestricted access
		PermissionSystemAdmin, // This permission grants access to everything
	},
}

// ============================================================================
// PERMISSION CHECKING FUNCTIONS
// ============================================================================

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission Permission) bool {
	// System admins have all permissions
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	// Check role-based permissions
	rolePerms, exists := RolePermissions[u.Role]
	if !exists {
		return false
	}
	
	for _, perm := range rolePerms {
		if perm == permission {
			return true
		}
	}
	
	// Check custom permissions in tenant access
	for _, tenantAccess := range u.TenantAccess {
		for _, perm := range tenantAccess.Permissions {
			if perm == permission {
				return true
			}
		}
	}
	
	return false
}

// HasPermissionInTenant checks if user has permission in specific tenant
func (u *User) HasPermissionInTenant(permission Permission, tenantID string) bool {
	// System admins have all permissions everywhere
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	// Enterprise admins with cross-tenant permissions
	if u.IsEnterpriseUser && (permission == PermissionCrossTenantView || permission == PermissionCrossTenantManage) {
		return u.HasPermission(permission)
	}
	
	// Check tenant-specific access
	for _, tenantAccess := range u.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			// Check role permissions for this tenant
			rolePerms, exists := RolePermissions[tenantAccess.Role]
			if exists {
				for _, perm := range rolePerms {
					if perm == permission {
						return true
					}
				}
			}
			
			// Check custom permissions for this tenant
			for _, perm := range tenantAccess.Permissions {
				if perm == permission {
					return true
				}
			}
		}
	}
	
	return false
}

// CanAccessYard checks if user can access specific yard in tenant
func (u *User) CanAccessYard(tenantID, yardLocation string) bool {
	// System admins can access any yard
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	// Enterprise admins can access any yard if they have tenant access
	if u.IsEnterpriseUser {
		for _, tenantAccess := range u.TenantAccess {
			if tenantAccess.TenantID == tenantID {
				return true // Enterprise admins can access all yards in accessible tenants
			}
		}
	}
	
	// Check specific yard access
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

// HasYardPermission checks specific yard-level permission
func (u *User) HasYardPermission(tenantID, yardLocation string, permission YardPermission) bool {
	// System admins have all permissions
	if u.Role == RoleSystemAdmin {
		return true
	}
	
	// Find tenant access
	for _, tenantAccess := range u.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			// Enterprise users with tenant access have all yard permissions
			if u.IsEnterpriseUser {
				return true
			}
			
			// Check specific yard permissions
			for _, yardAccess := range tenantAccess.YardAccess {
				if yardAccess.YardLocation == yardLocation {
					return yardAccess.HasPermission(permission)
				}
			}
		}
	}
	
	return false
}

// ============================================================================
// YARD ACCESS PERMISSION CHECKING
// ============================================================================

// YardPermission represents yard-level permissions
type YardPermission string

const (
	YardPermissionViewWorkOrders   YardPermission = "VIEW_WORK_ORDERS"
	YardPermissionCreateWorkOrders YardPermission = "CREATE_WORK_ORDERS"
	YardPermissionApproveOrders    YardPermission = "APPROVE_ORDERS"
	YardPermissionViewInventory    YardPermission = "VIEW_INVENTORY"
	YardPermissionManageTransport  YardPermission = "MANAGE_TRANSPORT"
	YardPermissionExportData       YardPermission = "EXPORT_DATA"
)

// HasPermission checks if yard access includes specific permission
func (ya *YardAccess) HasPermission(permission YardPermission) bool {
	switch permission {
	case YardPermissionViewWorkOrders:
		return ya.CanViewWorkOrders
	case YardPermissionCreateWorkOrders:
		return ya.CanCreateWorkOrders
	case YardPermissionApproveOrders:
		return ya.CanApproveOrders
	case YardPermissionViewInventory:
		return ya.CanViewInventory
	case YardPermissionManageTransport:
		return ya.CanManageTransport
	case YardPermissionExportData:
		return ya.CanExportData
	default:
		return false
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// GetAllPermissions returns all available permissions in the system
func GetAllPermissions() []Permission {
	return []Permission{
		// Inventory permissions
		PermissionInventoryRead,
		PermissionInventoryWrite,
		PermissionInventoryExport,
		PermissionInventoryDelete,
		
		// Work order permissions
		PermissionWorkOrderView,
		PermissionWorkOrderCreate,
		PermissionWorkOrderUpdate,
		PermissionWorkOrderApprove,
		PermissionWorkOrderDelete,
		PermissionWorkOrderInvoice,
		
		// Customer permissions
		PermissionCustomerView,
		PermissionCustomerCreate,
		PermissionCustomerUpdate,
		PermissionCustomerDelete,
		PermissionCustomerManage,
		
		// Transport permissions
		PermissionTransportView,
		PermissionTransportManage,
		PermissionLogisticsRead,
		PermissionLogisticsWrite,
		
		// User management permissions
		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserManagement,
		PermissionRoleManagement,
		
		// Admin permissions
		PermissionCrossTenantView,
		PermissionCrossTenantManage,
		PermissionTenantManagement,
		PermissionSystemAdmin,
		PermissionEnterpriseReports,
		PermissionAuditView,
		
		// Data permissions
		PermissionDataExport,
		PermissionReportGenerate,
		PermissionAnalytics,
	}
}

// GetPermissionsForRole returns default permissions for a role
func GetPermissionsForRole(role UserRole) []Permission {
	permissions, exists := RolePermissions[role]
	if !exists {
		return []Permission{}
	}
	
	// Return a copy to prevent modifications
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result
}

// ValidatePermission checks if a permission string is valid
func ValidatePermission(permission string) bool {
	allPerms := GetAllPermissions()
	for _, perm := range allPerms {
		if string(perm) == permission {
			return true
		}
	}
	return false
}

// PermissionDescription returns human-readable description of permission
func PermissionDescription(permission Permission) string {
	descriptions := map[Permission]string{
		// Inventory
		PermissionInventoryRead:   "View inventory items and details",
		PermissionInventoryWrite:  "Create and update inventory items",
		PermissionInventoryExport: "Export inventory data to files",
		PermissionInventoryDelete: "Delete inventory items",
		
		// Work orders
		PermissionWorkOrderView:    "View work orders and details",
		PermissionWorkOrderCreate:  "Create new work orders",
		PermissionWorkOrderUpdate:  "Update existing work orders",
		PermissionWorkOrderApprove: "Approve and process work orders",
		PermissionWorkOrderDelete:  "Delete work orders",
		PermissionWorkOrderInvoice: "Generate invoices for work orders",
		
		// Customers
		PermissionCustomerView:   "View customer information",
		PermissionCustomerCreate: "Create new customer accounts",
		PermissionCustomerUpdate: "Update customer information",
		PermissionCustomerDelete: "Delete customer accounts",
		PermissionCustomerManage: "Full customer management access",
		
		// Transport
		PermissionTransportView:   "View transport and logistics information",
		PermissionTransportManage: "Manage transport operations",
		PermissionLogisticsRead:   "View logistics data",
		PermissionLogisticsWrite:  "Update logistics information",
		
		// Users
		PermissionUserView:       "View user accounts and profiles",
		PermissionUserCreate:     "Create new user accounts",
		PermissionUserUpdate:     "Update user information",
		PermissionUserDelete:     "Delete user accounts",
		PermissionUserManagement: "Full user management access",
		PermissionRoleManagement: "Manage user roles and permissions",
		
		// Admin
		PermissionCrossTenantView:    "View data across multiple tenants",
		PermissionCrossTenantManage:  "Manage operations across tenants",
		PermissionTenantManagement:   "Manage tenant configurations",
		PermissionSystemAdmin:        "Full system administration access",
		PermissionEnterpriseReports:  "Generate enterprise-wide reports",
		PermissionAuditView:          "View audit logs and security events",
		
		// Data
		PermissionDataExport:     "Export system data to files",
		PermissionReportGenerate: "Generate reports and analytics",
		PermissionAnalytics:      "Access analytics and business intelligence",
	}
	
	if desc, exists := descriptions[permission]; exists {
		return desc
	}
	
	return string(permission)
}

// String implements the Stringer interface for Permission
func (p Permission) String() string {
	return string(p)
}

// MarshalText implements encoding.TextMarshaler for Permission
func (p Permission) MarshalText() ([]byte, error) {
	return []byte(p), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Permission
func (p *Permission) UnmarshalText(data []byte) error {
	*p = Permission(data)
	return nil
}

// IsValid checks if the permission is in the list of valid permissions
func (p Permission) IsValid() bool {
	return ValidatePermission(string(p))
}

// Category returns the category of the permission (e.g., "INVENTORY", "WORK_ORDER")
func (p Permission) Category() string {
	parts := strings.Split(string(p), "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "UNKNOWN"
}

// ============================================================================
// LEGACY CFM ACCESS LEVEL MAPPING
// ============================================================================

// CFMAccessLevel represents the legacy CFM access levels (1-5+)
type CFMAccessLevel int

const (
	CFMAccessLevelOperator        CFMAccessLevel = 1
	CFMAccessLevelSeniorOperator  CFMAccessLevel = 2
	CFMAccessLevelManager         CFMAccessLevel = 3
	CFMAccessLevelAdmin           CFMAccessLevel = 4
	CFMAccessLevelEnterpriseAdmin CFMAccessLevel = 5
	CFMAccessLevelSystemAdmin     CFMAccessLevel = 6
)

// MapCFMAccessToRole maps legacy CFM access levels to new user roles
func MapCFMAccessToRole(accessLevel CFMAccessLevel) UserRole {
	switch accessLevel {
	case CFMAccessLevelOperator, CFMAccessLevelSeniorOperator:
		return RoleOperator
	case CFMAccessLevelManager:
		return RoleManager
	case CFMAccessLevelAdmin:
		return RoleAdmin
	case CFMAccessLevelEnterpriseAdmin:
		return RoleEnterpriseAdmin
	case CFMAccessLevelSystemAdmin:
		return RoleSystemAdmin
	default:
		return RoleOperator // Safe default
	}
}

// MapRoleToCFMAccess maps new user roles back to CFM access levels
func MapRoleToCFMAccess(role UserRole) CFMAccessLevel {
	switch role {
	case RoleCustomerContact:
		return CFMAccessLevelOperator // Customer contacts map to basic access
	case RoleOperator:
		return CFMAccessLevelOperator
	case RoleManager:
		return CFMAccessLevelManager
	case RoleAdmin:
		return CFMAccessLevelAdmin
	case RoleEnterpriseAdmin:
		return CFMAccessLevelEnterpriseAdmin
	case RoleSystemAdmin:
		return CFMAccessLevelSystemAdmin
	default:
		return CFMAccessLevelOperator
	}
}

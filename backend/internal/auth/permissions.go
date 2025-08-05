// backend/internal/auth/permissions.go
// Permission calculation logic
package auth

import "context"

// Permission calculation utilities and business logic
// Note: Permission type and constants are defined in models.go to avoid redeclaration

// PermissionCalculator handles complex permission logic
type PermissionCalculator struct{}

// GetRolePermissions returns base permissions for a role
func (pc *PermissionCalculator) GetRolePermissions(role UserRole) []Permission {
	switch role {
	case RoleSystemAdmin:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionApproveWorkOrder,
			PermissionManageTransport,
			PermissionExportData,
			PermissionUserManagement,
			PermissionCrossTenantView,
		}
	case RoleEnterpriseAdmin:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionApproveWorkOrder,
			PermissionManageTransport,
			PermissionExportData,
			PermissionUserManagement,
			PermissionCrossTenantView,
		}
	case RoleAdmin:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionApproveWorkOrder,
			PermissionManageTransport,
			PermissionExportData,
			PermissionUserManagement,
		}
	case RoleManager:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionApproveWorkOrder,
			PermissionManageTransport,
			PermissionExportData,
		}
	case RoleOperator:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionManageTransport,
		}
	case RoleCustomerContact:
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
		}
	default:
		return []Permission{}
	}
}

// GetEnterprisePermissions returns permissions for enterprise users
func (pc *PermissionCalculator) GetEnterprisePermissions(role UserRole) []Permission {
	basePermissions := pc.GetRolePermissions(role)
	
	// Enterprise users always get cross-tenant permissions
	if role == RoleEnterpriseAdmin || role == RoleSystemAdmin {
		return append(basePermissions, PermissionCrossTenantView, PermissionUserManagement)
	}
	
	return basePermissions
}

// GetCrossTenantPermissions returns permissions for cross-tenant operations
func (pc *PermissionCalculator) GetCrossTenantPermissions(role UserRole) []Permission {
	if role == RoleEnterpriseAdmin || role == RoleSystemAdmin {
		return []Permission{
			PermissionCrossTenantView,
			PermissionUserManagement,
		}
	}
	return []Permission{}
}

// ValidatePermissionForRole checks if a role can have a specific permission
func (pc *PermissionCalculator) ValidatePermissionForRole(role UserRole, permission Permission) bool {
	rolePermissions := pc.GetRolePermissions(role)
	for _, p := range rolePermissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CalculateTenantPermissions determines permissions for a user in a specific tenant
func (pc *PermissionCalculator) CalculateTenantPermissions(user *User, tenantID string) []Permission {
	// Enterprise users get full permissions
	if user.IsEnterpriseUser {
		return pc.GetEnterprisePermissions(user.Role)
	}
	
	// Find tenant-specific permissions
	for _, access := range user.TenantAccess {
		if access.TenantID == tenantID {
			return access.Permissions
		}
	}
	
	// Fallback to role-based permissions
	return pc.GetRolePermissions(user.Role)
}

// HasPermissionInTenant checks if user has a specific permission in a tenant
func (pc *PermissionCalculator) HasPermissionInTenant(user *User, tenantID string, permission Permission) bool {
	permissions := pc.CalculateTenantPermissions(user, tenantID)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// PermissionService provides permission checking capabilities
type PermissionService interface {
	CheckPermission(ctx context.Context, userID int, tenantID string, permission Permission) error
	CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error
	GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error)
}

// permissionService implements PermissionService
type permissionService struct {
	repo       Repository
	calculator *PermissionCalculator
}

// NewPermissionService creates a new permission service
func NewPermissionService(repo Repository) PermissionService {
	return &permissionService{
		repo:       repo,
		calculator: &PermissionCalculator{},
	}
}

func (ps *permissionService) CheckPermission(ctx context.Context, userID int, tenantID string, permission Permission) error {
	user, err := ps.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if !ps.calculator.HasPermissionInTenant(user, tenantID, permission) {
		return ErrPermissionDenied
	}
	
	return nil
}

func (ps *permissionService) CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error {
	user, err := ps.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if !user.HasAccessToYard(tenantID, yardLocation) {
		return ErrYardAccessDenied
	}
	
	return nil
}

func (ps *permissionService) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	user, err := ps.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return ps.calculator.CalculateTenantPermissions(user, tenantID), nil
}

// ============================================================================
// LEGACY CFM SUPPORT FUNCTIONS
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

// GetPermissionsForRole returns default permissions for a role (legacy bridge compatibility)
func GetPermissionsForRole(role UserRole) []Permission {
	calc := &PermissionCalculator{}
	return calc.GetRolePermissions(role)
}

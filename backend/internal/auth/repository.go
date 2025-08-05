// backend/internal/auth/repository.go
package auth

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// Complete Repository interface with all methods
type Repository interface {
	// User management
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id int) error
	
	// Session management
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	InvalidateSession(ctx context.Context, sessionID string) error
	InvalidateUserSessions(ctx context.Context, userID int) error
	CleanupExpiredSessions(ctx context.Context) error
	
	// Multi-tenant user queries
	GetEnterpriseUsers(ctx context.Context) ([]User, error)
	GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error)
	GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error)
	GetUsersByRole(ctx context.Context, role UserRole) ([]User, error)
	GetCustomerContacts(ctx context.Context, customerID int) ([]User, error)
	
	// Permission and access queries
	GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error)
	GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error)
	ValidateCustomerExists(ctx context.Context, customerID int) error
	
	// Search and analytics
	SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
	
	// Tenant access management
	UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error
	UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error
	
	// Legacy tenant support (for backward compatibility)
	GetUserTenants(ctx context.Context, userID int) ([]LegacyTenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*LegacyTenant, error)
	ListTenants(ctx context.Context) ([]LegacyTenant, error)
	
	// Transaction support
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
}

// Legacy tenant structure for backward compatibility
type LegacyTenant struct {
	ID           int    `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Slug         string `json:"slug" db:"slug"`
	DatabaseName string `json:"database_name" db:"database_name"`
	Active       bool   `json:"active" db:"active"`
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// ============================================================================
// USER MANAGEMENT
// ============================================================================

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE username = $1 AND is_active = true`
	
	return r.scanUser(ctx, query, username)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE email = $1`
	
	return r.scanUser(ctx, query, email)
}

func (r *repository) GetUserByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE id = $1`
	
	return r.scanUser(ctx, query, id)
}

func (r *repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO auth.users (
			username, email, full_name, password_hash, role, access_level,
			is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
			contact_type, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`
	
	tenantAccessJson, err := user.TenantAccess.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tenant access: %w", err)
	}
	
	err = r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.FullName,
		user.PasswordHash,
		user.Role,
		user.AccessLevel,
		user.IsEnterpriseUser,
		tenantAccessJson,
		user.PrimaryTenantID,
		user.CustomerID,
		user.ContactType,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE auth.users 
		SET email = $2, full_name = $3, password_hash = $4, role = $5, 
			access_level = $6, is_enterprise_user = $7, tenant_access = $8,
			primary_tenant_id = $9, customer_id = $10, contact_type = $11,
			is_active = $12, updated_at = $13
		WHERE id = $1`
	
	tenantAccessJson, err := user.TenantAccess.Value()
	if err != nil {
		return fmt.Errorf("failed to serialize tenant access: %w", err)
	}
	
	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FullName,
		user.PasswordHash,
		user.Role,
		user.AccessLevel,
		user.IsEnterpriseUser,
		tenantAccessJson,
		user.PrimaryTenantID,
		user.CustomerID,
		user.ContactType,
		user.IsActive,
		user.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

func (r *repository) DeleteUser(ctx context.Context, id int) error {
	// Soft delete by setting is_active = false
	query := `UPDATE auth.users SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	// Also invalidate all user sessions
	return r.InvalidateUserSessions(ctx, id)
}

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

func (r *repository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO auth.sessions (
			id, user_id, tenant_id, token, refresh_token, 
			tenant_context, is_active, expires_at, refresh_expires_at,
			created_at, last_used_at, user_agent, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	tenantContextJson, err := r.serializeTenantContext(session.TenantContext)
	if err != nil {
		return fmt.Errorf("failed to serialize tenant context: %w", err)
	}
	
	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TenantID,
		session.Token,
		session.RefreshToken,
		tenantContextJson,
		session.IsActive,
		session.ExpiresAt,
		session.RefreshExpiresAt,
		session.CreatedAt,
		session.LastUsedAt,
		session.UserAgent,
		session.IPAddress,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	return nil
}

func (r *repository) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, tenant_id, token, refresh_token,
			   tenant_context, is_active, expires_at, refresh_expires_at,
			   created_at, last_used_at, user_agent, ip_address
		FROM auth.sessions 
		WHERE id = $1`
	
	return r.scanSession(ctx, query, sessionID)
}

func (r *repository) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT id, user_id, tenant_id, token, refresh_token,
			   tenant_context, is_active, expires_at, refresh_expires_at,
			   created_at, last_used_at, user_agent, ip_address
		FROM auth.sessions 
		WHERE token = $1 AND is_active = true AND expires_at > NOW()`
	
	return r.scanSession(ctx, query, token)
}

func (r *repository) UpdateSession(ctx context.Context, session *Session) error {
	query := `
		UPDATE auth.sessions 
		SET last_used_at = $2, is_active = $3, tenant_context = $4
		WHERE id = $1`
	
	tenantContextJson, err := r.serializeTenantContext(session.TenantContext)
	if err != nil {
		return fmt.Errorf("failed to serialize tenant context: %w", err)
	}
	
	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.LastUsedAt,
		session.IsActive,
		tenantContextJson,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	return nil
}

func (r *repository) InvalidateSession(ctx context.Context, sessionID string) error {
	query := `UPDATE auth.sessions SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

func (r *repository) InvalidateUserSessions(ctx context.Context, userID int) error {
	query := `UPDATE auth.sessions SET is_active = false WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *repository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM auth.sessions WHERE expires_at < NOW() - INTERVAL '7 days'`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// ============================================================================
// MULTI-TENANT USER QUERIES
// ============================================================================

func (r *repository) GetEnterpriseUsers(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE is_enterprise_user = true AND is_active = true
		ORDER BY full_name`
	
	return r.scanUsers(ctx, query)
}

func (r *repository) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE (primary_tenant_id = $1 OR tenant_access::text LIKE '%' || $1 || '%')
		  AND is_active = true
		ORDER BY role, full_name`
	
	return r.scanUsers(ctx, query, tenantID)
}

func (r *repository) GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE customer_id = $1 AND is_active = true
		ORDER BY contact_type, full_name`
	
	return r.scanUsers(ctx, query, customerID)
}

func (r *repository) GetUsersByRole(ctx context.Context, role UserRole) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE role = $1 AND is_active = true
		ORDER BY full_name`
	
	return r.scanUsers(ctx, query, role)
}

func (r *repository) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE customer_id = $1 AND role = 'CUSTOMER_CONTACT' AND is_active = true
		ORDER BY contact_type, full_name`
	
	return r.scanUsers(ctx, query, customerID)
}

// ============================================================================
// PERMISSION AND ACCESS QUERIES
// ============================================================================

func (r *repository) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Find tenant access for the user
	for _, access := range user.TenantAccess {
		if access.TenantID == tenantID {
			return access.Permissions, nil
		}
	}
	
	// Enterprise users get default permissions
	if user.IsEnterpriseUser {
		return []Permission{
			PermissionViewInventory,
			PermissionCreateWorkOrder,
			PermissionManageTransport,
			PermissionExportData,
		}, nil
	}
	
	return []Permission{}, nil
}

func (r *repository) GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE (
			primary_tenant_id = $1 OR 
			tenant_access::text LIKE '%' || $1 || '%'
		) AND tenant_access::text LIKE '%' || $2 || '%'
		AND is_active = true
		ORDER BY role, full_name`
	
	return r.scanUsers(ctx, query, tenantID, yardLocation)
}

func (r *repository) ValidateCustomerExists(ctx context.Context, customerID int) error {
	// This should query the customer table in the appropriate schema
	// For now, we'll assume customers are in the same schema
	query := `SELECT 1 FROM customers WHERE customer_id = $1 AND deleted = false LIMIT 1`
	
	var exists int
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("customer with ID %d does not exist", customerID)
		}
		return fmt.Errorf("failed to validate customer exists: %w", err)
	}
	
	return nil
}

// ============================================================================
// SEARCH AND ANALYTICS
// ============================================================================

func (r *repository) SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error) {
	// Build dynamic query based on filters
	baseQuery := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE is_active = true`
	
	countQuery := `SELECT COUNT(*) FROM auth.users WHERE is_active = true`
	
	args := []interface{}{}
	argIndex := 1
	conditions := []string{}
	
	// Add search conditions
	if filters.Query != "" {
		condition := fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d OR username ILIKE $%d)", argIndex, argIndex, argIndex)
		conditions = append(conditions, condition)
		args = append(args, "%"+filters.Query+"%")
		argIndex++
	}
	
	if filters.TenantID != "" {
		condition := fmt.Sprintf("primary_tenant_id = $%d", argIndex)
		conditions = append(conditions, condition)
		args = append(args, filters.TenantID)
		argIndex++
	}
	
	if filters.Role != "" {
		condition := fmt.Sprintf("role = $%d", argIndex)
		conditions = append(conditions, condition)
		args = append(args, filters.Role)
		argIndex++
	}
	
	if filters.CustomerID != nil {
		condition := fmt.Sprintf("customer_id = $%d", argIndex)
		conditions = append(conditions, condition)
		args = append(args, *filters.CustomerID)
		argIndex++
	}
	
	if filters.OnlyCustomerContacts {
		conditions = append(conditions, "role = 'CUSTOMER_CONTACT'")
	}
	
	if len(filters.AccessibleTenants) > 0 {
		placeholders := make([]string, len(filters.AccessibleTenants))
		for i, tenant := range filters.AccessibleTenants {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, tenant)
			argIndex++
		}
		condition := fmt.Sprintf("primary_tenant_id = ANY(ARRAY[%s])", strings.Join(placeholders, ","))
		conditions = append(conditions, condition)
	}
	
	// Apply conditions to both queries
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}
	
	// Get total count first (before adding pagination)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}
	
	// Add ordering and pagination to main query
	baseQuery += " ORDER BY full_name"
	
	if filters.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
		
		if filters.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}
	
	// Get users
	users, err := r.scanUsers(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

func (r *repository) GetUserStats(ctx context.Context) (*UserStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE is_active = true) as active_users,
			COUNT(*) FILTER (WHERE role = 'CUSTOMER_CONTACT') as customer_contacts,
			COUNT(*) FILTER (WHERE is_enterprise_user = true) as enterprise_users
		FROM auth.users`
	
	stats := &UserStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.CustomerContacts,
		&stats.EnterpriseUsers,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	
	return stats, nil
}

// ============================================================================
// TENANT ACCESS MANAGEMENT
// ============================================================================

func (r *repository) UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error {
	tenantAccessJson, err := tenantAccess.Value()
	if err != nil {
		return fmt.Errorf("failed to serialize tenant access: %w", err)
	}
	
	query := `UPDATE auth.users SET tenant_access = $2, updated_at = NOW() WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, userID, tenantAccessJson)
	if err != nil {
		return fmt.Errorf("failed to update user tenant access: %w", err)
	}
	
	return nil
}

func (r *repository) UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// Find and update the specific tenant's yard access
	for i, access := range user.TenantAccess {
		if access.TenantID == tenantID {
			user.TenantAccess[i].YardAccess = yardAccess
			break
		}
	}
	
	return r.UpdateUserTenantAccess(ctx, userID, user.TenantAccess)
}

// ============================================================================
// LEGACY TENANT SUPPORT
// ============================================================================

func (r *repository) GetUserTenants(ctx context.Context, userID int) ([]LegacyTenant, error) {
	// Legacy support - get tenants from user's tenant access
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	tenantIDs := make([]string, 0, len(user.TenantAccess))
	for _, access := range user.TenantAccess {
		tenantIDs = append(tenantIDs, access.TenantID)
	}
	
	if len(tenantIDs) == 0 {
		return []LegacyTenant{}, nil
	}
	
	placeholders := make([]string, len(tenantIDs))
	args := make([]interface{}, len(tenantIDs))
	for i, tenantID := range tenantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = tenantID
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, slug, 'tenant' as database_name, active
		FROM auth.tenants 
		WHERE slug IN (%s) AND active = true
		ORDER BY name`, strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}
	defer rows.Close()
	
	var tenants []LegacyTenant
	for rows.Next() {
		var tenant LegacyTenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.DatabaseName, &tenant.Active)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

func (r *repository) GetTenantBySlug(ctx context.Context, slug string) (*LegacyTenant, error) {
	query := `
		SELECT id, name, slug, 'tenant' as database_name, active
		FROM auth.tenants 
		WHERE slug = $1 AND active = true`
	
	tenant := &LegacyTenant{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.DatabaseName, &tenant.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant with slug '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}
	
	return tenant, nil
}

func (r *repository) ListTenants(ctx context.Context) ([]LegacyTenant, error) {
	query := `
		SELECT id, name, slug, 'tenant' as database_name, active
		FROM auth.tenants 
		WHERE active = true
		ORDER BY name`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()
	
	var tenants []LegacyTenant
	for rows.Next() {
		var tenant LegacyTenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.DatabaseName, &tenant.Active)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

// ============================================================================
// TRANSACTION SUPPORT
// ============================================================================

func (r *repository) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

func (r *repository) scanUser(ctx context.Context, query string, args ...interface{}) (*User, error) {
	user := &User{}
	var tenantAccessJson []byte
	
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.Role,
		&user.AccessLevel,
		&user.IsEnterpriseUser,
		&tenantAccessJson,
		&user.PrimaryTenantID,
		&user.CustomerID,
		&user.ContactType,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}
	
	// Deserialize tenant access
	if len(tenantAccessJson) > 0 {
		if err := user.TenantAccess.Scan(tenantAccessJson); err != nil {
			return nil, fmt.Errorf("failed to deserialize tenant access: %w", err)
		}
	}
	
	return user, nil
}

func (r *repository) scanUsers(ctx context.Context, query string, args ...interface{}) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		user := User{}
		var tenantAccessJson []byte
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.FullName,
			&user.PasswordHash,
			&user.Role,
			&user.AccessLevel,
			&user.IsEnterpriseUser,
			&tenantAccessJson,
			&user.PrimaryTenantID,
			&user.CustomerID,
			&user.ContactType,
			&user.IsActive,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		
		// Deserialize tenant access
		if len(tenantAccessJson) > 0 {
			if err := user.TenantAccess.Scan(tenantAccessJson); err != nil {
				return nil, fmt.Errorf("failed to deserialize tenant access: %w", err)
			}
		}
		
		users = append(users, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}
	
	return users, nil
}

func (r *repository) scanSession(ctx context.Context, query string, args ...interface{}) (*Session, error) {
	session := &Session{}
	var tenantContextJson []byte
	
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&session.ID,
		&session.UserID,
		&session.TenantID,
		&session.Token,
		&session.RefreshToken,
		&tenantContextJson,
		&session.IsActive,
		&session.ExpiresAt,
		&session.RefreshExpiresAt,
		&session.CreatedAt,
		&session.LastUsedAt,
		&session.UserAgent,
		&session.IPAddress,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to scan session: %w", err)
	}
	
	// Deserialize tenant context
	if len(tenantContextJson) > 0 {
		tenantContext := &TenantAccess{}
		if err := tenantContext.Scan(tenantContextJson); err != nil {
			return nil, fmt.Errorf("failed to deserialize tenant context: %w", err)
		}
		session.TenantContext = tenantContext
	}
	
	return session, nil
}

func (r *repository) serializeTenantContext(tenantContext *TenantAccess) ([]byte, error) {
	if tenantContext == nil {
		return nil, nil
	}
	
	// Use the TenantAccess Value() method for consistent serialization
	value, err := tenantContext.Value()
	if err != nil {
		return nil, err
	}
	
	// Convert driver.Value to []byte
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}
	
	return nil, fmt.Errorf("unexpected value type from TenantAccess.Value()")
}

// TenantAccess needs to implement driver.Valuer for individual instances
func (ta *TenantAccess) Value() (driver.Value, error) {
	if ta == nil {
		return nil, nil
	}
	return json.Marshal(ta)
}

// TenantAccess needs to implement sql.Scanner for individual instances  
func (ta *TenantAccess) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TenantAccess", value)
	}
	
	return json.Unmarshal(bytes, ta)
}

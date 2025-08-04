// backend/internal/auth/repository.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository interface {
	// User management
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error
	UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error
	DeleteUser(ctx context.Context, id int) error
	
	// Session management
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
	GetSessionByID(ctx context.Context, sessionID string) (*Session, error)
	UpdateSessionActivity(ctx context.Context, sessionID string) error
	InvalidateSession(ctx context.Context, token string) error
	InvalidateUserSessions(ctx context.Context, userID int) error
	CleanupExpiredSessions(ctx context.Context) error
	
	// Enterprise user queries
	GetEnterpriseUsers(ctx context.Context) ([]User, error)
	GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error)
	GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error)
	GetUsersByRole(ctx context.Context, role UserRole) ([]User, error)
	
	// Tenant management (legacy compatibility)
	GetUserTenants(ctx context.Context, userID int) ([]LegacyTenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*LegacyTenant, error)
	ListTenants(ctx context.Context) ([]LegacyTenant, error)
	
	// Permission queries
	GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error)
	GetRolePermissions(ctx context.Context, role UserRole) ([]Permission, error)
	
	// Yard access queries
	GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error)
	
	// Customer contact queries
	GetCustomerContacts(ctx context.Context, customerID int) ([]User, error)
	ValidateCustomerExists(ctx context.Context, customerID int) error
	
	// Search and analytics
	GetUserStats(ctx context.Context) (*UserStats, error)
	SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error)
	
	// Transaction support
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
}

// repository implements the Repository interface
type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sqlx.DB) Repository {
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
		FROM enterprise.users 
		WHERE username = $1 AND is_active = true`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	
	return &user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE email = $1 AND is_active = true`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

func (r *repository) GetUserByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE id = $1`
	
	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO enterprise.users (
			username, email, full_name, password_hash, role, access_level,
			is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
			contact_type, is_active, created_at, updated_at
		) VALUES (
			:username, :email, :full_name, :password_hash, :role, :access_level,
			:is_enterprise_user, :tenant_access, :primary_tenant_id, :customer_id,
			:contact_type, :is_active, :created_at, :updated_at
		) RETURNING id`
	
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true
	
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare create user query: %w", err)
	}
	defer stmt.Close()
	
	err = stmt.GetContext(ctx, &user.ID, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "username") {
					return ErrUserAlreadyExists
				}
				if strings.Contains(pqErr.Detail, "email") {
					return ErrUserAlreadyExists
				}
			case "23503": // foreign_key_violation
				if strings.Contains(pqErr.Detail, "customer_id") {
					return ErrCustomerNotFound
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE enterprise.users SET
			username = :username, email = :email, full_name = :full_name,
			role = :role, access_level = :access_level, is_enterprise_user = :is_enterprise_user,
			tenant_access = :tenant_access, primary_tenant_id = :primary_tenant_id,
			customer_id = :customer_id, contact_type = :contact_type,
			is_active = :is_active, last_login_at = :last_login_at, updated_at = :updated_at
		WHERE id = :id`
	
	user.UpdatedAt = time.Now()
	
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

func (r *repository) UpdateUserTenantAccess(ctx context.Context, userID int, tenantAccess TenantAccessList) error {
	query := `
		UPDATE enterprise.users 
		SET tenant_access = $1, updated_at = NOW()
		WHERE id = $2`
	
	_, err := r.db.ExecContext(ctx, query, tenantAccess, userID)
	if err != nil {
		return fmt.Errorf("failed to update user tenant access: %w", err)
	}
	
	return nil
}

func (r *repository) UpdateUserYardAccess(ctx context.Context, userID int, tenantID string, yardAccess []YardAccess) error {
	// Get current user to update specific tenant's yard access
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// Update the specific tenant's yard access
	for i, tenantAccess := range user.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			user.TenantAccess[i].YardAccess = yardAccess
			break
		}
	}
	
	return r.UpdateUserTenantAccess(ctx, userID, user.TenantAccess)
}

func (r *repository) DeleteUser(ctx context.Context, id int) error {
	// Soft delete by setting is_active = false
	query := `UPDATE enterprise.users SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

func (r *repository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO enterprise.sessions (
			id, user_id, tenant_id, token, expires_at, is_active,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TenantID, session.Token,
		session.ExpiresAt, session.IsActive, session.IPAddress,
		session.UserAgent, session.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	return nil
}

func (r *repository) GetSession(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT id, user_id, tenant_id, token, expires_at, is_active,
		       ip_address, user_agent, created_at
		FROM enterprise.sessions 
		WHERE token = $1 AND is_active = true AND expires_at > NOW()`
	
	var session Session
	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &session, nil
}

func (r *repository) GetSessionByID(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, tenant_id, token, expires_at, is_active,
		       ip_address, user_agent, created_at
		FROM enterprise.sessions 
		WHERE id = $1 AND is_active = true AND expires_at > NOW()`
	
	var session Session
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}
	
	return &session, nil
}

func (r *repository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	query := `UPDATE enterprise.sessions SET created_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	
	return nil
}

func (r *repository) InvalidateSession(ctx context.Context, token string) error {
	query := `UPDATE enterprise.sessions SET is_active = false WHERE token = $1`
	
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}
	
	return nil
}

func (r *repository) InvalidateUserSessions(ctx context.Context, userID int) error {
	query := `UPDATE enterprise.sessions SET is_active = false WHERE user_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}
	
	return nil
}

func (r *repository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM enterprise.sessions WHERE expires_at < NOW() - INTERVAL '7 days'`
	
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	
	return nil
}

// ============================================================================
// ENTERPRISE USER QUERIES
// ============================================================================

func (r *repository) GetEnterpriseUsers(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE is_enterprise_user = true AND is_active = true
		ORDER BY full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise users: %w", err)
	}
	
	return users, nil
}

func (r *repository) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE (primary_tenant_id = $1 OR tenant_access::text LIKE '%' || $1 || '%')
		  AND is_active = true
		ORDER BY role, full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by tenant: %w", err)
	}
	
	return users, nil
}

func (r *repository) GetUsersByCustomer(ctx context.Context, customerID int) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE customer_id = $1 AND is_active = true
		ORDER BY contact_type, full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by customer: %w", err)
	}
	
	return users, nil
}

func (r *repository) GetUsersByRole(ctx context.Context, role UserRole) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE role = $1 AND is_active = true
		ORDER BY full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	
	return users, nil
}

// ============================================================================
// LEGACY TENANT MANAGEMENT (Compatibility with existing models)
// ============================================================================

// LegacyTenant for compatibility with existing models.Tenant
type LegacyTenant struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	DatabaseType string    `json:"database_type" db:"database_type"`
	DatabaseName *string   `json:"database_name" db:"database_name"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (r *repository) GetUserTenants(ctx context.Context, userID int) ([]LegacyTenant, error) {
	// For compatibility with legacy code that expects this method
	// This method bridges to the new tenant access system
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	var tenants []LegacyTenant
	for _, tenantAccess := range user.TenantAccess {
		// Convert to legacy format
		tenant := LegacyTenant{
			ID:           tenantAccess.TenantID,
			Name:         tenantAccess.TenantID, // Use tenant ID as name for now
			Slug:         tenantAccess.TenantID,
			DatabaseType: "tenant",
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

func (r *repository) GetTenantBySlug(ctx context.Context, slug string) (*LegacyTenant, error) {
	// For compatibility - return a basic tenant structure
	// In production, this would query an actual tenants table
	return &LegacyTenant{
		ID:           slug,
		Name:         slug,
		Slug:         slug,
		DatabaseType: "tenant",
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (r *repository) ListTenants(ctx context.Context) ([]LegacyTenant, error) {
	// For compatibility - return basic tenant list
	// In production, this would query an actual tenants table
	tenants := []LegacyTenant{
		{
			ID:           "houston",
			Name:         "Houston Division",
			Slug:         "houston",
			DatabaseType: "tenant",
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "longbeach",
			Name:         "Long Beach Division",
			Slug:         "longbeach",
			DatabaseType: "tenant",
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	
	return tenants, nil
}

// ============================================================================
// PERMISSION QUERIES
// ============================================================================

func (r *repository) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Get permissions for the specific tenant
	for _, tenantAccess := range user.TenantAccess {
		if tenantAccess.TenantID == tenantID {
			return tenantAccess.Permissions, nil
		}
	}
	
	// If no specific tenant access, return role-based permissions
	return r.GetRolePermissions(ctx, user.Role)
}

func (r *repository) GetRolePermissions(ctx context.Context, role UserRole) ([]Permission, error) {
	query := `
		SELECT permission 
		FROM enterprise.role_permissions 
		WHERE role = $1
		ORDER BY permission`
	
	var permissions []Permission
	err := r.db.SelectContext(ctx, &permissions, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	
	return permissions, nil
}

// ============================================================================
// YARD ACCESS QUERIES
// ============================================================================

func (r *repository) GetUsersWithYardAccess(ctx context.Context, tenantID, yardLocation string) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE tenant_access::text LIKE '%' || $1 || '%'
		  AND tenant_access::text LIKE '%' || $2 || '%'
		  AND is_active = true
		ORDER BY role, full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, tenantID, yardLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with yard access: %w", err)
	}
	
	// Filter users who actually have access to the specific yard
	var filteredUsers []User
	for _, user := range users {
		if user.HasAccessToYard(tenantID, yardLocation) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	
	return filteredUsers, nil
}

// ============================================================================
// CUSTOMER CONTACT QUERIES
// ============================================================================

func (r *repository) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE customer_id = $1 AND role = 'CUSTOMER_CONTACT' AND is_active = true
		ORDER BY 
			CASE contact_type 
				WHEN 'PRIMARY' THEN 1 
				WHEN 'APPROVER' THEN 2 
				WHEN 'BILLING' THEN 3 
				ELSE 4 
			END,
			full_name`
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer contacts: %w", err)
	}
	
	return users, nil
}

func (r *repository) ValidateCustomerExists(ctx context.Context, customerID int) error {
	query := `SELECT 1 FROM store.customers WHERE id = $1 LIMIT 1`
	
	var exists int
	err := r.db.GetContext(ctx, &exists, query, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("failed to validate customer exists: %w", err)
	}
	
	return nil
}

// ============================================================================
// SEARCH AND ANALYTICS
// ============================================================================

func (r *repository) GetUserStats(ctx context.Context) (*UserStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active THEN 1 END) as active_users,
			COUNT(CASE WHEN role = 'CUSTOMER_CONTACT' THEN 1 END) as customer_contacts,
			COUNT(CASE WHEN is_enterprise_user THEN 1 END) as enterprise_users,
			COUNT(CASE WHEN last_login_at > NOW() - INTERVAL '30 days' THEN 1 END) as recent_logins
		FROM enterprise.users`
	
	var stats UserStats
	err := r.db.GetContext(ctx, &stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	
	return &stats, nil
}

func (r *repository) SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error) {
	query := `
		SELECT id, username, email, full_name, password_hash, role, access_level,
		       is_enterprise_user, tenant_access, primary_tenant_id, customer_id,
		       contact_type, is_active, last_login_at, created_at, updated_at
		FROM enterprise.users 
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.Query != "" {
		query += fmt.Sprintf(` AND (
			full_name ILIKE $%d OR 
			username ILIKE $%d OR 
			email ILIKE $%d
		)`, argIndex, argIndex, argIndex)
		searchTerm := "%" + filter.Query + "%"
		args = append(args, searchTerm)
		argIndex++
	}
	
	if filter.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, *filter.Role)
		argIndex++
	}
	
	if filter.TenantID != nil {
		query += fmt.Sprintf(" AND (primary_tenant_id = $%d OR tenant_access::text LIKE '%%' || $%d || '%%')", argIndex, argIndex)
		args = append(args, *filter.TenantID)
		argIndex++
	}
	
	if filter.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	
	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}
	
	if filter.IsEnterpriseUser != nil {
		query += fmt.Sprintf(" AND is_enterprise_user = $%d", argIndex)
		args = append(args, *filter.IsEnterpriseUser)
		argIndex++
	}
	
	// Add ordering
	orderBy := "full_name"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "username", "email", "role", "created_at", "last_login_at":
			orderBy = filter.SortBy
		}
	}
	
	sortOrder := "ASC"
	if filter.SortOrder == "desc" {
		sortOrder = "DESC"
	}
	
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, sortOrder)
	
	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
		
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
		}
	}
	
	var users []User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	
	return users, nil
}

// ============================================================================
// TRANSACTION SUPPORT
// ============================================================================

func (r *repository) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// SUPPORTING TYPES
// ============================================================================

type UserStats struct {
	TotalUsers       int `json:"total_users" db:"total_users"`
	ActiveUsers      int `json:"active_users" db:"active_users"`
	CustomerContacts int `json:"customer_contacts" db:"customer_contacts"`
	EnterpriseUsers  int `json:"enterprise_users" db:"enterprise_users"`
	RecentLogins     int `json:"recent_logins" db:"recent_logins"`
}

type UserSearchFilter struct {
	Query            string    `json:"query"`
	Role             *UserRole `json:"role"`
	TenantID         *string   `json:"tenant_id"`
	CustomerID       *int      `json:"customer_id"`
	IsActive         *bool     `json:"is_active"`
	IsEnterpriseUser *bool     `json:"is_enterprise_user"`
	SortBy           string    `json:"sort_by"`
	SortOrder        string    `json:"sort_order"`
	Limit            int       `json:"limit"`
	Offset           int       `json:"offset"`
}

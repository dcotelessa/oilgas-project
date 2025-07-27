// backend/internal/models/auth.go
package models

import (
	"time"
	"github.com/google/uuid"
)

// User represents a user in the auth.users table
type User struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	Username            string     `json:"username,omitempty" db:"username"`
	FirstName           *string    `json:"first_name" db:"first_name"`
	LastName            *string    `json:"last_name" db:"last_name"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	Role                string     `json:"role" db:"role"`                                  // user, operator, manager, admin, super-admin
	Company             string     `json:"company" db:"company"`
	TenantID            string     `json:"tenant_id" db:"tenant_id"`                        // References auth.tenants(slug)
	Active              bool       `json:"active" db:"active"`
	EmailVerified       bool       `json:"email_verified" db:"email_verified"`
	LastLogin           *time.Time `json:"last_login" db:"last_login"`
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`                    // Security field
	LockedUntil         *time.Time `json:"-" db:"locked_until"`                             // Security field
	PasswordChangedAt   *time.Time `json:"-" db:"password_changed_at"`
	Settings            map[string]interface{} `json:"settings" db:"settings"`           // JSONB field
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Tenant represents a tenant organization in auth.tenants table
type Tenant struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Slug         string                 `json:"slug" db:"slug"`                    // URL-friendly unique identifier
	DatabaseType string                 `json:"database_type" db:"database_type"` // tenant, main, test
	DatabaseName *string                `json:"database_name" db:"database_name"` // Physical database name
	Active       bool                   `json:"active" db:"active"`
	Settings     map[string]interface{} `json:"settings" db:"settings"`           // JSONB field
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Session represents a user session in auth.sessions table
type Session struct {
	ID           string     `json:"id" db:"id"`                        // VARCHAR(255) session identifier
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`          // References auth.tenants(slug)
	IPAddress    *string    `json:"ip_address" db:"ip_address"`        // INET type
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
}

// UserTenantRole represents multi-tenant role assignments
type UserTenantRole struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TenantID  string     `json:"tenant_id" db:"tenant_id"`          // References auth.tenants(slug)
	Role      string     `json:"role" db:"role"`                    // user, operator, manager, admin
	Active    bool       `json:"active" db:"active"`
	GrantedBy *uuid.UUID `json:"granted_by" db:"granted_by"`        // UUID referencing auth.users(id)
	GrantedAt time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
}

// Request/Response structs for API
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TenantID string `json:"tenant_id,omitempty"`
}

type LoginResponse struct {
	User      *User     `json:"user"`
	Tenant    *Tenant   `json:"tenant"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Role      string `json:"role" binding:"required"`
	Company   string `json:"company" binding:"required"`
	TenantID  string `json:"tenant_id" binding:"required"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UserWithTenant represents a user with their tenant information (for joins)
type UserWithTenant struct {
	User
	TenantName string `json:"tenant_name" db:"tenant_name"`
	TenantSlug string `json:"tenant_slug" db:"tenant_slug"`
}

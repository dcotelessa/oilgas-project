package auth

import (
	"time"
)

// User represents a user with multi-tenant support
type User struct {
	ID               int       `json:"id" db:"id"`
	Email            string    `json:"email" db:"email"`
	Username         string    `json:"username" db:"username"`
	FirstName        string    `json:"first_name" db:"first_name"`
	LastName         string    `json:"last_name" db:"last_name"`
	PasswordHash     string    `json:"-" db:"password_hash"`
	Role             string    `json:"role" db:"role"`
	Company          string    `json:"company" db:"company"`
	TenantID         string    `json:"tenant_id" db:"tenant_id"`
	Active           bool      `json:"active" db:"active"`
	EmailVerified    bool      `json:"email_verified" db:"email_verified"`
	LastLogin        *time.Time `json:"last_login" db:"last_login"`
	FailedLoginAttempts int    `json:"-" db:"failed_login_attempts"`
	LockedUntil      *time.Time `json:"-" db:"locked_until"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Tenant represents a tenant organization
type Tenant struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	Code         string    `json:"code" db:"code"`
	DatabaseName string    `json:"database_name" db:"database_name"`
	DatabaseType string    `json:"database_type" db:"database_type"`
	Active       bool      `json:"active" db:"active"`
	Settings     map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserTenant represents user-tenant relationships with roles
type UserTenant struct {
	UserID   int    `json:"user_id" db:"user_id"`
	TenantID int    `json:"tenant_id" db:"tenant_id"`
	Role     string `json:"role" db:"role"`
	Active   bool   `json:"active" db:"active"`
}

// Session represents user sessions
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
}

// Auth response structs for API
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
	Company  string `json:"company" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

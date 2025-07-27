
// internal/auth/models.go
package auth

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Username  string    `json:"username" db:"username"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Tenant struct {
	ID           int    `json:"id" db:"id"`
	Code         string `json:"code" db:"code"`         // e.g., "longbeach", "houston"
	Name         string `json:"name" db:"name"`         // e.g., "Long Beach Location"
	DatabaseName string `json:"database_name" db:"database_name"` // e.g., "oilgas_longbeach"
	Active       bool   `json:"active" db:"active"`
}

type UserTenant struct {
	UserID   int    `json:"user_id" db:"user_id"`
	TenantID int    `json:"tenant_id" db:"tenant_id"`
	Role     string `json:"role" db:"role"` // "admin", "operator", "viewer"
	Active   bool   `json:"active" db:"active"`
}

type Session struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
	TenantID  int    `json:"tenant_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

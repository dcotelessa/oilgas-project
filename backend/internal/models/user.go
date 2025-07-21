package models

import "time"

type User struct {
    UserID    int       `json:"user_id" db:"user_id"`
    Username  string    `json:"username" db:"username"`
    Email     string    `json:"email" db:"email"`
    FirstName *string   `json:"first_name" db:"first_name"`
    LastName  *string   `json:"last_name" db:"last_name"`
    TenantID  *int      `json:"tenant_id" db:"tenant_id"`
    Active    bool      `json:"active" db:"active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
    Username  string  `json:"username" binding:"required"`
    Email     string  `json:"email" binding:"required,email"`
    FirstName *string `json:"first_name"`
    LastName  *string `json:"last_name"`
    TenantID  int     `json:"tenant_id" binding:"required"`
    Role      string  `json:"role" binding:"required"`
}

type UserWithTenant struct {
    User
    TenantName string `json:"tenant_name" db:"tenant_name"`
    TenantSlug string `json:"tenant_slug" db:"tenant_slug"`
    Role       string `json:"role" db:"role"`
}

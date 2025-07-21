package models

import (
    "time"
    "encoding/json"
)

type Tenant struct {
    TenantID    int                    `json:"tenant_id" db:"tenant_id"`
    TenantName  string                 `json:"tenant_name" db:"tenant_name"`
    TenantSlug  string                 `json:"tenant_slug" db:"tenant_slug"`
    Description *string                `json:"description" db:"description"`
    ContactEmail *string               `json:"contact_email" db:"contact_email"`
    Phone       *string                `json:"phone" db:"phone"`
    Address     *string                `json:"address" db:"address"`
    Active      bool                   `json:"active" db:"active"`
    Settings    json.RawMessage        `json:"settings" db:"settings"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type CustomerTenantAssignment struct {
    ID               int       `json:"id" db:"id"`
    CustomerID       int       `json:"customer_id" db:"customer_id"`
    TenantID         int       `json:"tenant_id" db:"tenant_id"`
    RelationshipType string    `json:"relationship_type" db:"relationship_type"`
    Active           bool      `json:"active" db:"active"`
    AssignedAt       time.Time `json:"assigned_at" db:"assigned_at"`
    AssignedBy       *int      `json:"assigned_by" db:"assigned_by"`
    Notes            *string   `json:"notes" db:"notes"`
}

type UserTenantRole struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    TenantID  int       `json:"tenant_id" db:"tenant_id"`
    Role      string    `json:"role" db:"role"`
    GrantedAt time.Time `json:"granted_at" db:"granted_at"`
    GrantedBy *int      `json:"granted_by" db:"granted_by"`
    Active    bool      `json:"active" db:"active"`
}

type TenantSetting struct {
    ID           int       `json:"id" db:"id"`
    TenantID     int       `json:"tenant_id" db:"tenant_id"`
    SettingKey   string    `json:"setting_key" db:"setting_key"`
    SettingValue *string   `json:"setting_value" db:"setting_value"`
    SettingType  string    `json:"setting_type" db:"setting_type"`
    Description  *string   `json:"description" db:"description"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
    UpdatedBy    *int      `json:"updated_by" db:"updated_by"`
}

// TenantContext holds the current tenant context for requests
type TenantContext struct {
    TenantID   int    `json:"tenant_id"`
    TenantSlug string `json:"tenant_slug"`
    UserRole   string `json:"user_role"`
    IsAdmin    bool   `json:"is_admin"`
}

// backend/internal/models/reference.go
package models

import "time"

// Grade represents oil & gas industry standard grades
type Grade struct {
	ID          int       `json:"id" db:"id"`
	Grade       string    `json:"grade" db:"grade"`
	Description *string   `json:"description" db:"description"`
	Strength    *int      `json:"strength" db:"strength"`
	Standard    *string   `json:"standard" db:"standard"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Size represents standard pipe sizes
type Size struct {
	ID          int       `json:"id" db:"id"`
	Size        string    `json:"size" db:"size"`
	NominalSize *string   `json:"nominal_size" db:"nominal_size"`
	OuterDiam   *float64  `json:"outer_diameter" db:"outer_diameter"`
	InnerDiam   *float64  `json:"inner_diameter" db:"inner_diameter"`
	Weight      *float64  `json:"weight_per_foot" db:"weight_per_foot"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Connection represents pipe connection types
type Connection struct {
	ID          int       `json:"id" db:"id"`
	Connection  string    `json:"connection" db:"connection"`
	Description *string   `json:"description" db:"description"`
	Category    *string   `json:"category" db:"category"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Location represents storage locations
type Location struct {
	ID          int       `json:"id" db:"id"`
	Location    string    `json:"location" db:"location"`
	Description *string   `json:"description" db:"description"`
	Capacity    *int      `json:"capacity" db:"capacity"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

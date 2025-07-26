// backend/internal/repository/tenant_interfaces.go
// Enhanced repository interfaces with tenant isolation
package repository

import (
	"context"
	"time"
)

// Enhanced Customer model with tenant support
type Customer struct {
	CustomerID      int       `json:"customer_id" db:"customer_id"`
	Customer        string    `json:"customer" db:"customer"`
	BillingAddress  *string   `json:"billing_address" db:"billing_address"`
	BillingCity     *string   `json:"billing_city" db:"billing_city"`
	BillingState    *string   `json:"billing_state" db:"billing_state"`
	BillingZipcode  *string   `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         *string   `json:"contact" db:"contact"`
	Phone           *string   `json:"phone" db:"phone"`
	Fax             *string   `json:"fax" db:"fax"`
	Email           *string   `json:"email" db:"email"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"`
	ImportedAt      *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted         bool      `json:"deleted" db:"deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Enhanced InventoryItem model with tenant support
type InventoryItem struct {
	ID           int        `json:"id" db:"id"`
	WorkOrder    *string    `json:"work_order" db:"work_order"`
	RNumber      *string    `json:"r_number" db:"r_number"`
	CustomerID   *int       `json:"customer_id" db:"customer_id"`
	Customer     *string    `json:"customer" db:"customer"`
	Joints       *int       `json:"joints" db:"joints"`
	Rack         *string    `json:"rack" db:"rack"`
	Size         *string    `json:"size" db:"size"`
	Weight       *float64   `json:"weight" db:"weight"`
	Grade        *string    `json:"grade" db:"grade"`
	Connection   *string    `json:"connection" db:"connection"`
	CTD          *string    `json:"ctd" db:"ctd"`
	WString      *string    `json:"w_string" db:"w_string"`
	Color        *string    `json:"color" db:"color"`
	DateIn       *time.Time `json:"date_in" db:"date_in"`
	DateOut      *time.Time `json:"date_out" db:"date_out"`
	WellIn       *string    `json:"well_in" db:"well_in"`
	LeaseIn      *string    `json:"lease_in" db:"lease_in"`
	WellOut      *string    `json:"well_out" db:"well_out"`
	LeaseOut     *string    `json:"lease_out" db:"lease_out"`
	Location     *string    `json:"location" db:"location"`
	Notes        *string    `json:"notes" db:"notes"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	ImportedAt   *time.Time `json:"imported_at,omitempty" db:"imported_at"`
	Deleted      bool       `json:"deleted" db:"deleted"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// New models for enhanced functionality
type CustomerRelatedData struct {
	InventoryCount     int                    `json:"inventory_count"`
	WorkOrderCount     int                    `json:"work_order_count"`
	RecentWorkOrders   []RecentWorkOrder      `json:"recent_work_orders,omitempty"`
}

type RecentWorkOrder struct {
	WorkOrder  string `json:"work_order"`
	LatestDate string `json:"latest_date"`
}

type InventorySummary struct {
	TotalJoints      int `json:"total_joints"`
	TotalWeight      float64 `json:"total_weight"`
	UniqueCustomers  int `json:"unique_customers"`
	UniqueWorkOrders int `json:"unique_work_orders"`
	UniqueSizes      int `json:"unique_sizes"`
	UniqueGrades     int `json:"unique_grades"`
}

type WorkOrder struct {
	WorkOrder    *string `json:"work_order"`
	Customer     *string `json:"customer"`
	CustomerID   *int    `json:"customer_id"`
	ItemCount    int     `json:"item_count"`
	TotalJoints  int     `json:"total_joints"`
	StartDate    *string `json:"start_date"`
	LatestDate   *string `json:"latest_date"`
	Locations    *string `json:"locations"`
	Sizes        *string `json:"sizes"`
	Grades       *string `json:"grades"`
}

type WorkOrderDetails struct {
	Summary WorkOrder                `json:"summary"`
	Items   []InventoryItem          `json:"items"`
	Related []string                 `json:"related_work_orders,omitempty"`
}

// Enhanced filter models
type CustomerFilters struct {
	Search string
	State  string
	City   string
	Limit  int
	Offset int
}

type InventoryFilters struct {
	CustomerID *int
	WorkOrder  string
	Size       string
	Grade      string
	Location   string
	DateFrom   *time.Time
	DateTo     *time.Time
	Available  *bool
	Search     string
	Limit      int
	Offset     int
}

type WorkOrderFilters struct {
	CustomerID *int
	Search     string
	Limit      int
	Offset     int
}

// Search result models
type SearchResult struct {
	Type   string      `json:"type"`
	ID     interface{} `json:"id"`
	Title  string      `json:"title"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"`
}

type SearchSummary struct {
	Customers  int `json:"customers"`
	Inventory  int `json:"inventory"`
	WorkOrders int `json:"work_orders"`
	Total      int `json:"total"`
}

type SearchResults struct {
	TenantID string         `json:"tenant"`
	Query    string         `json:"query"`
	Results  []SearchResult `json:"results"`
	Summary  SearchSummary  `json:"summary"`
}

// Tenant-aware repository interfaces
type TenantCustomerRepository interface {
	// Tenant-specific methods
	GetAllForTenant(ctx context.Context, tenantID string) ([]Customer, error)
	GetByIDForTenant(ctx context.Context, tenantID string, id int) (*Customer, error)
	SearchForTenant(ctx context.Context, tenantID, query string) ([]Customer, error)
	GetAllWithFiltersForTenant(ctx context.Context, tenantID string, filters CustomerFilters) ([]Customer, error)
	GetCountForTenant(ctx context.Context, tenantID string, filters CustomerFilters) (int, error)
	GetRelatedDataForTenant(ctx context.Context, tenantID string, customerID int) (*CustomerRelatedData, error)
	
	// Tenant-aware CRUD
	CreateForTenant(ctx context.Context, tenantID string, customer *Customer) error
	UpdateForTenant(ctx context.Context, tenantID string, customer *Customer) error
	DeleteForTenant(ctx context.Context, tenantID string, id int) error
}

type TenantInventoryRepository interface {
	// Tenant-specific methods
	GetAllForTenant(ctx context.Context, tenantID string, filters InventoryFilters) ([]InventoryItem, error)
	GetByIDForTenant(ctx context.Context, tenantID string, id int) (*InventoryItem, error)
	GetByWorkOrderForTenant(ctx context.Context, tenantID, workOrder string) ([]InventoryItem, error)
	GetAvailableForTenant(ctx context.Context, tenantID string) ([]InventoryItem, error)
	SearchForTenant(ctx context.Context, tenantID, query string) ([]InventoryItem, error)
	GetCountForTenant(ctx context.Context, tenantID string, filters InventoryFilters) (int, error)
	GetSummaryForTenant(ctx context.Context, tenantID string, filters InventoryFilters) (*InventorySummary, error)
	
	// Work order methods
	GetWorkOrdersForTenant(ctx context.Context, tenantID string, filters WorkOrderFilters) ([]WorkOrder, error)
	GetWorkOrderCountForTenant(ctx context.Context, tenantID string, filters WorkOrderFilters) (int, error)
	GetWorkOrderDetailsForTenant(ctx context.Context, tenantID, workOrderID string) (*WorkOrderDetails, error)
	SearchWorkOrdersForTenant(ctx context.Context, tenantID, query string) ([]WorkOrder, error)
	
	// Tenant-aware CRUD
	CreateForTenant(ctx context.Context, tenantID string, item *InventoryItem) error
	UpdateForTenant(ctx context.Context, tenantID string, item *InventoryItem) error
	DeleteForTenant(ctx context.Context, tenantID string, id int) error
}

// Tenant database manager interface
type TenantDatabaseManager interface {
	GetTenantDB(tenantID string) (*sql.DB, error)
	ValidateTenantExists(tenantID string) bool
	GetAllTenantIDs() ([]string, error)
	GetConnectionStats(tenantID string) (*TenantConnectionStats, error)
	CloseAllConnections()
}

type TenantConnectionStats struct {
	TenantID        string        `json:"tenant_id"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	OpenConnections int           `json:"open_connections"`
	IdleConnections int           `json:"idle_connections"`
	InUseConns      int           `json:"in_use_connections"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	LastActivity    time.Time     `json:"last_activity"`
}

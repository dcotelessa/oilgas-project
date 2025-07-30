// backend/internal/repository/interfaces.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"oilgas-backend/internal/models"
)

// Repository interfaces - Legacy (non-tenant-aware)
type CustomerRepository interface {
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, id int) (*models.Customer, error)
	Search(ctx context.Context, query string) ([]models.Customer, error)
	Create(ctx context.Context, customer *models.Customer) error
	Update(ctx context.Context, customer *models.Customer) error
	Delete(ctx context.Context, id int) error
}

type InventoryRepository interface {
	GetAll(ctx context.Context, filters models.InventoryFilters) ([]models.InventoryItem, error)
	GetByID(ctx context.Context, id int) (*models.InventoryItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) ([]models.InventoryItem, error)
	GetAvailable(ctx context.Context) ([]models.InventoryItem, error)
	Search(ctx context.Context, query string) ([]models.InventoryItem, error)
	Create(ctx context.Context, item *models.InventoryItem) error
	Update(ctx context.Context, item *models.InventoryItem) error
	Delete(ctx context.Context, id int) error
}

type ReceivedRepository interface {
	GetAll(ctx context.Context, filters models.ReceivedFilters) ([]models.ReceivedItem, error)
	GetByID(ctx context.Context, id int) (*models.ReceivedItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) ([]models.ReceivedItem, error)
	GetInProduction(ctx context.Context) ([]models.ReceivedItem, error)
	GetPending(ctx context.Context) ([]models.ReceivedItem, error)
	Create(ctx context.Context, item *models.ReceivedItem) error
	Update(ctx context.Context, item *models.ReceivedItem) error
	MarkComplete(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

// Tenant-aware repository interfaces (Fixed methods to match service usage)
type TenantCustomerRepository interface {
	// Updated signature to include filters parameter
	GetAllForTenant(ctx context.Context, tenantID string, filters models.CustomerFilters) ([]models.Customer, error)
	GetByIDForTenant(ctx context.Context, tenantID string, id int) (*models.Customer, error)
	SearchForTenant(ctx context.Context, tenantID, query string) ([]models.Customer, error)
	GetCountForTenant(ctx context.Context, tenantID string, filters models.CustomerFilters) (int, error)
	GetRelatedDataForTenant(ctx context.Context, tenantID string, customerID int) (*models.CustomerRelatedData, error)
	CreateForTenant(ctx context.Context, tenantID string, customer *models.Customer) error
	UpdateForTenant(ctx context.Context, tenantID string, customer *models.Customer) error
	DeleteForTenant(ctx context.Context, tenantID string, id int) error
}

type TenantInventoryRepository interface {
	GetAllForTenant(ctx context.Context, tenantID string, filters models.InventoryFilters) ([]models.InventoryItem, error)
	GetByIDForTenant(ctx context.Context, tenantID string, id int) (*models.InventoryItem, error)
	GetByWorkOrderForTenant(ctx context.Context, tenantID, workOrder string) ([]models.InventoryItem, error)
	GetAvailableForTenant(ctx context.Context, tenantID string) ([]models.InventoryItem, error)
	SearchForTenant(ctx context.Context, tenantID, query string) ([]models.InventoryItem, error)
	GetCountForTenant(ctx context.Context, tenantID string, filters models.InventoryFilters) (int, error)
	GetSummaryForTenant(ctx context.Context, tenantID string, filters models.InventoryFilters) (*models.InventorySummary, error)
	GetWorkOrdersForTenant(ctx context.Context, tenantID string, filters models.WorkOrderFilters) ([]models.WorkOrder, error)
	GetWorkOrderCountForTenant(ctx context.Context, tenantID string, filters models.WorkOrderFilters) (int, error)
	GetWorkOrderDetailsForTenant(ctx context.Context, tenantID, workOrderID string) (*models.WorkOrderDetails, error)
	SearchWorkOrdersForTenant(ctx context.Context, tenantID, query string) ([]models.WorkOrder, error)
	CreateForTenant(ctx context.Context, tenantID string, item *models.InventoryItem) error
	UpdateForTenant(ctx context.Context, tenantID string, item *models.InventoryItem) error
	DeleteForTenant(ctx context.Context, tenantID string, id int) error
}

type TenantReceivedRepository interface {
	GetAllForTenant(ctx context.Context, tenantID string, filters models.ReceivedFilters) ([]models.ReceivedItem, error)
	GetByIDForTenant(ctx context.Context, tenantID string, id int) (*models.ReceivedItem, error)
	GetByWorkOrderForTenant(ctx context.Context, tenantID, workOrder string) ([]models.ReceivedItem, error)
	GetInProductionForTenant(ctx context.Context, tenantID string) ([]models.ReceivedItem, error)
	GetPendingForTenant(ctx context.Context, tenantID string) ([]models.ReceivedItem, error)
	CreateForTenant(ctx context.Context, tenantID string, item *models.ReceivedItem) error
	UpdateForTenant(ctx context.Context, tenantID string, item *models.ReceivedItem) error
	MarkCompleteForTenant(ctx context.Context, tenantID string, id int) error
	DeleteForTenant(ctx context.Context, tenantID string, id int) error
}

// Search repository interface
type SearchRepository interface {
	SearchAll(ctx context.Context, query string) (*models.SearchResults, error)
	SearchAllForTenant(ctx context.Context, tenantID, query string) (*models.SearchResults, error)
}

// Reference data repository interfaces
type ReferenceRepository interface {
	GetGrades(ctx context.Context) ([]models.Grade, error)
	GetSizes(ctx context.Context) ([]models.Size, error)
	GetConnections(ctx context.Context) ([]models.Connection, error)
	GetLocationsForTenant(ctx context.Context, tenantID string) ([]models.Location, error)
}

// Tenant database management
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

// Repository collections for dependency injection
type Repositories struct {
	Customer  CustomerRepository
	Inventory InventoryRepository
	Received  ReceivedRepository
	Reference ReferenceRepository
	Search    SearchRepository
}

type TenantRepositories struct {
	Customer  TenantCustomerRepository
	Inventory TenantInventoryRepository
	Received  TenantReceivedRepository
	Search    SearchRepository
	Reference ReferenceRepository
	DBManager TenantDatabaseManager
}

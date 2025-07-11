// backend/internal/repository/interfaces.go
package repository

import (
	"context"
	"oilgas-backend/internal/models"
)

// CustomerRepository handles customer data operations
type CustomerRepository interface {
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, id int) (*models.Customer, error)
	Create(ctx context.Context, customer *models.Customer) error
	Update(ctx context.Context, customer *models.Customer) error
	Delete(ctx context.Context, id int) error
	
	// Business logic queries
	ExistsByName(ctx context.Context, name string, excludeID ...int) (bool, error)
	HasActiveInventory(ctx context.Context, customerID int) (bool, error)
	GetTotalCount(ctx context.Context) (int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error)
}

// InventoryRepository handles inventory data operations
type InventoryRepository interface {
	GetByID(ctx context.Context, id int) (*models.InventoryItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) ([]models.InventoryItem, error)
	Create(ctx context.Context, item *models.InventoryItem) error
	Update(ctx context.Context, item *models.InventoryItem) error
	Delete(ctx context.Context, id int) error
	
	// Filtering and search
	GetFiltered(ctx context.Context, filters InventoryFilters) ([]models.InventoryItem, *models.Pagination, error)
	Search(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error)
	
	// Analytics
	GetSummary(ctx context.Context) (*models.InventorySummary, error)
	GetByCustomer(ctx context.Context, customerID int) ([]models.InventoryItem, error)
	GetRecentActivity(ctx context.Context, days int) ([]models.InventoryItem, error)
}

// GradeRepository handles grade reference data
type GradeRepository interface {
	GetAll(ctx context.Context) ([]models.Grade, error)
	Create(ctx context.Context, grade *models.Grade) error
	Delete(ctx context.Context, gradeName string) error
	IsInUse(ctx context.Context, gradeName string) (bool, error)
	GetUsageStats(ctx context.Context, gradeName string) (*models.GradeUsage, error)
}

// ReceivedRepository handles received items (incoming pipe tracking)
type ReceivedRepository interface {
	GetByID(ctx context.Context, id int) (*models.ReceivedItem, error)
	Create(ctx context.Context, item *models.ReceivedItem) error
	Update(ctx context.Context, item *models.ReceivedItem) error
	Delete(ctx context.Context, id int) error
	
	GetFiltered(ctx context.Context, filters ReceivedFilters) ([]models.ReceivedItem, *models.Pagination, error)
	UpdateStatus(ctx context.Context, id int, status models.WorkflowState, notes string) error
	CanDelete(ctx context.Context, id int) (bool, string, error)
	GetByWorkOrder(ctx context.Context, workOrder string) (*models.ReceivedItem, error)
}

// Filter structs
type InventoryFilters struct {
	CustomerID     *int   `json:"customer_id,omitempty"`
	Grade          string `json:"grade,omitempty"`
	Size           string `json:"size,omitempty"`
	Color          string `json:"color,omitempty"`
	Location       string `json:"location,omitempty"`
	Rack           string `json:"rack,omitempty"`
	Connection     string `json:"connection,omitempty"`
	MinJoints      *int   `json:"min_joints,omitempty"`
	MaxJoints      *int   `json:"max_joints,omitempty"`
	DateFrom       string `json:"date_from,omitempty"`
	DateTo         string `json:"date_to,omitempty"`
	IncludeDeleted bool   `json:"include_deleted"`
	
	// Pagination
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

type ReceivedFilters struct {
	CustomerID *int   `json:"customer_id,omitempty"`
	WorkOrder      *string `json:"work_order,omitempty"`
	Grade          *string `json:"grade,omitempty"`
	Size           *string `json:"size,omitempty"`
	Connection     *string `json:"connection,omitempty"`
	Status     string `json:"status,omitempty"`
	DateFrom   string `json:"date_from,omitempty"`
	DateTo     string `json:"date_to,omitempty"`
	
	// Pagination  
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

// Pagination defaults
const (
	DefaultPage     = 1
	DefaultPerPage  = 50
	MaxPerPage      = 1000
	DefaultOrderBy  = "created_at"
	DefaultOrderDir = "DESC"
)

// NormalizePagination sets defaults for inventory filters
func (f *InventoryFilters) NormalizePagination() {
	if f.Page < 1 {
		f.Page = DefaultPage
	}
	if f.PerPage < 1 {
		f.PerPage = DefaultPerPage
	}
	if f.PerPage > MaxPerPage {
		f.PerPage = MaxPerPage
	}
	if f.OrderBy == "" {
		f.OrderBy = "date_in"
	}
	if f.OrderDir != "ASC" && f.OrderDir != "DESC" {
		f.OrderDir = DefaultOrderDir
	}
}

// NormalizePagination sets defaults for received filters
func (f *ReceivedFilters) NormalizePagination() {
	if f.Page < 1 {
		f.Page = DefaultPage
	}
	if f.PerPage < 1 {
		f.PerPage = DefaultPerPage
	}
	if f.PerPage > MaxPerPage {
		f.PerPage = MaxPerPage
	}
	if f.OrderBy == "" {
		f.OrderBy = "date_received"
	}
	if f.OrderDir != "ASC" && f.OrderDir != "DESC" {
		f.OrderDir = DefaultOrderDir
	}
}

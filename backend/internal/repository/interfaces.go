package repository

import (
	"context"
	"oilgas-backend/internal/models"
)

// WorkflowRepository defines the interface for workflow data operations
type WorkflowRepository interface {
	// Dashboard operations
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
	GetJobSummaries(ctx context.Context) ([]models.JobSummary, error)
	GetRecentActivity(ctx context.Context, limit int) ([]models.Job, error)
	
	// Job operations
	GetJobs(ctx context.Context, filters JobFilters) ([]models.Job, *models.Pagination, error)
	GetJobByID(ctx context.Context, id int) (*models.Job, error)
	GetJobByWorkOrder(ctx context.Context, workOrder string) (*models.Job, error)
	CreateJob(ctx context.Context, job *models.Job) error
	UpdateJob(ctx context.Context, job *models.Job) error
	UpdateJobState(ctx context.Context, workOrder string, state models.WorkflowState) error
	DeleteJob(ctx context.Context, id int) error
	
	// Inspection operations
	GetInspectionResults(ctx context.Context, workOrder string) ([]models.InspectionResult, error)
	CreateInspectionResult(ctx context.Context, result *models.InspectionResult) error
	UpdateInspectionResult(ctx context.Context, result *models.InspectionResult) error
	
	// Inventory operations
	GetInventory(ctx context.Context, filters InventoryFilters) ([]models.InventoryItem, *models.Pagination, error)
	GetInventoryByCustomer(ctx context.Context, customerID int) ([]models.InventoryItem, error)
	CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error
	UpdateInventoryItem(ctx context.Context, item *models.InventoryItem) error
	ShipInventory(ctx context.Context, items []int, shipmentDetails map[string]interface{}) error
	
	// Customer operations
	GetCustomers(ctx context.Context, includeDeleted bool) ([]models.Customer, error)
	GetCustomerByID(ctx context.Context, id int) (*models.Customer, error)
	CreateCustomer(ctx context.Context, customer *models.Customer) error
	UpdateCustomer(ctx context.Context, customer *models.Customer) error
	DeleteCustomer(ctx context.Context, id int) error
	
	// Pipe size operations
	GetPipeSizes(ctx context.Context, customerID int) ([]models.PipeSize, error)
	CreatePipeSize(ctx context.Context, size *models.PipeSize) error
	UpdatePipeSize(ctx context.Context, size *models.PipeSize) error
	DeletePipeSize(ctx context.Context, id int) error
	
	// Grades
	GetGrades(ctx context.Context) ([]string, error)
}

// JobFilters represents filters for job queries
type JobFilters struct {
	CustomerID   *int                  `json:"customer_id,omitempty"`
	State        *models.WorkflowState `json:"state,omitempty"`
	WorkOrder    string                `json:"work_order,omitempty"`
	Grade        string                `json:"grade,omitempty"`
	Size         string                `json:"size,omitempty"`
	DateFrom     string                `json:"date_from,omitempty"`
	DateTo       string                `json:"date_to,omitempty"`
	IncludeDeleted bool                `json:"include_deleted"`
	
	// Pagination
	Page     int `json:"page"`
	PerPage  int `json:"per_page"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

// InventoryFilters represents filters for inventory queries
type InventoryFilters struct {
	CustomerID     *int   `json:"customer_id,omitempty"`
	Grade          string `json:"grade,omitempty"`
	Size           string `json:"size,omitempty"`
	Color          string `json:"color,omitempty"`
	CN             *models.ColorNumber `json:"cn,omitempty"`
	MinJoints      *int   `json:"min_joints,omitempty"`
	MaxJoints      *int   `json:"max_joints,omitempty"`
	Rack           string `json:"rack,omitempty"`
	IncludeShipped bool   `json:"include_shipped"`
	
	// Pagination
	Page     int `json:"page"`
	PerPage  int `json:"per_page"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

// Default values for pagination
const (
	DefaultPage     = 1
	DefaultPerPage  = 50
	MaxPerPage      = 1000
	DefaultOrderBy  = "created_at"
	DefaultOrderDir = "DESC"
)

// NormalizePagination sets defaults and validates pagination parameters
func (f *JobFilters) NormalizePagination() {
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
		f.OrderBy = DefaultOrderBy
	}
	if f.OrderDir != "ASC" && f.OrderDir != "DESC" {
		f.OrderDir = DefaultOrderDir
	}
}

// NormalizePagination sets defaults and validates pagination parameters
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
		f.OrderBy = "datein"
	}
	if f.OrderDir != "ASC" && f.OrderDir != "DESC" {
		f.OrderDir = DefaultOrderDir
	}
}

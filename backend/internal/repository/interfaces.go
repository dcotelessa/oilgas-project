// internal/repository/interfaces.go
package repository

import (
	"context"
	"time"
)

// Customer domain
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
	Deleted         bool      `json:"deleted" db:"deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type CustomerRepository interface {
	GetAll(ctx context.Context) ([]Customer, error)
	GetByID(ctx context.Context, id int) (*Customer, error)
	Search(ctx context.Context, query string) ([]Customer, error)
	Create(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id int) error
}

// Inventory domain
type InventoryItem struct {
	ID           int       `json:"id" db:"id"`
	WorkOrder    *string   `json:"work_order" db:"work_order"`
	CustomerID   *int      `json:"customer_id" db:"customer_id"`
	Customer     *string   `json:"customer" db:"customer"`
	Joints       *int      `json:"joints" db:"joints"`
	Size         *string   `json:"size" db:"size"`
	Weight       *float64  `json:"weight" db:"weight"`
	Grade        *string   `json:"grade" db:"grade"`
	Connection   *string   `json:"connection" db:"connection"`
	DateIn       *time.Time `json:"date_in" db:"date_in"`
	DateOut      *time.Time `json:"date_out" db:"date_out"`
	WellIn       *string   `json:"well_in" db:"well_in"`
	LeaseIn      *string   `json:"lease_in" db:"lease_in"`
	WellOut      *string   `json:"well_out" db:"well_out"`
	LeaseOut     *string   `json:"lease_out" db:"lease_out"`
	Location     *string   `json:"location" db:"location"`
	Notes        *string   `json:"notes" db:"notes"`
	Deleted      bool      `json:"deleted" db:"deleted"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type InventoryRepository interface {
	GetAll(ctx context.Context, filters InventoryFilters) ([]InventoryItem, error)
	GetByID(ctx context.Context, id int) (*InventoryItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) ([]InventoryItem, error)
	GetAvailable(ctx context.Context) ([]InventoryItem, error)
	Search(ctx context.Context, query string) ([]InventoryItem, error)
	Create(ctx context.Context, item *InventoryItem) error
	Update(ctx context.Context, item *InventoryItem) error
	Delete(ctx context.Context, id int) error
}

type InventoryFilters struct {
	CustomerID *int
	Grade      *string
	Size       *string
	Location   *string
	DateFrom   *time.Time
	DateTo     *time.Time
	Available  *bool
	Limit      int
	Offset     int
}

// Received items domain
type ReceivedItem struct {
	ID             int       `json:"id" db:"id"`
	WorkOrder      *string   `json:"work_order" db:"work_order"`
	CustomerID     *int      `json:"customer_id" db:"customer_id"`
	Customer       *string   `json:"customer" db:"customer"`
	Joints         *int      `json:"joints" db:"joints"`
	Size           *string   `json:"size" db:"size"`
	Weight         *float64  `json:"weight" db:"weight"`
	Grade          *string   `json:"grade" db:"grade"`
	Connection     *string   `json:"connection" db:"connection"`
	Well           *string   `json:"well" db:"well"`
	Lease          *string   `json:"lease" db:"lease"`
	OrderedBy      *string   `json:"ordered_by" db:"ordered_by"`
	Notes          *string   `json:"notes" db:"notes"`
	DateReceived   *time.Time `json:"date_received" db:"date_received"`
	InProduction   bool      `json:"in_production" db:"in_production"`
	Complete       bool      `json:"complete" db:"complete"`
	Deleted        bool      `json:"deleted" db:"deleted"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type ReceivedRepository interface {
	GetAll(ctx context.Context, filters ReceivedFilters) ([]ReceivedItem, error)
	GetByID(ctx context.Context, id int) (*ReceivedItem, error)
	GetByWorkOrder(ctx context.Context, workOrder string) ([]ReceivedItem, error)
	GetInProduction(ctx context.Context) ([]ReceivedItem, error)
	GetPending(ctx context.Context) ([]ReceivedItem, error)
	Create(ctx context.Context, item *ReceivedItem) error
	Update(ctx context.Context, item *ReceivedItem) error
	MarkComplete(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

type ReceivedFilters struct {
	CustomerID   *int
	InProduction *bool
	Complete     *bool
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
}

// Repository collection
type Repositories struct {
	Customer  CustomerRepository
	Inventory InventoryRepository
	Received  ReceivedRepository
}

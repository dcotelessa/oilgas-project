// backend/internal/models/models.go
package models

import (
	"time"
)

// Customer represents a customer in the oil & gas inventory system
type Customer struct {
	CustomerID      int       `json:"customer_id" db:"customer_id"`
	Customer        string    `json:"customer" db:"customer"`
	BillingAddress  string    `json:"billing_address" db:"billing_address"`
	BillingCity     string    `json:"billing_city" db:"billing_city"`
	BillingState    string    `json:"billing_state" db:"billing_state"`
	BillingZipcode  string    `json:"billing_zipcode" db:"billing_zipcode"`
	Contact         string    `json:"contact" db:"contact"`
	Phone           string    `json:"phone" db:"phone"`
	Fax             string    `json:"fax" db:"fax"`
	Email           string    `json:"email" db:"email"`
	Color1          string    `json:"color1" db:"color1"`
	Color2          string    `json:"color2" db:"color2"`
	Color3          string    `json:"color3" db:"color3"`
	Color4          string    `json:"color4" db:"color4"`
	Color5          string    `json:"color5" db:"color5"`
	Loss1           string    `json:"loss1" db:"loss1"`
	Loss2           string    `json:"loss2" db:"loss2"`
	Loss3           string    `json:"loss3" db:"loss3"`
	Loss4           string    `json:"loss4" db:"loss4"`
	Loss5           string    `json:"loss5" db:"loss5"`
	WSColor1        string    `json:"ws_color1" db:"wscolor1"`
	WSColor2        string    `json:"ws_color2" db:"wscolor2"`
	WSColor3        string    `json:"ws_color3" db:"wscolor3"`
	WSColor4        string    `json:"ws_color4" db:"wscolor4"`
	WSColor5        string    `json:"ws_color5" db:"wscolor5"`
	WSLoss1         string    `json:"ws_loss1" db:"wsloss1"`
	WSLoss2         string    `json:"ws_loss2" db:"wsloss2"`
	WSLoss3         string    `json:"ws_loss3" db:"wsloss3"`
	WSLoss4         string    `json:"ws_loss4" db:"wsloss4"`
	WSLoss5         string    `json:"ws_loss5" db:"wsloss5"`
	Deleted         bool      `json:"deleted" db:"deleted"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// InventoryItem represents an inventory item
type InventoryItem struct {
	ID         int        `json:"id" db:"id"`
	Username   string     `json:"username" db:"username"`
	WorkOrder  string     `json:"work_order" db:"work_order"`
	RNumber    int        `json:"r_number" db:"r_number"`
	CustomerID int        `json:"customer_id" db:"customer_id"`
	Customer   string     `json:"customer" db:"customer"`
	Joints     int        `json:"joints" db:"joints"`
	Rack       string     `json:"rack" db:"rack"`
	Size       string     `json:"size" db:"size"`
	Weight     string     `json:"weight" db:"weight"`
	Grade      string     `json:"grade" db:"grade"`
	Connection string     `json:"connection" db:"connection"`
	CTD        bool       `json:"ctd" db:"ctd"`
	WString    bool       `json:"w_string" db:"w_string"`
	SWGCC      string     `json:"swgcc" db:"swgcc"`
	Color      string     `json:"color" db:"color"`
	CustomerPO string     `json:"customer_po" db:"customer_po"`
	Fletcher   string     `json:"fletcher" db:"fletcher"`
	DateIn     *time.Time `json:"date_in" db:"date_in"`
	DateOut    *time.Time `json:"date_out" db:"date_out"`
	WellIn     string     `json:"well_in" db:"well_in"`
	LeaseIn    string     `json:"lease_in" db:"lease_in"`
	WellOut    string     `json:"well_out" db:"well_out"`
	LeaseOut   string     `json:"lease_out" db:"lease_out"`
	Trucking   string     `json:"trucking" db:"trucking"`
	Trailer    string     `json:"trailer" db:"trailer"`
	Location   string     `json:"location" db:"location"`
	Notes      string     `json:"notes" db:"notes"`
	PCode      string     `json:"pcode" db:"pcode"`
	CN         int        `json:"cn" db:"cn"`
	OrderedBy  string     `json:"ordered_by" db:"ordered_by"`
	Deleted    bool       `json:"deleted" db:"deleted"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ReceivedItem represents items received into inventory
type ReceivedItem struct {
	ID                int        `json:"id" db:"id"`
	WorkOrder         string     `json:"work_order" db:"work_order"`
	CustomerID        int        `json:"customer_id" db:"customer_id"`
	Customer          string     `json:"customer" db:"customer"`
	Joints            int        `json:"joints" db:"joints"`
	Rack              string     `json:"rack" db:"rack"`
	SizeID            int        `json:"size_id" db:"size_id"`
	Size              string     `json:"size" db:"size"`
	Weight            string     `json:"weight" db:"weight"`
	Grade             string     `json:"grade" db:"grade"`
	Connection        string     `json:"connection" db:"connection"`
	CTD               bool       `json:"ctd" db:"ctd"`
	WString           bool       `json:"w_string" db:"w_string"`
	Well              string     `json:"well" db:"well"`
	Lease             string     `json:"lease" db:"lease"`
	OrderedBy         string     `json:"ordered_by" db:"ordered_by"`
	Notes             string     `json:"notes" db:"notes"`
	CustomerPO        string     `json:"customer_po" db:"customer_po"`
	DateReceived      *time.Time `json:"date_received" db:"date_received"`
	Background        string     `json:"background" db:"background"`
	Norm              string     `json:"norm" db:"norm"`
	Services          string     `json:"services" db:"services"`
	BillToID          string     `json:"bill_to_id" db:"bill_to_id"`
	EnteredBy         string     `json:"entered_by" db:"entered_by"`
	WhenEntered       *time.Time `json:"when_entered" db:"when_entered"`
	Trucking          string     `json:"trucking" db:"trucking"`
	Trailer           string     `json:"trailer" db:"trailer"`
	InProduction      *time.Time `json:"in_production" db:"in_production"`
	InspectedDate     *time.Time `json:"inspected_date" db:"inspected_date"`
	ThreadingDate     *time.Time `json:"threading_date" db:"threading_date"`
	StraightenRequired bool      `json:"straighten_required" db:"straighten_required"`
	ExcessMaterial    bool       `json:"excess_material" db:"excess_material"`
	Complete          bool       `json:"complete" db:"complete"`
	InspectedBy       string     `json:"inspected_by" db:"inspected_by"`
	UpdatedBy         string     `json:"updated_by" db:"updated_by"`
	WhenUpdated       *time.Time `json:"when_updated" db:"when_updated"`
	Deleted           bool       `json:"deleted" db:"deleted"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// FletcherItem represents items in fletcher operations (threading/inspection)
type FletcherItem struct {
	ID         int        `json:"id" db:"id"`
	Username   string     `json:"username" db:"username"`
	Fletcher   string     `json:"fletcher" db:"fletcher"`
	RNumber    int        `json:"r_number" db:"r_number"`
	CustomerID int        `json:"customer_id" db:"customer_id"`
	Customer   string     `json:"customer" db:"customer"`
	Joints     int        `json:"joints" db:"joints"`
	Size       string     `json:"size" db:"size"`
	Weight     string     `json:"weight" db:"weight"`
	Grade      string     `json:"grade" db:"grade"`
	Connection string     `json:"connection" db:"connection"`
	CTD        bool       `json:"ctd" db:"ctd"`
	WString    bool       `json:"w_string" db:"w_string"`
	SWGCC      string     `json:"swgcc" db:"swgcc"`
	Color      string     `json:"color" db:"color"`
	CustomerPO string     `json:"customer_po" db:"customer_po"`
	DateIn     *time.Time `json:"date_in" db:"date_in"`
	DateOut    *time.Time `json:"date_out" db:"date_out"`
	WellIn     string     `json:"well_in" db:"well_in"`
	LeaseIn    string     `json:"lease_in" db:"lease_in"`
	WellOut    string     `json:"well_out" db:"well_out"`
	LeaseOut   string     `json:"lease_out" db:"lease_out"`
	Trucking   string     `json:"trucking" db:"trucking"`
	Trailer    string     `json:"trailer" db:"trailer"`
	Location   string     `json:"location" db:"location"`
	Notes      string     `json:"notes" db:"notes"`
	PCode      string     `json:"pcode" db:"pcode"`
	CN         int        `json:"cn" db:"cn"`
	OrderedBy  string     `json:"ordered_by" db:"ordered_by"`
	Deleted    bool       `json:"deleted" db:"deleted"`
	Complete   bool       `json:"complete" db:"complete"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// BakeoutItem represents items in bakeout process
type BakeoutItem struct {
	ID         int    `json:"id" db:"id"`
	Fletcher   string `json:"fletcher" db:"fletcher"`
	Joints     int    `json:"joints" db:"joints"`
	Color      string `json:"color" db:"color"`
	Size       string `json:"size" db:"size"`
	Weight     string `json:"weight" db:"weight"`
	Grade      string `json:"grade" db:"grade"`
	Connection string `json:"connection" db:"connection"`
	CTD        bool   `json:"ctd" db:"ctd"`
	SWGCC      string `json:"swgcc" db:"swgcc"`
	CustomerID int    `json:"customer_id" db:"customer_id"`
	Accept     int    `json:"accept" db:"accept"`
	Reject     int    `json:"reject" db:"reject"`
	Pin        int    `json:"pin" db:"pin"`
	Cplg       int    `json:"cplg" db:"cplg"`
	PC         int    `json:"pc" db:"pc"`
	Trucking   string `json:"trucking" db:"trucking"`
	Trailer    string `json:"trailer" db:"trailer"`
	DateIn     *time.Time `json:"date_in" db:"date_in"`
	CN         int    `json:"cn" db:"cn"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// InspectedItem represents inspected inventory items
type InspectedItem struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	WorkOrder string `json:"work_order" db:"work_order"`
	Color     string `json:"color" db:"color"`
	Joints    int    `json:"joints" db:"joints"`
	Accept    int    `json:"accept" db:"accept"`
	Reject    int    `json:"reject" db:"reject"`
	Pin       int    `json:"pin" db:"pin"`
	Cplg      int    `json:"cplg" db:"cplg"`
	PC        int    `json:"pc" db:"pc"`
	Complete  bool   `json:"complete" db:"complete"`
	Rack      string `json:"rack" db:"rack"`
	RepPin    int    `json:"rep_pin" db:"rep_pin"`
	RepCplg   int    `json:"rep_cplg" db:"rep_cplg"`
	RepPC     int    `json:"rep_pc" db:"rep_pc"`
	Deleted   bool   `json:"deleted" db:"deleted"`
	CN        int    `json:"cn" db:"cn"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Grade represents available pipe grades
type Grade struct {
	Grade string `json:"grade" db:"grade"`
}

// SWGC represents Size, Weight, Grade, Connection configurations
type SWGC struct {
	SizeID          int    `json:"size_id" db:"size_id"`
	CustomerID      int    `json:"customer_id" db:"customer_id"`
	Size            string `json:"size" db:"size"`
	Weight          string `json:"weight" db:"weight"`
	Connection      string `json:"connection" db:"connection"`
	PCodeReceive    string `json:"pcode_receive" db:"pcode_receive"`
	PCodeInventory  string `json:"pcode_inventory" db:"pcode_inventory"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// User represents system users
type User struct {
	UserID    int    `json:"user_id" db:"user_id"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"-" db:"password"` // Never expose password in JSON
	Access    int    `json:"access" db:"access"`
	FullName  string `json:"full_name" db:"full_name"`
	Email     string `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Temporary tables for processing
type TempItem struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	WorkOrder string `json:"work_order" db:"work_order"`
	Color     string `json:"color" db:"color"`
	Joints    int    `json:"joints" db:"joints"`
	Accept    int    `json:"accept" db:"accept"`
	Reject    int    `json:"reject" db:"reject"`
	Pin       int    `json:"pin" db:"pin"`
	Cplg      int    `json:"cplg" db:"cplg"`
	PC        int    `json:"pc" db:"pc"`
	Rack      string `json:"rack" db:"rack"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RNumber represents R numbers used in the system
type RNumber struct {
	RNumber int `json:"r_number" db:"r_number"`
}

// WKNumber represents work order numbers
type WKNumber struct {
	WKNumber int `json:"wk_number" db:"wk_number"`
}

// InventorySummary represents dashboard summary data
type InventorySummary struct {
	TotalItems        int                    `json:"total_items"`
	TotalJoints       int                    `json:"total_joints"`
	ItemsByGrade      map[string]int         `json:"items_by_grade"`
	ItemsByCustomer   map[string]int         `json:"items_by_customer"`
	ItemsByLocation   map[string]int         `json:"items_by_location"`
	RecentActivity    []InventoryItem        `json:"recent_activity"`
	TopCustomers      []Customer             `json:"top_customers"`
	LowStock          []InventoryItem        `json:"low_stock"`
	PendingInspection int                    `json:"pending_inspection"`
	InProduction      int                    `json:"in_production"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// SearchResult represents search results across multiple tables
type SearchResult struct {
	Type        string      `json:"type"`        // "inventory", "customer", "received"
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Relevance   float64     `json:"relevance"`
}

// JobSummary represents a simplified view of work items for dashboard
type JobSummary struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	Customer   string    `json:"customer"`
	Grade      string    `json:"grade"`
	Size       string    `json:"size"`
	Joints     int       `json:"joints"`
	Status     string    `json:"status"` // "active", "completed", "pending"
	CreatedAt  time.Time `json:"created_at"`
}

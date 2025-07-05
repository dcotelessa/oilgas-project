// backend/internal/models/reference_models.go
package models

import (
	"fmt"
	"strings"
	"time"
)

// Grade represents available pipe grades
type Grade struct {
	Grade string `json:"grade" db:"grade"`
}

// SWGC represents Size, Weight, Grade, Connection configurations
type SWGC struct {
	SizeID          int       `json:"size_id" db:"size_id"`
	CustomerID      int       `json:"customer_id" db:"customer_id"`
	Size            string    `json:"size" db:"size"`
	Weight          string    `json:"weight" db:"weight"`
	Connection      string    `json:"connection" db:"connection"`
	PCodeReceive    string    `json:"pcode_receive" db:"pcode_receive"`
	PCodeInventory  string    `json:"pcode_inventory" db:"pcode_inventory"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// User represents system users
type User struct {
	UserID    int       `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // Never expose password in JSON
	Access    int       `json:"access" db:"access"`
	FullName  string    `json:"full_name" db:"full_name"`
	Email     string    `json:"email" db:"email"`
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

// TempInvItem represents temporary inventory items
type TempInvItem struct {
	ID         int        `json:"id" db:"id"`
	Username   string     `json:"username" db:"username"`
	WorkOrder  string     `json:"work_order" db:"work_order"`
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
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// TestItem represents test records
type TestItem struct {
	ID   int    `json:"id" db:"id"`
	Test string `json:"test" db:"test"`
}

// Standard oil & gas grades
var StandardGrades = []string{
	"J55", "JZ55", "K55", "L80", "N80", 
	"P105", "P110", "Q125", "T95", "C90", "C95", "S135",
}

// Standard pipe sizes (OD in inches)
var StandardSizes = []string{
	"2 3/8\"", "2 7/8\"", "3 1/2\"", "4\"", "4 1/2\"", "5\"", "5 1/2\"",
	"6 5/8\"", "7\"", "7 5/8\"", "8 5/8\"", "9 5/8\"", "10 3/4\"", 
	"11 3/4\"", "13 3/8\"", "16\"", "18 5/8\"", "20\"",
}

// Standard connection types
var StandardConnections = []string{
	"LTC", "STC", "BTC", "EUE", "NUE", "PREMIUM", "VAM", "NEW VAM", "TENARIS",
}

// Grade validation
func IsValidGrade(grade string) bool {
	for _, validGrade := range StandardGrades {
		if strings.EqualFold(grade, validGrade) {
			return true
		}
	}
	return false
}

// Size validation
func IsValidSize(size string) bool {
	for _, validSize := range StandardSizes {
		if strings.EqualFold(size, validSize) {
			return true
		}
	}
	return false
}

// Connection validation
func IsValidConnection(connection string) bool {
	for _, validConnection := range StandardConnections {
		if strings.EqualFold(connection, validConnection) {
			return true
		}
	}
	return false
}

// User access levels
const (
	AccessAdmin    = 1
	AccessOperator = 2
	AccessViewer   = 3
)

// GetAccessLevelName returns human-readable access level
func (u *User) GetAccessLevelName() string {
	switch u.Access {
	case AccessAdmin:
		return "Administrator"
	case AccessOperator:
		return "Operator"
	case AccessViewer:
		return "Viewer"
	default:
		return "Unknown"
	}
}

// CanWrite returns true if user can modify data
func (u *User) CanWrite() bool {
	return u.Access <= AccessOperator
}

// CanDelete returns true if user can delete data
func (u *User) CanDelete() bool {
	return u.Access == AccessAdmin
}

// IsActive returns true if user account is active
func (u *User) IsActive() bool {
	return u.Username != "" && u.Access > 0
}

// SWGC methods
func (s *SWGC) GetSpecification() string {
	return fmt.Sprintf("%s %s %s", s.Size, s.Weight, s.Connection)
}

func (s *SWGC) Matches(size, weight, connection string) bool {
	return strings.EqualFold(s.Size, size) &&
		   strings.EqualFold(s.Weight, weight) &&
		   strings.EqualFold(s.Connection, connection)
}

// Grade methods  
func (g *Grade) IsStandard() bool {
	return IsValidGrade(g.Grade)
}

func (g *Grade) GetDescription() string {
	descriptions := map[string]string{
		"J55":  "Basic grade steel casing",
		"JZ55": "Enhanced J55 grade",
		"L80":  "Higher strength grade",
		"N80":  "Medium strength grade",
		"P105": "High performance grade",
		"P110": "Premium performance grade",
		"Q125": "Ultra-high strength grade",
		"T95":  "Intermediate strength grade",
		"C90":  "Corrosion resistant grade",
		"C95":  "Enhanced corrosion resistant grade",
		"S135": "Super high strength grade",
	}
	
	if desc, exists := descriptions[strings.ToUpper(g.Grade)]; exists {
		return desc
	}
	
	return "Standard oil & gas grade"
}

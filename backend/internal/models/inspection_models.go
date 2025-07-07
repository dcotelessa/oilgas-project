// backend/internal/models/inspection_models.go
package models

import (
	"fmt"
	"strings"
	"time"
)

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
	ID         int        `json:"id" db:"id"`
	Fletcher   string     `json:"fletcher" db:"fletcher"`
	Joints     int        `json:"joints" db:"joints"`
	Color      string     `json:"color" db:"color"`
	Size       string     `json:"size" db:"size"`
	Weight     string     `json:"weight" db:"weight"`
	Grade      string     `json:"grade" db:"grade"`
	Connection string     `json:"connection" db:"connection"`
	CTD        bool       `json:"ctd" db:"ctd"`
	SWGCC      string     `json:"swgcc" db:"swgcc"`
	CustomerID int        `json:"customer_id" db:"customer_id"`
	Accept     int        `json:"accept" db:"accept"`
	Reject     int        `json:"reject" db:"reject"`
	Pin        int        `json:"pin" db:"pin"`
	Cplg       int        `json:"cplg" db:"cplg"`
	PC         int        `json:"pc" db:"pc"`
	Trucking   string     `json:"trucking" db:"trucking"`
	Trailer    string     `json:"trailer" db:"trailer"`
	DateIn     *time.Time `json:"date_in" db:"date_in"`
	CN         int        `json:"cn" db:"cn"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// InspectedItem represents inspected inventory items
type InspectedItem struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	WorkOrder string    `json:"work_order" db:"work_order"`
	Color     string    `json:"color" db:"color"`
	Joints    int       `json:"joints" db:"joints"`
	Accept    int       `json:"accept" db:"accept"`
	Reject    int       `json:"reject" db:"reject"`
	Pin       int       `json:"pin" db:"pin"`
	Cplg      int       `json:"cplg" db:"cplg"`
	PC        int       `json:"pc" db:"pc"`
	Complete  bool      `json:"complete" db:"complete"`
	Rack      string    `json:"rack" db:"rack"`
	RepPin    int       `json:"rep_pin" db:"rep_pin"`
	RepCplg   int       `json:"rep_cplg" db:"rep_cplg"`
	RepPC     int       `json:"rep_pc" db:"rep_pc"`
	Deleted   bool      `json:"deleted" db:"deleted"`
	CN        int       `json:"cn" db:"cn"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TempItem represents temporary processing items
type TempItem struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	WorkOrder string    `json:"work_order" db:"work_order"`
	Color     string    `json:"color" db:"color"`
	Joints    int       `json:"joints" db:"joints"`
	Accept    int       `json:"accept" db:"accept"`
	Reject    int       `json:"reject" db:"reject"`
	Pin       int       `json:"pin" db:"pin"`
	Cplg      int       `json:"cplg" db:"cplg"`
	PC        int       `json:"pc" db:"pc"`
	Rack      string    `json:"rack" db:"rack"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// InspectionResult represents inspection data for workflow
type InspectionResult struct {
	ID          int         `json:"id" db:"id"`
	WorkOrder   string      `json:"work_order" db:"work_order"`
	Color       string      `json:"color" db:"color"`
	CN          ColorNumber `json:"cn" db:"cn"`
	Joints      int         `json:"joints" db:"joints"`
	Accept      int         `json:"accept" db:"accept"`
	Reject      int         `json:"reject" db:"reject"`
	Pin         int         `json:"pin" db:"pin"`
	Coupling    int         `json:"cplg" db:"cplg"`
	PC          int         `json:"pc" db:"pc"`
	Rack        string      `json:"rack,omitempty" db:"rack"`
	RepairPin   int         `json:"repair_pin" db:"rep_pin"`
	RepairCplg  int         `json:"repair_cplg" db:"rep_cplg"`
	RepairPC    int         `json:"repair_pc" db:"rep_pc"`
	Complete    bool        `json:"complete" db:"complete"`
}

// FletcherItem methods
func (f *FletcherItem) IsComplete() bool {
	return f.Complete
}

func (f *FletcherItem) GetDaysInProcess() int {
	if f.DateIn == nil {
		return 0
	}
	
	endDate := time.Now()
	if f.DateOut != nil {
		endDate = *f.DateOut
	}
	
	return int(endDate.Sub(*f.DateIn).Hours() / 24)
}

func (f *FletcherItem) GetStatus() string {
	if f.Deleted {
		return "deleted"
	}
	if f.Complete {
		return "completed"
	}
	if f.DateIn != nil {
		return "in_process"
	}
	return "pending"
}

// BakeoutItem methods
func (b *BakeoutItem) GetAcceptanceRate() float64 {
	if b.Joints == 0 {
		return 0.0
	}
	return float64(b.Accept) / float64(b.Joints) * 100
}

func (b *BakeoutItem) GetRejectionRate() float64 {
	if b.Joints == 0 {
		return 0.0
	}
	return float64(b.Reject) / float64(b.Joints) * 100
}

func (b *BakeoutItem) GetTotalDefects() int {
	return b.Pin + b.Cplg + b.PC
}

func (b *BakeoutItem) HasDefects() bool {
	return b.GetTotalDefects() > 0
}

// InspectedItem methods
func (i *InspectedItem) GetQualityMetrics() map[string]int {
	return map[string]int{
		"accept":   i.Accept,
		"reject":   i.Reject,
		"pin":      i.Pin,
		"coupling": i.Cplg,
		"pc":       i.PC,
	}
}

func (i *InspectedItem) GetRepairMetrics() map[string]int {
	return map[string]int{
		"rep_pin":  i.RepPin,
		"rep_cplg": i.RepCplg,
		"rep_pc":   i.RepPC,
	}
}

func (i *InspectedItem) NeedsRepair() bool {
	return i.RepPin > 0 || i.RepCplg > 0 || i.RepPC > 0
}

func (i *InspectedItem) GetAcceptanceRate() float64 {
	if i.Joints == 0 {
		return 0.0
	}
	return float64(i.Accept) / float64(i.Joints) * 100
}

// InspectionResult methods
func (ir *InspectionResult) GetTotalProcessed() int {
	return ir.Accept + ir.Reject
}

func (ir *InspectionResult) GetQualityRate() float64 {
	total := ir.GetTotalProcessed()
	if total == 0 {
		return 0.0
	}
	return float64(ir.Accept) / float64(total) * 100
}

func (ir *InspectionResult) HasRepairWork() bool {
	return ir.RepairPin > 0 || ir.RepairCplg > 0 || ir.RepairPC > 0
}

func (ir *InspectionResult) GetRepairSummary() string {
	repairs := []string{}
	
	if ir.RepairPin > 0 {
		repairs = append(repairs, fmt.Sprintf("%d pin repairs", ir.RepairPin))
	}
	if ir.RepairCplg > 0 {
		repairs = append(repairs, fmt.Sprintf("%d coupling repairs", ir.RepairCplg))
	}
	if ir.RepairPC > 0 {
		repairs = append(repairs, fmt.Sprintf("%d PC repairs", ir.RepairPC))
	}
	
	if len(repairs) == 0 {
		return "No repairs needed"
	}
	
	return strings.Join(repairs, ", ")
}

// InspectionItem represents items that have been inspected
type InspectionItem struct {
	ID             int        `json:"id" db:"id"`
	WorkOrder      string     `json:"work_order" db:"work_order"`
	CustomerID     int        `json:"customer_id" db:"customer_id"`
	Customer       string     `json:"customer" db:"customer"`
	Joints         int        `json:"joints" db:"joints"`
	Size           string     `json:"size" db:"size"`
	Weight         string     `json:"weight" db:"weight"`
	Grade          string     `json:"grade" db:"grade"`
	Connection     string     `json:"connection" db:"connection"`
	PassedJoints   int        `json:"passed_joints" db:"passed_joints"`
	FailedJoints   int        `json:"failed_joints" db:"failed_joints"`
	InspectionDate *time.Time `json:"inspection_date" db:"inspection_date"`
	Inspector      string     `json:"inspector" db:"inspector"`
	Notes          string     `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

func (i *InspectionItem) GetPassRate() float64 {
	if i.Joints == 0 {
		return 0
	}
	return float64(i.PassedJoints) / float64(i.Joints) * 100
}

func (i *InspectionItem) IsComplete() bool {
	return i.PassedJoints+i.FailedJoints == i.Joints
}

func (i *InspectionItem) GetStatus() string {
	if !i.IsComplete() {
		return "in_progress"
	}
	
	passRate := i.GetPassRate()
	switch {
	case passRate >= 95:
		return "excellent"
	case passRate >= 85:
		return "good"
	case passRate >= 70:
		return "fair"
	default:
		return "poor"
	}
}

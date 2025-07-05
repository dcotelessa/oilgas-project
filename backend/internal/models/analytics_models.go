// backend/internal/models/analytics_models.go
package models

import "time"

// DashboardStats represents aggregated dashboard data
type DashboardStats struct {
	ActiveInventory   int                    `json:"active_inventory"`
	TotalCustomers    int                    `json:"total_customers"`
	ActiveJobs        int                    `json:"active_jobs"`
	TotalJoints       int                    `json:"total_joints"`
	TopCustomers      []CustomerStats        `json:"top_customers"`
	InventoryByGrade  map[string]int         `json:"inventory_by_grade"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// JobSummary represents workflow state summaries
type JobSummary struct {
	Status      string  `json:"status"`
	Count       int     `json:"count"`
	TotalJoints int     `json:"total_joints"`
	AvgDays     float64 `json:"avg_days"`
}

// ActivityItem represents a recent activity entry
type ActivityItem struct {
	Type         string    `json:"type"`         // "inventory", "received", "customer"
	ReferenceID  string    `json:"reference_id"` // ID of the related item
	Title        string    `json:"title"`        // Customer name or main identifier
	Description  string    `json:"description"`  // What happened
	CustomerID   int       `json:"customer_id"`
	ActivityTime time.Time `json:"activity_time"`
}

// CustomerActivity represents customer activity analytics
type CustomerActivity struct {
	Days            int `json:"days"`
	ActiveCustomers int `json:"active_customers"`
	TotalActivities int `json:"total_activities"`
	TotalJoints     int `json:"total_joints"`
}

// GradeAnalytics represents grade-based analytics
type GradeAnalytics struct {
	GradeDistribution map[string]GradeStats `json:"grade_distribution"`
	LastUpdated       time.Time             `json:"last_updated"`
}

// GradeStats represents statistics for a specific grade
type GradeStats struct {
	ItemCount         int     `json:"item_count"`
	TotalJoints       int     `json:"total_joints"`
	AvgJointsPerItem  float64 `json:"avg_joints_per_item"`
	CustomerCount     int     `json:"customer_count"`
}

// InventoryTrends represents inventory trends over time
type InventoryTrends struct {
	Days       int                     `json:"days"`
	DailyStats []DailyInventoryStats   `json:"daily_stats"`
}

// DailyInventoryStats represents daily inventory statistics
type DailyInventoryStats struct {
	Date            time.Time `json:"date"`
	ItemsAdded      int       `json:"items_added"`
	JointsAdded     int       `json:"joints_added"`
	CustomersActive int       `json:"customers_active"`
}

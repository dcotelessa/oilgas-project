// backend/internal/services/workflow_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
)

type WorkflowService struct {
	repos *repository.Repositories
	cache CacheService
}

func NewWorkflowService(repos *repository.Repositories, cache CacheService) *WorkflowService {
	return &WorkflowService{
		repos: repos,
		cache: cache,
	}
}

type WorkflowStats struct {
	TotalItems      int                    `json:"total_items"`
	TotalJoints     int                    `json:"total_joints"`
	ItemsByCustomer map[string]int         `json:"items_by_customer"`
	ItemsByGrade    map[string]int         `json:"items_by_grade"`
	ItemsByStatus   map[string]int         `json:"items_by_status"`
	ItemsByLocation map[string]int         `json:"items_by_location"`
	RecentActivity  []WorkflowActivity     `json:"recent_activity"`
	Alerts          []WorkflowAlert        `json:"alerts"`
	LastUpdated     time.Time              `json:"last_updated"`
}

type WorkflowActivity struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CustomerID  int       `json:"customer_id"`
	Customer    string    `json:"customer"`
	Timestamp   time.Time `json:"timestamp"`
}

type WorkflowAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	ItemID      *int      `json:"item_id,omitempty"`
	CustomerID  *int      `json:"customer_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *WorkflowService) GetDashboardStats(ctx context.Context) (*WorkflowStats, error) {
	// Try cache first
	if cached := s.cache.GetWorkflowStats(); cached != nil {
		return cached, nil
	}

	// Use your existing GetSummary method which already has the data we need
	summary, err := s.repos.Inventory.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	// Get recent items for activity and alerts
	recentItemsFilters := repository.InventoryFilters{
	    Page:    1,
	    PerPage: 20,
	}
	recentItems, _, err := s.repos.Inventory.GetFiltered(ctx, recentItemsFilters)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent items: %w", err)
	}

	// Convert summary to workflow stats format
	stats := &WorkflowStats{
		TotalItems:      summary.TotalItems,
		TotalJoints:     summary.TotalJoints,
		ItemsByCustomer: summary.ItemsByCustomer,
		ItemsByGrade:    summary.ItemsByGrade,
		ItemsByStatus:   make(map[string]int),
		ItemsByLocation: summary.ItemsByLocation,
		RecentActivity:  []WorkflowActivity{},
		Alerts:          []WorkflowAlert{},
		LastUpdated:     summary.LastUpdated,
	}

	// Process recent items for status counts and activities
	for _, item := range recentItems {
		// Count by status
		status := determineItemStatus(&item)
		stats.ItemsByStatus[status]++
	}

	// Get recent activity from the summary's recent activity
	for _, item := range summary.RecentActivity {
		status := determineItemStatus(&item)
		activity := WorkflowActivity{
			ID:          item.ID,
			CustomerID:  item.CustomerID,
			Customer:    item.Customer,
			Timestamp:   item.CreatedAt,
		}

		// Set activity type and description based on status
		switch status {
		case "pending":
			activity.Type = "received"
			activity.Description = fmt.Sprintf("Received %d joints of %s %s from %s", 
				item.Joints, item.Size, item.Grade, item.Customer)
		case "in_progress", "threading":
			activity.Type = "started"
			activity.Description = fmt.Sprintf("Started processing %d joints of %s %s", 
				item.Joints, item.Size, item.Grade)
		case "completed":
			activity.Type = "completed"
			activity.Description = fmt.Sprintf("Completed processing %d joints of %s %s", 
				item.Joints, item.Size, item.Grade)
		default:
			activity.Type = "updated"
			activity.Description = fmt.Sprintf("Updated %d joints of %s %s at %s", 
				item.Joints, item.Size, item.Grade, item.Location)
		}

		stats.RecentActivity = append(stats.RecentActivity, activity)
	}

	// Generate alerts based on recent items
	alerts := s.generateAlerts(recentItems)
	stats.Alerts = alerts

	// Cache the results
	s.cache.SetWorkflowStats(stats, 5*time.Minute)

	return stats, nil
}

func (s *WorkflowService) generateAlerts(items []models.InventoryItem) []WorkflowAlert {
	alerts := []WorkflowAlert{}
	now := time.Now()

	for _, item := range items {
		// Check for overdue items
		if item.DateIn != nil && item.DateOut == nil {
			daysSinceStart := int(now.Sub(*item.DateIn).Hours() / 24)
			if daysSinceStart > 7 {
				alerts = append(alerts, WorkflowAlert{
					Type:       "overdue",
					Severity:   "warning",
					Message:    fmt.Sprintf("Work order %s has been in progress for %d days", item.WorkOrder, daysSinceStart),
					ItemID:     &item.ID,
					CustomerID: &item.CustomerID,
					CreatedAt:  now,
				})
			}
		}

		// Check for missing location
		if item.Location == "" {
			alerts = append(alerts, WorkflowAlert{
				Type:       "missing_data",
				Severity:   "info",
				Message:    fmt.Sprintf("Item %d is missing location information", item.ID),
				ItemID:     &item.ID,
				CustomerID: &item.CustomerID,
				CreatedAt:  now,
			})
		}

		// Check for items without rack assignment
		if item.Rack == "" && item.Location != "" {
			alerts = append(alerts, WorkflowAlert{
				Type:       "missing_data",
				Severity:   "info",
				Message:    fmt.Sprintf("Item %d at %s needs rack assignment", item.ID, item.Location),
				ItemID:     &item.ID,
				CustomerID: &item.CustomerID,
				CreatedAt:  now,
			})
		}
	}

	// Check for high-volume customers
	customerCounts := make(map[int]int)
	for _, item := range items {
		customerCounts[item.CustomerID]++
	}

	for customerID, count := range customerCounts {
		if count > 10 { // Adjust threshold as needed
			alerts = append(alerts, WorkflowAlert{
				Type:       "high_volume",
				Severity:   "info",
				Message:    fmt.Sprintf("Customer has %d active items in inventory", count),
				CustomerID: &customerID,
				CreatedAt:  now,
			})
		}
	}

	return alerts
}

func (s *WorkflowService) GetCustomerWorkflow(ctx context.Context, customerID int) (*CustomerWorkflowSummary, error) {
	cacheKey := fmt.Sprintf("customer_workflow_%d", customerID)
	if cached := s.cache.Get(cacheKey); cached != nil {
		if summary, ok := cached.(*CustomerWorkflowSummary); ok {
			return summary, nil
		}
	}

	// Get customer info
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	customerFilters := repository.InventoryFilters{
	    CustomerID: &customerID,
	    Page:       1,
	    PerPage:    100,
	}
	items, pagination, err := s.repos.Inventory.GetFiltered(ctx, customerFilters)
	if err != nil {
	    return nil, fmt.Errorf("failed to get customer items: %w", err)
	}

	// Extract total from pagination
	total := 0
	if pagination != nil {
	    total = pagination.Total
	}

	summary := &CustomerWorkflowSummary{
	    Customer:      *customer,
	    TotalItems:    total,  // Now total is defined
	    ItemsByStatus: make(map[string]int),
	    ItemsByGrade:  make(map[string]int),
	    Items:         items,
	}

	// Calculate status and grade distributions
	for _, item := range items {
		status := determineItemStatus(&item)
		summary.ItemsByStatus[status]++
		summary.ItemsByGrade[item.Grade]++
	}

	s.cache.Set(cacheKey, summary, 10*time.Minute)
	return summary, nil
}

func (s *WorkflowService) SearchInventory(ctx context.Context, query string, limit, offset int) ([]models.InventoryItem, int, error) {
	// Use your existing search method
	return s.repos.Inventory.Search(ctx, query, limit, offset)
}

type CustomerWorkflowSummary struct {
	Customer      models.Customer       `json:"customer"`
	TotalItems    int                   `json:"total_items"`
	ItemsByStatus map[string]int        `json:"items_by_status"`
	ItemsByGrade  map[string]int        `json:"items_by_grade"`
	Items         []models.InventoryItem `json:"items"`
}

// Enhanced status determination based on oil & gas workflow
func determineItemStatus(item *models.InventoryItem) string {
	if item.Deleted {
		return "deleted"
	}
	
	// More detailed status based on oil & gas workflow
	if item.DateOut != nil {
		return "completed"
	}
	
	if item.DateIn != nil {
		// Item is being processed - check specific stages
		if item.Fletcher != "" {
			return "threading" // In fletcher/threading stage
		}
		return "in_progress" // General processing
	}
	
	// Check location-based status
	if item.Location != "" {
		switch item.Location {
		case "SHIPPING", "LOADOUT":
			return "ready_to_ship"
		case "INSPECTION":
			return "inspecting"
		case "REPAIR":
			return "repairing"
		case "THREADING", "FLETCHER":
			return "threading"
		}
	}
	
	// Check if item has been received but not started
	if item.RNumber > 0 {
		return "received" // Has R-number but not started processing
	}
	
	return "pending" // Default status
}

// Cache interface
type CacheService interface {
	GetWorkflowStats() *WorkflowStats
	SetWorkflowStats(stats *WorkflowStats, ttl time.Duration)
	Get(key string) interface{}
	Set(key string, value interface{}, ttl time.Duration)
}

// backend/internal/services/analytics_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

type AnalyticsService interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
	GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error)
	GetInventoryAnalytics(ctx context.Context) (*InventoryAnalytics, error)
	GetGradeDistribution(ctx context.Context) (map[string]int, error)
	GetLocationUtilization(ctx context.Context) (*LocationStats, error)
	GetJobSummaries(ctx context.Context) ([]models.JobSummary, error)
	GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error)
	RefreshCache(ctx context.Context) error
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	customerRepo  repository.CustomerRepository
	inventoryRepo repository.InventoryRepository
	receivedRepo  repository.ReceivedRepository
	cache         *cache.Cache
}

func NewAnalyticsService(
	analyticsRepo repository.AnalyticsRepository,
	customerRepo repository.CustomerRepository,
	inventoryRepo repository.InventoryRepository,
	receivedRepo repository.ReceivedRepository,
	cache *cache.Cache,
) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		customerRepo:  customerRepo,
		inventoryRepo: inventoryRepo,
		receivedRepo:  receivedRepo,
		cache:         cache,
	}
}

func (s *analyticsService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	cacheKey := "dashboard:stats"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if stats, ok := cached.(*models.DashboardStats); ok {
			return stats, nil
		}
	}

	// Get from analytics repository
	stats, err := s.analyticsRepo.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	// Cache for 5 minutes
	s.cache.SetWithTTL(cacheKey, stats, 5*time.Minute)

	return stats, nil
}

func (s *analyticsService) GetCustomerActivity(ctx context.Context, days int, customerID *int) (*models.CustomerActivity, error) {
	cacheKey := fmt.Sprintf("analytics:customer_activity:%d", days)
	if customerID != nil {
		cacheKey = fmt.Sprintf("analytics:customer_activity:%d:%d", days, *customerID)
	}

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if activity, ok := cached.(*models.CustomerActivity); ok {
			return activity, nil
		}
	}

	// Get from repository
	activity, err := s.analyticsRepo.GetCustomerActivity(ctx, days, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer activity: %w", err)
	}

	// Cache for 10 minutes
	s.cache.SetWithTTL(cacheKey, activity, 10*time.Minute)

	return activity, nil
}

func (s *analyticsService) GetInventoryAnalytics(ctx context.Context) (*InventoryAnalytics, error) {
	cacheKey := "analytics:inventory"

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if analytics, ok := cached.(*InventoryAnalytics); ok {
			return analytics, nil
		}
	}

	// Build inventory analytics from multiple sources
	// Get inventory summary
	summary, err := s.inventoryRepo.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	// Get grade analytics
	gradeAnalytics, err := s.analyticsRepo.GetGradeAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade analytics: %w", err)
	}

	// Combine into inventory analytics
	analytics := &InventoryAnalytics{
		TotalItems:        summary.TotalItems,
		TotalJoints:       summary.TotalJoints,
		ItemsByGrade:      summary.ItemsByGrade,
		ItemsByLocation:   summary.ItemsByLocation,
		GradeDistribution: gradeAnalytics.GradeDistribution,
		LastUpdated:       time.Now(),
	}

	// Cache for 15 minutes
	s.cache.SetWithTTL(cacheKey, analytics, 15*time.Minute)

	return analytics, nil
}

func (s *analyticsService) GetGradeDistribution(ctx context.Context) (map[string]int, error) {
	cacheKey := "analytics:grade_distribution"

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if distribution, ok := cached.(map[string]int); ok {
			return distribution, nil
		}
	}

	// Get grade analytics
	gradeAnalytics, err := s.analyticsRepo.GetGradeAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade analytics: %w", err)
	}

	// Extract just the distribution
	distribution := make(map[string]int)
	for grade, stats := range gradeAnalytics.GradeDistribution {
		distribution[grade] = stats.ItemCount
	}

	// Cache for 30 minutes
	s.cache.SetWithTTL(cacheKey, distribution, 30*time.Minute)

	return distribution, nil
}

func (s *analyticsService) GetLocationUtilization(ctx context.Context) (*LocationStats, error) {
	cacheKey := "analytics:location_utilization"

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if stats, ok := cached.(*LocationStats); ok {
			return stats, nil
		}
	}

	// Get inventory summary for location data
	summary, err := s.inventoryRepo.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	// Calculate location utilization
	totalItems := summary.TotalItems
	locationStats := &LocationStats{
		TotalLocations: len(summary.ItemsByLocation),
		LocationUsage:  make(map[string]LocationUsage),
		LastUpdated:    time.Now(),
	}

	for location, itemCount := range summary.ItemsByLocation {
		utilization := 0.0
		if totalItems > 0 {
			utilization = float64(itemCount) / float64(totalItems) * 100
		}

		locationStats.LocationUsage[location] = LocationUsage{
			ItemCount:    itemCount,
			Utilization:  utilization,
			IsOverloaded: itemCount > 100, // Configurable threshold
		}
	}

	// Cache for 20 minutes
	s.cache.SetWithTTL(cacheKey, locationStats, 20*time.Minute)

	return locationStats, nil
}

func (s *analyticsService) GetJobSummaries(ctx context.Context) ([]models.JobSummary, error) {
	cacheKey := "analytics:job_summaries"

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if summaries, ok := cached.([]models.JobSummary); ok {
			return summaries, nil
		}
	}

	// Get from repository
	summaries, err := s.analyticsRepo.GetJobSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job summaries: %w", err)
	}

	// Cache for 5 minutes
	s.cache.SetWithTTL(cacheKey, summaries, 5*time.Minute)

	return summaries, nil
}

func (s *analyticsService) GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error) {
	cacheKey := fmt.Sprintf("analytics:recent_activity:%d", limit)

	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if activity, ok := cached.([]models.ActivityItem); ok {
			return activity, nil
		}
	}

	// Get from repository
	activity, err := s.analyticsRepo.GetRecentActivity(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Cache for 2 minutes (recent activity changes frequently)
	s.cache.SetWithTTL(cacheKey, activity, 2*time.Minute)

	return activity, nil
}

func (s *analyticsService) RefreshCache(ctx context.Context) error {
	// Clear all analytics-related cache entries
	keys := []string{
		"dashboard:stats",
		"analytics:inventory",
		"analytics:grade_distribution",
		"analytics:location_utilization",
		"analytics:job_summaries",
	}

	for _, key := range keys {
		s.cache.Delete(key)
	}

	// Pre-warm cache with fresh data
	_, err := s.GetDashboardStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh dashboard stats: %w", err)
	}

	_, err = s.GetInventoryAnalytics(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh inventory analytics: %w", err)
	}

	return nil
}

// Service-specific types (defined here to avoid model pollution)
type InventoryAnalytics struct {
	TotalItems        int                              `json:"total_items"`
	TotalJoints       int                              `json:"total_joints"`
	ItemsByGrade      map[string]int                   `json:"items_by_grade"`
	ItemsByLocation   map[string]int                   `json:"items_by_location"`
	GradeDistribution map[string]models.GradeStats     `json:"grade_distribution"`
	LastUpdated       time.Time                        `json:"last_updated"`
}

type LocationStats struct {
	TotalLocations int                          `json:"total_locations"`
	LocationUsage  map[string]LocationUsage     `json:"location_usage"`
	LastUpdated    time.Time                    `json:"last_updated"`
}

type LocationUsage struct {
	ItemCount    int     `json:"item_count"`
	Utilization  float64 `json:"utilization_percent"`
	IsOverloaded bool    `json:"is_overloaded"`
}

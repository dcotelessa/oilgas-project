package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

type WorkflowService struct {
	repo  repository.WorkflowRepository
	cache cache.Cache
}

func NewWorkflowService(repo repository.WorkflowRepository, cache cache.Cache) *WorkflowService {
	return &WorkflowService{
		repo:  repo,
		cache: cache,
	}
}

// Dashboard operations with caching
func (s *WorkflowService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	// Try cache first
	if stats, exists := s.cache.GetDashboardStats(); exists {
		return stats, nil
	}

	// Fetch from repository
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	// Cache for 5 minutes
	s.cache.CacheDashboardStats(stats)
	return stats, nil
}

func (s *WorkflowService) GetJobSummaries(ctx context.Context) ([]models.JobSummary, error) {
	// Try cache first
	cacheKey := "job_summaries"
	if cached, exists := s.cache.Get(cacheKey); exists {
		if summaries, ok := cached.([]models.JobSummary); ok {
			return summaries, nil
		}
	}

	summaries, err := s.repo.GetJobSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job summaries: %w", err)
	}

	// Cache for 2 minutes
	s.cache.Set(cacheKey, summaries, 2*time.Minute)
	return summaries, nil
}

func (s *WorkflowService) GetRecentActivity(ctx context.Context, limit int) ([]models.Job, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.GetRecentActivity(ctx, limit)
}

// Job operations
func (s *WorkflowService) GetJobs(ctx context.Context, filters repository.JobFilters) ([]models.Job, *models.Pagination, error) {
	filters.NormalizePagination()
	
	jobs, pagination, err := s.repo.GetJobs(ctx, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	// Set current state for each job
	for i := range jobs {
		jobs[i].CurrentState = jobs[i].GetCurrentState()
	}

	return jobs, pagination, nil
}

func (s *WorkflowService) GetJobByID(ctx context.Context, id int) (*models.Job, error) {
	// Try cache first
	if job, exists := s.cache.GetJob(id); exists {
		return job, nil
	}

	job, err := s.repo.GetJobByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job %d: %w", id, err)
	}

	if job != nil {
		job.CurrentState = job.GetCurrentState()
		s.cache.CacheJob(id, job)
	}

	return job, nil
}

func (s *WorkflowService) GetJobByWorkOrder(ctx context.Context, workOrder string) (*models.Job, error) {
	job, err := s.repo.GetJobByWorkOrder(ctx, workOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get job %s: %w", workOrder, err)
	}

	if job != nil {
		job.CurrentState = job.GetCurrentState()
		s.cache.CacheJob(job.ID, job)
	}

	return job, nil
}

func (s *WorkflowService) CreateJob(ctx context.Context, job *models.Job) error {
	// Validate job data
	if err := s.validateJob(job); err != nil {
		return fmt.Errorf("invalid job data: %w", err)
	}

	// Set initial state
	job.CurrentState = models.StateReceiving
	if job.DateReceived == nil {
		now := time.Now()
		job.DateReceived = &now
	}

	err := s.repo.CreateJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateJobCaches()
	
	log.Printf("Created job %s for customer %s", job.WorkOrder, job.Customer)
	return nil
}

func (s *WorkflowService) UpdateJob(ctx context.Context, job *models.Job) error {
	if err := s.validateJob(job); err != nil {
		return fmt.Errorf("invalid job data: %w", err)
	}

	// Update current state
	job.CurrentState = job.GetCurrentState()
	now := time.Now()
	job.UpdatedAt = &now

	err := s.repo.UpdateJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("job_%d", job.ID))
	s.invalidateJobCaches()

	return nil
}

func (s *WorkflowService) AdvanceJobToProduction(ctx context.Context, workOrder string) error {
	job, err := s.GetJobByWorkOrder(ctx, workOrder)
	if err != nil {
		return err
	}

	if job.GetCurrentState() != models.StateReceiving {
		return fmt.Errorf("job %s is not in RECEIVING state", workOrder)
	}

	now := time.Now()
	job.InProduction = &now
	
	return s.UpdateJob(ctx, job)
}

func (s *WorkflowService) AdvanceJobToInspection(ctx context.Context, workOrder string, inspectedBy string) error {
	job, err := s.GetJobByWorkOrder(ctx, workOrder)
	if err != nil {
		return err
	}

	if job.GetCurrentState() != models.StateProduction {
		return fmt.Errorf("job %s is not in PRODUCTION state", workOrder)
	}

	now := time.Now()
	job.Inspected = &now
	job.InspectedBy = inspectedBy
	job.Complete = true

	return s.UpdateJob(ctx, job)
}

func (s *WorkflowService) MoveJobToInventory(ctx context.Context, workOrder string) error {
	job, err := s.GetJobByWorkOrder(ctx, workOrder)
	if err != nil {
		return err
	}

	if job.GetCurrentState() != models.StateInspection {
		return fmt.Errorf("job %s is not in INSPECTION state", workOrder)
	}

	now := time.Now()
	job.DateIn = &now

	return s.UpdateJob(ctx, job)
}

// Inspection operations
func (s *WorkflowService) GetInspectionResults(ctx context.Context, workOrder string) ([]models.InspectionResult, error) {
	return s.repo.GetInspectionResults(ctx, workOrder)
}

func (s *WorkflowService) CreateInspectionResult(ctx context.Context, result *models.InspectionResult) error {
	if err := s.validateInspectionResult(result); err != nil {
		return fmt.Errorf("invalid inspection result: %w", err)
	}

	err := s.repo.CreateInspectionResult(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to create inspection result: %w", err)
	}

	// Invalidate caches
	s.invalidateJobCaches()
	return nil
}

// Inventory operations
func (s *WorkflowService) GetInventory(ctx context.Context, filters repository.InventoryFilters) ([]models.InventoryItem, *models.Pagination, error) {
	filters.NormalizePagination()
	return s.repo.GetInventory(ctx, filters)
}

func (s *WorkflowService) GetInventoryByCustomer(ctx context.Context, customerID int) ([]models.InventoryItem, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("inventory_customer_%d", customerID)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if items, ok := cached.([]models.InventoryItem); ok {
			return items, nil
		}
	}

	items, err := s.repo.GetInventoryByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory for customer %d: %w", customerID, err)
	}

	// Cache for 5 minutes
	s.cache.Set(cacheKey, items, 5*time.Minute)
	return items, nil
}

func (s *WorkflowService) ShipInventory(ctx context.Context, itemIDs []int, shipmentDetails map[string]interface{}) error {
	if len(itemIDs) == 0 {
		return fmt.Errorf("no items specified for shipping")
	}

	err := s.repo.ShipInventory(ctx, itemIDs, shipmentDetails)
	if err != nil {
		return fmt.Errorf("failed to ship inventory: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateInventoryCaches()
	s.invalidateJobCaches()

	log.Printf("Shipped %d inventory items", len(itemIDs))
	return nil
}

// Customer operations
func (s *WorkflowService) GetCustomers(ctx context.Context, includeDeleted bool) ([]models.Customer, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("customers_deleted_%v", includeDeleted)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if customers, ok := cached.([]models.Customer); ok {
			return customers, nil
		}
	}

	customers, err := s.repo.GetCustomers(ctx, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	// Cache for 10 minutes
	s.cache.Set(cacheKey, customers, 10*time.Minute)
	return customers, nil
}

func (s *WorkflowService) GetCustomerByID(ctx context.Context, id int) (*models.Customer, error) {
	// Try cache first
	if customer, exists := s.cache.GetCustomer(id); exists {
		return customer, nil
	}

	customer, err := s.repo.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer %d: %w", id, err)
	}

	if customer != nil {
		s.cache.CacheCustomer(id, customer)
	}

	return customer, nil
}

func (s *WorkflowService) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("invalid customer data: %w", err)
	}

	err := s.repo.CreateCustomer(ctx, customer)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	// Invalidate customer cache
	s.cache.Delete("customers_deleted_false")
	s.cache.Delete("customers_deleted_true")

	log.Printf("Created customer: %s", customer.Name)
	return nil
}

// Grades
func (s *WorkflowService) GetGrades(ctx context.Context) ([]string, error) {
	// Try cache first
	cacheKey := "grades"
	if cached, exists := s.cache.Get(cacheKey); exists {
		if grades, ok := cached.([]string); ok {
			return grades, nil
		}
	}

	grades, err := s.repo.GetGrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	// Cache for 30 minutes (grades don't change often)
	s.cache.Set(cacheKey, grades, 30*time.Minute)
	return grades, nil
}

// Validation functions
func (s *WorkflowService) validateJob(job *models.Job) error {
	if job.WorkOrder == "" {
		return fmt.Errorf("work order is required")
	}
	if job.CustomerID <= 0 {
		return fmt.Errorf("valid customer ID is required")
	}
	if job.Customer == "" {
		return fmt.Errorf("customer name is required")
	}
	if job.Size == "" {
		return fmt.Errorf("pipe size is required")
	}
	if job.Weight == "" {
		return fmt.Errorf("pipe weight is required")
	}
	if job.Grade == "" {
		return fmt.Errorf("pipe grade is required")
	}
	if job.Connection == "" {
		return fmt.Errorf("pipe connection is required")
	}
	if job.Joints <= 0 {
		return fmt.Errorf("joints count must be positive")
	}
	return nil
}

func (s *WorkflowService) validateInspectionResult(result *models.InspectionResult) error {
	if result.WorkOrder == "" {
		return fmt.Errorf("work order is required")
	}
	if result.Color == "" {
		return fmt.Errorf("color is required")
	}
	if result.Joints < 0 {
		return fmt.Errorf("joints count cannot be negative")
	}
	if result.Accept < 0 {
		return fmt.Errorf("accept count cannot be negative")
	}
	if result.Reject < 0 {
		return fmt.Errorf("reject count cannot be negative")
	}
	if result.Accept+result.Reject > result.Joints {
		return fmt.Errorf("accept + reject cannot exceed total joints")
	}
	return nil
}

func (s *WorkflowService) validateCustomer(customer *models.Customer) error {
	if customer.Name == "" {
		return fmt.Errorf("customer name is required")
	}
	// Email validation if provided
	if customer.Email != "" {
		// Basic email validation
		if len(customer.Email) < 5 || !contains(customer.Email, "@") {
			return fmt.Errorf("invalid email format")
		}
	}
	return nil
}

// Cache invalidation helpers
func (s *WorkflowService) invalidateJobCaches() {
	// Invalidate dashboard and summary caches
	s.cache.Delete("dashboard_stats")
	s.cache.Delete("job_summaries")
	
	// Note: Individual job caches are invalidated when jobs are updated
}

func (s *WorkflowService) invalidateInventoryCaches() {
	// Pattern-based cache invalidation would be ideal here
	// For now, we'll clear specific known patterns
	keys := []string{
		"dashboard_stats",
		"job_summaries",
	}
	
	for _, key := range keys {
		s.cache.Delete(key)
	}
	
	// Could also iterate through cache keys matching pattern "inventory_customer_*"
	// but that depends on cache implementation
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

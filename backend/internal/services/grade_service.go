// backend/internal/services/grade_service.go
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/pkg/validation"
)

type GradeService interface {
	// Basic CRUD operations
	GetAll(ctx context.Context) ([]models.Grade, error)
	GetByName(ctx context.Context, gradeName string) (*models.Grade, error)
	Create(ctx context.Context, req *validation.GradeValidation) (*models.Grade, error)
	Update(ctx context.Context, gradeName string, req *validation.GradeValidation) (*models.Grade, error)
	Delete(ctx context.Context, gradeName string) error

	// Business logic operations
	IsInUse(ctx context.Context, gradeName string) (bool, error)
	GetUsageStats(ctx context.Context, gradeName string) (*models.GradeUsage, error)
	GetGradeDistribution(ctx context.Context) (map[string]int, error)
	
	// Validation and normalization
	ValidateGrade(ctx context.Context, gradeName string) error
	NormalizeGradeName(gradeName string) string
	GetSimilarGrades(ctx context.Context, gradeName string) ([]string, error)
	
	// Analytics and reporting
	GetPopularGrades(ctx context.Context, limit int) ([]GradePopularity, error)
	GetGradeMetrics(ctx context.Context) (*GradeMetrics, error)
	GetUnusedGrades(ctx context.Context) ([]models.Grade, error)
	
	// Cache management
	RefreshCache(ctx context.Context) error
}

type gradeService struct {
	gradeRepo     repository.GradeRepository
	inventoryRepo repository.InventoryRepository
	receivedRepo  repository.ReceivedRepository
	cache         *cache.Cache
}

func NewGradeService(
	gradeRepo repository.GradeRepository,
	cache *cache.Cache,
) GradeService {
	return &gradeService{
		gradeRepo: gradeRepo,
		cache:     cache,
	}
}

// Basic CRUD operations

func (s *gradeService) GetAll(ctx context.Context) ([]models.Grade, error) {
	cacheKey := "grades:all"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if grades, ok := cached.([]models.Grade); ok {
			return grades, nil
		}
	}

	// Get from repository
	gradeNames, err := s.gradeRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	// Convert to Grade models
	grades := make([]models.Grade, len(gradeNames))
	for i, gradeName := range gradeNames {
		grades[i] = models.Grade{Grade: gradeName}
	}

	// Cache for 30 minutes (grades don't change often)
	s.cache.SetWithTTL(cacheKey, grades, 30*time.Minute)

	return grades, nil
}

func (s *gradeService) GetByName(ctx context.Context, gradeName string) (*models.Grade, error) {
	// Normalize the grade name
	normalizedName := s.NormalizeGradeName(gradeName)
	cacheKey := fmt.Sprintf("grade:%s", normalizedName)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if grade, ok := cached.(*models.Grade); ok {
			return grade, nil
		}
	}

	// Get all grades and check if this one exists
	grades, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	for _, grade := range grades {
		if strings.EqualFold(grade.Grade, normalizedName) {
			foundGrade := &models.Grade{Grade: grade.Grade}
			
			// Cache for 30 minutes
			s.cache.SetWithTTL(cacheKey, foundGrade, 30*time.Minute)
			
			return foundGrade, nil
		}
	}

	return nil, fmt.Errorf("grade %s not found", gradeName)
}

func (s *gradeService) Create(ctx context.Context, req *validation.GradeValidation) (*models.Grade, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Normalize grade name
	normalizedName := s.NormalizeGradeName(req.Grade)

	// Check if grade already exists
	existing, _ := s.GetByName(ctx, normalizedName)
	if existing != nil {
		return nil, fmt.Errorf("grade %s already exists", normalizedName)
	}

	// Create in repository
	if err := s.gradeRepo.Create(ctx, &models.Grade{
		Grade:       normalizedName,
		Description: req.Description,
	}); err != nil {
		return nil, fmt.Errorf("failed to create grade: %w", err)
	}

	// Create the response model
	grade := &models.Grade{
		Grade:       normalizedName,
		Description: req.Description,
	}

	// Invalidate cache
	s.cache.Delete("grades:all")
	s.cache.Delete(fmt.Sprintf("grade:%s", normalizedName))

	return grade, nil
}

func (s *gradeService) Update(ctx context.Context, gradeName string, req *validation.GradeValidation) (*models.Grade, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Normalize names
	normalizedName := s.NormalizeGradeName(gradeName)
	newNormalizedName := s.NormalizeGradeName(req.Grade)

	// Check if grade exists
	existing, err := s.GetByName(ctx, normalizedName)
	if err != nil {
		return nil, fmt.Errorf("grade %s not found: %w", gradeName, err)
	}

	// If name is changing, check new name doesn't exist
	if normalizedName != newNormalizedName {
		if existingNew, _ := s.GetByName(ctx, newNormalizedName); existingNew != nil {
			return nil, fmt.Errorf("grade %s already exists", newNormalizedName)
		}
	}

	// Check if grade is in use before allowing changes
	inUse, err := s.IsInUse(ctx, normalizedName)
	if err != nil {
		return nil, fmt.Errorf("failed to check grade usage: %w", err)
	}

	if inUse && normalizedName != newNormalizedName {
		return nil, fmt.Errorf("cannot rename grade %s because it is in use", normalizedName)
	}

	// Update the grade (implementation depends on your repository interface)
	updatedGrade := &models.Grade{
		Grade:       newNormalizedName,
		Description: req.Description,
	}

	// For now, we'll delete and recreate since the repository interface might not support update
	if normalizedName != newNormalizedName {
		if err := s.gradeRepo.Delete(ctx, normalizedName); err != nil {
			return nil, fmt.Errorf("failed to delete old grade: %w", err)
		}
		
		if err := s.gradeRepo.Create(ctx, updatedGrade); err != nil {
			return nil, fmt.Errorf("failed to create updated grade: %w", err)
		}
	}

	// Invalidate caches
	s.cache.Delete("grades:all")
	s.cache.Delete(fmt.Sprintf("grade:%s", normalizedName))
	s.cache.Delete(fmt.Sprintf("grade:%s", newNormalizedName))

	return updatedGrade, nil
}

func (s *gradeService) Delete(ctx context.Context, gradeName string) error {
	// Normalize grade name
	normalizedName := s.NormalizeGradeName(gradeName)

	// Check if grade exists
	_, err := s.GetByName(ctx, normalizedName)
	if err != nil {
		return fmt.Errorf("grade %s not found: %w", gradeName, err)
	}

	// Check if grade is in use
	inUse, err := s.IsInUse(ctx, normalizedName)
	if err != nil {
		return fmt.Errorf("failed to check grade usage: %w", err)
	}

	if inUse {
		return fmt.Errorf("cannot delete grade %s because it is in use", normalizedName)
	}

	// Delete from repository
	if err := s.gradeRepo.Delete(ctx, normalizedName); err != nil {
		return fmt.Errorf("failed to delete grade: %w", err)
	}

	// Invalidate caches
	s.cache.Delete("grades:all")
	s.cache.Delete(fmt.Sprintf("grade:%s", normalizedName))

	return nil
}

// Business logic operations

func (s *gradeService) IsInUse(ctx context.Context, gradeName string) (bool, error) {
	// Normalize grade name
	normalizedName := s.NormalizeGradeName(gradeName)
	cacheKey := fmt.Sprintf("grade:in_use:%s", normalizedName)
	
	// Try cache first (short TTL since usage can change)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if inUse, ok := cached.(bool); ok {
			return inUse, nil
		}
	}

	// Check with repository
	inUse, err := s.gradeRepo.IsInUse(ctx, normalizedName)
	if err != nil {
		return false, fmt.Errorf("failed to check if grade is in use: %w", err)
	}

	// Cache for 2 minutes
	s.cache.SetWithTTL(cacheKey, inUse, 2*time.Minute)

	return inUse, nil
}

func (s *gradeService) GetUsageStats(ctx context.Context, gradeName string) (*models.GradeUsage, error) {
	// Normalize grade name
	normalizedName := s.NormalizeGradeName(gradeName)
	cacheKey := fmt.Sprintf("grade:usage_stats:%s", normalizedName)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if stats, ok := cached.(*models.GradeUsage); ok {
			return stats, nil
		}
	}

	// Get from repository
	stats, err := s.gradeRepo.GetUsageStats(ctx, normalizedName)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// Cache for 10 minutes
	s.cache.SetWithTTL(cacheKey, stats, 10*time.Minute)

	return stats, nil
}

func (s *gradeService) GetGradeDistribution(ctx context.Context) (map[string]int, error) {
	cacheKey := "grades:distribution"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if distribution, ok := cached.(map[string]int); ok {
			return distribution, nil
		}
	}

	// Get all grades and their usage
	grades, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	distribution := make(map[string]int)
	for _, grade := range grades {
		stats, err := s.GetUsageStats(ctx, grade.Grade)
		if err != nil {
			// If we can't get stats, assume 0 usage
			distribution[grade.Grade] = 0
		} else {
			distribution[grade.Grade] = stats.InventoryCount + stats.ReceivedCount
		}
	}

	// Cache for 15 minutes
	s.cache.SetWithTTL(cacheKey, distribution, 15*time.Minute)

	return distribution, nil
}

// Validation and normalization

func (s *gradeService) ValidateGrade(ctx context.Context, gradeName string) error {
	// Normalize the grade name
	normalizedName := s.NormalizeGradeName(gradeName)

	// Check if it's a valid grade
	if err := validation.ValidateGrade(normalizedName); err != nil {
		return fmt.Errorf("invalid grade: %w", err)
	}

	// Check if grade exists in our system
	_, err := s.GetByName(ctx, normalizedName)
	if err != nil {
		return fmt.Errorf("grade %s not found in system: %w", gradeName, err)
	}

	return nil
}

func (s *gradeService) NormalizeGradeName(gradeName string) string {
	// Use the validation package's normalization
	return validation.NormalizeGrade(gradeName)
}

func (s *gradeService) GetSimilarGrades(ctx context.Context, gradeName string) ([]string, error) {
	// Get all grades
	grades, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	normalizedInput := strings.ToUpper(strings.TrimSpace(gradeName))
	var similar []string

	for _, grade := range grades {
		normalizedGrade := strings.ToUpper(strings.TrimSpace(grade.Grade))
		
		// Check for similar patterns
		if strings.Contains(normalizedGrade, normalizedInput) ||
			strings.Contains(normalizedInput, normalizedGrade) ||
			s.isGradeSimilar(normalizedInput, normalizedGrade) {
			similar = append(similar, grade.Grade)
		}
	}

	return similar, nil
}

// Analytics and reporting

func (s *gradeService) GetPopularGrades(ctx context.Context, limit int) ([]GradePopularity, error) {
	cacheKey := fmt.Sprintf("grades:popular:%d", limit)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if popular, ok := cached.([]GradePopularity); ok {
			return popular, nil
		}
	}

	// Get grade distribution
	distribution, err := s.GetGradeDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade distribution: %w", err)
	}

	// Convert to popularity slice and sort
	var popular []GradePopularity
	for grade, count := range distribution {
		popular = append(popular, GradePopularity{
			Grade: grade,
			Count: count,
		})
	}

	// Sort by count descending
	for i := 0; i < len(popular)-1; i++ {
		for j := i + 1; j < len(popular); j++ {
			if popular[j].Count > popular[i].Count {
				popular[i], popular[j] = popular[j], popular[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(popular) > limit {
		popular = popular[:limit]
	}

	// Cache for 20 minutes
	s.cache.SetWithTTL(cacheKey, popular, 20*time.Minute)

	return popular, nil
}

func (s *gradeService) GetGradeMetrics(ctx context.Context) (*GradeMetrics, error) {
	cacheKey := "grades:metrics"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if metrics, ok := cached.(*GradeMetrics); ok {
			return metrics, nil
		}
	}

	// Get all grades
	grades, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	// Get distribution
	distribution, err := s.GetGradeDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade distribution: %w", err)
	}

	// Calculate metrics
	totalGrades := len(grades)
	totalUsage := 0
	usedGrades := 0

	for _, count := range distribution {
		totalUsage += count
		if count > 0 {
			usedGrades++
		}
	}

	metrics := &GradeMetrics{
		TotalGrades:    totalGrades,
		UsedGrades:     usedGrades,
		UnusedGrades:   totalGrades - usedGrades,
		TotalUsage:     totalUsage,
		AverageUsage:   float64(totalUsage) / float64(totalGrades),
		LastUpdated:    time.Now(),
	}

	// Cache for 30 minutes
	s.cache.SetWithTTL(cacheKey, metrics, 30*time.Minute)

	return metrics, nil
}

func (s *gradeService) GetUnusedGrades(ctx context.Context) ([]models.Grade, error) {
	cacheKey := "grades:unused"
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if unused, ok := cached.([]models.Grade); ok {
			return unused, nil
		}
	}

	// Get all grades
	grades, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	// Check usage for each grade
	var unused []models.Grade
	for _, grade := range grades {
		inUse, err := s.IsInUse(ctx, grade.Grade)
		if err != nil {
			continue // Skip if we can't check usage
		}
		
		if !inUse {
			unused = append(unused, grade)
		}
	}

	// Cache for 15 minutes
	s.cache.SetWithTTL(cacheKey, unused, 15*time.Minute)

	return unused, nil
}

// Cache management

func (s *gradeService) RefreshCache(ctx context.Context) error {
	// Clear all grade-related caches
	patterns := []string{
		"grades:",
		"grade:",
	}

	for _, pattern := range patterns {
		s.cache.DeletePattern(pattern)
	}

	// Pre-warm cache with fresh data
	_, err := s.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh grades cache: %w", err)
	}

	_, err = s.GetGradeDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh grade distribution cache: %w", err)
	}

	return nil
}

// Helper methods

func (s *gradeService) isGradeSimilar(grade1, grade2 string) bool {
	// Simple similarity check based on common patterns
	// You could implement more sophisticated similarity algorithms here
	
	// Check if they share common prefixes/suffixes
	if len(grade1) >= 2 && len(grade2) >= 2 {
		if grade1[:2] == grade2[:2] { // Same first two characters
			return true
		}
	}
	
	// Check for common grade patterns (J55/L80/etc.)
	commonPrefixes := []string{"J", "L", "N", "P", "Q", "S", "T"}
	for _, prefix := range commonPrefixes {
		if strings.HasPrefix(grade1, prefix) && strings.HasPrefix(grade2, prefix) {
			return true
		}
	}
	
	return false
}

// Type definitions

type GradePopularity struct {
	Grade string `json:"grade"`
	Count int    `json:"count"`
}

type GradeMetrics struct {
	TotalGrades   int       `json:"total_grades"`
	UsedGrades    int       `json:"used_grades"`
	UnusedGrades  int       `json:"unused_grades"`
	TotalUsage    int       `json:"total_usage"`
	AverageUsage  float64   `json:"average_usage"`
	LastUpdated   time.Time `json:"last_updated"`
}

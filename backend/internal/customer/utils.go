// backend/internal/customer/utils.go
package customer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationUtils provides common validation utilities
type ValidationUtils struct{}

// NormalizeCompanyCode ensures company codes follow consistent format
func (v *ValidationUtils) NormalizeCompanyCode(code string) string {
	code = strings.TrimSpace(strings.ToUpper(code))
	// Remove any non-alphanumeric characters
	reg := regexp.MustCompile(`[^A-Z0-9]`)
	return reg.ReplaceAllString(code, "")
}

// ValidateEmail performs basic email validation
func (v *ValidationUtils) ValidateEmail(email string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return reg.MatchString(email)
}

// ValidateTaxID validates tax ID format (basic US EIN format)
func (v *ValidationUtils) ValidateTaxID(taxID string) bool {
	if taxID == "" {
		return true // Optional field
	}
	// US EIN format: XX-XXXXXXX or XXXXXXXXX
	reg := regexp.MustCompile(`^\d{2}-?\d{7}$`)
	return reg.MatchString(taxID)
}

// FormatTaxID standardizes tax ID format
func (v *ValidationUtils) FormatTaxID(taxID string) string {
	taxID = strings.ReplaceAll(taxID, "-", "")
	if len(taxID) == 9 {
		return taxID[:2] + "-" + taxID[2:]
	}
	return taxID
}

// SearchOptimizer provides search optimization utilities
type SearchOptimizer struct{}

// OptimizeSearchTerms cleans and optimizes search terms
func (s *SearchOptimizer) OptimizeSearchTerms(term string) string {
	term = strings.TrimSpace(term)
	if term == "" {
		return term
	}
	
	// Remove excessive whitespace
	reg := regexp.MustCompile(`\s+`)
	term = reg.ReplaceAllString(term, " ")
	
	// Remove special characters that might interfere with SQL LIKE
	reg = regexp.MustCompile(`[%_\\]`)
	term = reg.ReplaceAllString(term, "")
	
	return term
}

// BuildSearchQuery creates optimized search queries
func (s *SearchOptimizer) BuildSearchQuery(filters SearchFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, "placeholder-tenant") // Will be replaced
	argIndex++
	
	if filters.Name != "" {
		optimized := s.OptimizeSearchTerms(filters.Name)
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR company_code ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, "%"+optimized+"%", "%"+optimized+"%")
		argIndex += 2
	}
	
	if filters.CompanyCode != "" {
		conditions = append(conditions, fmt.Sprintf("company_code ILIKE $%d", argIndex))
		args = append(args, "%"+filters.CompanyCode+"%")
		argIndex++
	}
	
	if len(filters.Status) > 0 {
		placeholders := make([]string, len(filters.Status))
		for i, status := range filters.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}
	
	return strings.Join(conditions, " AND "), args
}

// PerformanceTracker tracks operation performance
type PerformanceTracker struct {
	operation string
	startTime time.Time
}

func NewPerformanceTracker(operation string) *PerformanceTracker {
	return &PerformanceTracker{
		operation: operation,
		startTime: time.Now(),
	}
}

func (p *PerformanceTracker) Duration() time.Duration {
	return time.Since(p.startTime)
}

func (p *PerformanceTracker) LogIfSlow(threshold time.Duration) {
	duration := p.Duration()
	if duration > threshold {
		// In production, use proper logging
		fmt.Printf("SLOW OPERATION: %s took %v (threshold: %v)\n", p.operation, duration, threshold)
	}
}

// CustomerMetrics provides business metrics utilities
type CustomerMetrics struct{}

// CalculateCustomerHealth returns a health score based on activity
func (c *CustomerMetrics) CalculateCustomerHealth(analytics *CustomerAnalytics) string {
	if analytics.TotalWorkOrders == 0 {
		return "new"
	}
	
	if analytics.ActiveOrders > 0 && analytics.TotalRevenue > 10000 {
		return "excellent"
	}
	
	if analytics.ActiveOrders > 0 {
		return "good"
	}
	
	// Check if last order was recent
	if analytics.LastOrderDate != nil {
		daysSinceLastOrder := time.Since(*analytics.LastOrderDate).Hours() / 24
		if daysSinceLastOrder < 30 {
			return "good"
		} else if daysSinceLastOrder < 90 {
			return "fair"
		}
	}
	
	return "at_risk"
}

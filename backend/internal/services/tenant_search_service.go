// backend/internal/services/tenant_search_service.go
// Cross-table search service with tenant isolation
package services

import (
	"context"
	"fmt"
	"strings"

	"oilgas-backend/internal/repository"
	"oilgas-backend/internal/models"
)

type TenantSearchService struct {
	customerRepo  repository.TenantCustomerRepository
	inventoryRepo repository.TenantInventoryRepository
}

func NewTenantSearchService(customerRepo repository.TenantCustomerRepository, inventoryRepo repository.TenantInventoryRepository) *TenantSearchService {
	return &TenantSearchService{
		customerRepo:  customerRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *TenantSearchService) GlobalSearchForTenant(ctx context.Context, tenantID, query string, limit int) (*models.SearchResults, error) {
	if err := s.validateTenantID(tenantID); err != nil {
		return nil, err
	}
	
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}
	
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	
	// Search across different entity types
	customerResults, err := s.customerRepo.SearchForTenant(ctx, tenantID, query)
	if err != nil {
		customerResults = []models.Customer{} // Continue with empty results
	}
	
	inventoryResults, err := s.inventoryRepo.SearchForTenant(ctx, tenantID, query)
	if err != nil {
		inventoryResults = []models.InventoryItem{} // Continue with empty results
	}
	
	workOrderResults, err := s.inventoryRepo.SearchWorkOrdersForTenant(ctx, tenantID, query)
	if err != nil {
		workOrderResults = []models.WorkOrder{} // Continue with empty results
	}
	
	// Convert to unified search results
	results := &models.SearchResults{
		TenantID: tenantID,
		Query:    query,
		Summary: models.SearchSummary{
			Customers:  len(customerResults),
			Inventory:  len(inventoryResults),
			WorkOrders: len(workOrderResults),
		},
	}
	
	// Add customer results
	for _, customer := range customerResults {
		if len(results.Results) >= limit {
			break
		}
		results.Results = append(results.Results, models.SearchResult{
			Type:   "customer",
			ID:     customer.CustomerID,
			Title:  customer.Customer,
			Detail: s.buildCustomerDetail(customer),
			Data:   customer,
		})
	}
	
	// Add inventory results
	for _, item := range inventoryResults {
		if len(results.Results) >= limit {
			break
		}
		results.Results = append(results.Results, models.SearchResult{
			Type:   "inventory",
			ID:     item.ID,
			Title:  s.buildInventoryTitle(item),
			Detail: s.buildInventoryDetail(item),
			Data:   item,
		})
	}
	
	// Add work order results
	for _, workOrder := range workOrderResults {
		if len(results.Results) >= limit {
			break
		}
		results.Results = append(results.Results, models.SearchResult{
			Type:   "work_order",
			ID:     *workOrder.WorkOrder,
			Title:  fmt.Sprintf("WO: %s", *workOrder.WorkOrder),
			Detail: s.buildWorkOrderDetail(workOrder),
			Data:   workOrder,
		})
	}
	
	results.Summary.Total = len(results.Results)
	
	return results, nil
}

func (s *TenantSearchService) buildCustomerDetail(customer models.Customer) string {
	var details []string
	
	if customer.Contact != nil && *customer.Contact != "" {
		details = append(details, "Contact: "+*customer.Contact)
	}
	
	if customer.BillingCity != nil && customer.BillingState != nil {
		if *customer.BillingCity != "" && *customer.BillingState != "" {
			details = append(details, fmt.Sprintf("Location: %s, %s", *customer.BillingCity, *customer.BillingState))
		}
	}
	
	if customer.Phone != nil && *customer.Phone != "" {
		details = append(details, "Phone: "+*customer.Phone)
	}
	
	if len(details) == 0 {
		return "Oil & Gas Customer"
	}
	
	return strings.Join(details, " • ")
}

func (s *TenantSearchService) buildInventoryTitle(item models.InventoryItem) string {
	if item.WorkOrder != nil && *item.WorkOrder != "" {
		return fmt.Sprintf("WO: %s", *item.WorkOrder)
	}
	return fmt.Sprintf("Inventory #%d", item.ID)
}

func (s *TenantSearchService) buildInventoryDetail(item models.InventoryItem) string {
	var details []string
	
	if item.Customer != nil {
		details = append(details, *item.Customer)
	}
	
	if item.Joints != nil && *item.Joints > 0 {
		details = append(details, fmt.Sprintf("%d joints", *item.Joints))
	}
	
	if item.Size != nil && item.Grade != nil && *item.Size != "" && *item.Grade != "" {
		details = append(details, fmt.Sprintf("%s %s", *item.Size, *item.Grade))
	}
	
	if item.Location != nil && *item.Location != "" {
		details = append(details, fmt.Sprintf("@ %s", *item.Location))
	}
	
	return strings.Join(details, " • ")
}

func (s *TenantSearchService) buildWorkOrderDetail(workOrder models.WorkOrder) string {
	var details []string
	
	if workOrder.Customer != nil {
		details = append(details, *workOrder.Customer)
	}
	
	if workOrder.ItemCount > 0 {
		details = append(details, fmt.Sprintf("%d items", workOrder.ItemCount))
	}
	
	if workOrder.TotalJoints > 0 {
		details = append(details, fmt.Sprintf("%d joints", workOrder.TotalJoints))
	}
	
	return strings.Join(details, " • ")
}

func (s *TenantSearchService) validateTenantID(tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if len(tenantID) < 2 || len(tenantID) > 20 {
		return fmt.Errorf("tenant ID must be between 2 and 20 characters")
	}
	return nil
}

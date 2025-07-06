// backend/internal/services/search_service.go
package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"oilgas-backend/internal/models"
	"oilgas-backend/internal/repository"
	"oilgas-backend/pkg/cache"
)

type SearchService interface {
	// Global search across all entities
	GlobalSearch(ctx context.Context, query string, limit, offset int) (*SearchResults, error)
	
	// Entity-specific searches
	SearchCustomers(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error)
	SearchInventory(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error)
	SearchReceived(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error)
	
	// Advanced search features
	SearchByWorkOrder(ctx context.Context, workOrder string) (*WorkOrderSearchResult, error)
	SearchByCustomer(ctx context.Context, customerQuery string) (*CustomerSearchResult, error)
	SearchBySpecs(ctx context.Context, size, grade, connection string) (*SpecSearchResult, error)
	
	// Search suggestions and autocomplete
	GetSuggestions(ctx context.Context, query string, entityType string) ([]SearchSuggestion, error)
	GetRecentSearches(ctx context.Context, userID string) ([]string, error)
	SaveRecentSearch(ctx context.Context, userID, query string) error
	
	// Search analytics
	GetPopularSearches(ctx context.Context, days int) ([]PopularSearch, error)
	GetSearchMetrics(ctx context.Context) (*SearchMetrics, error)
}

type searchService struct {
	customerRepo  repository.CustomerRepository
	inventoryRepo repository.InventoryRepository
	receivedRepo  repository.ReceivedRepository
	gradeRepo     repository.GradeRepository
	cache         *cache.Cache
}

func NewSearchService(
	customerRepo repository.CustomerRepository,
	inventoryRepo repository.InventoryRepository,
	receivedRepo repository.ReceivedRepository,
	gradeRepo repository.GradeRepository,
	cache *cache.Cache,
) SearchService {
	return &searchService{
		customerRepo:  customerRepo,
		inventoryRepo: inventoryRepo,
		receivedRepo:  receivedRepo,
		gradeRepo:     gradeRepo,
		cache:         cache,
	}
}

// Global search across all entities

func (s *searchService) GlobalSearch(ctx context.Context, query string, limit, offset int) (*SearchResults, error) {
	cacheKey := fmt.Sprintf("search:global:%s:%d:%d", query, limit, offset)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if results, ok := cached.(*SearchResults); ok {
			return results, nil
		}
	}

	// Perform searches across all entities in parallel
	results := &SearchResults{
		Query:       query,
		TotalHits:   0,
		Results:     []SearchResult{},
		SearchTime:  time.Now(),
	}

	// Search customers
	customers, customerTotal, err := s.SearchCustomers(ctx, query, limit/3, 0)
	if err == nil {
		for _, customer := range customers {
			results.Results = append(results.Results, SearchResult{
				Type:        "customer",
				ID:          customer.CustomerID,
				Title:       customer.Customer,
				Description: fmt.Sprintf("%s, %s %s", customer.BillingCity, customer.BillingState, customer.BillingZipcode),
				Data:        customer,
				Relevance:   s.calculateCustomerRelevance(customer, query),
			})
		}
		results.TotalHits += customerTotal
	}

	// Search inventory
	inventoryFilters := make(map[string]interface{})
	inventory, inventoryTotal, err := s.SearchInventory(ctx, query, inventoryFilters, limit/3, 0)
	if err == nil {
		for _, item := range inventory {
			results.Results = append(results.Results, SearchResult{
				Type:        "inventory",
				ID:          item.ID,
				Title:       fmt.Sprintf("WO-%s: %s", item.WorkOrder, item.Customer),
				Description: fmt.Sprintf("%d joints of %s %s %s", item.Joints, item.Size, item.Weight, item.Grade),
				Data:        item,
				Relevance:   s.calculateInventoryRelevance(item, query),
			})
		}
		results.TotalHits += inventoryTotal
	}

	// Search received items
	receivedFilters := make(map[string]interface{})
	received, receivedTotal, err := s.SearchReceived(ctx, query, receivedFilters, limit/3, 0)
	if err == nil {
		for _, item := range received {
			results.Results = append(results.Results, SearchResult{
				Type:        "received",
				ID:          item.ID,
				Title:       fmt.Sprintf("WO-%s: %s", item.WorkOrder, item.Customer),
				Description: fmt.Sprintf("%d joints of %s %s %s - %s", item.Joints, item.Size, item.Weight, item.Grade, item.CurrentState),
				Data:        item,
				Relevance:   s.calculateReceivedRelevance(item, query),
			})
		}
		results.TotalHits += receivedTotal
	}

	// Sort results by relevance
	sort.Slice(results.Results, func(i, j int) bool {
		return results.Results[i].Relevance > results.Results[j].Relevance
	})

	// Apply limit and offset
	if offset >= len(results.Results) {
		results.Results = []SearchResult{}
	} else {
		end := offset + limit
		if end > len(results.Results) {
			end = len(results.Results)
		}
		results.Results = results.Results[offset:end]
	}

	// Cache for 2 minutes
	s.cache.SetWithTTL(cacheKey, results, 2*time.Minute)

	return results, nil
}

// Entity-specific searches

func (s *searchService) SearchCustomers(ctx context.Context, query string, limit, offset int) ([]models.Customer, int, error) {
	cacheKey := fmt.Sprintf("search:customers:%s:%d:%d", query, limit, offset)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedCustomerSearchResult); ok {
			return result.Customers, result.Total, nil
		}
	}

	// Search using repository
	customers, total, err := s.customerRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	// Cache result
	result := CachedCustomerSearchResult{Customers: customers, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 5*time.Minute)

	return customers, total, nil
}

func (s *searchService) SearchInventory(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]models.InventoryItem, int, error) {
	cacheKey := fmt.Sprintf("search:inventory:%s:%v:%d:%d", query, filters, limit, offset)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedInventorySearchResult); ok {
			return result.Items, result.Total, nil
		}
	}

	// Use repository search method
	items, total, err := s.inventoryRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search inventory: %w", err)
	}

	// Apply additional filters if provided
	if len(filters) > 0 {
		items = s.applyInventoryFilters(items, filters)
		total = len(items)
	}

	// Cache result
	result := CachedInventorySearchResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 3*time.Minute)

	return items, total, nil
}

func (s *searchService) SearchReceived(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]models.ReceivedItem, int, error) {
	cacheKey := fmt.Sprintf("search:received:%s:%v:%d:%d", query, filters, limit, offset)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(CachedReceivedSearchResult); ok {
			return result.Items, result.Total, nil
		}
	}

	// Build repository filters
	repoFilters := s.buildReceivedFilters(query, filters)
	repoFilters.Page = offset/limit + 1
	repoFilters.PerPage = limit

	// Search using repository
	items, pagination, err := s.receivedRepo.GetFiltered(ctx, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search received items: %w", err)
	}

	total := 0
	if pagination != nil {
		total = pagination.Total
	}

	// Set current state for each item
	for i := range items {
		items[i].CurrentState = items[i].GetCurrentState()
	}

	// Cache result
	result := CachedReceivedSearchResult{Items: items, Total: total}
	s.cache.SetWithTTL(cacheKey, result, 3*time.Minute)

	return items, total, nil
}

// Advanced search features

func (s *searchService) SearchByWorkOrder(ctx context.Context, workOrder string) (*WorkOrderSearchResult, error) {
	cacheKey := fmt.Sprintf("search:workorder:%s", workOrder)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*WorkOrderSearchResult); ok {
			return result, nil
		}
	}

	result := &WorkOrderSearchResult{
		WorkOrder: workOrder,
		Found:     false,
	}

	// Search in received items first
	receivedFilters := repository.ReceivedFilters{
		WorkOrder: &workOrder,
		Page:      1,
		PerPage:   1,
	}

	receivedItems, _, err := s.receivedRepo.GetFiltered(ctx, receivedFilters)
	if err == nil && len(receivedItems) > 0 {
		result.Found = true
		result.ReceivedItem = &receivedItems[0]
		result.ReceivedItem.CurrentState = result.ReceivedItem.GetCurrentState()
	}

	// Search in inventory
	inventoryFilters := repository.InventoryFilters{
		WorkOrder: &workOrder,
		Page:      1,
		PerPage:   10,
	}

	inventoryItems, _, err := s.inventoryRepo.GetFiltered(ctx, inventoryFilters)
	if err == nil && len(inventoryItems) > 0 {
		result.Found = true
		result.InventoryItems = inventoryItems
	}

	// Cache result
	s.cache.SetWithTTL(cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *searchService) SearchByCustomer(ctx context.Context, customerQuery string) (*CustomerSearchResult, error) {
	cacheKey := fmt.Sprintf("search:customer_detailed:%s", customerQuery)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*CustomerSearchResult); ok {
			return result, nil
		}
	}

	result := &CustomerSearchResult{
		Query: customerQuery,
	}

	// Find matching customers
	customers, _, err := s.SearchCustomers(ctx, customerQuery, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	if len(customers) == 0 {
		return result, nil
	}

	// Take the best match
	customer := customers[0]
	result.Customer = &customer

	// Get customer's inventory
	inventoryFilters := repository.InventoryFilters{
		CustomerID: &customer.CustomerID,
		Page:       1,
		PerPage:    50,
	}

	inventory, _, err := s.inventoryRepo.GetFiltered(ctx, inventoryFilters)
	if err == nil {
		result.InventoryItems = inventory
	}

	// Get customer's received items
	receivedFilters := repository.ReceivedFilters{
		CustomerID: &customer.CustomerID,
		Page:       1,
		PerPage:    50,
	}

	received, _, err := s.receivedRepo.GetFiltered(ctx, receivedFilters)
	if err == nil {
		for i := range received {
			received[i].CurrentState = received[i].GetCurrentState()
		}
		result.ReceivedItems = received
	}

	// Cache result
	s.cache.SetWithTTL(cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *searchService) SearchBySpecs(ctx context.Context, size, grade, connection string) (*SpecSearchResult, error) {
	cacheKey := fmt.Sprintf("search:specs:%s:%s:%s", size, grade, connection)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if result, ok := cached.(*SpecSearchResult); ok {
			return result, nil
		}
	}

	result := &SpecSearchResult{
		Size:       size,
		Grade:      grade,
		Connection: connection,
	}

	// Build filters for inventory search
	inventoryFilters := repository.InventoryFilters{
		Page:    1,
		PerPage: 100,
	}

	if size != "" {
		inventoryFilters.Size = &size
	}
	if grade != "" {
		inventoryFilters.Grade = &grade
	}
	if connection != "" {
		inventoryFilters.Connection = &connection
	}

	// Search inventory
	inventory, _, err := s.inventoryRepo.GetFiltered(ctx, inventoryFilters)
	if err == nil {
		result.InventoryItems = inventory
	}

	// Build filters for received search
	receivedFilters := repository.ReceivedFilters{
		Page:    1,
		PerPage: 100,
	}

	if size != "" {
		receivedFilters.Size = &size
	}
	if grade != "" {
		receivedFilters.Grade = &grade
	}
	if connection != "" {
		receivedFilters.Connection = &connection
	}

	// Search received items
	received, _, err := s.receivedRepo.GetFiltered(ctx, receivedFilters)
	if err == nil {
		for i := range received {
			received[i].CurrentState = received[i].GetCurrentState()
		}
		result.ReceivedItems = received
	}

	// Calculate totals
	result.TotalInventory = len(result.InventoryItems)
	result.TotalReceived = len(result.ReceivedItems)

	// Cache result
	s.cache.SetWithTTL(cacheKey, result, 10*time.Minute)

	return result, nil
}

// Search suggestions and autocomplete

func (s *searchService) GetSuggestions(ctx context.Context, query string, entityType string) ([]SearchSuggestion, error) {
	cacheKey := fmt.Sprintf("search:suggestions:%s:%s", query, entityType)
	
	// Try cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if suggestions, ok := cached.([]SearchSuggestion); ok {
			return suggestions, nil
		}
	}

	var suggestions []SearchSuggestion

	switch entityType {
	case "customer":
		suggestions = s.getCustomerSuggestions(ctx, query)
	case "grade":
		suggestions = s.getGradeSuggestions(ctx, query)
	case "size":
		suggestions = s.getSizeSuggestions(ctx, query)
	case "connection":
		suggestions = s.getConnectionSuggestions(ctx, query)
	case "workorder":
		suggestions = s.getWorkOrderSuggestions(ctx, query)
	default:
		// Global suggestions
		suggestions = append(suggestions, s.getCustomerSuggestions(ctx, query)...)
		suggestions = append(suggestions, s.getGradeSuggestions(ctx, query)...)
		suggestions = append(suggestions, s.getSizeSuggestions(ctx, query)...)
	}

	// Limit to top 10 suggestions
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	// Cache for 30 minutes
	s.cache.SetWithTTL(cacheKey, suggestions, 30*time.Minute)

	return suggestions, nil
}

func (s *searchService) GetRecentSearches(ctx context.Context, userID string) ([]string, error) {
	cacheKey := fmt.Sprintf("search:recent:%s", userID)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if searches, ok := cached.([]string); ok {
			return searches, nil
		}
	}

	// Return empty slice for now - in production, you'd store this in database
	return []string{}, nil
}

func (s *searchService) SaveRecentSearch(ctx context.Context, userID, query string) error {
	cacheKey := fmt.Sprintf("search:recent:%s", userID)
	
	// Get existing searches
	recentSearches, _ := s.GetRecentSearches(ctx, userID)
	
	// Add new search to beginning, remove duplicates
	newSearches := []string{query}
	for _, search := range recentSearches {
		if search != query && len(newSearches) < 10 {
			newSearches = append(newSearches, search)
		}
	}

	// Cache for 7 days
	s.cache.SetWithTTL(cacheKey, newSearches, 7*24*time.Hour)

	return nil
}

// Search analytics

func (s *searchService) GetPopularSearches(ctx context.Context, days int) ([]PopularSearch, error) {
	cacheKey := fmt.Sprintf("search:popular:%d", days)
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if searches, ok := cached.([]PopularSearch); ok {
			return searches, nil
		}
	}

	// Mock data for now - in production, you'd analyze search logs
	popular := []PopularSearch{
		{Query: "J55", Count: 45, Category: "grade"},
		{Query: "5 1/2\"", Count: 38, Category: "size"},
		{Query: "LTC", Count: 32, Category: "connection"},
		{Query: "Oil Company", Count: 28, Category: "customer"},
		{Query: "L80", Count: 25, Category: "grade"},
	}

	// Cache for 1 hour
	s.cache.SetWithTTL(cacheKey, popular, time.Hour)

	return popular, nil
}

func (s *searchService) GetSearchMetrics(ctx context.Context) (*SearchMetrics, error) {
	cacheKey := "search:metrics"
	
	if cached, exists := s.cache.Get(cacheKey); exists {
		if metrics, ok := cached.(*SearchMetrics); ok {
			return metrics, nil
		}
	}

	// Mock metrics for now
	metrics := &SearchMetrics{
		TotalSearches:     1250,
		UniqueQueries:     485,
		AverageResultTime: 120, // milliseconds
		TopCategories: map[string]int{
			"customer":   450,
			"inventory":  380,
			"received":   290,
			"workorder":  130,
		},
		LastUpdated: time.Now(),
	}

	// Cache for 1 hour
	s.cache.SetWithTTL(cacheKey, metrics, time.Hour)

	return metrics, nil
}

// Helper methods

func (s *searchService) calculateCustomerRelevance(customer models.Customer, query string) float64 {
	score := 0.0
	query = strings.ToLower(query)
	customerName := strings.ToLower(customer.Customer)

	// Exact match gets highest score
	if customerName == query {
		score = 1.0
	} else if strings.HasPrefix(customerName, query) {
		score = 0.8
	} else if strings.Contains(customerName, query) {
		score = 0.6
	} else {
		// Check other fields
		if strings.Contains(strings.ToLower(customer.BillingCity), query) {
			score = 0.4
		} else if strings.Contains(strings.ToLower(customer.BillingState), query) {
			score = 0.3
		}
	}

	return score
}

func (s *searchService) calculateInventoryRelevance(item models.InventoryItem, query string) float64 {
	score := 0.0
	query = strings.ToLower(query)

	// Check work order
	if strings.Contains(strings.ToLower(item.WorkOrder), query) {
		score += 0.4
	}

	// Check customer name
	if strings.Contains(strings.ToLower(item.Customer), query) {
		score += 0.3
	}

	// Check specifications
	if strings.Contains(strings.ToLower(item.Size), query) ||
		strings.Contains(strings.ToLower(item.Grade), query) ||
		strings.Contains(strings.ToLower(item.Connection), query) {
		score += 0.3
	}

	return score
}

func (s *searchService) calculateReceivedRelevance(item models.ReceivedItem, query string) float64 {
	score := 0.0
	query = strings.ToLower(query)

	// Check work order
	if strings.Contains(strings.ToLower(item.WorkOrder), query) {
		score += 0.4
	}

	// Check customer name
	if strings.Contains(strings.ToLower(item.Customer), query) {
		score += 0.3
	}

	// Check specifications
	if strings.Contains(strings.ToLower(item.Size), query) ||
		strings.Contains(strings.ToLower(item.Grade), query) ||
		strings.Contains(strings.ToLower(item.Connection), query) {
		score += 0.3
	}

	return score
}

func (s *searchService) applyInventoryFilters(items []models.InventoryItem, filters map[string]interface{}) []models.InventoryItem {
	var filtered []models.InventoryItem

	for _, item := range items {
		include := true

		if customerID, ok := filters["customer_id"]; ok {
			if id, ok := customerID.(int); ok && item.CustomerID != id {
				include = false
			}
		}

		if grade, ok := filters["grade"]; ok {
			if g, ok := grade.(string); ok && !strings.EqualFold(item.Grade, g) {
				include = false
			}
		}

		if size, ok := filters["size"]; ok {
			if s, ok := size.(string); ok && !strings.EqualFold(item.Size, s) {
				include = false
			}
		}

		if include {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (s *searchService) buildReceivedFilters(query string, filters map[string]interface{}) repository.ReceivedFilters {
	repoFilters := repository.ReceivedFilters{
		Page:    1,
		PerPage: 50,
	}

	// If query looks like a work order, search by work order
	if strings.HasPrefix(strings.ToUpper(query), "WO-") || 
		strings.HasPrefix(strings.ToUpper(query), "WK-") {
		workOrder := strings.TrimPrefix(strings.ToUpper(query), "WO-")
		workOrder = strings.TrimPrefix(workOrder, "WK-")
		repoFilters.WorkOrder = &workOrder
	}

	// Apply additional filters
	if customerID, ok := filters["customer_id"]; ok {
		if id, ok := customerID.(int); ok {
			repoFilters.CustomerID = &id
		}
	}

	if status, ok := filters["status"]; ok {
		if s, ok := status.(string); ok {
			repoFilters.Status = &s
		}
	}

	return repoFilters
}

func (s *searchService) getCustomerSuggestions(ctx context.Context, query string) []SearchSuggestion {
	customers, _, err := s.SearchCustomers(ctx, query, 5, 0)
	if err != nil {
		return []SearchSuggestion{}
	}

	var suggestions []SearchSuggestion
	for _, customer := range customers {
		suggestions = append(suggestions, SearchSuggestion{
			Text:        customer.Customer,
			Type:        "customer",
			Description: fmt.Sprintf("Customer in %s, %s", customer.BillingCity, customer.BillingState),
		})
	}

	return suggestions
}

func (s *searchService) getGradeSuggestions(ctx context.Context, query string) []SearchSuggestion {
	grades, err := s.gradeRepo.GetAll(ctx)
	if err != nil {
		return []SearchSuggestion{}
	}

	var suggestions []SearchSuggestion
	query = strings.ToLower(query)

	for _, grade := range grades {
		if strings.Contains(strings.ToLower(grade.Grade), query) {
			suggestions = append(suggestions, SearchSuggestion{
				Text:        grade.Grade,
				Type:        "grade",
				Description: "Pipe grade",
			})
		}
	}

	return suggestions
}

func (s *searchService) getSizeSuggestions(ctx context.Context, query string) []SearchSuggestion {
	// Common pipe sizes
	commonSizes := []string{
		"4 1/2\"", "5\"", "5 1/2\"", "7\"", "7 5/8\"", "8 5/8\"", "9 5/8\"",
		"10 3/4\"", "11 3/4\"", "13 3/8\"", "16\"", "18 5/8\"", "20\"",
	}

	var suggestions []SearchSuggestion
	query = strings.ToLower(query)

	for _, size := range commonSizes {
		if strings.Contains(strings.ToLower(size), query) {
			suggestions = append(suggestions, SearchSuggestion{
				Text:        size,
				Type:        "size",
				Description: "Pipe size",
			})
		}
	}

	return suggestions
}

func (s *searchService) getConnectionSuggestions(ctx context.Context, query string) []SearchSuggestion {
	// Common connection types
	commonConnections := []string{
		"LTC", "BTC", "EUE", "PREMIUM", "VAM", "NEW VAM", "HYDRIL",
	}

	var suggestions []SearchSuggestion
	query = strings.ToLower(query)

	for _, connection := range commonConnections {
		if strings.Contains(strings.ToLower(connection), query) {
			suggestions = append(suggestions, SearchSuggestion{
				Text:        connection,
				Type:        "connection",
				Description: "Connection type",
			})
		}
	}

	return suggestions
}

func (s *searchService) getWorkOrderSuggestions(ctx context.Context, query string) []SearchSuggestion {
	// This would typically search recent work orders
	// For now, return empty suggestions
	return []SearchSuggestion{}
}

// Type definitions

type SearchResults struct {
	Query      string         `json:"query"`
	TotalHits  int            `json:"total_hits"`
	Results    []SearchResult `json:"results"`
	SearchTime time.Time      `json:"search_time"`
}

type SearchResult struct {
	Type        string      `json:"type"`
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Relevance   float64     `json:"relevance"`
}

type WorkOrderSearchResult struct {
	WorkOrder      string                  `json:"work_order"`
	Found          bool                    `json:"found"`
	ReceivedItem   *models.ReceivedItem    `json:"received_item,omitempty"`
	InventoryItems []models.InventoryItem  `json:"inventory_items,omitempty"`
}

type CustomerSearchResult struct {
	Query          string                  `json:"query"`
	Customer       *models.Customer        `json:"customer,omitempty"`
	InventoryItems []models.InventoryItem  `json:"inventory_items,omitempty"`
	ReceivedItems  []models.ReceivedItem   `json:"received_items,omitempty"`
}

type SpecSearchResult struct {
	Size           string                  `json:"size"`
	Grade          string                  `json:"grade"`
	Connection     string                  `json:"connection"`
	InventoryItems []models.InventoryItem  `json:"inventory_items"`
	ReceivedItems  []models.ReceivedItem   `json:"received_items"`
	TotalInventory int                     `json:"total_inventory"`
	TotalReceived  int                     `json:"total_received"`
}

type SearchSuggestion struct {
	Text        string `json:"text"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type PopularSearch struct {
	Query    string `json:"query"`
	Count    int    `json:"count"`
	Category string `json:"category"`
}

type SearchMetrics struct {
	TotalSearches     int            `json:"total_searches"`
	UniqueQueries     int            `json:"unique_queries"`
	AverageResultTime int            `json:"average_result_time_ms"`
	TopCategories     map[string]int `json:"top_categories"`
	LastUpdated       time.Time      `json:"last_updated"`
}

// Cached result types
type CachedCustomerSearchResult struct {
	Customers []models.Customer `json:"customers"`
	Total     int               `json:"total"`
}

type CachedInventorySearchResult struct {
	Items []models.InventoryItem `json:"items"`
	Total int                    `json:"total"`
}

type CachedReceivedSearchResult struct {
	Items []models.ReceivedItem `json:"items"`
	Total int                   `json:"total"`
}

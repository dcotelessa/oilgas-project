// backend/internal/models/search_models.go
package models

import "time"

// SearchResult represents search results across multiple tables
type SearchResult struct {
	Type        string      `json:"type"`        // "inventory", "customer", "received"
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Relevance   float64     `json:"relevance"`
	MatchedOn   []string    `json:"matched_on"`   // Fields that matched the search
	CustomerID  int         `json:"customer_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// GlobalSearchResults represents results from searching across multiple domains
type GlobalSearchResults struct {
	Query           string              `json:"query"`
	TotalResults    int                 `json:"total_results"`
	Customers       []CustomerResult    `json:"customers"`
	Inventory       []InventoryResult   `json:"inventory"`
	Received        []ReceivedResult    `json:"received"`
	ExecutionTimeMs int                 `json:"execution_time_ms"`
	SearchedAt      time.Time           `json:"searched_at"`
}

// Specific result types for better type safety
type CustomerResult struct {
	CustomerID int       `json:"customer_id"`
	Name       string    `json:"name"`
	Contact    string    `json:"contact"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Relevance  float64   `json:"relevance"`
	MatchedOn  []string  `json:"matched_on"`
	CreatedAt  time.Time `json:"created_at"`
}

type InventoryResult struct {
	ID          int       `json:"id"`
	WorkOrder   string    `json:"work_order"`
	Customer    string    `json:"customer"`
	CustomerID  int       `json:"customer_id"`
	Joints      int       `json:"joints"`
	Size        string    `json:"size"`
	Grade       string    `json:"grade"`
	Connection  string    `json:"connection"`
	Location    string    `json:"location"`
	Rack        string    `json:"rack"`
	CustomerPO  string    `json:"customer_po"`
	Relevance   float64   `json:"relevance"`
	MatchedOn   []string  `json:"matched_on"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReceivedResult struct {
	ID           int       `json:"id"`
	WorkOrder    string    `json:"work_order"`
	Customer     string    `json:"customer"`
	CustomerID   int       `json:"customer_id"`
	Joints       int       `json:"joints"`
	Size         string    `json:"size"`
	Grade        string    `json:"grade"`
	CurrentState string    `json:"current_state"`
	Well         string    `json:"well"`
	Lease        string    `json:"lease"`
	Relevance    float64   `json:"relevance"`
	MatchedOn    []string  `json:"matched_on"`
	CreatedAt    time.Time `json:"created_at"`
}

// SearchSuggestion represents autocomplete suggestions
type SearchSuggestion struct {
	Type        string  `json:"type"`        // "customer", "grade", "size", "location"
	Value       string  `json:"value"`       // The suggested value
	Label       string  `json:"label"`       // Display label
	Count       int     `json:"count"`       // How many times this appears
	Category    string  `json:"category"`    // Grouping category
	Relevance   float64 `json:"relevance"`
}

// SearchFilters represents advanced search filters
type SearchFilters struct {
	Query          string    `json:"query"`
	Type           string    `json:"type"`            // "all", "customers", "inventory", "received"
	CustomerID     *int      `json:"customer_id,omitempty"`
	Grade          string    `json:"grade,omitempty"`
	Size           string    `json:"size,omitempty"`
	DateFrom       *time.Time `json:"date_from,omitempty"`
	DateTo         *time.Time `json:"date_to,omitempty"`
	MinJoints      *int      `json:"min_joints,omitempty"`
	MaxJoints      *int      `json:"max_joints,omitempty"`
	Location       string    `json:"location,omitempty"`
	WorkflowState  string    `json:"workflow_state,omitempty"`
	
	// Pagination
	Page           int       `json:"page"`
	PerPage        int       `json:"per_page"`
	SortBy         string    `json:"sort_by"`         // "relevance", "date", "customer", "joints"
	SortOrder      string    `json:"sort_order"`      // "asc", "desc"
}

// SearchMetrics represents search performance metrics
type SearchMetrics struct {
	QueryExecutionTime time.Duration `json:"query_execution_time"`
	TotalQueries       int           `json:"total_queries"`
	CacheHits          int           `json:"cache_hits"`
	CacheMisses        int           `json:"cache_misses"`
	IndexesUsed        []string      `json:"indexes_used"`
	ResultCount        int           `json:"result_count"`
}

// Default values for search
const (
	DefaultSearchLimit     = 20
	MaxSearchLimit         = 100
	DefaultSearchSortBy    = "relevance"
	DefaultSearchSortOrder = "desc"
)

// Normalize sets default values for search filters
func (sf *SearchFilters) Normalize() {
	if sf.Page < 1 {
		sf.Page = 1
	}
	if sf.PerPage < 1 {
		sf.PerPage = DefaultSearchLimit
	}
	if sf.PerPage > MaxSearchLimit {
		sf.PerPage = MaxSearchLimit
	}
	if sf.SortBy == "" {
		sf.SortBy = DefaultSearchSortBy
	}
	if sf.SortOrder != "asc" && sf.SortOrder != "desc" {
		sf.SortOrder = DefaultSearchSortOrder
	}
	if sf.Type == "" {
		sf.Type = "all"
	}
}

// GetOffset calculates the offset for pagination
func (sf *SearchFilters) GetOffset() int {
	return (sf.Page - 1) * sf.PerPage
}

// HasDateRange returns true if date filters are set
func (sf *SearchFilters) HasDateRange() bool {
	return sf.DateFrom != nil || sf.DateTo != nil
}

// HasJointsRange returns true if joints filters are set
func (sf *SearchFilters) HasJointsRange() bool {
	return sf.MinJoints != nil || sf.MaxJoints != nil
}

// CreateSearchResult creates a SearchResult from any domain object
func CreateSearchResult(resultType string, id int, title, description string, data interface{}, relevance float64, matchedOn []string) SearchResult {
	return SearchResult{
		Type:        resultType,
		ID:          id,
		Title:       title,
		Description: description,
		Data:        data,
		Relevance:   relevance,
		MatchedOn:   matchedOn,
		CreatedAt:   time.Now(),
	}
}

// CreateCustomerResult creates a CustomerResult from Customer model
func CreateCustomerResult(customer Customer, relevance float64, matchedOn []string) CustomerResult {
	return CustomerResult{
		CustomerID: customer.CustomerID,
		Name:       customer.Customer,
		Contact:    customer.Contact,
		Phone:      customer.Phone,
		Email:      customer.Email,
		City:       customer.BillingCity,
		State:      customer.BillingState,
		Relevance:  relevance,
		MatchedOn:  matchedOn,
		CreatedAt:  customer.CreatedAt,
	}
}

// CreateInventoryResult creates an InventoryResult from InventoryItem model
func CreateInventoryResult(item InventoryItem, relevance float64, matchedOn []string) InventoryResult {
	return InventoryResult{
		ID:         item.ID,
		WorkOrder:  item.WorkOrder,
		Customer:   item.Customer,
		CustomerID: item.CustomerID,
		Joints:     item.Joints,
		Size:       item.Size,
		Grade:      item.Grade,
		Connection: item.Connection,
		Location:   item.Location,
		Rack:       item.Rack,
		CustomerPO: item.CustomerPO,
		Relevance:  relevance,
		MatchedOn:  matchedOn,
		CreatedAt:  item.CreatedAt,
	}
}

// CreateReceivedResult creates a ReceivedResult from ReceivedItem model
func CreateReceivedResult(item ReceivedItem, relevance float64, matchedOn []string) ReceivedResult {
	return ReceivedResult{
		ID:           item.ID,
		WorkOrder:    item.WorkOrder,
		Customer:     item.Customer,
		CustomerID:   item.CustomerID,
		Joints:       item.Joints,
		Size:         item.Size,
		Grade:        item.Grade,
		CurrentState: string(item.GetCurrentState()),
		Well:         item.Well,
		Lease:        item.Lease,
		Relevance:    relevance,
		MatchedOn:    matchedOn,
		CreatedAt:    item.CreatedAt,
	}
}

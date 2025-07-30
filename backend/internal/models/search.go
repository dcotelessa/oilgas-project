// backend/internal/models/search.go
package models

import "time"

// SearchResult represents a unified search result
type SearchResult struct {
	Type   string      `json:"type"`
	ID     interface{} `json:"id"`
	Title  string      `json:"title"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"`
}

// SearchSummary provides search result counts by type
type SearchSummary struct {
	Customers  int `json:"customers"`
	Inventory  int `json:"inventory"`
	WorkOrders int `json:"work_orders"`
	Received   int `json:"received"`
	Total      int `json:"total"`
}

// SearchResults contains comprehensive search response
type SearchResults struct {
	TenantID string         `json:"tenant_id"`
	Query    string         `json:"query"`
	Results  []SearchResult `json:"results"`
	Summary  SearchSummary  `json:"summary"`
}

// Filter types for repository queries
type CustomerFilters struct {
	Search     string     `json:"search"`
	Deleted    *bool      `json:"deleted"`
	State      string     `json:"state"`
	City       string     `json:"city"`
	DateFrom   *time.Time `json:"date_from"`
	DateTo     *time.Time `json:"date_to"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type InventoryFilters struct {
	CustomerID *int       `json:"customer_id"`
	WorkOrder  string     `json:"work_order"`
	Size       *string    `json:"size"`
	Grade      *string    `json:"grade"`
	Location   *string    `json:"location"`
	DateFrom   *time.Time `json:"date_from"`
	DateTo     *time.Time `json:"date_to"`
	Available  *bool      `json:"available"`
	Search     string     `json:"search"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type WorkOrderFilters struct {
	CustomerID *int       `json:"customer_id"`
	Search     string     `json:"search"`
	Status     string     `json:"status"`
	DateFrom   *time.Time `json:"date_from"`
	DateTo     *time.Time `json:"date_to"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type ReceivedFilters struct {
	CustomerID   *int       `json:"customer_id"`
	WorkOrder    string     `json:"work_order"`
	InProduction *bool      `json:"in_production"`
	Complete     *bool      `json:"complete"`
	DateFrom     *time.Time `json:"date_from"`
	DateTo       *time.Time `json:"date_to"`
	Search       string     `json:"search"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

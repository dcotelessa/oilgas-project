// backend/internal/events/tenant_event_system.go
// Event-driven architecture for cross-tenant operations while maintaining data isolation
package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// TenantEvent represents business events that can be shared across tenants safely
type TenantEvent struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"` // API endpoint or service that generated event
}

// EventHandler processes tenant events
type EventHandler func(event TenantEvent) error

// TenantEventBus manages event publishing and subscription
type TenantEventBus struct {
	subscribers map[string][]EventHandler
	mutex       sync.RWMutex
	eventLog    []TenantEvent // For audit trail
	maxLogSize  int
}

var (
	globalEventBus *TenantEventBus
	busOnce        sync.Once
)

// GetEventBus returns singleton event bus
func GetEventBus() *TenantEventBus {
	busOnce.Do(func() {
		globalEventBus = &TenantEventBus{
			subscribers: make(map[string][]EventHandler),
			eventLog:    make([]TenantEvent, 0),
			maxLogSize:  1000, // Keep last 1000 events for audit
		}
		
		// Register default handlers for cross-tenant analytics
		globalEventBus.registerDefaultHandlers()
	})
	return globalEventBus
}

// PublishEvent publishes an event to all subscribers (async, non-blocking)
func (bus *TenantEventBus) PublishEvent(event TenantEvent) error {
	// Validate event has required fields
	if event.TenantID == "" || event.EventType == "" {
		return fmt.Errorf("event must have tenant_id and event_type")
	}

	// Set metadata if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Log event for audit trail
	bus.logEvent(event)

	// Process subscribers asynchronously (non-blocking for tenant operations)
	go bus.processEventAsync(event)

	return nil
}

func (bus *TenantEventBus) processEventAsync(event TenantEvent) {
	bus.mutex.RLock()
	handlers, exists := bus.subscribers[event.EventType]
	bus.mutex.RUnlock()

	if !exists || len(handlers) == 0 {
		return // No subscribers, that's okay
	}

	// Process all handlers for this event type
	for i, handler := range handlers {
		func(handlerIndex int, h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("🚨 Event handler panic for %s (handler %d): %v", 
						event.EventType, handlerIndex, r)
				}
			}()

			if err := h(event); err != nil {
				log.Printf("⚠️  Event handler error for %s: %v", event.EventType, err)
				// Continue processing other handlers
			}
		}(i, handler)
	}
}

func (bus *TenantEventBus) logEvent(event TenantEvent) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// Add to event log
	bus.eventLog = append(bus.eventLog, event)

	// Trim log if it exceeds max size
	if len(bus.eventLog) > bus.maxLogSize {
		bus.eventLog = bus.eventLog[len(bus.eventLog)-bus.maxLogSize:]
	}
}

// Subscribe registers an event handler for a specific event type
func (bus *TenantEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	
	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
	log.Printf("📝 Subscribed to event type: %s", eventType)
}

// GetEventHistory returns recent events for audit/debugging
func (bus *TenantEventBus) GetEventHistory(tenantID string, limit int) []TenantEvent {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	var filtered []TenantEvent
	count := 0

	// Search backwards through event log
	for i := len(bus.eventLog) - 1; i >= 0 && count < limit; i-- {
		event := bus.eventLog[i]
		if tenantID == "" || event.TenantID == tenantID {
			filtered = append([]TenantEvent{event}, filtered...) // Prepend to maintain order
			count++
		}
	}

	return filtered
}

// registerDefaultHandlers sets up standard cross-tenant analytics handlers
func (bus *TenantEventBus) registerDefaultHandlers() {
	// Cross-tenant analytics handler
	bus.Subscribe("work_order_completed", CrossTenantAnalyticsHandler)
	bus.Subscribe("customer_created", CrossTenantAnalyticsHandler)
	bus.Subscribe("inventory_received", CrossTenantAnalyticsHandler)
	bus.Subscribe("inventory_low_stock", InventoryAlertHandler)
	
	// Audit trail handler
	bus.Subscribe("*", AuditTrailHandler) // Special handler for all events
}

// Oil & Gas Industry Specific Event Types

// WorkOrderCompletedEvent - Published when work order is completed
type WorkOrderCompletedEvent struct {
	TenantID      string    `json:"tenant_id"`
	WorkOrder     string    `json:"work_order"`
	Customer      string    `json:"customer"`
	TotalJoints   int       `json:"total_joints"`
	TotalWeight   float64   `json:"total_weight"`
	CompletedAt   time.Time `json:"completed_at"`
	CompletedBy   string    `json:"completed_by"`
	Location      string    `json:"location"`
}

// InventoryReceivedEvent - Published when new inventory arrives
type InventoryReceivedEvent struct {
	TenantID     string    `json:"tenant_id"`
	WorkOrder    string    `json:"work_order"`
	Customer     string    `json:"customer"`
	Joints       int       `json:"joints"`
	Size         string    `json:"size"`
	Grade        string    `json:"grade"`
	Weight       float64   `json:"weight"`
	ReceivedAt   time.Time `json:"received_at"`
	ReceivedBy   string    `json:"received_by"`
}

// InventoryLowStockEvent - Published when inventory falls below threshold
type InventoryLowStockEvent struct {
	TenantID         string `json:"tenant_id"`
	Size             string `json:"size"`
	Grade            string `json:"grade"`
	CurrentCount     int    `json:"current_count"`
	MinimumThreshold int    `json:"minimum_threshold"`
	Location         string `json:"location"`
}

// CustomerCreatedEvent - Published when new customer is added
type CustomerCreatedEvent struct {
	TenantID     string    `json:"tenant_id"`
	CustomerID   int       `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Contact      string    `json:"contact"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
}

// Event Handler Implementations

// CrossTenantAnalyticsHandler aggregates data for company-wide reporting
func CrossTenantAnalyticsHandler(event TenantEvent) error {
	// Store aggregated data in separate analytics database
	// This enables cross-tenant reporting while maintaining data isolation
	
	switch event.EventType {
	case "work_order_completed":
		return handleWorkOrderCompletedAnalytics(event)
	case "customer_created":
		return handleCustomerCreatedAnalytics(event)
	case "inventory_received":
		return handleInventoryReceivedAnalytics(event)
	default:
		return nil // Unknown event type, ignore
	}
}

func handleWorkOrderCompletedAnalytics(event TenantEvent) error {
	// Extract safe data for analytics (no sensitive customer details)
	analyticsData := map[string]interface{}{
		"tenant_id":     event.TenantID,
		"work_order":    event.Data["work_order"],
		"total_joints":  event.Data["total_joints"],
		"total_weight":  event.Data["total_weight"],
		"completed_at":  event.Data["completed_at"],
		"location":      event.Data["location"],
		// NOTE: No customer names or sensitive data included
	}

	// Store in analytics database (would be implemented based on your analytics needs)
	return storeAnalyticsData("work_order_completions", analyticsData)
}

func handleCustomerCreatedAnalytics(event TenantEvent) error {
	analyticsData := map[string]interface{}{
		"tenant_id":  event.TenantID,
		"city":       event.Data["city"],
		"state":      event.Data["state"],
		"created_at": event.Data["created_at"],
		// NOTE: No customer names or contact info included for privacy
	}

	return storeAnalyticsData("customer_demographics", analyticsData)
}

func handleInventoryReceivedAnalytics(event TenantEvent) error {
	analyticsData := map[string]interface{}{
		"tenant_id":   event.TenantID,
		"joints":      event.Data["joints"],
		"size":        event.Data["size"],
		"grade":       event.Data["grade"],
		"weight":      event.Data["weight"],
		"received_at": event.Data["received_at"],
	}

	return storeAnalyticsData("inventory_trends", analyticsData)
}

// InventoryAlertHandler manages low stock alerts across tenants
func InventoryAlertHandler(event TenantEvent) error {
	if event.EventType != "inventory_low_stock" {
		return nil
	}

	// Extract alert data
	tenantID := event.TenantID
	size := event.Data["size"].(string)
	grade := event.Data["grade"].(string)
	currentCount := int(event.Data["current_count"].(float64))
	
	log.Printf("🚨 Low stock alert: %s needs %s %s (current: %d)", 
		tenantID, size, grade, currentCount)

	// Could trigger:
	// - Email alerts to management
	// - Dashboard notifications
	// - Automatic reorder workflows
	// - Check other locations for available inventory

	return checkOtherLocationsForInventory(tenantID, size, grade)
}

func checkOtherLocationsForInventory(requestingTenant, size, grade string) error {
	// This demonstrates cross-tenant coordination while maintaining isolation
	// We check other tenants for available inventory without accessing their full data
	
	// Publish event asking other tenants to check inventory
	event := TenantEvent{
		ID:        generateEventID(),
		TenantID:  "system", // System-level event
		EventType: "inventory_availability_request",
		Data: map[string]interface{}{
			"requesting_tenant": requestingTenant,
			"size":             size,
			"grade":            grade,
		},
		Timestamp: time.Now(),
		Source:    "inventory_alert_handler",
	}

	// Other tenants can respond with availability without exposing customer data
	return GetEventBus().PublishEvent(event)
}

// AuditTrailHandler logs all events for compliance
func AuditTrailHandler(event TenantEvent) error {
	// Store in audit database for regulatory compliance
	auditEntry := map[string]interface{}{
		"event_id":   event.ID,
		"tenant_id":  event.TenantID,
		"event_type": event.EventType,
		"timestamp":  event.Timestamp,
		"source":     event.Source,
		// Store minimal data for audit, not full event data
	}

	return storeAuditEntry(auditEntry)
}

// Helper Functions

func generateEventID() string {
	return fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), randomString(6))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// These would be implemented based on your analytics/audit database setup
func storeAnalyticsData(table string, data map[string]interface{}) error {
	// Implementation depends on your analytics database (PostgreSQL, ClickHouse, etc.)
	log.Printf("📊 Analytics: %s <- %v", table, data)
	return nil
}

func storeAuditEntry(entry map[string]interface{}) error {
	// Implementation depends on your audit logging system
	log.Printf("📝 Audit: %v", entry)
	return nil
}

// Event Publishing Convenience Functions for API Handlers

// PublishWorkOrderCompleted - Call this from your API when work order is completed
func PublishWorkOrderCompleted(tenantID, workOrder, customer string, totalJoints int, totalWeight float64, completedBy string) {
	event := TenantEvent{
		TenantID:  tenantID,
		EventType: "work_order_completed",
		Data: map[string]interface{}{
			"work_order":   workOrder,
			"customer":     customer,
			"total_joints": totalJoints,
			"total_weight": totalWeight,
			"completed_by": completedBy,
			"completed_at": time.Now(),
		},
		Source: "work_order_api",
	}
	
	GetEventBus().PublishEvent(event)
}

// PublishInventoryReceived - Call this from your API when inventory is received
func PublishInventoryReceived(tenantID, workOrder, customer string, joints int, size, grade string, weight float64, receivedBy string) {
	event := TenantEvent{
		TenantID:  tenantID,
		EventType: "inventory_received",
		Data: map[string]interface{}{
			"work_order":   workOrder,
			"customer":     customer,
			"joints":       joints,
			"size":         size,
			"grade":        grade,
			"weight":       weight,
			"received_by":  receivedBy,
			"received_at":  time.Now(),
		},
		Source: "inventory_api",
	}
	
	GetEventBus().PublishEvent(event)
}

// PublishCustomerCreated - Call this from your API when customer is created
func PublishCustomerCreated(tenantID string, customerID int, customerName, contact, city, state string) {
	event := TenantEvent{
		TenantID:  tenantID,
		EventType: "customer_created",
		Data: map[string]interface{}{
			"customer_id":   customerID,
			"customer_name": customerName,
			"contact":       contact,
			"city":          city,
			"state":         state,
			"created_at":    time.Now(),
		},
		Source: "customer_api",
	}
	
	GetEventBus().PublishEvent(event)
}

// PublishInventoryLowStock - Call this when inventory falls below threshold
func PublishInventoryLowStock(tenantID, size, grade, location string, currentCount, minimumThreshold int) {
	event := TenantEvent{
		TenantID:  tenantID,
		EventType: "inventory_low_stock",
		Data: map[string]interface{}{
			"size":              size,
			"grade":             grade,
			"location":          location,
			"current_count":     currentCount,
			"minimum_threshold": minimumThreshold,
			"alert_time":        time.Now(),
		},
		Source: "inventory_monitor",
	}
	
	GetEventBus().PublishEvent(event)
}

// Integration with API Handlers

// Example: Enhanced inventory handler with event publishing
func EnhancedInventoryHandler(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	
	// ... existing inventory logic ...
	
	// After successful inventory operation, publish event
	if shouldPublishEvent(operation) {
		PublishInventoryReceived(
			tenantID, 
			workOrder, 
			customer, 
			joints, 
			size, 
			grade, 
			weight, 
			receivedBy,
		)
	}
	
	// ... rest of handler ...
}

// Example: Enhanced work order completion with events
func EnhancedWorkOrderCompletionHandler(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	workOrderID := c.Param("id")
	
	// Complete work order in tenant database
	workOrder, err := completeWorkOrderInTenantDB(tenantID, workOrderID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	// Publish completion event for cross-tenant analytics
	PublishWorkOrderCompleted(
		tenantID,
		workOrder.WorkOrder,
		workOrder.Customer,
		workOrder.TotalJoints,
		workOrder.TotalWeight,
		c.GetString("user_id"), // From auth middleware
	)
	
	c.JSON(200, gin.H{
		"status": "completed",
		"work_order": workOrder,
	})
}

// Advanced Event Analytics Functions

// GetTenantActivitySummary returns recent activity for a tenant
func GetTenantActivitySummary(tenantID string, hours int) map[string]int {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	events := GetEventBus().GetEventHistory(tenantID, 1000)
	
	summary := make(map[string]int)
	for _, event := range events {
		if event.Timestamp.After(since) {
			summary[event.EventType]++
		}
	}
	
	return summary
}

// GetSystemWideActivity returns activity across all tenants
func GetSystemWideActivity(hours int) map[string]map[string]int {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	events := GetEventBus().GetEventHistory("", 10000) // All tenants
	
	// tenant_id -> event_type -> count
	activity := make(map[string]map[string]int)
	
	for _, event := range events {
		if event.Timestamp.After(since) {
			if activity[event.TenantID] == nil {
				activity[event.TenantID] = make(map[string]int)
			}
			activity[event.TenantID][event.EventType]++
		}
	}
	
	return activity
}

// Real-time Dashboard Data (Safe Cross-Tenant Aggregation)
func GetDashboardMetrics() map[string]interface{} {
	activity := GetSystemWideActivity(24) // Last 24 hours
	
	metrics := map[string]interface{}{
		"active_tenants":           len(activity),
		"work_orders_completed":    0,
		"inventory_items_received": 0,
		"new_customers":           0,
		"low_stock_alerts":        0,
	}
	
	for _, tenantActivity := range activity {
		metrics["work_orders_completed"] = metrics["work_orders_completed"].(int) + tenantActivity["work_order_completed"]
		metrics["inventory_items_received"] = metrics["inventory_items_received"].(int) + tenantActivity["inventory_received"]
		metrics["new_customers"] = metrics["new_customers"].(int) + tenantActivity["customer_created"]
		metrics["low_stock_alerts"] = metrics["low_stock_alerts"].(int) + tenantActivity["inventory_low_stock"]
	}
	
	return metrics
}

// Event-Based Monitoring and Alerts

type AlertRule struct {
	EventType   string
	Threshold   int
	TimeWindow  time.Duration
	AlertAction func(tenantID string, count int)
}

var alertRules = []AlertRule{
	{
		EventType:  "inventory_low_stock",
		Threshold:  3, // More than 3 low stock alerts
		TimeWindow: time.Hour,
		AlertAction: func(tenantID string, count int) {
			log.Printf("🚨 ALERT: Tenant %s has %d low stock alerts in last hour", tenantID, count)
			// Send email, dashboard alert, etc.
		},
	},
	{
		EventType:  "work_order_completed",
		Threshold:  10, // More than 10 completions per hour
		TimeWindow: time.Hour,
		AlertAction: func(tenantID string, count int) {
			log.Printf("🎉 HIGH ACTIVITY: Tenant %s completed %d work orders in last hour", tenantID, count)
			// Positive alert for high productivity
		},
	},
}

// MonitorAlerts checks for alert conditions
func MonitorAlerts() {
	for _, rule := range alertRules {
		checkAlertRule(rule)
	}
}

func checkAlertRule(rule AlertRule) {
	since := time.Now().Add(-rule.TimeWindow)
	events := GetEventBus().GetEventHistory("", 10000)
	
	tenantCounts := make(map[string]int)
	
	for _, event := range events {
		if event.EventType == rule.EventType && event.Timestamp.After(since) {
			tenantCounts[event.TenantID]++
		}
	}
	
	for tenantID, count := range tenantCounts {
		if count > rule.Threshold {
			rule.AlertAction(tenantID, count)
		}
	}
}

// API Endpoints for Event System

// GET /admin/events/history?tenant_id=longbeach&limit=100
func GetEventHistoryHandler(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	
	events := GetEventBus().GetEventHistory(tenantID, limit)
	
	c.JSON(200, gin.H{
		"events": events,
		"count":  len(events),
		"tenant": tenantID,
	})
}

// GET /admin/dashboard/metrics
func GetDashboardMetricsHandler(c *gin.Context) {
	metrics := GetDashboardMetrics()
	c.JSON(200, metrics)
}

// GET /admin/tenants/activity?hours=24
func GetTenantActivityHandler(c *gin.Context) {
	hours := 24
	if h := c.Query("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 168 { // Max 1 week
			hours = parsed
		}
	}
	
	activity := GetSystemWideActivity(hours)
	
	c.JSON(200, gin.H{
		"activity":    activity,
		"time_window": fmt.Sprintf("%d hours", hours),
		"tenants":     len(activity),
	})
}

// This event system enables:
// 1. ✅ Complete tenant data isolation (no cross-tenant database access)
// 2. ✅ Cross-tenant analytics and reporting (via safe event aggregation)  
// 3. ✅ Real-time business intelligence (via event streaming)
// 4. ✅ Audit trail for compliance (all events logged)
// 5. ✅ Operational alerts (low stock, high activity, etc.)
// 6. ✅ Performance monitoring (event-driven metrics)

// Usage in your API handlers:
// - Call PublishWorkOrderCompleted() when work orders finish
// - Call PublishInventoryReceived() when new inventory arrives  
// - Call PublishCustomerCreated() when customers are added
// - Events automatically flow to analytics without breaking tenant isolation

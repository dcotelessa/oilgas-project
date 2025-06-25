package cache

import (
	"fmt"
	"sync"
	"time"
)
import "oilgas-backend/internal/models"

type Customer = models.Customer
type InventoryItem = models.InventoryItem
type Grade = models.Grade
type SearchResult = models.SearchResult

// Cache item with expiration
type cacheItem struct {
	value      interface{}
	expiration time.Time
	lastAccess time.Time
}

// Cache configuration
type Config struct {
	TTL             time.Duration
	CleanupInterval time.Duration
	MaxSize         int
}

// In-memory cache with TTL and LRU-like eviction
type Cache struct {
	items           map[string]*cacheItem
	mutex           sync.RWMutex
	config          Config
	stopCleanup     chan bool
	stats           Stats
	statsMutex      sync.RWMutex
}

// Cache statistics
type Stats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Items       int   `json:"items"`
	Evictions   int64 `json:"evictions"`
	Cleanups    int64 `json:"cleanups"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// New creates a new cache instance
func New(config Config) *Cache {
	cache := &Cache{
		items:       make(map[string]*cacheItem),
		config:      config,
		stopCleanup: make(chan bool),
	}

	// Start background cleanup
	go cache.cleanup()

	return cache
}

// Generic cache operations
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict items
	if len(c.items) >= c.config.MaxSize {
		c.evictLRU()
	}

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(c.config.TTL),
		lastAccess: time.Now(),
	}

	c.updateStats(func(s *Stats) { s.Items = len(c.items) })
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	item, exists := c.items[key]
	c.mutex.RUnlock()

	if !exists {
		c.updateStats(func(s *Stats) { s.Misses++ })
		return nil, false
	}

	// Check expiration
	if time.Now().After(item.expiration) {
		c.Delete(key)
		c.updateStats(func(s *Stats) { s.Misses++ })
		return nil, false
	}

	// Update last access time
	c.mutex.Lock()
	item.lastAccess = time.Now()
	c.mutex.Unlock()

	c.updateStats(func(s *Stats) { s.Hits++ })
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
	c.updateStats(func(s *Stats) { s.Items = len(c.items) })
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*cacheItem)
	c.updateStats(func(s *Stats) { s.Items = 0 })
}

// Domain-specific cache methods for oil & gas inventory

// Customer cache operations
func (c *Cache) CacheCustomer(id int, customer *Customer) {
	c.Set(fmt.Sprintf("customer:%d", id), customer)
}

func (c *Cache) GetCustomer(id int) (*Customer, bool) {
	value, exists := c.Get(fmt.Sprintf("customer:%d", id))
	if !exists {
		return nil, false
	}
	
	customer, ok := value.(*Customer)
	if !ok {
		c.Delete(fmt.Sprintf("customer:%d", id))
		return nil, false
	}
	
	return customer, true
}

func (c *Cache) DeleteCustomer(id int) {
	c.Delete(fmt.Sprintf("customer:%d", id))
}

// Inventory cache operations
func (c *Cache) CacheInventoryItem(id int, item *InventoryItem) {
	c.Set(fmt.Sprintf("inventory:%d", id), item)
}

func (c *Cache) GetInventoryItem(id int) (*InventoryItem, bool) {
	value, exists := c.Get(fmt.Sprintf("inventory:%d", id))
	if !exists {
		return nil, false
	}
	
	item, ok := value.(*InventoryItem)
	if !ok {
		c.Delete(fmt.Sprintf("inventory:%d", id))
		return nil, false
	}
	
	return item, true
}

func (c *Cache) DeleteInventoryItem(id int) {
	c.Delete(fmt.Sprintf("inventory:%d", id))
}

// Customer inventory cache (for frequently accessed customer inventory lists)
func (c *Cache) CacheCustomerInventory(customerID int, items []*InventoryItem) {
	c.Set(fmt.Sprintf("customer_inventory:%d", customerID), items)
}

func (c *Cache) GetCustomerInventory(customerID int) ([]*InventoryItem, bool) {
	value, exists := c.Get(fmt.Sprintf("customer_inventory:%d", customerID))
	if !exists {
		return nil, false
	}
	
	items, ok := value.([]*InventoryItem)
	if !ok {
		c.Delete(fmt.Sprintf("customer_inventory:%d", customerID))
		return nil, false
	}
	
	return items, true
}

// Grade cache operations
func (c *Cache) CacheGrades(grades []Grade) {
	c.Set("grades:all", grades)
}

func (c *Cache) GetGrades() ([]Grade, bool) {
	value, exists := c.Get("grades:all")
	if !exists {
		return nil, false
	}
	
	grades, ok := value.([]Grade)
	if !ok {
		c.Delete("grades:all")
		return nil, false
	}
	
	return grades, true
}

// Search result caching (for expensive queries)
func (c *Cache) CacheSearchResults(query string, results interface{}) {
	c.Set(fmt.Sprintf("search:%s", query), results)
}

func (c *Cache) GetSearchResults(query string) (interface{}, bool) {
	return c.Get(fmt.Sprintf("search:%s", query))
}

// Background cleanup goroutine
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.performCleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) performCleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expired := 0

	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
			expired++
		}
	}

	c.updateStats(func(s *Stats) {
		s.Items = len(c.items)
		s.Cleanups++
		s.LastCleanup = now
	})

	if expired > 0 {
		fmt.Printf("Cache cleanup: removed %d expired items\n", expired)
	}
}

// LRU-like eviction when cache is full
func (c *Cache) evictLRU() {
	if len(c.items) == 0 {
		return
	}

	// Find oldest accessed item
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.lastAccess
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		c.updateStats(func(s *Stats) { s.Evictions++ })
	}
}

// Statistics and monitoring
func (c *Cache) GetStats() Stats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	
	c.mutex.RLock()
	itemCount := len(c.items)
	c.mutex.RUnlock()
	
	stats := c.stats
	stats.Items = itemCount
	return stats
}

func (c *Cache) updateStats(fn func(*Stats)) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	fn(&c.stats)
}

func (c *Cache) GetHitRatio() float64 {
	stats := c.GetStats()
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(stats.Hits) / float64(total)
}

// Cache invalidation helpers
func (c *Cache) InvalidateCustomerData(customerID int) {
	c.DeleteCustomer(customerID)
	c.Delete(fmt.Sprintf("customer_inventory:%d", customerID))
	// Clear any search results that might contain this customer
	c.invalidateSearchResults()
}

func (c *Cache) invalidateSearchResults() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for key := range c.items {
		if len(key) > 7 && key[:7] == "search:" {
			delete(c.items, key)
		}
	}
}

// Shutdown cleanup
func (c *Cache) Close() {
	close(c.stopCleanup)
	c.Clear()
}

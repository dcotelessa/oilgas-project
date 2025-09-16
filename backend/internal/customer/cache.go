// backend/internal/customer/cache.go
package customer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type inMemoryCache struct {
	customers map[string]*cacheEntry
	mutex     sync.RWMutex
	ttl       time.Duration
}

type cacheEntry struct {
	customer  *Customer
	expiresAt time.Time
}

func NewInMemoryCache(ttl time.Duration) CacheService {
	cache := &inMemoryCache{
		customers: make(map[string]*cacheEntry),
		ttl:       ttl,
	}
	
	// Cleanup routine
	go func() {
		ticker := time.NewTicker(ttl)
		defer ticker.Stop()
		
		for range ticker.C {
			cache.cleanup()
		}
	}()
	
	return cache
}

func (c *inMemoryCache) GetCustomer(tenantID string, customerID int) (*Customer, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	key := fmt.Sprintf("%s:%d", tenantID, customerID)
	entry, exists := c.customers[key]
	
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	
	return entry.customer, true
}

func (c *inMemoryCache) CacheCustomer(tenantID string, customer *Customer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := fmt.Sprintf("%s:%d", tenantID, customer.ID)
	c.customers[key] = &cacheEntry{
		customer:  customer,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *inMemoryCache) InvalidateCustomer(tenantID string, customerID int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	key := fmt.Sprintf("%s:%d", tenantID, customerID)
	delete(c.customers, key)
}

func (c *inMemoryCache) InvalidateCustomerSearch(tenantID string, filters SearchFilters) {
	// For search invalidation, we could implement more sophisticated cache keys
	// For now, we'll keep it simple and rely on TTL
}

func (c *inMemoryCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	for key, entry := range c.customers {
		if now.After(entry.expiresAt) {
			delete(c.customers, key)
		}
	}
}

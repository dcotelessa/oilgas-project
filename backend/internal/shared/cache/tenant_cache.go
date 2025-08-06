// backend/internal/shared/cache/tenant_cache.go
package cache

import (
    "context"
    "sync"
    "time"
)

// TenantAwareCache provides multi-tenant caching with Go's native structures
type TenantAwareCache struct {
    // Separate cache per tenant for isolation
    tenantCaches map[string]*TenantCache
    mutex        sync.RWMutex
    config       CacheConfig
}

type TenantCache struct {
    data     map[string]*CacheEntry
    mutex    sync.RWMutex
    tenantID string
}

type CacheEntry struct {
    Value     interface{}
    ExpiresAt time.Time
    CreatedAt time.Time
}

type CacheConfig struct {
    DefaultTTL    time.Duration // 15 minutes for most data
    MaxEntries    int           // 10,000 entries per tenant
    CleanupInterval time.Duration // 5 minutes
}

func NewTenantAwareCache(config CacheConfig) *TenantAwareCache {
    cache := &TenantAwareCache{
        tenantCaches: make(map[string]*TenantCache),
        config:       config,
    }
    
    // Start cleanup goroutine
    go cache.startCleanup()
    
    return cache
}

// Get retrieves a value from tenant-specific cache
func (tc *TenantAwareCache) Get(tenantID, key string) (interface{}, bool) {
    tc.mutex.RLock()
    tenantCache, exists := tc.tenantCaches[tenantID]
    tc.mutex.RUnlock()
    
    if !exists {
        return nil, false
    }
    
    tenantCache.mutex.RLock()
    defer tenantCache.mutex.RUnlock()
    
    entry, exists := tenantCache.data[key]
    if !exists {
        return nil, false
    }
    
    // Check expiration
    if time.Now().After(entry.ExpiresAt) {
        delete(tenantCache.data, key) // Lazy cleanup
        return nil, false
    }
    
    return entry.Value, true
}

// Set stores a value in tenant-specific cache
func (tc *TenantAwareCache) Set(tenantID, key string, value interface{}, ttl time.Duration) {
    tc.mutex.Lock()
    tenantCache, exists := tc.tenantCaches[tenantID]
    if !exists {
        tenantCache = &TenantCache{
            data:     make(map[string]*CacheEntry),
            tenantID: tenantID,
        }
        tc.tenantCaches[tenantID] = tenantCache
    }
    tc.mutex.Unlock()
    
    if ttl == 0 {
        ttl = tc.config.DefaultTTL
    }
    
    entry := &CacheEntry{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
        CreatedAt: time.Now(),
    }
    
    tenantCache.mutex.Lock()
    defer tenantCache.mutex.Unlock()
    
    // Implement simple LRU if we're at capacity
    if len(tenantCache.data) >= tc.config.MaxEntries {
        tc.evictOldest(tenantCache)
    }
    
    tenantCache.data[key] = entry
}

// Cache-friendly service patterns
func (tc *TenantAwareCache) GetCustomer(tenantID string, customerID int) (*Customer, bool) {
    key := fmt.Sprintf("customer:%d", customerID)
    if value, found := tc.Get(tenantID, key); found {
        return value.(*Customer), true
    }
    return nil, false
}

func (tc *TenantAwareCache) CacheCustomer(tenantID string, customer *Customer) {
    key := fmt.Sprintf("customer:%d", customer.ID)
    tc.Set(tenantID, key, customer, 15*time.Minute)
}

// Invalidation patterns for data consistency
func (tc *TenantAwareCache) InvalidateCustomer(tenantID string, customerID int) {
    key := fmt.Sprintf("customer:%d", customerID)
    tc.Delete(tenantID, key)
}

func (tc *TenantAwareCache) InvalidatePattern(tenantID, pattern string) {
    // Invalidate all keys matching pattern (e.g., "customer:*")
    tc.mutex.RLock()
    tenantCache, exists := tc.tenantCaches[tenantID]
    tc.mutex.RUnlock()
    
    if !exists {
        return
    }
    
    tenantCache.mutex.Lock()
    defer tenantCache.mutex.Unlock()
    
    for key := range tenantCache.data {
        if matched, _ := filepath.Match(pattern, key); matched {
            delete(tenantCache.data, key)
        }
    }
}

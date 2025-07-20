package cache

import (
	"sync"
	"time"
)

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*Item
}

type Item struct {
	Value      interface{}
	Expiration int64
}

func New() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]*Item),
	}
}

func NewWithDefaultExpiration(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
	cache := New()
	
	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				cache.DeleteExpired()
			}
		}
	}()
	
	return cache
}

func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	
	c.items[key] = &Item{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *MemoryCache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, found := c.items[key]
	if !found {
		return nil
	}
	
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		delete(c.items, key)
		return nil
	}
	
	return item.Value
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MemoryCache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now().UnixNano()
	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
}

func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

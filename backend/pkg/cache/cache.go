// pkg/cache/cache.go - Simple in-memory cache implementation
package cache

import (
	"sync"
	"time"
)

type Cache struct {
	items             map[string]*Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

type Item struct {
	Value      interface{}
	Expiration int64
}

func NewWithDefaultExpiration(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:             make(map[string]*Item),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}

	// Start cleanup goroutine
	go c.cleanupLoop()
	return c
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = &Item{
		Value:      value,
		Expiration: expiration,
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.mu.RUnlock()
		c.Delete(key)
		return nil, false
	}

	c.mu.RUnlock()
	return item.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) ItemCount() int {
	c.mu.RLock()
	count := len(c.items)
	c.mu.RUnlock()
	return count
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) DeleteExpired() {
	c.mu.Lock()
	now := time.Now().UnixNano()
	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}

func (c *Cache) Stop() {
	close(c.stopCleanup)
}

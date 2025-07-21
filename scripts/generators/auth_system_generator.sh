#!/bin/bash
# scripts/generators/auth_system_generator.sh
# Generates complete authentication system components for Phase 3

set -e
echo "🔐 Generating authentication system components..."

# Detect backend directory
BACKEND_DIR=""
if [ -d "backend" ] && [ -f "backend/go.mod" ]; then
    BACKEND_DIR="backend/"
elif [ -f "go.mod" ]; then
    BACKEND_DIR=""
else
    echo "❌ Error: Cannot find go.mod file"
    exit 1
fi

# Get module name for imports
MODULE_NAME=""
if [ -f "${BACKEND_DIR}go.mod" ]; then
    MODULE_NAME=$(grep "^module " "${BACKEND_DIR}go.mod" | awk '{print $2}')
fi
if [ -z "$MODULE_NAME" ]; then
    MODULE_NAME="dcotelessa/oil-gas-inventory"
    echo "⚠️  Using default module name: $MODULE_NAME"
fi

echo "📦 Using module name: $MODULE_NAME"

# Create directories
mkdir -p "${BACKEND_DIR}internal/auth"
mkdir -p "${BACKEND_DIR}internal/handlers"
mkdir -p "${BACKEND_DIR}pkg/cache"

# Create in-memory cache implementation
cat > "${BACKEND_DIR}pkg/cache/memory_cache.go" << 'EOF'
package cache

import (
	"sync"
	"time"
)

type MemoryCache struct {
	items map[string]*cacheItem
	mutex sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

func New(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items:             make(map[string]*cacheItem),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}
	go cache.cleanupExpired()
	return cache
}

func (c *MemoryCache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil
	}
	
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil
	}
	
	return item.value
}

func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	
	c.mutex.Lock()
	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
	c.mutex.Unlock()
}

func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	delete(c.items, key)
	c.mutex.Unlock()
}

func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			now := time.Now().UnixNano()
			for key, item := range c.items {
				if item.expiration > 0 && now > item.expiration {
					delete(c.items, key)
				}
			}
			c.mutex.Unlock()
		case <-c.stopCleanup:
			return
		}
	}
}
EOF

# Create session helpers
cat > "${BACKEND_DIR}internal/auth/session_helpers.go" << EOF
package auth

import (
	"strings"
	"github.com/gin-gonic/gin"
)

func ExtractSessionID(c *gin.Context) string {
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		return cookie
	}

	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}
EOF

echo "✅ Authentication system components generated"
echo "   - In-memory cache implementation"
echo "   - Session helper functions"
echo "   - Directory structure created"

package auth_test

import (
	"testing"
	"time"

	"github.com/dcotelessa/oil-gas-inventory/pkg/cache"
)

func TestMemoryCache(t *testing.T) {
	cache := cache.NewWithDefaultExpiration(5*time.Minute, 1*time.Minute)
	
	// Test basic operations
	cache.Set("test_key", "test_value", 1*time.Second)
	
	value := cache.Get("test_key")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}
	
	// Test expiration
	time.Sleep(2 * time.Second)
	value = cache.Get("test_key")
	if value != nil {
		t.Errorf("Expected nil after expiration, got %v", value)
	}
}

func TestCacheCount(t *testing.T) {
	cache := cache.New()
	
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	
	if cache.Count() != 2 {
		t.Errorf("Expected count 2, got %d", cache.Count())
	}
	
	cache.Delete("key1")
	
	if cache.Count() != 1 {
		t.Errorf("Expected count 1 after delete, got %d", cache.Count())
	}
}

# In-Memory Cache Implementation

Copy the complete cache implementation from the `inmemory_cache` artifact to `cache.go` in this directory.

The cache provides:
- Thread-safe operations with TTL
- Automatic cleanup of expired items
- Size-based eviction (LRU-like)
- Specialized methods for oil & gas domain objects
- Performance monitoring and statistics

## Usage

```go
import "your-app/pkg/cache"

// Initialize cache
cache := cache.New(cache.Config{
    TTL:             5 * time.Minute,
    CleanupInterval: 10 * time.Minute,
    MaxSize:         1000,
})

// Use cache
cache.CacheCustomer(123, customerData)
customer, exists := cache.GetCustomer(123)
```

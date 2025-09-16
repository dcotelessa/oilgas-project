// backend/internal/database/tenant_connection.go
package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	connectionPool = make(map[string]*sql.DB)
	poolMutex      sync.RWMutex
)

// GetTenantDB returns a database connection for the specified tenant with optimized settings
func GetTenantDB(tenantID string) (*sql.DB, error) {
	poolMutex.RLock()
	if db, exists := connectionPool[tenantID]; exists {
		poolMutex.RUnlock()
		return db, nil
	}
	poolMutex.RUnlock()

	return createOptimizedTenantConnection(tenantID)
}

func createOptimizedTenantConnection(tenantID string) (*sql.DB, error) {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := connectionPool[tenantID]; exists {
		return db, nil
	}

	// Get tenant-specific database URL
	var tenantURL string
	switch tenantID {
	case "longbeach":
		tenantURL = os.Getenv("LONGBEACH_DB_URL")
		if tenantURL == "" {
			return nil, fmt.Errorf("LONGBEACH_DB_URL environment variable not set")
		}
	case "bakersfield":
		tenantURL = os.Getenv("BAKERSFIELD_DB_URL")
		if tenantURL == "" {
			return nil, fmt.Errorf("BAKERSFIELD_DB_URL environment variable not set")
		}
	default:
		return nil, fmt.Errorf("unsupported tenant: %s", tenantID)
	}

	db, err := sql.Open("postgres", tenantURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to tenant database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Configure optimized connection pool per tenant
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)        // Connections live max 1 hour
	db.SetConnMaxIdleTime(15 * time.Minute) // Idle connections cleaned after 15 min

	connectionPool[tenantID] = db
	return db, nil
}

func buildTenantConnectionString(baseURL, tenantID string) string {
	// Example: postgresql://user:pass@host:5432/maindb
	// Becomes: postgresql://user:pass@host:5432/oilgas_longbeach
	
	parts := strings.Split(baseURL, "/")
	if len(parts) >= 4 {
		// Replace database name with tenant database
		// Handle potential query parameters
		lastPart := parts[len(parts)-1]
		var params string
		if strings.Contains(lastPart, "?") {
			dbAndParams := strings.Split(lastPart, "?")
			params = "?" + dbAndParams[1]
		}
		
		// Replace with tenant database name
		parts[len(parts)-1] = fmt.Sprintf("oilgas_%s%s", tenantID, params)
		return strings.Join(parts, "/")
	}
	
	// Fallback: append tenant database name
	return fmt.Sprintf("%s/oilgas_%s", strings.TrimSuffix(baseURL, "/"), tenantID)
}

// CloseTenantConnections closes all tenant database connections
// (Adding this function that was missing from optimized version)
func CloseTenantConnections() {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	for tenantID, db := range connectionPool {
		if db != nil {
			db.Close()
		}
		delete(connectionPool, tenantID)
	}
}

// GetConnectionStats returns connection statistics for monitoring
func GetConnectionStats(tenantID string) (*ConnectionStats, error) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	
	db, exists := connectionPool[tenantID]
	if !exists {
		return nil, fmt.Errorf("no connection pool found for tenant %s", tenantID)
	}
	
	stats := db.Stats()
	return &ConnectionStats{
		TenantID:        tenantID,
		MaxOpenConns:    stats.MaxOpenConnections,
		OpenConnections: stats.OpenConnections,
		InUse:          stats.InUse,
		Idle:           stats.Idle,
		WaitCount:      stats.WaitCount,
		WaitDuration:   stats.WaitDuration,
		MaxIdleClosed:  stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}, nil
}

// ConnectionStats holds connection pool statistics
type ConnectionStats struct {
	TenantID          string        `json:"tenant_id"`
	MaxOpenConns      int           `json:"max_open_conns"`
	OpenConnections   int           `json:"open_connections"`
	InUse            int           `json:"in_use"`
	Idle             int           `json:"idle"`
	WaitCount        int64         `json:"wait_count"`
	WaitDuration     time.Duration `json:"wait_duration"`
	MaxIdleClosed    int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// GetAllTenantIDs returns all tenant IDs with active connections
func GetAllTenantIDs() []string {
	poolMutex.RLock()
	defer poolMutex.RUnlock()
	
	tenantIDs := make([]string, 0, len(connectionPool))
	for tenantID := range connectionPool {
		tenantIDs = append(tenantIDs, tenantID)
	}
	return tenantIDs
}

// ValidateTenantExists checks if a tenant database exists and is accessible
func ValidateTenantExists(tenantID string) bool {
	_, err := GetTenantDB(tenantID)
	return err == nil
}

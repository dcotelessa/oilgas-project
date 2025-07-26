// backend/internal/database/optimized_tenant_connection.go
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

// GetTenantDB with optimized connection routing
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

	// Get base connection string
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Dynamic database name routing
	tenantURL := buildTenantConnectionString(databaseURL, tenantID)

	db, err := sql.Open("postgres", tenantURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to tenant database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Configure connection pool per tenant
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)

	connectionPool[tenantID] = db
	return db, nil
}

func buildTenantConnectionString(baseURL, tenantID string) string {
	// Example: postgresql://user:pass@host:5432/maindb
	// Becomes: postgresql://user:pass@host:5432/oilgas_longbeach
	
	parts := strings.Split(baseURL, "/")
	if len(parts) >= 4 {
		// Replace database name with tenant database
		parts[len(parts)-1] = fmt.Sprintf("oilgas_%s", tenantID)
		return strings.Join(parts, "/")
	}
	
	// Fallback: append tenant database name
	return fmt.Sprintf("%s_oilgas_%s", baseURL, tenantID)
}

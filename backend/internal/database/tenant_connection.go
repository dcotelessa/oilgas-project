// backend/internal/database/tenant_connection.go
package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var (
	connectionPool = make(map[string]*sql.DB)
	poolMutex      sync.RWMutex
)

// GetTenantDB returns a database connection for the specified tenant
func GetTenantDB(tenantID string) (*sql.DB, error) {
	poolMutex.RLock()
	if db, exists := connectionPool[tenantID]; exists {
		poolMutex.RUnlock()
		return db, nil
	}
	poolMutex.RUnlock()

	// Create new connection
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := connectionPool[tenantID]; exists {
		return db, nil
	}

	db, err := createTenantConnection(tenantID)
	if err != nil {
		return nil, err
	}

	connectionPool[tenantID] = db
	return db, nil
}

func createTenantConnection(tenantID string) (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Replace database name in connection string
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid DATABASE_URL format")
	}

	dbName := fmt.Sprintf("oilgas_%s", tenantID)
	parts[len(parts)-1] = dbName
	tenantURL := strings.Join(parts, "/")

	db, err := sql.Open("postgres", tenantURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to tenant database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Configure connection pool for tenant
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db, nil
}

// CloseTenantConnections closes all tenant database connections
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

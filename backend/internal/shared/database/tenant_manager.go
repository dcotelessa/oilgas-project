// backend/internal/shared/database/tenant_manager.go
package database

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

// TenantManager handles connections to multiple tenant databases
type TenantManager struct {
    authDB      *sqlx.DB
    tenantDBs   map[string]*sqlx.DB
    tenantMutex sync.RWMutex
    config      DatabaseConfig
}

type DatabaseConfig struct {
    AuthDBConnString string
    TenantDBConfigs  map[string]TenantDBConfig
    MaxOpenConns     int
    MaxIdleConns     int
    ConnMaxLifetime  time.Duration
}

type TenantDBConfig struct {
    ConnectionString string
    DatabaseName     string
    Location         string
}

func NewTenantManager(config DatabaseConfig) (*TenantManager, error) {
    tm := &TenantManager{
        tenantDBs: make(map[string]*sqlx.DB),
        config:    config,
    }
    
    // Connect to central auth database
    authDB, err := sqlx.Connect("postgres", config.AuthDBConnString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to auth database: %w", err)
    }
    
    // Configure connection pool for auth DB
    authDB.SetMaxOpenConns(config.MaxOpenConns)
    authDB.SetMaxIdleConns(config.MaxIdleConns)
    authDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    tm.authDB = authDB
    
    // Connect to all tenant databases
    for tenantID, dbConfig := range config.TenantDBConfigs {
        if err := tm.connectTenantDB(tenantID, dbConfig); err != nil {
            return nil, fmt.Errorf("failed to connect to tenant %s database: %w", tenantID, err)
        }
    }
    
    return tm, nil
}

func (tm *TenantManager) connectTenantDB(tenantID string, config TenantDBConfig) error {
    db, err := sqlx.Connect("postgres", config.ConnectionString)
    if err != nil {
        return err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(tm.config.MaxOpenConns)
    db.SetMaxIdleConns(tm.config.MaxIdleConns)
    db.SetConnMaxLifetime(tm.config.ConnMaxLifetime)
    
    tm.tenantMutex.Lock()
    tm.tenantDBs[tenantID] = db
    tm.tenantMutex.Unlock()
    
    return nil
}

// GetAuthDB returns the central auth database connection
func (tm *TenantManager) GetAuthDB() *sqlx.DB {
    return tm.authDB
}

// GetTenantDB returns a specific tenant's database connection
func (tm *TenantManager) GetTenantDB(tenantID string) (*sqlx.DB, error) {
    tm.tenantMutex.RLock()
    defer tm.tenantMutex.RUnlock()
    
    db, exists := tm.tenantDBs[tenantID]
    if !exists {
        return nil, fmt.Errorf("tenant database not found: %s", tenantID)
    }
    
    return db, nil
}

// GetAllTenantDBs returns all tenant database connections (for enterprise operations)
func (tm *TenantManager) GetAllTenantDBs() map[string]*sqlx.DB {
    tm.tenantMutex.RLock()
    defer tm.tenantMutex.RUnlock()
    
    // Return a copy to prevent external modification
    result := make(map[string]*sqlx.DB)
    for k, v := range tm.tenantDBs {
        result[k] = v
    }
    return result
}

// ValidateTenantAccess checks if a user can access a specific tenant
func (tm *TenantManager) ValidateTenantAccess(ctx context.Context, userID int, tenantID string) error {
    query := `
        SELECT COUNT(*) FROM user_tenant_access 
        WHERE user_id = $1 AND tenant_id = $2 AND is_active = true`
    
    var count int
    err := tm.authDB.GetContext(ctx, &count, query, userID, tenantID)
    if err != nil {
        return fmt.Errorf("failed to check tenant access: %w", err)
    }
    
    if count == 0 {
        return fmt.Errorf("user %d does not have access to tenant %s", userID, tenantID)
    }
    
    return nil
}

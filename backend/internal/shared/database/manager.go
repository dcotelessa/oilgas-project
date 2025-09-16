// backend/internal/shared/database/manager.go
package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// DatabaseManager handles connections to multiple tenant databases
type DatabaseManager struct {
	centralDB    *sql.DB
	tenantDBs    map[string]*sql.DB
	mutex        sync.RWMutex
	config       *Config
}

type Config struct {
	CentralDBURL string
	TenantDBs    map[string]string // tenantID -> connection string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

func NewDatabaseManager(config *Config) (*DatabaseManager, error) {
	// Connect to central auth database
	centralDB, err := sql.Open("postgres", config.CentralDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to central database: %w", err)
	}
	
	centralDB.SetMaxOpenConns(config.MaxOpenConns)
	centralDB.SetMaxIdleConns(config.MaxIdleConns)
	centralDB.SetConnMaxLifetime(config.MaxLifetime)
	
	manager := &DatabaseManager{
		centralDB: centralDB,
		tenantDBs: make(map[string]*sql.DB),
		config:    config,
	}
	
	// Initialize tenant databases
	for tenantID, connStr := range config.TenantDBs {
		if err := manager.addTenantDB(tenantID, connStr); err != nil {
			return nil, fmt.Errorf("failed to connect to tenant database %s: %w", tenantID, err)
		}
	}
	
	return manager, nil
}

func (dm *DatabaseManager) addTenantDB(tenantID, connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	
	db.SetMaxOpenConns(dm.config.MaxOpenConns)
	db.SetMaxIdleConns(dm.config.MaxIdleConns)
	db.SetConnMaxLifetime(dm.config.MaxLifetime)
	
	if err := db.Ping(); err != nil {
		return err
	}
	
	dm.mutex.Lock()
	dm.tenantDBs[tenantID] = db
	dm.mutex.Unlock()
	
	return nil
}

func (dm *DatabaseManager) GetCentralDB() *sql.DB {
	return dm.centralDB
}

func (dm *DatabaseManager) GetTenantDB(tenantID string) (*sql.DB, error) {
	dm.mutex.RLock()
	db, exists := dm.tenantDBs[tenantID]
	dm.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("tenant database not found: %s", tenantID)
	}
	
	return db, nil
}

func (dm *DatabaseManager) Close() error {
	var errs []error
	
	if err := dm.centralDB.Close(); err != nil {
		errs = append(errs, err)
	}
	
	dm.mutex.Lock()
	for tenantID, db := range dm.tenantDBs {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close tenant DB %s: %w", tenantID, err))
		}
	}
	dm.mutex.Unlock()
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}
	
	return nil
}

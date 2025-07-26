// backend/internal/tenant/manager.go
package tenant

import (
	"database/sql"
	"fmt"
	"sync"
	
	_ "github.com/lib/pq"
)

type DatabaseManager struct {
	connections map[string]*sql.DB
	mutex       sync.RWMutex
	baseConnStr string
}

func NewDatabaseManager(baseConnStr string) *DatabaseManager {
	return &DatabaseManager{
		connections: make(map[string]*sql.DB),
		baseConnStr: baseConnStr,
	}
}

func (dm *DatabaseManager) GetConnection(databaseName string) (*sql.DB, error) {
	dm.mutex.RLock()
	if conn, exists := dm.connections[databaseName]; exists {
		dm.mutex.RUnlock()
		return conn, nil
	}
	dm.mutex.RUnlock()
	
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if conn, exists := dm.connections[databaseName]; exists {
		return conn, nil
	}
	
	// Create new connection
	connStr := fmt.Sprintf("%s/%s?sslmode=disable", dm.baseConnStr, databaseName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	
	// Configure connection pool for each tenant database
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	
	dm.connections[databaseName] = db
	return db, nil
}

func (dm *DatabaseManager) CloseAll() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	for _, conn := range dm.connections {
		conn.Close()
	}
	dm.connections = make(map[string]*sql.DB)
}

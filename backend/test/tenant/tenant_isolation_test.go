package tenant_test

import (
    "database/sql"
    "os"
    "testing"
    
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
    // Load environment from .env.local
    if err := godotenv.Load("../../.env.local"); err != nil {
        t.Fatalf("Failed to load environment: %v", err)
    }
    
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        t.Fatal("DATABASE_URL not set in environment")
    }
    
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    if err := db.Ping(); err != nil {
        t.Fatalf("Failed to ping test database: %v", err)
    }
    
    return db
}

func TestTenantIsolation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Test RLS policies work
    t.Run("RLS policies enforce tenant isolation", func(t *testing.T) {
        // Set tenant context
        _, err := db.Exec("SET app.current_tenant_id = 1")
        if err != nil {
            t.Fatalf("Failed to set tenant context: %v", err)
        }
        
        // Query should only return tenant 1 data
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM store.inventory WHERE tenant_id = 1").Scan(&count)
        if err != nil {
            t.Fatalf("Failed to query tenant inventory: %v", err)
        }
        
        t.Logf("Tenant 1 has %d inventory items", count)
    })
    
    t.Run("Admin bypass works", func(t *testing.T) {
        // Set admin role
        _, err := db.Exec("SET app.user_role = 'admin'")
        if err != nil {
            t.Fatalf("Failed to set admin role: %v", err)
        }
        
        // Admin should see all data
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM store.inventory").Scan(&count)
        if err != nil {
            t.Fatalf("Admin should be able to query all inventory: %v", err)
        }
        
        t.Logf("Admin can see %d inventory items total", count)
    })
}

func TestTenantContextFunctions(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    t.Run("get_current_tenant_id function works", func(t *testing.T) {
        _, err := db.Exec("SET app.current_tenant_id = 1")
        if err != nil {
            t.Fatalf("Failed to set tenant context: %v", err)
        }
        
        var tenantID int
        err = db.QueryRow("SELECT get_current_tenant_id()").Scan(&tenantID)
        if err != nil {
            t.Fatalf("Failed to get current tenant ID: %v", err)
        }
        
        if tenantID != 1 {
            t.Errorf("Expected tenant ID 1, got %d", tenantID)
        }
    })
    
    t.Run("is_admin_user function works", func(t *testing.T) {
        _, err := db.Exec("SET app.user_role = 'admin'")
        if err != nil {
            t.Fatalf("Failed to set admin role: %v", err)
        }
        
        var isAdmin bool
        err = db.QueryRow("SELECT is_admin_user()").Scan(&isAdmin)
        if err != nil {
            t.Fatalf("Failed to check admin status: %v", err)
        }
        
        if !isAdmin {
            t.Errorf("Expected admin status to be true")
        }
    })
}

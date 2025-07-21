package tenant_test

import (
    "database/sql"
    "os"
    "testing"
    
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
    // Find .env.local file - try multiple possible locations
    envPaths := []string{
        "../../../.env.local",  // From backend/test/tenant/ -> root
        "../../.env.local",     // From backend/test/ -> root
        ".env.local",           // Current directory
        "../.env.local",        // One level up
    }
    
    var envFile string
    for _, path := range envPaths {
        if _, err := os.Stat(path); err == nil {
            envFile = path
            break
        }
    }
    
    if envFile == "" {
        t.Skip("Skipping test - .env.local not found in any expected location")
    }
    
    // Load environment from .env.local
    if err := godotenv.Load(envFile); err != nil {
        t.Skipf("Skipping test - failed to load %s: %v", envFile, err)
    }
    
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        t.Skip("Skipping test - DATABASE_URL not set")
    }
    
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        t.Skipf("Skipping test - database connection failed: %v", err)
    }
    
    if err := db.Ping(); err != nil {
        t.Skipf("Skipping test - database ping failed: %v", err)
    }
    
    return db
}

func TestBasicTenantSetup(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Test that tenants table exists and has data
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM store.tenants").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query tenants table: %v", err)
    }
    
    if count == 0 {
        t.Error("No tenants found - run 'make setup' first")
    } else {
        t.Logf("✅ Found %d tenants", count)
    }
}

func TestTenantIsolationFunctions(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Test RLS functions exist
    t.Run("RLS functions exist", func(t *testing.T) {
        var result bool
        
        // Test get_current_tenant_id function
        err := db.QueryRow("SELECT get_current_tenant_id() IS NOT NULL").Scan(&result)
        if err != nil {
            t.Errorf("get_current_tenant_id function not working: %v", err)
        }
        
        // Test is_admin_user function  
        err = db.QueryRow("SELECT is_admin_user() IS NOT NULL").Scan(&result)
        if err != nil {
            t.Errorf("is_admin_user function not working: %v", err)
        }
        
        t.Log("✅ RLS functions are working")
    })
}

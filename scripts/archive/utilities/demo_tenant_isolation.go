package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func main() {
    // Load environment
    if err := godotenv.Load(".env"); err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    // Connect to database
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    fmt.Println("🧪 Tenant Isolation Demo")
    fmt.Println("=======================")

    // Demo 1: Show all tenants
    fmt.Println("\n📋 All Tenants:")
    showAllTenants(db)

    // Demo 2: Test tenant context switching
    fmt.Println("\n🔄 Testing Tenant Context:")
    testTenantContext(db)

    // Demo 3: Show RLS in action
    fmt.Println("\n🛡️  Row-Level Security Demo:")
    testRLSPolicies(db)

    fmt.Println("\n✅ Demo completed!")
}

func showAllTenants(db *sql.DB) {
    rows, err := db.Query("SELECT tenant_id, tenant_name, tenant_slug, active FROM store.tenants ORDER BY tenant_id")
    if err != nil {
        log.Printf("Error querying tenants: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name, slug string
        var active bool
        rows.Scan(&id, &name, &slug, &active)
        status := "✅"
        if !active {
            status = "❌"
        }
        fmt.Printf("   %s [%d] %s (%s)\n", status, id, name, slug)
    }
}

func testTenantContext(db *sql.DB) {
    // Get first tenant
    var tenantID int
    var tenantName string
    err := db.QueryRow("SELECT tenant_id, tenant_name FROM store.tenants WHERE active = true LIMIT 1").Scan(&tenantID, &tenantName)
    if err != nil {
        fmt.Printf("   ❌ No active tenants found\n")
        return
    }

    // Set tenant context
    _, err = db.Exec("SET app.current_tenant_id = $1", tenantID)
    if err != nil {
        fmt.Printf("   ❌ Failed to set tenant context: %v\n", err)
        return
    }

    _, err = db.Exec("SET app.user_role = 'admin'")
    if err != nil {
        fmt.Printf("   ❌ Failed to set user role: %v\n", err)
        return
    }

    fmt.Printf("   ✅ Tenant context set to: %s (ID: %d)\n", tenantName, tenantID)

    // Test context retrieval
    var currentTenant int
    err = db.QueryRow("SELECT get_current_tenant_id()").Scan(&currentTenant)
    if err != nil {
        fmt.Printf("   ❌ Failed to get current tenant: %v\n", err)
        return
    }

    fmt.Printf("   ✅ Current tenant function returns: %d\n", currentTenant)
}

func testRLSPolicies(db *sql.DB) {
    // Count total inventory without RLS context
    var totalInventory int
    err := db.QueryRow("SELECT COUNT(*) FROM store.inventory").Scan(&totalInventory)
    if err != nil {
        fmt.Printf("   ❌ Failed to count inventory: %v\n", err)
        return
    }

    fmt.Printf("   📦 Total inventory items (with current context): %d\n", totalInventory)

    // Test with different tenant contexts
    tenants := []int{}
    rows, err := db.Query("SELECT tenant_id FROM store.tenants WHERE active = true ORDER BY tenant_id")
    if err != nil {
        return
    }
    defer rows.Close()

    for rows.Next() {
        var tid int
        rows.Scan(&tid)
        tenants = append(tenants, tid)
    }

    for _, tid := range tenants {
        _, err = db.Exec("SET app.current_tenant_id = $1", tid)
        if err != nil {
            continue
        }

        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM store.inventory").Scan(&count)
        if err != nil {
            continue
        }

        fmt.Printf("   🔒 Tenant %d can see %d inventory items\n", tid, count)
    }
}

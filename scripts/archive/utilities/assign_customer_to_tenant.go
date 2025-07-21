package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func main() {
    customerID := flag.Int("customer_id", 0, "Customer ID (required)")
    tenantID := flag.Int("tenant_id", 0, "Tenant ID (required)")
    relationshipType := flag.String("type", "primary", "Relationship type (primary, secondary, billing_only)")
    assignedBy := flag.Int("assigned_by", 1, "User ID who is making the assignment")
    flag.Parse()

    if *customerID == 0 || *tenantID == 0 {
        fmt.Println("Usage: go run assign_customer_to_tenant.go -customer_id 1 -tenant_id 1 [options]")
        fmt.Println("Options:")
        fmt.Println("  -customer_id  Customer ID (required)")
        fmt.Println("  -tenant_id    Tenant ID (required)")
        fmt.Println("  -type         Relationship type (primary, secondary, billing_only)")
        fmt.Println("  -assigned_by  User ID making the assignment")
        os.Exit(1)
    }

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

    // Verify customer exists
    var customerName string
    err = db.QueryRow("SELECT customer FROM store.customers WHERE customer_id = $1", *customerID).Scan(&customerName)
    if err != nil {
        log.Fatalf("Customer ID %d not found: %v", *customerID, err)
    }

    // Verify tenant exists
    var tenantName string
    err = db.QueryRow("SELECT tenant_name FROM store.tenants WHERE tenant_id = $1", *tenantID).Scan(&tenantName)
    if err != nil {
        log.Fatalf("Tenant ID %d not found: %v", *tenantID, err)
    }

    // Assign customer to tenant
    query := `
        INSERT INTO store.customer_tenant_assignments (customer_id, tenant_id, relationship_type, assigned_by)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (customer_id, tenant_id) 
        DO UPDATE SET 
            relationship_type = EXCLUDED.relationship_type,
            assigned_by = EXCLUDED.assigned_by,
            assigned_at = CURRENT_TIMESTAMP,
            active = true
    `

    _, err = db.Exec(query, *customerID, *tenantID, *relationshipType, *assignedBy)
    if err != nil {
        log.Fatalf("Failed to assign customer to tenant: %v", err)
    }

    fmt.Printf("✅ Customer assigned to tenant successfully!\n")
    fmt.Printf("   Customer: %s (ID: %d)\n", customerName, *customerID)
    fmt.Printf("   Tenant: %s (ID: %d)\n", tenantName, *tenantID)
    fmt.Printf("   Relationship: %s\n", *relationshipType)
}

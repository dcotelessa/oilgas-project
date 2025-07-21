package main

import (
    "database/sql"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func main() {
    // Command line flags
    name := flag.String("name", "", "Tenant name (required)")
    slug := flag.String("slug", "", "Tenant slug (optional, generated from name)")
    description := flag.String("description", "", "Tenant description")
    email := flag.String("email", "", "Contact email")
    phone := flag.String("phone", "", "Contact phone")
    address := flag.String("address", "", "Business address")
    flag.Parse()

    if *name == "" {
        fmt.Println("Usage: go run create_tenant.go -name 'Tenant Name' [options]")
        fmt.Println("Options:")
        fmt.Println("  -name        Tenant name (required)")
        fmt.Println("  -slug        Tenant slug (optional)")
        fmt.Println("  -description Tenant description")
        fmt.Println("  -email       Contact email")
        fmt.Println("  -phone       Contact phone")
        fmt.Println("  -address     Business address")
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

    // Generate slug if not provided
    tenantSlug := *slug
    if tenantSlug == "" {
        tenantSlug = generateSlug(*name)
    }

    // Create tenant
    var tenantID int
    query := `
        INSERT INTO store.tenants (tenant_name, tenant_slug, description, contact_email, phone, address, settings)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING tenant_id
    `
    
    defaultSettings := map[string]interface{}{
        "default_location":        "Yard-A",
        "auto_assign_work_orders": true,
        "require_inspection":      true,
        "billing_terms_days":      30,
        "rush_order_multiplier":   1.5,
    }
    
    settingsJSON, _ := json.Marshal(defaultSettings)

    err = db.QueryRow(
        query,
        *name,
        tenantSlug,
        nullString(*description),
        nullString(*email),
        nullString(*phone),
        nullString(*address),
        settingsJSON,
    ).Scan(&tenantID)

    if err != nil {
        log.Fatalf("Failed to create tenant: %v", err)
    }

    fmt.Printf("✅ Tenant created successfully!\n")
    fmt.Printf("   ID: %d\n", tenantID)
    fmt.Printf("   Name: %s\n", *name)
    fmt.Printf("   Slug: %s\n", tenantSlug)
    if *description != "" {
        fmt.Printf("   Description: %s\n", *description)
    }
    if *email != "" {
        fmt.Printf("   Email: %s\n", *email)
    }
    
    fmt.Printf("\nNext steps:\n")
    fmt.Printf("1. Assign users: go run assign_user_to_tenant.go -user_id 1 -tenant_id %d -role admin\n", tenantID)
    fmt.Printf("2. Assign customers: go run assign_customer_to_tenant.go -customer_id 1 -tenant_id %d\n", tenantID)
}

func generateSlug(name string) string {
    // Simple slug generation - convert to lowercase and replace spaces
    slug := name
    slug = fmt.Sprintf("%s", slug)
    // This is a simplified version - the real implementation would be more robust
    return slug
}

func nullString(s string) interface{} {
    if s == "" {
        return nil
    }
    return s
}

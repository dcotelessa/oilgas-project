// backend/cmd/migrate/main.go - Migration utility example
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	
	"oilgas-backend/internal/customer"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: migrate <tenant_id> <legacy_data.json>")
		os.Exit(1)
	}
	
	tenantID := os.Args[1]
	filePath := os.Args[2]
	
	// Connect to central auth database (for customer migrations)
	centralAuthURL := os.Getenv("CENTRAL_AUTH_DB_URL")
	if centralAuthURL == "" {
		log.Fatal("CENTRAL_AUTH_DB_URL environment variable not set")
	}
	db, err := sql.Open("postgres", centralAuthURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	// Setup services
	repo := customer.NewRepository(db)
	cache := customer.NewInMemoryCache(time.Hour)
	authSvc := &mockAuthService{} // Replace with actual auth service
	
	customerSvc := customer.NewService(repo, authSvc, cache)
	migrationSvc := customer.NewCustomerMigrationService(customerSvc)
	
	// Load legacy data
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Failed to open legacy data file:", err)
	}
	defer file.Close()
	
	var legacyCustomers []customer.LegacyCustomer
	if err := json.NewDecoder(file).Decode(&legacyCustomers); err != nil {
		log.Fatal("Failed to parse legacy data:", err)
	}
	
	// Run migration
	ctx := context.Background()
	result, err := migrationSvc.BatchMigrateFromLegacy(ctx, tenantID, legacyCustomers)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	
	// Report results
	fmt.Printf("Migration completed in %v\n", result.Duration)
	fmt.Printf("Total: %d, Successful: %d, Failed: %d\n", result.Total, result.Successful, result.Failed)
	
	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  %s: %s\n", err.LegacyID, err.Error)
		}
	}
	
	fmt.Printf("\nSuccessfully migrated %d customers\n", result.Successful)
}

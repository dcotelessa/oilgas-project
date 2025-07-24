// scripts/utilities/database_validation.go
package main

import (
	"context"
	"fmt"
)

func ValidateDatabase() {
	fmt.Println("🗄️ Validating Database Schema and Connectivity...")
	
	pool := getDBConnection()
	defer pool.Close()
	
	ctx := context.Background()
	
	// Check core tables exist
	tables := []string{"customers", "inventory", "received", "grade", "sizes"}
	for _, table := range tables {
		var count int
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM store.%s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Error accessing store.%s: %v\n", table, err)
		} else {
			fmt.Printf("✅ Table store.%s exists with %d records\n", table, count)
		}
	}
	
	// Check for required indexes
	indexes := map[string]string{
		"idx_inventory_customer_id": "Performance index for customer lookups",
		"idx_inventory_work_order":  "Performance index for work order lookups",
		"idx_received_customer_id":  "Performance index for received items",
	}
	
	for indexName, description := range indexes {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE schemaname = 'store' AND indexname = $1
			)
		`, indexName).Scan(&exists)
		
		if err != nil {
			fmt.Printf("❌ Error checking index %s: %v\n", indexName, err)
		} else if exists {
			fmt.Printf("✅ Index %s exists (%s)\n", indexName, description)
		} else {
			fmt.Printf("⚠️  Index %s missing (%s)\n", indexName, description)
		}
	}
	
	fmt.Println("✅ Database validation complete!")
}

func ValidateRLS() {
	fmt.Println("🔐 Validating Row-Level Security...")
	
	pool := getDBConnection()
	defer pool.Close()
	
	ctx := context.Background()
	
	// Check RLS status on core tables
	tables := []string{"customers", "inventory", "received"}
	for _, table := range tables {
		var rlsEnabled bool
		err := pool.QueryRow(ctx, `
			SELECT COALESCE(relrowsecurity, false)
			FROM pg_class 
			WHERE relname = $1 AND relnamespace = (
				SELECT oid FROM pg_namespace WHERE nspname = 'store'
			)
		`, table).Scan(&rlsEnabled)
		
		if err != nil {
			fmt.Printf("❌ Error checking RLS for %s: %v\n", table, err)
			continue
		}
		
		if rlsEnabled {
			fmt.Printf("✅ RLS enabled on store.%s\n", table)
		} else {
			fmt.Printf("⚠️  RLS disabled on store.%s (single-tenant mode)\n", table)
		}
	}
	
	fmt.Println("📋 Note: Full RLS implementation planned for Step 10 (multi-tenant)")
	fmt.Println("✅ RLS validation complete!")
}

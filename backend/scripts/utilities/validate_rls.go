package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment
	godotenv.Load(".env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("🔐 Validating Row-Level Security Implementation...")
	
	ctx := context.Background()
	
	// Check if RLS is enabled on key tables
	tables := []string{"customers", "inventory", "received"}
	for _, table := range tables {
		var rlsEnabled bool
		err = pool.QueryRow(ctx, `
			SELECT relrowsecurity 
			FROM pg_class 
			WHERE relname = $1 AND relnamespace = (
				SELECT oid FROM pg_namespace WHERE nspname = 'store'
			)
		`, table).Scan(&rlsEnabled)
		
		if err != nil {
			fmt.Printf("❌ Error checking RLS status for %s: %v\n", table, err)
			continue
		}
		
		if rlsEnabled {
			fmt.Printf("✅ Row-Level Security is ENABLED on %s table\n", table)
		} else {
			fmt.Printf("❌ Row-Level Security is DISABLED on %s table\n", table)
		}
	}

	// Check tenant count
	var tenantCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM store.tenants WHERE is_active = true").Scan(&tenantCount)
	if err != nil {
		fmt.Printf("❌ Error counting tenants: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d active tenants\n", tenantCount)
	}

	// Check auth tables exist
	var userCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth.users WHERE deleted_at IS NULL").Scan(&userCount)
	if err != nil {
		fmt.Printf("❌ Error checking auth.users: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d users in auth system\n", userCount)
	}

	fmt.Println("✅ Row-Level Security validation complete!")
}

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
	godotenv.Load(".env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/oil_gas_inventory?sslmode=disable"
	}

	fmt.Printf("🔌 Connecting to: %s\n", databaseURL)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping: %v", err)
	}

	fmt.Println("✅ Connected successfully")

	// Test 1: List current schemas
	fmt.Println("\n📋 Current schemas:")
	rows, err := db.Query("SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
	if err != nil {
		fmt.Printf("❌ Failed to query schemas: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var schema string
			if err := rows.Scan(&schema); err != nil {
				fmt.Printf("❌ Failed to scan schema: %v\n", err)
			} else {
				fmt.Printf("  📁 %s\n", schema)
			}
		}
	}

	// Test 2: Try to create store schema
	fmt.Println("\n🔧 Attempting to create 'store' schema...")
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS store")
	if err != nil {
		fmt.Printf("❌ Failed to create store schema: %v\n", err)
		return
	}
	fmt.Println("✅ Store schema creation command executed")

	// Test 3: Verify store schema was created
	fmt.Println("\n🔍 Checking if store schema exists...")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = 'store'").Scan(&count)
	if err != nil {
		fmt.Printf("❌ Failed to check store schema: %v\n", err)
	} else {
		if count > 0 {
			fmt.Println("✅ Store schema confirmed to exist")
		} else {
			fmt.Println("❌ Store schema does not exist after creation attempt")
		}
	}

	// Test 4: List schemas again
	fmt.Println("\n📋 Schemas after creation attempt:")
	rows, err = db.Query("SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
	if err != nil {
		fmt.Printf("❌ Failed to query schemas: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var schema string
			if err := rows.Scan(&schema); err != nil {
				fmt.Printf("❌ Failed to scan schema: %v\n", err)
			} else {
				fmt.Printf("  📁 %s\n", schema)
			}
		}
	}

	// Test 5: Try to create a simple table in store schema
	fmt.Println("\n🔧 Attempting to create simple table in store schema...")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS store.simple_test (id SERIAL PRIMARY KEY, name VARCHAR(50))")
	if err != nil {
		fmt.Printf("❌ Failed to create table in store schema: %v\n", err)
	} else {
		fmt.Println("✅ Table creation in store schema succeeded")
	}

	// Test 6: Check if table exists
	fmt.Println("\n🔍 Checking if table was created...")
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'store' AND table_name = 'simple_test'").Scan(&count)
	if err != nil {
		fmt.Printf("❌ Failed to check table: %v\n", err)
	} else {
		if count > 0 {
			fmt.Println("✅ Table confirmed to exist in store schema")
		} else {
			fmt.Println("❌ Table does not exist in store schema")
		}
	}

	// Clean up
	fmt.Println("\n🧹 Cleaning up...")
	db.Exec("DROP TABLE IF EXISTS store.simple_test")
	db.Exec("DROP SCHEMA IF EXISTS store")
	
	fmt.Println("✅ Test complete")
}

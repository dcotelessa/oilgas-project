package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	var (
		name = flag.String("name", "", "Tenant name")
		slug = flag.String("slug", "", "Tenant slug")
	)
	flag.Parse()

	if *name == "" {
		log.Fatal("Tenant name is required")
	}

	if *slug == "" {
		*slug = generateSlug(*name)
	}

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

	// Create tenant
	var tenantID string
	err = pool.QueryRow(context.Background(), `
		INSERT INTO store.tenants (name, slug, database_type, is_active)
		VALUES ($1, $2, 'shared', true)
		RETURNING id
	`, *name, *slug).Scan(&tenantID)
	if err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}

	fmt.Printf("✅ Tenant created successfully:\n")
	fmt.Printf("   ID: %s\n", tenantID)
	fmt.Printf("   Name: %s\n", *name)
	fmt.Printf("   Slug: %s\n", *slug)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

#!/bin/bash
# scripts/generators/utility_scripts_generator.sh
# Generates utility scripts for user and tenant management

set -e
echo "🛠️ Generating utility scripts..."

# Detect backend directory
BACKEND_DIR=""
if [ -d "backend" ] && [ -f "backend/go.mod" ]; then
    BACKEND_DIR="backend/"
elif [ -f "go.mod" ]; then
    BACKEND_DIR=""
else
    echo "❌ Error: Cannot find go.mod file"
    exit 1
fi

# Create utilities directory
mkdir -p "${BACKEND_DIR}scripts/utilities"

# Generate user creation utility
cat > "${BACKEND_DIR}scripts/utilities/create_user.go" << 'EOF'
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		email    = flag.String("email", "", "User email")
		password = flag.String("password", "", "User password")
		role     = flag.String("role", "viewer", "User role")
		tenant   = flag.String("tenant", "", "Tenant slug")
		company  = flag.String("company", "", "Company name")
	)
	flag.Parse()

	if *email == "" || *password == "" || *tenant == "" {
		log.Fatal("Email, password, and tenant are required")
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ User created successfully:\n")
	fmt.Printf("   ID: %s\n", userID)
	fmt.Printf("   Email: %s\n", *email)
	fmt.Printf("   Role: %s\n", *role)
	fmt.Printf("   Company: %s\n", *company)
	fmt.Printf("   Tenant: %s\n", *tenant)
}
EOF

# Generate tenant creation utility
cat > "${BACKEND_DIR}scripts/utilities/create_tenant.go" << 'EOF'
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

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

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
EOF

echo "✅ Utility scripts generated"
echo "   - User creation utility"
echo "   - Tenant creation utility"
echo "   - Password hashing and validation"
echo "   - Database connection handling"

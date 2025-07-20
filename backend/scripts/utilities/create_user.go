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
		role     = flag.String("role", "viewer", "User role (admin, operator, customer_manager, customer_user, viewer)")
		tenant   = flag.String("tenant", "default", "Tenant slug")
		company  = flag.String("company", "", "Company name")
	)
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("Email and password are required")
	}

	// Load environment variables from multiple possible locations
	envFiles := []string{
		"../../.env.local",  // From backend/scripts/utilities/ to root
		"../.env.local",     // From backend/ to root
		".env.local",        // Current directory
		"../../.env",        // From backend/scripts/utilities/ to root
		"../.env",           // From backend/ to root
		".env",              // Current directory
	}
	
	loaded := false
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			log.Printf("Loaded environment from: %s", envFile)
			loaded = true
			break
		}
	}
	
	if !loaded {
		log.Printf("Warning: No .env file found, using system environment")
	}

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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Get tenant ID
	var tenantID string
	err = pool.QueryRow(context.Background(), 
		"SELECT id FROM store.tenants WHERE slug = $1", *tenant).Scan(&tenantID)
	if err != nil {
		log.Fatalf("Failed to find tenant '%s': %v", *tenant, err)
	}

	// Create user
	_, err = pool.Exec(context.Background(), `
		INSERT INTO auth.users (email, password_hash, role, company, tenant_id, email_verified)
		VALUES ($1, $2, $3, $4, $5, true)
	`, *email, string(hashedPassword), *role, *company, tenantID)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ User created successfully:\n")
	fmt.Printf("   Email: %s\n", *email)
	fmt.Printf("   Role: %s\n", *role)
	fmt.Printf("   Tenant: %s\n", *tenant)
	if *company != "" {
		fmt.Printf("   Company: %s\n", *company)
	}
}

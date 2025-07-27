// cmd/tools/user/create.go  
// User creation tool that works with your current migrator structure
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	_ "github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Company  string `json:"company"`
	TenantID string `json:"tenant_id"`
}

func main() {
	var (
		email    = flag.String("email", "", "User email")
		password = flag.String("password", "", "User password (will prompt if not provided)")
		role     = flag.String("role", "user", "User role (user, operator, manager, admin, super-admin)")
		company  = flag.String("company", "", "Company name")
		tenantID = flag.String("tenant", "", "Tenant ID")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Validate required fields
	if *email == "" {
		fmt.Print("Email: ")
		fmt.Scanln(email)
	}

	if *company == "" {
		fmt.Print("Company: ")
		fmt.Scanln(company)
	}

	if *tenantID == "" {
		fmt.Print("Tenant ID: ")
		fmt.Scanln(tenantID)
	}

	// Get password securely
	if *password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
		}
		fmt.Println() // New line after password input
		*password = string(passwordBytes)
	}

	// Validate inputs
	if err := validateInputs(*email, *password, *role, *company, *tenantID); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	// Connect to database
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Create user
	ctx := context.Background()
	user, err := createUser(ctx, db, *email, *password, *role, *company, *tenantID)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ User created successfully!\n")
	fmt.Printf("   ID: %s\n", user.ID)
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Role: %s\n", user.Role)
	fmt.Printf("   Company: %s\n", user.Company)
	fmt.Printf("   Tenant: %s\n", user.TenantID)
}

// createUser creates a new user (simplified service layer)
func createUser(ctx context.Context, db *sql.DB, email, password, role, company, tenantID string) (*User, error) {
	// Ensure auth schema exists first
	if err := ensureAuthSchema(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to ensure auth schema: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if user already exists
	var existingID string
	err = db.QueryRowContext(ctx, 
		"SELECT id FROM auth.users WHERE email = $1", 
		email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			return nil, fmt.Errorf("user with email %s already exists", email)
		}
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Ensure tenant exists (create if super-admin)
	if err := ensureTenant(ctx, db, tenantID, company, role); err != nil {
		return nil, fmt.Errorf("failed to ensure tenant: %w", err)
	}

	// Create user in transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user
	var userID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO auth.users (email, password_hash, role, company, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id`,
		email, string(passwordHash), role, company, tenantID,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &User{
		ID:       userID,
		Email:    email,
		Role:     role,
		Company:  company,
		TenantID: tenantID,
	}, nil
}

// ensureAuthSchema creates auth schema if it doesn't exist
func ensureAuthSchema(ctx context.Context, db *sql.DB) error {
	// Check if auth schema exists
	var exists bool
	err := db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'auth')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check auth schema: %w", err)
	}

	if exists {
		return nil // Schema already exists
	}

	fmt.Println("🔄 Creating auth schema...")

	// Create auth schema with basic tables
	authSchema := `
	-- Create auth schema
	CREATE SCHEMA auth;

	-- Create tenants table
	CREATE TABLE auth.tenants (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) NOT NULL UNIQUE,
		database_type VARCHAR(50) DEFAULT 'tenant',
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Create users table
	CREATE TABLE auth.users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		company VARCHAR(255) NOT NULL,
		tenant_id VARCHAR(100) NOT NULL,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		
		CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug)
	);

	-- Create indexes
	CREATE INDEX idx_users_email ON auth.users(email);
	CREATE INDEX idx_users_tenant_id ON auth.users(tenant_id);
	CREATE INDEX idx_tenants_slug ON auth.tenants(slug);

	-- Insert default system tenant
	INSERT INTO auth.tenants (name, slug, database_type) 
	VALUES ('System Administration', 'system', 'main');`

	if _, err := db.ExecContext(ctx, authSchema); err != nil {
		return fmt.Errorf("failed to create auth schema: %w", err)
	}

	fmt.Println("✅ Auth schema created")
	return nil
}

// ensureTenant creates tenant if it doesn't exist (for super-admin)
func ensureTenant(ctx context.Context, db *sql.DB, tenantID, company, role string) error {
	// Check if tenant exists
	var exists bool
	err := db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE slug = $1)", 
		tenantID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}

	if exists {
		return nil // Tenant already exists
	}

	// Create tenant if super-admin or system tenant
	if role == "super-admin" || tenantID == "system" {
		_, err := db.ExecContext(ctx, `
			INSERT INTO auth.tenants (name, slug, database_type, created_at, updated_at)
			VALUES ($1, $2, 'main', NOW(), NOW())`,
			company, tenantID,
		)
		if err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		fmt.Printf("✅ Created tenant: %s (%s)\n", company, tenantID)
	} else {
		return fmt.Errorf("tenant '%s' does not exist. Create it first or use super-admin role", tenantID)
	}

	return nil
}

// connectDatabase establishes database connection
func connectDatabase() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// validateInputs validates user input
func validateInputs(email, password, role, company, tenantID string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	validRoles := []string{"user", "operator", "manager", "admin", "super-admin"}
	roleValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return fmt.Errorf("invalid role. Valid roles: %s", strings.Join(validRoles, ", "))
	}

	if company == "" {
		return fmt.Errorf("company is required")
	}

	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate tenant ID format (lowercase, no spaces)
	if strings.ToLower(tenantID) != tenantID || strings.Contains(tenantID, " ") {
		return fmt.Errorf("tenant ID must be lowercase with no spaces")
	}

	return nil
}

// showHelp displays usage information
func showHelp() {
	fmt.Println("Create User Tool - Oil & Gas Inventory System")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/tools/user/create.go [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create super-admin")
	fmt.Println("  go run cmd/tools/user/create.go \\")
	fmt.Println("    --email=admin@company.com \\")
	fmt.Println("    --role=super-admin \\")
	fmt.Println("    --company=\"Your Company\" \\")
	fmt.Println("    --tenant=system")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - DATABASE_URL environment variable must be set")
	fmt.Println("  - Creates auth schema automatically if needed")
}

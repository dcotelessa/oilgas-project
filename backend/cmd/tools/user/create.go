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
		email       = flag.String("email", "", "User email")
		password    = flag.String("password", "", "User password (will prompt if not provided)")
		role        = flag.String("role", "user", "User role (user, operator, manager, admin, super-admin)")
		company     = flag.String("company", "", "Company name")
		tenantID    = flag.String("tenant", "", "Tenant ID")
		bootstrap   = flag.Bool("bootstrap", false, "Create bootstrap admin with access to all tenants")
		help        = flag.Bool("help", false, "Show help")
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

	if *tenantID == "" && !*bootstrap {
		fmt.Print("Tenant ID: ")
		fmt.Scanln(tenantID)
	}

	// Get password securely
	if *password == "" {
		if *bootstrap {
			// For bootstrap, use a default secure password
			*password = "PetrosAdmin2024!"
			fmt.Println("Using default bootstrap password: PetrosAdmin2024!")
		} else {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatalf("Failed to read password: %v", err)
			}
			fmt.Println() // New line after password input
			*password = string(passwordBytes)
		}
	}

	// Validate inputs
	if *bootstrap {
		if err := validateBootstrapInputs(*email, *password, *company); err != nil {
			log.Fatalf("Validation error: %v", err)
		}
	} else {
		if err := validateInputs(*email, *password, *role, *company, *tenantID); err != nil {
			log.Fatalf("Validation error: %v", err)
		}
	}

	// Connect to database  
	db, err := connectDatabase()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Create user
	ctx := context.Background()
	var user *User
	if *bootstrap {
		user, err = createBootstrapAdmin(ctx, db, *email, *password, *company)
	if err != nil {
			log.Fatalf("Failed to create bootstrap admin: %v", err)
		}
	} else {
		user, err = createUser(ctx, db, *email, *password, *role, *company, *tenantID)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}

	if *bootstrap {
		fmt.Printf("🎉 Bootstrap admin created successfully!\n")
		fmt.Printf("   ID: %s\n", user.ID)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Role: %s\n", user.Role)
		fmt.Printf("   Company: %s\n", user.Company)
		fmt.Printf("   Primary Tenant: %s\n", user.TenantID)
		fmt.Printf("   Multi-tenant Access: longbeach, bakersfield, colorado\n")
	} else {
		fmt.Printf("✅ User created successfully!\n")
		fmt.Printf("   ID: %s\n", user.ID)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Role: %s\n", user.Role)
		fmt.Printf("   Company: %s\n", user.Company)
		fmt.Printf("   Tenant: %s\n", user.TenantID)
	}
}

// createBootstrapAdmin creates a bootstrap admin with access to all tenants
func createBootstrapAdmin(ctx context.Context, db *sql.DB, email, password, company string) (*User, error) {
	// Ensure auth schema exists first
	if err := ensureAuthSchema(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to ensure auth schema: %w", err)
	}

	// Ensure all tenants exist
	if err := ensureAllTenants(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to ensure tenants: %w", err)
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

	// Create admin in transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build tenant access JSON for super-admin
	tenantAccessJSON := `[
		{"tenant_id": "longbeach", "role": "ADMIN", "permissions": ["VIEW_INVENTORY", "MANAGE_CUSTOMERS", "MANAGE_USERS"], "yard_access": [], "can_read": true, "can_write": true, "can_delete": true, "can_approve": true},
		{"tenant_id": "bakersfield", "role": "ADMIN", "permissions": ["VIEW_INVENTORY", "MANAGE_CUSTOMERS", "MANAGE_USERS"], "yard_access": [], "can_read": true, "can_write": true, "can_delete": true, "can_approve": true},
		{"tenant_id": "colorado", "role": "ADMIN", "permissions": ["VIEW_INVENTORY", "MANAGE_CUSTOMERS", "MANAGE_USERS"], "yard_access": [], "can_read": true, "can_write": true, "can_delete": true, "can_approve": true}
	]`

	// Insert admin user with system as primary tenant
	var userID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO auth.users (
			username, email, full_name, password_hash, role, 
			access_level, is_enterprise_user, tenant_access, primary_tenant_id, 
			contact_type, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id`,
		email, // username = email for super-admin
		email, // email
		company + " Administrator", // full_name
		string(passwordHash), // password_hash
		"SYSTEM_ADMIN", // role
		10, // access_level (highest)
		true, // is_enterprise_user
		tenantAccessJSON, // tenant_access (multi-tenant permissions)
		"system", // primary_tenant_id
		"PRIMARY", // contact_type (default for system admin)
		true, // is_active
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert admin user: %w", err)
	}

	// Grant access to all operational tenants
	tenantSlugs := []string{"longbeach", "bakersfield", "colorado"}
	for _, tenantSlug := range tenantSlugs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO auth.user_tenant_access (user_id, tenant_id, role, is_active, created_at, updated_at)
			VALUES ($1, $2, 'admin', true, NOW(), NOW())`,
			userID, tenantSlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to grant access to tenant %s: %w", tenantSlug, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &User{
		ID:       fmt.Sprintf("%d", userID),
		Email:    email,
		Role:     "SYSTEM_ADMIN",
		Company:  company,
		TenantID: "system",
	}, nil
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
	var userID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO auth.users (username, email, full_name, password_hash, role, access_level, is_enterprise_user, primary_tenant_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id`,
		email, email, role + " User", string(passwordHash), role, 5, false, tenantID, true,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &User{
		ID:       fmt.Sprintf("%d", userID),
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
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Create users table (full schema matching auth repository)
	CREATE TABLE auth.users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255),
		email VARCHAR(255) NOT NULL UNIQUE,
		full_name VARCHAR(255),
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		access_level INTEGER DEFAULT 1,
		is_enterprise_user BOOLEAN DEFAULT false,
		tenant_access JSONB,
		primary_tenant_id VARCHAR(100) NOT NULL,
		customer_id INTEGER,
		contact_type VARCHAR(50),
		is_active BOOLEAN DEFAULT true,
		last_login_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		
		CONSTRAINT fk_users_tenant FOREIGN KEY (primary_tenant_id) REFERENCES auth.tenants(slug)
	);

	-- Create user_tenant_access table for multi-tenant permissions
	CREATE TABLE auth.user_tenant_access (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		tenant_id VARCHAR(100) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		permissions TEXT[],
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		
		CONSTRAINT fk_user_tenant_access_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
		CONSTRAINT fk_user_tenant_access_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug) ON DELETE CASCADE,
		CONSTRAINT uk_user_tenant_access UNIQUE (user_id, tenant_id)
	);

	-- Create sessions table for JWT token management
	CREATE TABLE auth.sessions (
		id VARCHAR(255) PRIMARY KEY,
		user_id INTEGER NOT NULL,
		tenant_id VARCHAR(100) NOT NULL,
		token TEXT NOT NULL,
		refresh_token TEXT,
		tenant_context JSONB,
		is_active BOOLEAN DEFAULT true,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		refresh_expires_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		user_agent TEXT DEFAULT '',
		ip_address VARCHAR(45) DEFAULT '',
		
		CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
		CONSTRAINT fk_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES auth.tenants(slug) ON DELETE CASCADE
	);

	-- Create indexes
	CREATE INDEX idx_users_email ON auth.users(email);
	CREATE INDEX idx_users_primary_tenant_id ON auth.users(primary_tenant_id);
	CREATE INDEX idx_users_active ON auth.users(is_active);
	CREATE INDEX idx_tenants_slug ON auth.tenants(slug);
	CREATE INDEX idx_tenants_active ON auth.tenants(is_active);
	CREATE INDEX idx_user_tenant_access_user ON auth.user_tenant_access(user_id);
	CREATE INDEX idx_user_tenant_access_tenant ON auth.user_tenant_access(tenant_id);
	CREATE INDEX idx_user_tenant_access_active ON auth.user_tenant_access(is_active);
	CREATE INDEX idx_sessions_user ON auth.sessions(user_id);
	CREATE INDEX idx_sessions_token ON auth.sessions(token);
	CREATE INDEX idx_sessions_expires_at ON auth.sessions(expires_at);

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

// ensureAllTenants creates all required tenant entries
func ensureAllTenants(ctx context.Context, db *sql.DB) error {
	tenants := []struct {
		Name         string
		Slug         string
		DatabaseType string
	}{
		{"System Administration", "system", "main"},
		{"Long Beach Operations", "longbeach", "tenant"},
		{"Bakersfield Operations", "bakersfield", "tenant"},
		{"Colorado Operations", "colorado", "tenant"},
	}

	for _, tenant := range tenants {
		// Check if tenant exists
		var exists bool
		err := db.QueryRowContext(ctx, 
			"SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE slug = $1)", 
			tenant.Slug).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check tenant %s: %w", tenant.Slug, err)
		}

		if exists {
			continue
		}

		// Create tenant
		_, err = db.ExecContext(ctx, `
			INSERT INTO auth.tenants (name, slug, database_type, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())`,
			tenant.Name, tenant.Slug, tenant.DatabaseType,
		)
		if err != nil {
			return fmt.Errorf("failed to create tenant %s: %w", tenant.Slug, err)
		}
		fmt.Printf("✅ Created tenant: %s (%s)\n", tenant.Name, tenant.Slug)
	}

	return nil
}

// connectDatabase establishes database connection
func connectDatabase() (*sql.DB, error) {
	// Prefer DEV URL for local development
	databaseURL := os.Getenv("DEV_CENTRAL_AUTH_DB_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("CENTRAL_AUTH_DB_URL")
		if databaseURL == "" {
			// Fallback to DATABASE_URL for backwards compatibility
			databaseURL = os.Getenv("DATABASE_URL")
			if databaseURL == "" {
				return nil, fmt.Errorf("DEV_CENTRAL_AUTH_DB_URL, CENTRAL_AUTH_DB_URL, or DATABASE_URL environment variable must be set")
			}
		}
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

// validateBootstrapInputs validates bootstrap admin input
func validateBootstrapInputs(email, password, company string) error {
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

	if company == "" {
		return fmt.Errorf("company is required")
	}

	return nil
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
	fmt.Println("  # Create bootstrap admin (recommended for initial setup)")
	fmt.Println("  go run cmd/tools/user/create.go \\")
	fmt.Println("    --bootstrap \\")
	fmt.Println("    --email=admin@company.com \\")
	fmt.Println("    --company=\"Your Company\"")
	fmt.Println()
	fmt.Println("  # Create regular user")
	fmt.Println("  go run cmd/tools/user/create.go \\")
	fmt.Println("    --email=user@company.com \\")
	fmt.Println("    --role=operator \\")
	fmt.Println("    --company=\"Your Company\" \\")
	fmt.Println("    --tenant=longbeach")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - CENTRAL_AUTH_DB_URL environment variable must be set")
	fmt.Println("  - Creates auth schema and tenants automatically if needed")
	fmt.Println("  - Bootstrap creates admin with access to all tenants")
}

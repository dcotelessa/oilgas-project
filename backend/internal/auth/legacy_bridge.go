// backend/internal/auth/legacy_bridge.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// LegacyAuthBridge handles migration from CFM EZPASSWORD system
// NOTE: ODBC dependencies removed for development - add back when needed for production migration
type LegacyAuthBridge struct {
	legacyDB    *sql.DB // Connection to legacy database (placeholder)
	newAuthRepo Repository
	service     Service
	dryRun      bool // If true, only log what would be done
}

// NewLegacyAuthBridge creates a new legacy bridge instance
// NOTE: ODBC connection disabled for development
func NewLegacyAuthBridge(legacyDSN string, newAuthRepo Repository, service Service) (*LegacyAuthBridge, error) {
	// For development: Don't attempt ODBC connection
	log.Printf("Legacy bridge created in development mode (ODBC disabled)")
	log.Printf("Legacy DSN would be: %s", legacyDSN)

	return &LegacyAuthBridge{
		legacyDB:    nil, // No actual connection in dev mode
		newAuthRepo: newAuthRepo,
		service:     service,
		dryRun:      true, // Default to dry run in dev mode
	}, nil
}

// Close closes the legacy database connection
func (b *LegacyAuthBridge) Close() error {
	if b.legacyDB != nil {
		return b.legacyDB.Close()
	}
	return nil
}

// SetDryRun enables/disables dry run mode
func (b *LegacyAuthBridge) SetDryRun(dryRun bool) {
	b.dryRun = dryRun
}

// ============================================================================
// CFM EZPASSWORD MODELS
// ============================================================================

// CFMUser represents a user from the CFM EZPASSWORD table
type CFMUser struct {
	USERNAME     string         `db:"USERNAME"`
	PASSWORD     string         `db:"PASSWORD"`      // Plain text password in CFM
	FULLNAME     string         `db:"FULLNAME"`
	ACCESS       int            `db:"ACCESS"`        // CFM access level (1-5+)
	EMAIL        sql.NullString `db:"EMAIL"`         // May not exist in all CFM systems
	ACTIVE       sql.NullBool   `db:"ACTIVE"`        // May not exist
	LASTLOGIN    sql.NullTime   `db:"LASTLOGIN"`     // May not exist
	CREATEDATE   sql.NullTime   `db:"CREATEDATE"`    // May not exist
	DEPARTMENT   sql.NullString `db:"DEPARTMENT"`    // May not exist
	LOCATION     sql.NullString `db:"LOCATION"`      // May indicate tenant/yard
}

// MigrationReport provides summary of migration results
type MigrationReport struct {
	TotalCFMUsers    int                    `json:"total_cfm_users"`
	MigratedUsers    int                    `json:"migrated_users"`
	SkippedUsers     int                    `json:"skipped_users"`
	ErrorUsers       int                    `json:"error_users"`
	UserDetails      []MigrationUserDetail  `json:"user_details"`
	TenantMappings   map[string]string      `json:"tenant_mappings"`
	RoleMappings     map[int]string         `json:"role_mappings"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
}

// MigrationUserDetail provides details about each migrated user
type MigrationUserDetail struct {
	CFMUsername    string    `json:"cfm_username"`
	CFMAccess      int       `json:"cfm_access"`
	NewUserID      int       `json:"new_user_id,omitempty"`
	NewRole        UserRole  `json:"new_role"`
	TenantID       string    `json:"tenant_id"`
	Status         string    `json:"status"` // "migrated", "skipped", "error"
	ErrorMessage   string    `json:"error_message,omitempty"`
	MigrationTime  time.Time `json:"migration_time"`
}

// CFMUserStats provides statistics about CFM users
type CFMUserStats struct {
	TotalUsers           int         `json:"total_users"`
	ActiveUsers          int         `json:"active_users"`
	UsersWithEmail       int         `json:"users_with_email"`
	UsersByAccessLevel   map[int]int `json:"users_by_access_level"`
}

// ============================================================================
// DEVELOPMENT MODE SIMULATION FUNCTIONS
// ============================================================================

// MigrateCFMUsers simulates CFM user migration for development
func (b *LegacyAuthBridge) MigrateCFMUsers(ctx context.Context, tenantMappings map[string]string) (*MigrationReport, error) {
	startTime := time.Now()
	
	log.Printf("Starting CFM user migration simulation (development mode)")
	
	report := &MigrationReport{
		UserDetails:    []MigrationUserDetail{},
		TenantMappings: tenantMappings,
		RoleMappings:   make(map[int]string),
		StartTime:      startTime,
	}

	// Build role mappings for report
	for i := 1; i <= 6; i++ {
		role := MapCFMAccessToRole(CFMAccessLevel(i))
		report.RoleMappings[i] = string(role)
	}

	// Simulate some sample CFM users for development
	sampleCFMUsers := b.generateSampleCFMUsers()
	report.TotalCFMUsers = len(sampleCFMUsers)
	
	log.Printf("Simulating migration of %d sample CFM users", len(sampleCFMUsers))

	// Simulate migration of each user
	for _, cfmUser := range sampleCFMUsers {
		detail := b.simulateMigrateUser(ctx, cfmUser, tenantMappings)
		report.UserDetails = append(report.UserDetails, detail)
		
		switch detail.Status {
		case "migrated":
			report.MigratedUsers++
		case "skipped":
			report.SkippedUsers++
		case "error":
			report.ErrorUsers++
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	log.Printf("Migration simulation completed: %d migrated, %d skipped, %d errors in %v",
		report.MigratedUsers, report.SkippedUsers, report.ErrorUsers, report.Duration)

	return report, nil
}

// generateSampleCFMUsers creates sample CFM users for development testing
func (b *LegacyAuthBridge) generateSampleCFMUsers() []CFMUser {
	return []CFMUser{
		{
			USERNAME: "cfm_admin",
			PASSWORD: "admin123",
			FULLNAME: "CFM Administrator",
			ACCESS:   5,
			EMAIL:    sql.NullString{String: "admin@cfm.com", Valid: true},
			ACTIVE:   sql.NullBool{Bool: true, Valid: true},
			LOCATION: sql.NullString{String: "HOUSTON", Valid: true},
		},
		{
			USERNAME: "cfm_manager",
			PASSWORD: "manager123",
			FULLNAME: "CFM Site Manager",
			ACCESS:   3,
			EMAIL:    sql.NullString{String: "manager@cfm.com", Valid: true},
			ACTIVE:   sql.NullBool{Bool: true, Valid: true},
			LOCATION: sql.NullString{String: "LB", Valid: true},
		},
		{
			USERNAME: "cfm_operator",
			PASSWORD: "operator123",
			FULLNAME: "CFM Operator",
			ACCESS:   2,
			EMAIL:    sql.NullString{String: "operator@cfm.com", Valid: true},
			ACTIVE:   sql.NullBool{Bool: true, Valid: true},
			LOCATION: sql.NullString{String: "HOUSTON", Valid: true},
		},
		{
			USERNAME: "cfm_inactive",
			PASSWORD: "inactive123",
			FULLNAME: "Inactive CFM User",
			ACCESS:   1,
			EMAIL:    sql.NullString{String: "inactive@cfm.com", Valid: true},
			ACTIVE:   sql.NullBool{Bool: false, Valid: true},
			LOCATION: sql.NullString{String: "HOUSTON", Valid: true},
		},
	}
}

// simulateMigrateUser simulates migrating a single CFM user
func (b *LegacyAuthBridge) simulateMigrateUser(ctx context.Context, cfmUser CFMUser, tenantMappings map[string]string) MigrationUserDetail {
	detail := MigrationUserDetail{
		CFMUsername:   cfmUser.USERNAME,
		CFMAccess:     cfmUser.ACCESS,
		MigrationTime: time.Now(),
	}

	// Determine tenant from location or use default
	tenantID := b.determineTenant(cfmUser, tenantMappings)
	detail.TenantID = tenantID

	// Map CFM access level to new role
	role := MapCFMAccessToRole(CFMAccessLevel(cfmUser.ACCESS))
	detail.NewRole = role

	// Check if user already exists (simulate check)
	existingUser, err := b.newAuthRepo.GetUserByUsername(ctx, cfmUser.USERNAME)
	if err == nil && existingUser != nil {
		detail.Status = "skipped"
		detail.ErrorMessage = "User already exists"
		log.Printf("Simulating skip of existing user: %s", cfmUser.USERNAME)
		return detail
	}

	// Skip inactive users
	if cfmUser.ACTIVE.Valid && !cfmUser.ACTIVE.Bool {
		detail.Status = "skipped"
		detail.ErrorMessage = "User is inactive in CFM"
		log.Printf("Simulating skip of inactive user: %s", cfmUser.USERNAME)
		return detail
	}

	// Simulate creating new user
	if b.dryRun {
		detail.Status = "migrated"
		detail.NewUserID = 0 // No actual ID in dry run
		log.Printf("SIMULATION: Would migrate user %s (access %d) as %s in tenant %s",
			cfmUser.USERNAME, cfmUser.ACCESS, role, tenantID)
	} else {
		// Actually create the user if not in dry run
		newUser, err := b.createNewUser(ctx, cfmUser, role, tenantID)
		if err != nil {
			detail.Status = "error"
			detail.ErrorMessage = err.Error()
			log.Printf("Error migrating user %s: %v", cfmUser.USERNAME, err)
		} else {
			detail.Status = "migrated"
			detail.NewUserID = newUser.ID
			log.Printf("Migrated user %s (ID: %d) as %s in tenant %s",
				cfmUser.USERNAME, newUser.ID, role, tenantID)
		}
	}

	return detail
}

// ============================================================================
// VALIDATION FUNCTIONS (DEVELOPMENT MODE)
// ============================================================================

// ValidateCFMDatabase simulates CFM database validation
func (b *LegacyAuthBridge) ValidateCFMDatabase(ctx context.Context) error {
	log.Printf("CFM database validation simulated (development mode)")
	log.Printf("In production, this would validate EZPASSWORD table accessibility")
	return nil
}

// GetCFMUserStats provides simulated statistics about CFM users
func (b *LegacyAuthBridge) GetCFMUserStats(ctx context.Context) (*CFMUserStats, error) {
	// Return simulated statistics for development
	stats := &CFMUserStats{
		TotalUsers:         4,
		ActiveUsers:        3,
		UsersWithEmail:     4,
		UsersByAccessLevel: map[int]int{1: 1, 2: 1, 3: 1, 5: 1},
	}

	log.Printf("CFM user stats simulation: %d total users, %d active", stats.TotalUsers, stats.ActiveUsers)
	return stats, nil
}

// ============================================================================
// AUTHENTICATION FUNCTIONS (DEVELOPMENT PLACEHOLDERS)
// ============================================================================

// AuthenticateViaCFM simulates CFM authentication for development
func (b *LegacyAuthBridge) AuthenticateViaCFM(ctx context.Context, username, password string) (*CFMUser, error) {
	log.Printf("CFM authentication simulation for user: %s", username)
	
	// Simulate checking against sample users
	sampleUsers := b.generateSampleCFMUsers()
	for _, user := range sampleUsers {
		if user.USERNAME == username {
			if user.PASSWORD == password {
				if user.ACTIVE.Valid && !user.ACTIVE.Bool {
					return nil, ErrUserInactive
				}
				return &user, nil
			}
			return nil, ErrInvalidCredentials
		}
	}
	
	return nil, ErrUserNotFound
}

// SyncUserFromCFM simulates user synchronization for development
func (b *LegacyAuthBridge) SyncUserFromCFM(ctx context.Context, username string) error {
	log.Printf("CFM user sync simulation for: %s", username)
	log.Printf("In production, this would sync user data from CFM EZPASSWORD")
	return nil
}

// ============================================================================
// UTILITY FUNCTIONS (SHARED BETWEEN DEV AND PROD)
// ============================================================================

// determineTenant determines which tenant a CFM user belongs to
func (b *LegacyAuthBridge) determineTenant(cfmUser CFMUser, tenantMappings map[string]string) string {
	// Try to determine from LOCATION field first
	if cfmUser.LOCATION.Valid && cfmUser.LOCATION.String != "" {
		location := strings.ToUpper(strings.TrimSpace(cfmUser.LOCATION.String))
		
		// Check tenant mappings
		for pattern, tenantID := range tenantMappings {
			if strings.Contains(location, strings.ToUpper(pattern)) {
				return tenantID
			}
		}
	}

	// Try to determine from DEPARTMENT field
	if cfmUser.DEPARTMENT.Valid && cfmUser.DEPARTMENT.String != "" {
		department := strings.ToUpper(strings.TrimSpace(cfmUser.DEPARTMENT.String))
		
		// Check tenant mappings
		for pattern, tenantID := range tenantMappings {
			if strings.Contains(department, strings.ToUpper(pattern)) {
				return tenantID
			}
		}
	}

	// Try to determine from username prefix (e.g., "LB_username" for Long Beach)
	username := strings.ToUpper(cfmUser.USERNAME)
	for pattern, tenantID := range tenantMappings {
		if strings.HasPrefix(username, strings.ToUpper(pattern)+"_") {
			return tenantID
		}
	}

	// Default tenant (first one in mappings, or "default")
	for _, tenantID := range tenantMappings {
		return tenantID
	}

	return "local-dev" // Development default
}

// createNewUser creates a new user in the auth system
func (b *LegacyAuthBridge) createNewUser(ctx context.Context, cfmUser CFMUser, role UserRole, tenantID string) (*User, error) {
	if b.dryRun {
		return &User{ID: 0}, nil // Return dummy user for dry run
	}

	// Hash the CFM password (which is stored in plain text)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfmUser.PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine email
	email := cfmUser.USERNAME + "@migrated.local" // Default email for migrated users
	if cfmUser.EMAIL.Valid && cfmUser.EMAIL.String != "" {
		email = cfmUser.EMAIL.String
	}

	// Create user based on role type
	if role == RoleCustomerContact {
		// For customer contacts, we need to link to a customer
		return b.createCustomerContactUser(ctx, cfmUser, hashedPassword, email, tenantID)
	} else {
		// Create enterprise user
		return b.createEnterpriseUser(ctx, cfmUser, role, hashedPassword, email, tenantID)
	}
}

// createEnterpriseUser creates an enterprise user
func (b *LegacyAuthBridge) createEnterpriseUser(ctx context.Context, cfmUser CFMUser, role UserRole, hashedPassword []byte, email, tenantID string) (*User, error) {
	// Build tenant access
	tenantAccess := []TenantAccess{
		{
			TenantID:    tenantID,
			Role:        role,
			Permissions: GetPermissionsForRole(role),
		},
	}

	// Enterprise admins get cross-tenant access
	if role == RoleEnterpriseAdmin {
		tenantAccess[0].Permissions = append(tenantAccess[0].Permissions, PermissionCrossTenantView)
	}

	req := CreateEnterpriseUserRequest{
		Username:        cfmUser.USERNAME,
		Email:           email,
		FullName:        cfmUser.FULLNAME,
		Password:        string(hashedPassword),
		Role:            role,
		IsEnterpriseUser: role == RoleEnterpriseAdmin,
		PrimaryTenantID: tenantID,
		TenantAccess:    tenantAccess,
	}

	return b.service.CreateEnterpriseUser(ctx, req)
}

// createCustomerContactUser creates a customer contact user
func (b *LegacyAuthBridge) createCustomerContactUser(ctx context.Context, cfmUser CFMUser, hashedPassword []byte, email, tenantID string) (*User, error) {
	// For customer contacts, we need to determine the customer ID
	// In development, use a default customer ID
	customerID := 1 // Default customer for development

	req := CreateCustomerContactRequest{
		CustomerID:  customerID,
		TenantID:    tenantID,
		Email:       email,
		FullName:    cfmUser.FULLNAME,
		Password:    string(hashedPassword),
		ContactType: ContactPrimary,
		YardAccess: []YardAccess{
			{
				YardLocation:       tenantID + "_main",
				CanViewWorkOrders:  true,
				CanViewInventory:   true,
				CanCreateWorkOrders: false,
			},
		},
	}

	return b.service.CreateCustomerContact(ctx, req)
}

// ============================================================================
// MIGRATION UTILITIES
// ============================================================================

// GenerateTenantMappings creates default tenant mappings for development
func GenerateTenantMappings() map[string]string {
	return map[string]string{
		"LB":        "longbeach",
		"HOUSTON":   "houston",
		"GALVESTON": "galveston",
		"CORPUS":    "corpus",
		"ADMIN":     "enterprise",
		"DEFAULT":   "local-dev",
	}
}

// MigrationOptions provides options for migration
type MigrationOptions struct {
	DryRun           bool              `json:"dry_run"`
	TenantMappings   map[string]string `json:"tenant_mappings"`
	SkipInactive     bool              `json:"skip_inactive"`
	SkipExisting     bool              `json:"skip_existing"`
	DefaultPassword  string            `json:"default_password"`
	ReportFile       string            `json:"report_file"`
}

// RunMigrationWithOptions runs migration simulation with specified options
func (b *LegacyAuthBridge) RunMigrationWithOptions(ctx context.Context, options MigrationOptions) (*MigrationReport, error) {
	b.SetDryRun(options.DryRun)
	
	// Use default tenant mappings if none provided
	tenantMappings := options.TenantMappings
	if len(tenantMappings) == 0 {
		tenantMappings = GenerateTenantMappings()
	}

	// Run migration simulation
	report, err := b.MigrateCFMUsers(ctx, tenantMappings)
	if err != nil {
		return nil, err
	}

	// Log report summary
	log.Printf("Migration simulation summary: %s", report.GetMigrationSummary())

	return report, nil
}

// GetMigrationSummary returns a human-readable summary
func (report *MigrationReport) GetMigrationSummary() string {
	successRate := float64(report.MigratedUsers) / float64(report.TotalCFMUsers) * 100
	
	return fmt.Sprintf(`
CFM User Migration Summary
=========================
Total CFM Users: %d
Successfully Migrated: %d
Skipped: %d
Errors: %d
Success Rate: %.1f%%
Duration: %v
`,
		report.TotalCFMUsers,
		report.MigratedUsers,
		report.SkippedUsers,
		report.ErrorUsers,
		successRate,
		report.Duration,
	)
}

// backend/internal/auth/service.go
package auth

import (
	"context"
	"errors"
	"time"
	
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
	// Import models from centralized location
	"oilgas-backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found") 
	ErrUserExists        = errors.New("user already exists")
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrSessionExpired    = errors.New("session expired")
)

// Service handles authentication operations using centralized models
type Service struct {
	// In-memory stores for testing - will be replaced with database repo
	users    map[string]*models.User    // email -> user
	tenants  map[string]*models.Tenant  // slug -> tenant  
	sessions map[string]*models.Session // session_id -> session
}

// NewService creates a new auth service
func NewService() *Service {
	return &Service{
		users:    make(map[string]*models.User),
		tenants:  make(map[string]*models.Tenant),
		sessions: make(map[string]*models.Session),
	}
}

// CreateUser creates a new user with UUID
func (s *Service) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	if _, exists := s.users[req.Email]; exists {
		return nil, ErrUserExists
	}
	
	// Validate tenant exists
	if _, exists := s.tenants[req.TenantID]; !exists {
		return nil, ErrTenantNotFound
	}
	
	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	
	// Generate new UUID for user
	userID := uuid.New()
	
	firstName := &req.FirstName
	lastName := &req.LastName
	if req.FirstName == "" {
		firstName = nil
	}
	if req.LastName == "" {
		lastName = nil
	}
	
	user := &models.User{
		ID:            userID,
		Email:         req.Email,
		Username:      req.Email, // Use email as username
		FirstName:     firstName,
		LastName:      lastName,
		PasswordHash:  passwordHash,
		Role:          req.Role,
		Company:       req.Company,
		TenantID:      req.TenantID,
		Active:        true,
		EmailVerified: true, // Auto-verify for demo
		Settings:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	s.users[req.Email] = user
	return user, nil
}

// GetUser retrieves a user by email
func (s *Service) GetUser(ctx context.Context, email string) (*models.User, error) {
	user, exists := s.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Login authenticates a user and creates a session
func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, exists := s.users[req.Email]
	if !exists {
		return nil, ErrInvalidCredentials
	}
	
	// Check password
	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	
	// Get tenant
	tenant, exists := s.tenants[user.TenantID]
	if !exists {
		return nil, ErrTenantNotFound
	}
	
	// Create session
	session := &models.Session{
		ID:           generateSessionID(),
		UserID:       user.ID, // UUID
		TenantID:     user.TenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	s.sessions[session.ID] = session
	
	// Update last login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now
	
	return &models.LoginResponse{
		User:      user,
		Tenant:    tenant,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// ValidateSession checks if a session is valid
func (s *Service) ValidateSession(ctx context.Context, sessionID string) (*models.User, *models.Tenant, error) {
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, nil, ErrSessionExpired
	}
	
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, sessionID)
		return nil, nil, ErrSessionExpired
	}
	
	user, exists := s.getUserByID(session.UserID)
	if !exists {
		return nil, nil, ErrUserNotFound
	}
	
	tenant, exists := s.tenants[session.TenantID]
	if !exists {
		return nil, nil, ErrTenantNotFound
	}
	
	// Update last activity
	session.LastActivity = time.Now()
	
	return user, tenant, nil
}

// CreateTenant creates a new tenant with UUID
func (s *Service) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	if _, exists := s.tenants[slug]; exists {
		return nil, errors.New("tenant already exists")
	}
	
	// Generate new UUID for tenant
	tenantID := uuid.New()
	
	dbName := "oilgas_" + slug
	tenant := &models.Tenant{
		ID:           tenantID,
		Name:         name,
		Slug:         slug,
		DatabaseType: "tenant",
		DatabaseName: &dbName,
		Active:       true,
		Settings:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	s.tenants[slug] = tenant
	return tenant, nil
}

// Helper methods
func (s *Service) getUserByID(id uuid.UUID) (*models.User, bool) {
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return nil, false
}

// ValidatePassword checks password strength
func (s *Service) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

// Utility functions
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateSessionID() string {
	return "sess_" + uuid.New().String()
}

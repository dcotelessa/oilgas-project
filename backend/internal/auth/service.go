package auth

import (
	"context"
	"errors"
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found") 
	ErrUserExists        = errors.New("user already exists")
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrSessionExpired    = errors.New("session expired")
)

// Service handles authentication operations
type Service struct {
	// In-memory stores for testing
	users    map[string]*User    // email -> user
	tenants  map[string]*Tenant  // slug -> tenant
	sessions map[string]*Session // session_id -> session
}

// NewService creates a new auth service
func NewService() *Service {
	return &Service{
		users:    make(map[string]*User),
		tenants:  make(map[string]*Tenant),
		sessions: make(map[string]*Session),
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
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
	
	user := &User{
		ID:           len(s.users) + 1, // Simple ID generation for testing
		Email:        req.Email,
		Username:     req.Email, // Use email as username for now
		PasswordHash: passwordHash,
		Role:         req.Role,
		Company:      req.Company,
		TenantID:     req.TenantID,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	s.users[req.Email] = user
	return user, nil
}

// GetUser retrieves a user by email
func (s *Service) GetUser(ctx context.Context, email string) (*User, error) {
	user, exists := s.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Login authenticates a user and creates a session
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
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
	session := &Session{
		ID:           generateSessionID(),
		UserID:       user.ID,
		TenantID:     user.TenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour sessions
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	s.sessions[session.ID] = session
	
	// Update last login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now
	
	return &LoginResponse{
		User:      user,
		Tenant:    tenant,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// ValidateSession checks if a session is valid
func (s *Service) ValidateSession(ctx context.Context, sessionID string) (*User, *Tenant, error) {
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

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(ctx context.Context, name, slug string) (*Tenant, error) {
	if _, exists := s.tenants[slug]; exists {
		return nil, errors.New("tenant already exists")
	}
	
	tenant := &Tenant{
		ID:           len(s.tenants) + 1,
		Name:         name,
		Slug:         slug,
		Code:         slug,
		DatabaseName: "oilgas_" + slug,
		DatabaseType: "tenant",
		Active:       true,
		Settings:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	s.tenants[slug] = tenant
	return tenant, nil
}

// Helper methods
func (s *Service) getUserByID(id int) (*User, bool) {
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
	return "sess_" + time.Now().Format("20060102150405") + "_" + time.Now().Format("000000")
}

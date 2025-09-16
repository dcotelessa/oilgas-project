// backend/internal/auth/service.go
package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"oilgas-backend/internal/shared/database"
)

type Service interface {
	Authenticate(ctx context.Context, email, password string) (*LoginResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (*User, *Session, error)
	CreateCustomerContact(ctx context.Context, req *CreateCustomerContactRequest) (*User, error)
	CreateEnterpriseUser(ctx context.Context, req *CreateEnterpriseUserRequest) (*User, error)
	UpdateUserTenantAccess(ctx context.Context, req *UpdateUserTenantAccessRequest) error
	ValidateUserTenantAccess(ctx context.Context, userID int, tenantID string) (bool, error)
	GetUserCustomerContext(ctx context.Context, userID int, tenantID string) (*int, error)
	GetUserByID(ctx context.Context, userID int) (*User, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

type service struct {
	dbManager  *database.DatabaseManager
	repository Repository
	jwtSecret  []byte
}

func NewService(dbManager *database.DatabaseManager, repository Repository) Service {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-change-in-production")
	}
	
	return &service{
		dbManager:  dbManager,
		repository: repository,
		jwtSecret:  jwtSecret,
	}
}

func (s *service) CreateCustomerContact(ctx context.Context, req *CreateCustomerContactRequest) (*User, error) {
	if err := s.validateCreateCustomerContactRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:            req.Email,
		FullName:         req.FullName,
		PasswordHash:     string(passwordHash),
		Role:             RoleCustomerContact,
		IsActive:         true,
		IsEnterpriseUser: false,
		PrimaryTenantID:  req.TenantID,
		CustomerID:       &req.CustomerID,
		ContactType:      req.ContactType,
		TenantAccess:     TenantAccessList{{
			TenantID:     req.TenantID,
			Role:         RoleCustomerContact,
			Permissions:  []Permission{PermissionViewInventory},
			YardAccess:   req.YardAccess,
			CanRead:      true,
			CanWrite:     false,
			CanDelete:    false,
			CanApprove:   false,
		}},
	}

	createdUser, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer contact: %w", err)
	}

	return createdUser, nil
}

func (s *service) ValidateUserTenantAccess(ctx context.Context, userID int, tenantID string) (bool, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		return false, fmt.Errorf("user account is inactive")
	}

	return user.CanAccessTenant(tenantID), nil
}

func (s *service) GetUserCustomerContext(ctx context.Context, userID int, tenantID string) (*int, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.IsCustomerContact() && user.CanAccessTenant(tenantID) {
		return user.CustomerID, nil
	}

	return nil, nil
}

func (s *service) Authenticate(ctx context.Context, email, password string) (*LoginResponse, error) {
	if err := s.validateCredentials(email, password); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	session, err := s.createSession(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	token, err := s.generateJWT(user, session)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var tenantContext *TenantAccess
	if len(user.TenantAccess) > 0 {
		tenantContext = &user.TenantAccess[0]
	}

	return &LoginResponse{
		Token:         token,
		User:          user.ToResponse(),
		TenantContext: tenantContext,
		ExpiresAt:     session.ExpiresAt,
		RefreshToken:  session.RefreshToken,
	}, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (*User, *Session, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("invalid session ID in token")
	}

	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, nil, fmt.Errorf("session expired")
	}

	user, err := s.repository.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	session.LastUsedAt = time.Now()
	s.repository.UpdateSession(ctx, session)

	return user, session, nil
}

func (s *service) CreateEnterpriseUser(ctx context.Context, req *CreateEnterpriseUserRequest) (*User, error) {
	if err := s.validateCreateEnterpriseUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Username:         req.Username,
		Email:            req.Email,
		FullName:         req.FullName,
		PasswordHash:     string(passwordHash),
		Role:             req.Role,
		IsActive:         true,
		IsEnterpriseUser: req.IsEnterpriseUser,
		PrimaryTenantID:  req.PrimaryTenantID,
		TenantAccess:     TenantAccessList(req.TenantAccess),
	}

	createdUser, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create enterprise user: %w", err)
	}

	return createdUser, nil
}

func (s *service) UpdateUserTenantAccess(ctx context.Context, req *UpdateUserTenantAccessRequest) error {
	if req.UserID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	if len(req.TenantAccess) == 0 {
		return fmt.Errorf("tenant access list cannot be empty")
	}

	return s.repository.UpdateUserTenantAccess(ctx, req.UserID, TenantAccessList(req.TenantAccess))
}

func (s *service) GetUserByID(ctx context.Context, userID int) (*User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.GetUserByID(ctx, userID)
}

func (s *service) InvalidateSession(ctx context.Context, sessionID string) error {
	return s.repository.InvalidateSession(ctx, sessionID)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	sessions, err := s.repository.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !sessions.IsActive || (sessions.RefreshExpiresAt != nil && time.Now().After(*sessions.RefreshExpiresAt)) {
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := s.repository.GetUserByID(ctx, sessions.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	newSession, err := s.createSession(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	s.repository.InvalidateSession(ctx, sessions.ID)

	token, err := s.generateJWT(user, newSession)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var tenantContext *TenantAccess
	if len(user.TenantAccess) > 0 {
		tenantContext = &user.TenantAccess[0]
	}

	return &LoginResponse{
		Token:         token,
		User:          user.ToResponse(),
		TenantContext: tenantContext,
		ExpiresAt:     newSession.ExpiresAt,
		RefreshToken:  newSession.RefreshToken,
	}, nil
}

func (s *service) createSession(ctx context.Context, user *User) (*Session, error) {
	sessionID, err := generateSecureID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	refreshToken, err := generateSecureID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	var tenantContext *TenantAccess
	if len(user.TenantAccess) > 0 {
		tenantContext = &user.TenantAccess[0]
	}

	session := &Session{
		ID:               sessionID,
		UserID:           user.ID,
		TenantID:         user.PrimaryTenantID,
		RefreshToken:     refreshToken,
		TenantContext:    tenantContext,
		IsActive:         true,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: &refreshExpiresAt,
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}

	err = s.repository.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *service) generateJWT(user *User, session *Session) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"session_id": session.ID,
		"tenant_id":  session.TenantID,
		"role":       user.Role,
		"exp":        session.ExpiresAt.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateSecureID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *service) validateCredentials(email, password string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func (s *service) validateCreateCustomerContactRequest(req *CreateCustomerContactRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if req.CustomerID <= 0 {
		return fmt.Errorf("invalid customer ID")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.ContactType == "" {
		req.ContactType = ContactPrimary
	}
	return nil
}

func (s *service) validateCreateEnterpriseUserRequest(req *CreateEnterpriseUserRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.Role == "" {
		return fmt.Errorf("role is required")
	}
	if len(req.TenantAccess) == 0 {
		return fmt.Errorf("tenant access is required")
	}
	return nil
}

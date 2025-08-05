// backend/internal/auth/service.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	// Authentication
	Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*User, *TenantAccess, error)
	
	// User management
	CreateCustomerContact(ctx context.Context, req CreateCustomerContactRequest) (*User, error)
	CreateEnterpriseUser(ctx context.Context, req CreateEnterpriseUserRequest) (*User, error)
	UpdateUser(ctx context.Context, userID int, updates UserUpdates) (*User, error)
	DeactivateUser(ctx context.Context, userID int) error
	GetUserByID(ctx context.Context, userID int) (*User, error)
	GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error)
	GetCustomerContacts(ctx context.Context, customerID int) ([]User, error)
	
	// User search and filtering
	SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
	
	// Tenant access management
	UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error
	UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error
	
	// Permission checking (missing methods that middleware needs)
	CheckPermission(ctx context.Context, check UserPermissionCheck) error
	CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error
	GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error)
	
	// Enterprise operations (missing methods that middleware needs)
	GetEnterpriseContext(ctx context.Context, userID int) (*EnterpriseContext, error)
	GetCustomerAccessContext(ctx context.Context, userID int) (*CustomerAccessContext, error)
}

type service struct {
	repo           Repository
	permissionSvc  PermissionService
	jwtSecret      []byte
	tokenExpiry    time.Duration
	calculator     *PermissionCalculator
}

func NewService(repo Repository, jwtSecret []byte, tokenExpiry time.Duration) Service {
	permissionSvc := NewPermissionService(repo)
	return &service{
		repo:          repo,
		permissionSvc: permissionSvc,
		jwtSecret:     jwtSecret,
		tokenExpiry:   tokenExpiry,
		calculator:    &PermissionCalculator{},
	}
}

func (s *service) Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	
	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Determine tenant context
	tenantID, tenantContext, err := s.determineTenantContext(user, req.TenantID)
	if err != nil {
		return nil, err
	}
	
	// Generate JWT token
	token, err := s.generateToken(user.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Create session
	session := &Session{
		ID:           s.generateSessionID(),
		UserID:       user.ID,
		TenantID:     tenantID,
		Token:        token,
		RefreshToken: refreshToken,
		TenantContext: &tenantContext,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(s.tokenExpiry),
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}
	
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.repo.UpdateUser(ctx, user)
	
	return &LoginResponse{
		Token:         token,
		User:          user.ToResponse(),
		TenantContext: &tenantContext,
		ExpiresAt:     time.Now().Add(s.tokenExpiry),
		RefreshToken:  refreshToken,
	}, nil
}

func (s *service) determineTenantContext(user *User, requestedTenantID string) (string, TenantAccess, error) {
	var tenantID string
	var tenantContext TenantAccess
	
	// Enterprise users can access any tenant
	if user.IsEnterpriseUser {
		if requestedTenantID != "" {
			tenantID = requestedTenantID
		} else {
			tenantID = user.PrimaryTenantID
		}
		
		// Fixed: Create TenantAccess properly with all required fields
		tenantContext = TenantAccess{
			TenantID:    tenantID,
			Role:        user.Role,
			Permissions: s.getEnterprisePermissions(user.Role),
			YardAccess:  []YardAccess{}, // Enterprise users get all yard access
			CanRead:     true,
			CanWrite:    user.Role != RoleCustomerContact,
			CanDelete:   user.Role == RoleEnterpriseAdmin || user.Role == RoleSystemAdmin,
			CanApprove:  user.Role != RoleCustomerContact,
		}
	} else {
		// Regular users must specify tenant or use primary
		if requestedTenantID != "" {
			tenantID = requestedTenantID
		} else {
			tenantID = user.PrimaryTenantID
		}
		
		// Find tenant access
		found := false
		for _, access := range user.TenantAccess {
			if access.TenantID == tenantID {
				tenantContext = access
				found = true
				break
			}
		}
		
		if !found {
			return "", TenantAccess{}, ErrTenantAccessDenied
		}
	}
	
	return tenantID, tenantContext, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (*User, *TenantAccess, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	
	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %w", err)
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token claims")
	}
	
	userID := int(claims["user_id"].(float64))
	sessionID := claims["session_id"].(string)
	
	// Get session and validate
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	
	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, nil, ErrInvalidToken
	}
	
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}
	
	if !user.IsActive {
		return nil, nil, ErrInvalidToken
	}
	
	// Update session last used
	session.LastUsedAt = time.Now()
	s.repo.UpdateSession(ctx, session)
	
	return user, session.TenantContext, nil
}

func (s *service) CreateCustomerContact(ctx context.Context, req CreateCustomerContactRequest) (*User, error) {
	// Check if user exists
	if existing, _ := s.repo.GetUserByEmail(ctx, req.Email); existing != nil {
		return nil, ErrUserExists
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create tenant access
	tenantAccess := TenantAccess{
		TenantID:    req.TenantID,
		Role:        RoleCustomerContact,
		Permissions: []Permission{PermissionViewInventory, PermissionCreateWorkOrder},
		YardAccess:  req.YardAccess,
		CanRead:     true,
		CanWrite:    true,
		CanDelete:   false,
		CanApprove:  false,
	}
	
	user := &User{
		Username:         req.Email,
		Email:            req.Email,
		FullName:         req.FullName,
		PasswordHash:     string(hashedPassword),
		Role:             RoleCustomerContact,
		AccessLevel:      1,
		IsEnterpriseUser: false,
		TenantAccess:     []TenantAccess{tenantAccess},
		PrimaryTenantID:  req.TenantID,
		CustomerID:       &req.CustomerID,
		ContactType:      req.ContactType,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	return s.repo.CreateUser(ctx, user)
}

func (s *service) CreateEnterpriseUser(ctx context.Context, req CreateEnterpriseUserRequest) (*User, error) {
	// Check if user exists
	if existing, _ := s.repo.GetUserByEmail(ctx, req.Email); existing != nil {
		return nil, ErrUserExists
	}
	
	// Validate role
	if !s.isValidRole(req.Role) {
		return nil, ErrInvalidUserRole
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	user := &User{
		Username:         req.Username,
		Email:            req.Email,
		FullName:         req.FullName,
		PasswordHash:     string(hashedPassword),
		Role:             req.Role,
		AccessLevel:      s.getRoleAccessLevel(req.Role),
		IsEnterpriseUser: req.IsEnterpriseUser,
		TenantAccess:     req.TenantAccess,
		PrimaryTenantID:  req.PrimaryTenantID,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	return s.repo.CreateUser(ctx, user)
}

func (s *service) UpdateUser(ctx context.Context, userID int, updates UserUpdates) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Apply updates
	if updates.Email != nil {
		user.Email = *updates.Email
	}
	if updates.FullName != nil {
		user.FullName = *updates.FullName
	}
	if updates.Role != nil {
		if !s.isValidRole(*updates.Role) {
			return nil, ErrInvalidUserRole
		}
		user.Role = *updates.Role
		user.AccessLevel = s.getRoleAccessLevel(*updates.Role)
	}
	if updates.IsActive != nil {
		user.IsActive = *updates.IsActive
	}
	if updates.IsEnterpriseUser != nil {
		user.IsEnterpriseUser = *updates.IsEnterpriseUser
	}
	if updates.PrimaryTenantID != nil {
		user.PrimaryTenantID = *updates.PrimaryTenantID
	}
	
	// Handle password change
	if updates.CurrentPassword != nil && updates.NewPassword != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*updates.CurrentPassword)); err != nil {
			return nil, ErrInvalidCredentials
		}
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updates.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}
	
	user.UpdatedAt = time.Now()
	
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *service) DeactivateUser(ctx context.Context, userID int) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	user.IsActive = false
	user.UpdatedAt = time.Now()
	
	// Invalidate all user sessions
	if err := s.repo.InvalidateUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}
	
	return s.repo.UpdateUser(ctx, user)
}

func (s *service) GetUserByID(ctx context.Context, userID int) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *service) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	return s.repo.GetUsersByTenant(ctx, tenantID)
}

func (s *service) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	return s.repo.GetCustomerContacts(ctx, customerID)
}

func (s *service) SearchUsers(ctx context.Context, filters UserSearchFilters) ([]User, int, error) {
	return s.repo.SearchUsers(ctx, filters)
}

func (s *service) GetUserStats(ctx context.Context) (*UserStats, error) {
	return s.repo.GetUserStats(ctx)
}

func (s *service) UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error {
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	
	user.TenantAccess = req.TenantAccess
	user.UpdatedAt = time.Now()
	
	return s.repo.UpdateUser(ctx, user)
}

func (s *service) UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error {
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	
	// Find and update the tenant access
	for i, access := range user.TenantAccess {
		if access.TenantID == req.TenantID {
			user.TenantAccess[i].YardAccess = req.YardAccess
			break
		}
	}
	
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(ctx, user)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// In a production system, you'd store and validate refresh tokens
	// For now, this is a placeholder implementation
	return nil, fmt.Errorf("refresh token functionality not implemented")
}

func (s *service) Logout(ctx context.Context, token string) error {
	// Parse token to get session ID
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	
	if err != nil {
		return ErrInvalidToken
	}
	
	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return ErrInvalidToken
	}
	
	return s.repo.InvalidateSession(ctx, sessionID)
}

// Helper methods

func (s *service) generateToken(userID int, tenantID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"session_id": s.generateSessionID(),
		"exp":        time.Now().Add(s.tokenExpiry).Unix(),
		"iat":        time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) generateRefreshToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) generateSessionID() string {
	return fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func (s *service) isValidRole(role UserRole) bool {
	validRoles := []UserRole{
		RoleCustomerContact,
		RoleOperator,
		RoleManager,
		RoleAdmin,
		RoleEnterpriseAdmin,
		RoleSystemAdmin,
	}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func (s *service) getRoleAccessLevel(role UserRole) int {
	switch role {
	case RoleCustomerContact:
		return 1
	case RoleOperator:
		return 2
	case RoleManager:
		return 3
	case RoleAdmin:
		return 4
	case RoleEnterpriseAdmin:
		return 5
	case RoleSystemAdmin:
		return 6
	default:
		return 1
	}
}

func (s *service) getEnterprisePermissions(role UserRole) []Permission {
	basePermissions := []Permission{
		PermissionViewInventory,
		PermissionCreateWorkOrder,
		PermissionManageTransport,
		PermissionExportData,
	}
	
	switch role {
	case RoleEnterpriseAdmin, RoleSystemAdmin:
		return append(basePermissions, 
			PermissionApproveWorkOrder,
			PermissionUserManagement,
			PermissionCrossTenantView,
		)
	case RoleAdmin, RoleManager:
		return append(basePermissions, PermissionApproveWorkOrder)
	default:
		return basePermissions
	}
}

func (s *service) getCrossTenantPermissions(role UserRole) []Permission {
	return s.calculator.GetCrossTenantPermissions(role)
}

// Missing method implementations that middleware needs

func (s *service) CheckPermission(ctx context.Context, check UserPermissionCheck) error {
	return s.permissionSvc.CheckPermission(ctx, check.UserID, check.TenantID, check.Permission)
}

func (s *service) CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error {
	return s.permissionSvc.CheckYardAccess(ctx, userID, tenantID, yardLocation)
}

func (s *service) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	return s.permissionSvc.GetUserPermissions(ctx, userID, tenantID)
}

func (s *service) GetEnterpriseContext(ctx context.Context, userID int) (*EnterpriseContext, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if !user.CanPerformCrossTenantOperation() {
		return nil, ErrEnterpriseAccessDenied
	}
	
	return &EnterpriseContext{
		UserID:            user.ID,
		AccessibleTenants: user.TenantAccess,
		IsEnterpriseAdmin: user.Role == RoleEnterpriseAdmin || user.Role == RoleSystemAdmin,
		CrossTenantPerms:  s.getCrossTenantPermissions(user.Role),
	}, nil
}

func (s *service) GetCustomerAccessContext(ctx context.Context, userID int) (*CustomerAccessContext, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if !user.IsCustomerContact() {
		return nil, ErrNotCustomerContact
	}
	
	return user.GetCustomerAccessContext(), nil
}

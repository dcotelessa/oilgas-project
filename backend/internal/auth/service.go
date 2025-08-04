// backend/internal/auth/service.go
package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
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
	UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error
	UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error
	DeactivateUser(ctx context.Context, userID int) error
	
	// Permission checking
	CheckPermission(ctx context.Context, check UserPermissionCheck) error
	CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error
	GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error)
	
	// Enterprise operations
	GetEnterpriseContext(ctx context.Context, userID int) (*EnterpriseContext, error)
	GetCustomerAccessContext(ctx context.Context, userID int) (*CustomerAccessContext, error)
	
	// User queries
	GetUser(ctx context.Context, userID int) (*User, error)
	GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error)
	GetCustomerContacts(ctx context.Context, customerID int) ([]User, error)
	SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
	
	// Session management
	InvalidateUserSessions(ctx context.Context, userID int) error
	CleanupExpiredSessions(ctx context.Context) error
}

type service struct {
	repo          Repository
	jwtSecret     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:          repo,
		jwtSecret:     []byte(jwtSecret),
		tokenExpiry:   24 * time.Hour,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

// Authentication
func (s *service) Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Determine tenant context
	tenantID, tenantContext, err := s.determineTenantContext(user, req.TenantID)
	if err != nil {
		return nil, err
	}
	
	// Create session
	session, err := s.createSession(ctx, user.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Generate JWT token
	token, err := s.generateJWT(user, tenantID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Generate refresh token
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.repo.UpdateUser(ctx, user)
	
	return &LoginResponse{
		Token:         token,
		User:          user.ToResponse(),
		TenantContext: tenantContext,
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
	tenantID := claims["tenant_id"].(string)
	sessionID := claims["session_id"].(string)
	
	// Validate session is still active
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found or expired")
	}
	
	if !session.IsActive {
		return nil, nil, fmt.Errorf("session is inactive")
	}
	
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	
	if !user.IsActive {
		return nil, nil, ErrUserInactive
	}
	
	// Get tenant context
	_, tenantContext, err := s.determineTenantContext(user, tenantID)
	if err != nil {
		return nil, nil, err
	}
	
	return user, &tenantContext, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	return s.repo.InvalidateSession(ctx, token)
}

// User management
func (s *service) CreateCustomerContact(ctx context.Context, req CreateCustomerContactRequest) (*User, error) {
	// Validate customer exists
	if err := s.repo.ValidateCustomerExists(ctx, req.CustomerID); err != nil {
		return nil, err
	}
	
	// Validate yard access
	if err := s.validateYardAccess(req.YardAccess); err != nil {
		return nil, err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	user := &User{
		Username:        req.Email, // Use email as username for customer contacts
		Email:           req.Email,
		FullName:        req.FullName,
		PasswordHash:    string(hashedPassword),
		Role:            RoleCustomerContact,
		AccessLevel:     1, // Basic customer access
		IsEnterpriseUser: false,
		CustomerID:      &req.CustomerID,
		ContactType:     req.ContactType,
		PrimaryTenantID: req.TenantID,
		TenantAccess: TenantAccessList{
			{
				TenantID:    req.TenantID,
				Role:        RoleCustomerContact,
				Permissions: s.getCustomerContactPermissions(),
				YardAccess:  req.YardAccess,
				CanRead:     true,
				CanWrite:    req.ContactType == ContactPrimary || req.ContactType == ContactApprover,
				CanDelete:   false,
				CanApprove:  req.ContactType == ContactApprover,
			},
		},
		IsActive: true,
	}
	
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *service) CreateEnterpriseUser(ctx context.Context, req CreateEnterpriseUserRequest) (*User, error) {
	// Validate role
	if !s.isValidRole(req.Role) {
		return nil, ErrInvalidUserRole
	}
	
	// Validate tenant access
	for _, tenantAccess := range req.TenantAccess {
		if err := s.validateYardAccess(tenantAccess.YardAccess); err != nil {
			return nil, err
		}
	}
	
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
		PrimaryTenantID:  req.PrimaryTenantID,
		TenantAccess:     req.TenantAccess,
		IsActive:         true,
	}
	
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *service) UpdateUserTenantAccess(ctx context.Context, req UpdateUserTenantAccessRequest) error {
	// Validate tenant access
	for _, tenantAccess := range req.TenantAccess {
		if err := s.validateYardAccess(tenantAccess.YardAccess); err != nil {
			return err
		}
	}
	
	return s.repo.UpdateUserTenantAccess(ctx, req.UserID, req.TenantAccess)
}

func (s *service) UpdateUserYardAccess(ctx context.Context, req UpdateUserYardAccessRequest) error {
	// Validate yard access
	if err := s.validateYardAccess(req.YardAccess); err != nil {
		return err
	}
	
	return s.repo.UpdateUserYardAccess(ctx, req.UserID, req.TenantID, req.YardAccess)
}

// Permission checking
func (s *service) CheckPermission(ctx context.Context, check UserPermissionCheck) error {
	user, err := s.repo.GetUserByID(ctx, check.UserID)
	if err != nil {
		return err
	}
	
	if !user.HasPermissionInTenant(check.TenantID, check.Permission) {
		return ErrPermissionDenied
	}
	
	return nil
}

func (s *service) CheckYardAccess(ctx context.Context, userID int, tenantID, yardLocation string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if !user.HasAccessToYard(tenantID, yardLocation) {
		return ErrYardAccessDenied
	}
	
	return nil
}

func (s *service) GetUserPermissions(ctx context.Context, userID int, tenantID string) ([]Permission, error) {
	return s.repo.GetUserPermissions(ctx, userID, tenantID)
}

// Enterprise operations
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
		IsEnterpriseAdmin: user.Role == RoleEnterpriseAdmin,
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

// User queries
func (s *service) GetUser(ctx context.Context, userID int) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *service) GetUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	return s.repo.GetUsersByTenant(ctx, tenantID)
}

func (s *service) GetCustomerContacts(ctx context.Context, customerID int) ([]User, error) {
	return s.repo.GetCustomerContacts(ctx, customerID)
}

func (s *service) SearchUsers(ctx context.Context, filter UserSearchFilter) ([]User, error) {
	return s.repo.SearchUsers(ctx, filter)
}

func (s *service) GetUserStats(ctx context.Context) (*UserStats, error) {
	return s.repo.GetUserStats(ctx)
}

// Session management
func (s *service) InvalidateUserSessions(ctx context.Context, userID int) error {
	return s.repo.InvalidateUserSessions(ctx, userID)
}

func (s *service) CleanupExpiredSessions(ctx context.Context) error {
	return s.repo.CleanupExpiredSessions(ctx)
}

// Helper methods
func (s *service) createSession(ctx context.Context, userID int, tenantID string) (*Session, error) {
	token, err := s.generateSessionToken()
	if err != nil {
		return nil, err
	}
	
	session := &Session{
		ID:        generateSessionID(),
		UserID:    userID,
		TenantID:  tenantID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.tokenExpiry),
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	
	return session, nil
}

func (s *service) generateJWT(user *User, tenantID, sessionID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"tenant_id":  tenantID,
		"session_id": sessionID,
		"exp":        time.Now().Add(s.tokenExpiry).Unix(),
		"iat":        time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
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

func (s *service) isValidRole(role UserRole) bool {
	validRoles := []UserRole{
		RoleCustomerContact, RoleOperator, RoleManager,
		RoleAdmin, RoleEnterpriseAdmin, RoleSystemAdmin,
	}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	
	return false
}

func (s *service) validateYardAccess(yardAccess []YardAccess) error {
	// Validate yard locations are not empty and permissions are valid
	for _, yard := range yardAccess {
		if yard.YardLocation == "" {
			return ErrInvalidYardAccess
		}
		
		// At least one permission should be granted
		if !yard.CanViewWorkOrders && !yard.CanViewInventory && 
		   !yard.CanCreateWorkOrders && !yard.CanApproveOrders &&
		   !yard.CanManageTransport && !yard.CanExportData {
			return ErrInvalidYardAccess
		}
	}
	
	return nil
}

func (s *service) getCustomerContactPermissions() []Permission {
	return []Permission{
		PermissionWorkOrderRead,
		PermissionInventoryRead,
		PermissionCustomerRead,
	}
}

func (s *service) getEnterprisePermissions(role UserRole) []Permission {
	switch role {
	case RoleEnterpriseAdmin:
		return []Permission{
			PermissionCustomerRead, PermissionCustomerWrite,
			PermissionInventoryRead, PermissionInventoryWrite, PermissionInventoryExport,
			PermissionWorkOrderRead, PermissionWorkOrderWrite, PermissionWorkOrderApprove,
			PermissionTransportRead, PermissionTransportWrite,
			PermissionCrossTenantView, PermissionUserManagement,
		}
	case RoleSystemAdmin:
		return []Permission{
			PermissionCustomerRead, PermissionCustomerWrite, PermissionCustomerDelete,
			PermissionInventoryRead, PermissionInventoryWrite, PermissionInventoryDelete, PermissionInventoryExport,
			PermissionWorkOrderRead, PermissionWorkOrderWrite, PermissionWorkOrderDelete, 
			PermissionWorkOrderApprove, PermissionWorkOrderInvoice,
			PermissionTransportRead, PermissionTransportWrite,
			PermissionCrossTenantView, PermissionUserManagement, PermissionSystemConfig,
		}
	case RoleAdmin:
		return []Permission{
			PermissionCustomerRead, PermissionCustomerWrite,
			PermissionInventoryRead, PermissionInventoryWrite, PermissionInventoryExport,
			PermissionWorkOrderRead, PermissionWorkOrderWrite, PermissionWorkOrderApprove,
			PermissionTransportRead, PermissionTransportWrite,
			PermissionUserManagement,
		}
	case RoleManager:
		return []Permission{
			PermissionCustomerRead, PermissionCustomerWrite,
			PermissionInventoryRead, PermissionInventoryWrite, PermissionInventoryExport,
			PermissionWorkOrderRead, PermissionWorkOrderWrite, PermissionWorkOrderApprove,
			PermissionTransportRead, PermissionTransportWrite,
		}
	case RoleOperator:
		return []Permission{
			PermissionInventoryRead, PermissionInventoryWrite,
			PermissionWorkOrderRead, PermissionWorkOrderWrite,
			PermissionTransportRead,
		}
	default:
		return []Permission{}
	}
}

func (s *service) getCrossTenantPermissions(role UserRole) []Permission {
	if role == RoleEnterpriseAdmin || role == RoleSystemAdmin {
		return []Permission{
			PermissionCrossTenantView,
			PermissionUserManagement,
		}
	}
	return []Permission{}
}

// Additional user update functionality
type UserUpdates struct {
	Email            *string          `json:"email"`
	FullName         *string          `json:"full_name"`
	Role             *UserRole        `json:"role"`
	IsActive         *bool            `json:"is_active"`
	IsEnterpriseUser *bool            `json:"is_enterprise_user"`
	PrimaryTenantID  *string          `json:"primary_tenant_id"`
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
	
	// Invalidate all user sessions
	if err := s.repo.InvalidateUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}
	
	return s.repo.UpdateUser(ctx, user)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// In a production system, you'd store and validate refresh tokens
	// For now, this is a placeholder implementation
	return nil, fmt.Errorf("refresh token functionality not implemented")
}

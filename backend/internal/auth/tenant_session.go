package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost        = 12
	SessionExpiration = 24 * time.Hour
	MaxFailedAttempts = 5
	LockoutDuration   = 30 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked     = errors.New("account is locked due to too many failed attempts")
	ErrSessionNotFound   = errors.New("session not found or expired")
	ErrNoTenantAccess    = errors.New("user has no access to specified tenant")
)

type Cache interface {
	Get(key string) interface{}
	Set(key string, value interface{}, duration time.Duration)
	Delete(key string)
}

type TenantSessionManager struct {
	pool  *pgxpool.Pool
	cache Cache
}

type TenantSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Company   string    `json:"company"`
	TenantSlug string   `json:"tenant_slug"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Company      string     `json:"company"`
	TenantID     string     `json:"tenant_id"`
	EmailVerified bool      `json:"email_verified"`
	LastLogin    *time.Time `json:"last_login"`
	FailedAttempts int      `json:"failed_attempts"`
	LockedUntil  *time.Time `json:"locked_until"`
}

func NewTenantSessionManager(pool *pgxpool.Pool, cache Cache) *TenantSessionManager {
	return &TenantSessionManager{
		pool:  pool,
		cache: cache,
	}
}

// Generate cryptographically secure session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate session ID: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// Create new session with tenant context
func (tsm *TenantSessionManager) CreateSession(ctx context.Context, user *User, tenantSlug, ipAddr, userAgent string) (*TenantSession, error) {
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(SessionExpiration)
	
	session := &TenantSession{
		ID:        sessionID,
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Role:      user.Role,
		Company:   user.Company,
		TenantSlug: tenantSlug,
		ExpiresAt: expiresAt,
		IPAddress: ipAddr,
		UserAgent: userAgent,
	}
	
	// Store in database
	_, err := tsm.pool.Exec(ctx, `
		INSERT INTO auth.sessions (id, user_id, tenant_id, email, role, company, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, sessionID, user.ID, user.TenantID, user.Email, user.Role, user.Company, expiresAt, ipAddr, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}
	
	// Cache the session
	cacheKey := fmt.Sprintf("session:%s", sessionID)
	tsm.cache.Set(cacheKey, session, SessionExpiration)
	
	return session, nil
}

// Validate session with cache-first lookup
func (tsm *TenantSessionManager) ValidateSession(ctx context.Context, sessionID string) (*TenantSession, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("session:%s", sessionID)
	if cached := tsm.cache.Get(cacheKey); cached != nil {
		if session, ok := cached.(*TenantSession); ok {
			if time.Now().Before(session.ExpiresAt) {
				return session, nil
			}
			// Expired, remove from cache
			tsm.cache.Delete(cacheKey)
		}
	}
	
	// Fallback to database
	var session TenantSession
	err := tsm.pool.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.tenant_id, s.email, s.role, s.company, 
		       t.slug, s.expires_at, s.ip_address, s.user_agent
		FROM auth.sessions s
		JOIN store.tenants t ON s.tenant_id = t.id
		WHERE s.id = $1 AND s.expires_at > NOW() AND s.deleted_at IS NULL
	`, sessionID).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.Email, 
		&session.Role, &session.Company, &session.TenantSlug, &session.ExpiresAt,
		&session.IPAddress, &session.UserAgent,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}
	
	// Update cache
	tsm.cache.Set(cacheKey, &session, time.Until(session.ExpiresAt))
	
	return &session, nil
}

// Login with tenant-aware authentication
func (tsm *TenantSessionManager) Login(ctx context.Context, email, password, tenantSlug, ipAddr, userAgent string) (*TenantSession, error) {
	// Get user with tenant info
	var user User
	
	query := `
		SELECT u.id, u.email, u.password_hash, u.role, u.company, u.tenant_id,
		       u.email_verified, u.last_login, u.failed_attempts, u.locked_until
		FROM auth.users u
		JOIN store.tenants t ON u.tenant_id = t.id
		WHERE u.email = $1 AND t.slug = $2 AND u.deleted_at IS NULL AND t.is_active = true
	`
	
	err := tsm.pool.QueryRow(ctx, query, email, tenantSlug).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Company,
		&user.TenantID, &user.EmailVerified, &user.LastLogin, &user.FailedAttempts,
		&user.LockedUntil,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed attempts
		tsm.incrementFailedAttempts(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}
	
	// Reset failed attempts and update last login
	tsm.resetFailedAttempts(ctx, user.ID)
	
	// Create session
	return tsm.CreateSession(ctx, &user, tenantSlug, ipAddr, userAgent)
}

// Set tenant context for RLS
func (tsm *TenantSessionManager) setTenantContext(ctx context.Context, session *TenantSession) error {
	_, err := tsm.pool.Exec(ctx, `
		SELECT 
			set_config('app.user_id', $1, true),
			set_config('app.user_role', $2, true),
			set_config('app.user_company', $3, true),
			set_config('app.tenant_id', $4, true)
	`, session.UserID, session.Role, session.Company, session.TenantID)
	
	return err
}

// Gin middleware for tenant-aware authentication
func (tsm *TenantSessionManager) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := extractSessionID(c)
		if sessionID == "" {
			c.JSON(401, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		
		session, err := tsm.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			if err == ErrSessionNotFound {
				c.JSON(401, gin.H{"error": "invalid or expired session"})
			} else {
				c.JSON(500, gin.H{"error": "authentication failed"})
			}
			c.Abort()
			return
		}
		
		// Set tenant context for RLS
		if err := tsm.setTenantContext(c.Request.Context(), session); err != nil {
			c.JSON(500, gin.H{"error": "failed to set tenant context"})
			c.Abort()
			return
		}
		
		// Store session in context
		c.Set("session", session)
		c.Set("user_id", session.UserID)
		c.Set("tenant_id", session.TenantID)
		c.Set("user_role", session.Role)
		c.Next()
	}
}

// Helper functions
func (tsm *TenantSessionManager) incrementFailedAttempts(ctx context.Context, userID string) {
	var lockedUntil *time.Time
	
	// Get current failed attempts
	var failedAttempts int
	tsm.pool.QueryRow(ctx, "SELECT failed_attempts FROM auth.users WHERE id = $1", userID).Scan(&failedAttempts)
	
	failedAttempts++
	if failedAttempts >= MaxFailedAttempts {
		lockTime := time.Now().Add(LockoutDuration)
		lockedUntil = &lockTime
	}
	
	tsm.pool.Exec(ctx, `
		UPDATE auth.users 
		SET failed_attempts = $1, locked_until = $2 
		WHERE id = $3
	`, failedAttempts, lockedUntil, userID)
}

func (tsm *TenantSessionManager) resetFailedAttempts(ctx context.Context, userID string) {
	tsm.pool.Exec(ctx, `
		UPDATE auth.users 
		SET failed_attempts = 0, locked_until = NULL, last_login = NOW() 
		WHERE id = $1
	`, userID)
}

func (tsm *TenantSessionManager) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := tsm.pool.Exec(ctx, `
		UPDATE auth.sessions 
		SET deleted_at = NOW() 
		WHERE id = $1
	`, sessionID)
	
	if err == nil {
		cacheKey := fmt.Sprintf("session:%s", sessionID)
		tsm.cache.Delete(cacheKey)
	}
	
	return err
}

// Extract session ID from cookie or Authorization header
func extractSessionID(c *gin.Context) string {
	// Try cookie first
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		return cookie
	}
	
	// Fallback to Authorization header
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	
	return ""
}

// Extract IP address from request
func ExtractIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// Extract user agent
func ExtractUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// backend/internal/auth/session_manager.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oilgas-backend/pkg/cache"
	"oilgas-backend/internal/models"
)

type TenantSessionManager struct {
	pool  *pgxpool.Pool
	cache *cache.MemoryCache
}

func NewTenantSessionManager(pool *pgxpool.Pool, cache *cache.MemoryCache) *TenantSessionManager {
	return &TenantSessionManager{
		pool:  pool,
		cache: cache,
	}
}

func (sm *TenantSessionManager) CreateSession(ctx context.Context, userID string, tenantID string) (*models.Session, error) {
	session := &models.Session{
		ID:           generateSessionID(),
		UserID:       uuid.MustParse(userID), // Convert string to UUID
		TenantID:     tenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	// Store in cache for fast access
	sm.cache.Set(session.ID, session, 24*time.Hour)
	
	return session, nil
}

func (sm *TenantSessionManager) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	// Try cache first
	if cached := sm.cache.Get(sessionID); cached != nil {
		if session, ok := cached.(*models.Session); ok {
			if time.Now().Before(session.ExpiresAt) {
				// Update last activity
				session.LastActivity = time.Now()
				sm.cache.Set(sessionID, session, time.Until(session.ExpiresAt))
				return session, nil
			}
		}
	}
	
	return nil, ErrSessionExpired
}

func (sm *TenantSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	sm.cache.Delete(sessionID)
	return nil
}

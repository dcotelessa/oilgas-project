// backend/internal/auth/session_manager.go
package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"oilgas-backend/internal/models"
)

type SessionManager struct {
	sessions map[string]*models.Session
	mutex    sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*models.Session),
	}
}

func (sm *SessionManager) CreateSession(userID uuid.UUID, tenantID string) string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionID := uuid.New().String()
	session := &models.Session{
		ID:           sessionID,
		UserID:       userID,
		TenantID:     tenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	
	sm.sessions[sessionID] = session
	return sessionID
}

func (sm *SessionManager) GetSession(sessionID string) (*models.Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	
	return session, true
}

func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	delete(sm.sessions, sessionID)
}

func (sm *SessionManager) CleanupExpired() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	now := time.Now()
	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, sessionID)
		}
	}
}

package manager

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// SessionManager manages client sessions
type SessionManager struct {
	storage  *storage.Storage
	sessions map[string]*models.ClientSession
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(storage *storage.Storage) *SessionManager {
	sm := &SessionManager{
		storage:  storage,
		sessions: make(map[string]*models.ClientSession),
	}
	sm.load()
	return sm
}

// load reads sessions from storage
func (sm *SessionManager) load() error {
	var sessions []*models.ClientSession
	if err := sm.storage.LoadJSON("sessions.json", &sessions); err != nil {
		return err
	}
	for _, sess := range sessions {
		sm.sessions[sess.SessionID] = sess
	}
	return nil
}

// save writes sessions to storage
func (sm *SessionManager) save() error {
	sessions := make([]*models.ClientSession, 0, len(sm.sessions))
	for _, sess := range sm.sessions {
		sessions = append(sessions, sess)
	}
	return sm.storage.SaveJSON("sessions.json", sessions)
}

// RegisterOrUpdate creates a new session or updates existing one
func (sm *SessionManager) RegisterOrUpdate(clientID, clientName, configID, ipAddress string) (*models.ClientSession, error) {
	// Check if client already has a session
	for _, sess := range sm.sessions {
		if sess.ClientID == clientID {
			// Update existing session
			sess.ClientName = clientName
			sess.IPAddress = ipAddress
			sess.LastConnected = time.Now()
			sess.LastActive = time.Now()
			if configID != "" {
				sess.ConfigID = configID
			}
			sm.save()
			return sess, nil
		}
	}

	// Create new session
	session := &models.ClientSession{
		SessionID:     uuid.New().String(),
		ClientID:      clientID,
		ClientName:    clientName,
		ConfigID:      configID,
		IPAddress:     ipAddress,
		LastConnected: time.Now(),
		LastActive:    time.Now(),
	}

	sm.sessions[session.SessionID] = session
	sm.save()
	return session, nil
}

// Get retrieves a session by session ID
func (sm *SessionManager) Get(sessionID string) (*models.ClientSession, error) {
	sess, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return sess, nil
}

// GetByClientID retrieves a session by client ID
func (sm *SessionManager) GetByClientID(clientID string) (*models.ClientSession, error) {
	for _, sess := range sm.sessions {
		if sess.ClientID == clientID {
			return sess, nil
		}
	}
	return nil, fmt.Errorf("session not found for client: %s", clientID)
}

// List returns all sessions
func (sm *SessionManager) List() []*models.ClientSession {
	sessions := make([]*models.ClientSession, 0, len(sm.sessions))
	for _, sess := range sm.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

// UpdateConfig updates the configuration for a session
func (sm *SessionManager) UpdateConfig(sessionID, configID string) error {
	sess, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	sess.ConfigID = configID
	sess.LastActive = time.Now()
	return sm.save()
}

// UpdateActivity updates the last activity time for a session
func (sm *SessionManager) UpdateActivity(sessionID string) error {
	sess, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	sess.LastActive = time.Now()
	return sm.save()
}

// Delete removes a session
func (sm *SessionManager) Delete(sessionID string) error {
	delete(sm.sessions, sessionID)
	return sm.save()
}

// CleanupInactive removes sessions inactive for more than the specified duration
func (sm *SessionManager) CleanupInactive(duration time.Duration) error {
	cutoff := time.Now().Add(-duration)
	for sessionID, sess := range sm.sessions {
		if sess.LastActive.Before(cutoff) {
			delete(sm.sessions, sessionID)
		}
	}
	return sm.save()
}

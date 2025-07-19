package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type SessionManager struct {
	sessions map[string]string
	mutex    *sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]string),
		mutex:    &sync.RWMutex{},
	}
}

func (m *SessionManager) AddSession(userId string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	token := m.generateToken()
	m.sessions[token] = userId

	return token
}

func (m *SessionManager) GetAuthenticatedUser(token string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.sessions[token]
}

func (m *SessionManager) UpdateSession(token string, userId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[token] = userId
}

func (m *SessionManager) DeleteSession(token string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, token)
}

func (m *SessionManager) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

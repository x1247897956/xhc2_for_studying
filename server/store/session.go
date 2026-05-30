package store

import (
	"sync"

	"xhc2_for_studying/protocol"
)

// Session holds the cipher context and extension map for an active implant
// session.
type Session struct {
	CipherCtx *protocol.CipherContext
	ExtMap    protocol.ExtensionMap
	BeaconID  string
}

// SessionStore maps session tokens to their corresponding Session objects,
// protected by a read-write mutex.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewSessionStore creates and returns an initialized SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// Set associates a session token with the given Session.
func (s *SessionStore) Set(token string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = session
}

// Get retrieves the Session for a given token, or nil if not found.
func (s *SessionStore) Get(token string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[token]
}

package store

import (
	"sync"

	"xhc2_for_studying/protocol"
)

// SessionStore 持有每个 sessionToken 对应的加密上下文。
// server 在 KeyExchange 握手后存入，后续 Register 和 Checkin 通过 X-Session-Token 头取出使用。
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*protocol.CipherContext
}

// NewSessionStore 创建空的 session 存储。
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*protocol.CipherContext),
	}
}

// Set 存入 BeaconID 对应的 CipherContext。
func (s *SessionStore) Set(beaconID string, ctx *protocol.CipherContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[beaconID] = ctx
}

// Get 取出 BeaconID 对应的 CipherContext。不存在返回 nil。
func (s *SessionStore) Get(beaconID string) *protocol.CipherContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[beaconID]
}

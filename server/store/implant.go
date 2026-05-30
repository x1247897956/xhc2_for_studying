package store

import (
	"sync"

	"xhc2_for_studying/protocol"
)

// ImplantRecord holds an implant's Age private key and its negotiated
// extension map, stored at generation time.
type ImplantRecord struct {
	ImplantAgePrivateKey string
	ExtMap               protocol.ExtensionMap
}

// MemoryImplantStore retains implant records indexed by the hex-encoded SHA-256
// digest of the implant's Age public key.
type MemoryImplantStore struct {
	mu       sync.RWMutex
	implants map[string]*ImplantRecord
}

// NewImplantStore creates and returns an initialized ImplantStore.
func NewImplantStore() ImplantStore {
	return &MemoryImplantStore{
		implants: make(map[string]*ImplantRecord),
	}
}

// Set stores an implant record keyed by the given public key digest.
func (s *MemoryImplantStore) Set(pubKeyDigest string, record *ImplantRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.implants[pubKeyDigest] = record
	return nil
}

// Get retrieves the implant record for pubKeyDigest. The boolean indicates
// whether the record was found.
func (s *MemoryImplantStore) Get(pubKeyDigest string) (*ImplantRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.implants[pubKeyDigest]
	return r, ok
}

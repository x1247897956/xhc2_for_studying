// Package store provides in-memory storage for beacons, tasks, implants, and
// sessions with concurrency-safe access.
package store

import (
	"errors"
	"sync"

	"xhc2_for_studying/server/core"
)

// ErrBeaconNotFound is returned when a beacon lookup by ID fails.
var ErrBeaconNotFound = errors.New("beacon not found")

// MemoryBeaconStore holds the in-memory registry of all active beacons,
// protected by a read-write mutex.
type MemoryBeaconStore struct {
	mu      sync.RWMutex
	beacons map[string]*core.Beacon
}

// NewBeaconStore creates and returns an initialized BeaconStore.
func NewBeaconStore() BeaconStore {
	return &MemoryBeaconStore{
		beacons: make(map[string]*core.Beacon),
	}
}

// Add inserts or overwrites the beacon record keyed by its ID. Nil values and
// empty IDs are silently ignored.
func (s *MemoryBeaconStore) Add(beacon *core.Beacon) error {
	if beacon == nil || beacon.ID == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	beaconCopy := *beacon
	s.beacons[beacon.ID] = &beaconCopy
	return nil
}

// Get returns a copy of the beacon identified by id, or ErrBeaconNotFound.
func (s *MemoryBeaconStore) Get(id string) (*core.Beacon, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	beacon, ok := s.beacons[id]
	if !ok {
		return nil, ErrBeaconNotFound
	}

	beaconCopy := *beacon
	return &beaconCopy, nil
}

// UpdateCheckIn updates the last check-in timestamp and, if non-empty, the
// remote address for the given beacon id.
func (s *MemoryBeaconStore) UpdateCheckIn(id string, lastCheckIn int64, remoteAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	beacon, ok := s.beacons[id]
	if !ok {
		return ErrBeaconNotFound
	}

	beacon.LastCheckIn = lastCheckIn
	if remoteAddress != "" {
		beacon.RemoteAddress = remoteAddress
	}

	return nil
}

// ListIDs returns all currently registered beacon IDs in no particular order.
func (s *MemoryBeaconStore) ListIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.beacons))
	for id := range s.beacons {
		ids = append(ids, id)
	}

	return ids
}

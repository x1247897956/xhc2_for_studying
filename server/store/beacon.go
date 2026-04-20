package store

import (
	"errors"
	"sync"
	
	"xhc2_for_studying/server/core"
)

var ErrBeaconNotFound = errors.New("beacon not found")

type BeaconStore struct {
	mu      sync.RWMutex
	beacons map[string]*core.Beacon
}

func NewBeaconStore() *BeaconStore {
	return &BeaconStore{
		beacons: make(map[string]*core.Beacon),
	}
}

func (s *BeaconStore) Add(beacon *core.Beacon) {
	if beacon == nil || beacon.ID == "" {
		return
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	beaconCopy := *beacon
	s.beacons[beacon.ID] = &beaconCopy
}

func (s *BeaconStore) Get(id string) (*core.Beacon, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	beacon, ok := s.beacons[id]
	if !ok {
		return nil, ErrBeaconNotFound
	}
	
	beaconCopy := *beacon
	return &beaconCopy, nil
}

func (s *BeaconStore) UpdateCheckIn(id string, lastCheckIn int64, remoteAddress string) error {
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

func (s *BeaconStore) ListIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ids := make([]string, 0, len(s.beacons))
	for id := range s.beacons {
		ids = append(ids, id)
	}
	
	return ids
}

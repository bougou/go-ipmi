package bmc

import "sync"

// SDRRepoStore tracks the active SDR repository reservation (§33.11).
type SDRRepoStore struct {
	mu            sync.Mutex
	reservationID uint16
	generation    uint16
}

// NewSDRRepoStore returns an empty SDR reservation tracker.
func NewSDRRepoStore() *SDRRepoStore {
	return &SDRRepoStore{}
}

// Reserve invalidates any prior reservation and returns a new non-zero ID.
func (s *SDRRepoStore) Reserve() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.generation++
	if s.generation == 0 {
		s.generation = 1
	}
	s.reservationID = s.generation
	return s.reservationID
}

// Validate reports whether id matches the active reservation.
func (s *SDRRepoStore) Validate(id uint16) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return id != 0 && id == s.reservationID
}

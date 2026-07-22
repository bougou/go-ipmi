package mock

import (
	"context"
	"sort"
	"sync"

	"github.com/bougou/go-ipmi/pkg/hal"
)

// Storage is the mock [hal.StorageHAL].
type Storage struct {
	mu  sync.RWMutex
	fru map[uint8][]byte
	sdr map[uint16][]byte
}

func (s *Storage) FRU() hal.FRUStore { return (*fruStore)(s) }
func (s *Storage) SDR() hal.SDRStore { return (*sdrStore)(s) }

type fruStore Storage

func (f *fruStore) Read(_ context.Context, deviceID uint8) ([]byte, error) {
	s := (*Storage)(f)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.fru[deviceID]
	if !ok {
		return nil, hal.ErrNotFound
	}
	return append([]byte(nil), v...), nil
}

func (f *fruStore) Write(_ context.Context, deviceID uint8, data []byte) error {
	s := (*Storage)(f)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fru == nil {
		s.fru = map[uint8][]byte{}
	}
	s.fru[deviceID] = append([]byte(nil), data...)
	return nil
}

func (f *fruStore) Delete(_ context.Context, deviceID uint8) error {
	s := (*Storage)(f)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.fru, deviceID)
	return nil
}

func (f *fruStore) DeviceIDs(_ context.Context) ([]uint8, error) {
	s := (*Storage)(f)
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]uint8, 0, len(s.fru))
	for id := range s.fru {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

type sdrStore Storage

func (d *sdrStore) Read(_ context.Context, recordID uint16) ([]byte, error) {
	s := (*Storage)(d)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.sdr[recordID]
	if !ok {
		return nil, hal.ErrNotFound
	}
	return append([]byte(nil), v...), nil
}

func (d *sdrStore) Write(_ context.Context, recordID uint16, data []byte) error {
	s := (*Storage)(d)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sdr == nil {
		s.sdr = map[uint16][]byte{}
	}
	s.sdr[recordID] = append([]byte(nil), data...)
	return nil
}

func (d *sdrStore) Delete(_ context.Context, recordID uint16) error {
	s := (*Storage)(d)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sdr, recordID)
	return nil
}

func (d *sdrStore) RecordIDs(_ context.Context) ([]uint16, error) {
	s := (*Storage)(d)
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]uint16, 0, len(s.sdr))
	for id := range s.sdr {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

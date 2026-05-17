package bmc

import (
	"errors"
	"fmt"
	"sync"
)

// ErrChannelNotFound is returned when the requested channel number is not configured.
var ErrChannelNotFound = errors.New("channel not found")

// ChannelMedium identifies the physical medium of a channel (LAN, serial, etc.).
type ChannelMedium uint8

const (
	ChannelMediumIPMBv10  ChannelMedium = 0x01
	ChannelMediumICMB     ChannelMedium = 0x02
	ChannelMediumLAN      ChannelMedium = 0x04
	ChannelMediumSerial   ChannelMedium = 0x05
	ChannelMediumSMBus    ChannelMedium = 0x06
	ChannelMediumSMBusv20 ChannelMedium = 0x07
	ChannelMediumUSBv1    ChannelMedium = 0x08
	ChannelMediumUSBv2    ChannelMedium = 0x09
	ChannelMediumSystemIF ChannelMedium = 0x0C
)

// ChannelAccessMode controls whether a channel accepts connections.
type ChannelAccessMode uint8

const (
	ChannelAccessDisabled    ChannelAccessMode = 0x00
	ChannelAccessPreBootOnly ChannelAccessMode = 0x01
	ChannelAccessAlways      ChannelAccessMode = 0x02
	ChannelAccessShared      ChannelAccessMode = 0x03
)

// Channel holds the configuration for a single IPMI channel.
type Channel struct {
	Number     uint8
	Medium     ChannelMedium
	AccessMode ChannelAccessMode
	// MaxPrivilege is the maximum privilege level allowed on this channel.
	MaxPrivilege PrivilegeLevel
	// PerMessageAuth and UserLevelAuth reflect the channel security settings.
	PerMessageAuth bool
	UserLevelAuth  bool
	// PEFAlerts controls whether PEF alerting is enabled on this channel.
	PEFAlerts bool
}

// ChannelStore holds the configuration for all BMC channels.
//
// Channel numbers follow the IPMI spec:
//   - 0x00 – primary IPMB
//   - 0x01-0x0B – implementation-specific
//   - 0x0E – current channel (self-reference, resolved by caller)
//   - 0x0F – system interface
type ChannelStore struct {
	mu       sync.RWMutex
	channels map[uint8]*Channel
}

// NewChannelStore returns a ChannelStore pre-populated with a default LAN channel (1)
// and the system interface (15 / 0x0F).
func NewChannelStore() *ChannelStore {
	s := &ChannelStore{channels: make(map[uint8]*Channel, 4)}
	// Channel 1: LAN
	s.channels[1] = &Channel{
		Number:         1,
		Medium:         ChannelMediumLAN,
		AccessMode:     ChannelAccessAlways,
		MaxPrivilege:   PrivilegeLevelAdministrator,
		PerMessageAuth: true,
		UserLevelAuth:  true,
	}
	// Channel 15: System Interface
	s.channels[0x0F] = &Channel{
		Number:       0x0F,
		Medium:       ChannelMediumSystemIF,
		AccessMode:   ChannelAccessAlways,
		MaxPrivilege: PrivilegeLevelAdministrator,
	}
	return s
}

// Get returns the channel at number n, or [ErrChannelNotFound].
func (s *ChannelStore) Get(n uint8) (*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.channels[n]
	if !ok {
		return nil, fmt.Errorf("channel %d: %w", n, ErrChannelNotFound)
	}
	return ch, nil
}

// Set adds or replaces the channel at number n.
func (s *ChannelStore) Set(ch *Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[ch.Number] = ch
}

// All returns a snapshot of all configured channels.
func (s *ChannelStore) All() []*Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Channel, 0, len(s.channels))
	for _, ch := range s.channels {
		out = append(out, ch)
	}
	return out
}

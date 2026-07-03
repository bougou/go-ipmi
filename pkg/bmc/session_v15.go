package bmc

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
)

// V15AuthType mirrors IPMI v1.5 authentication type codes.
type V15AuthType uint8

const (
	V15AuthTypeNone     V15AuthType = 0x00
	V15AuthTypeMD2      V15AuthType = 0x01
	V15AuthTypeMD5      V15AuthType = 0x02
	V15AuthTypePassword V15AuthType = 0x04
	V15AuthTypeOEM      V15AuthType = 0x05
)

// v15InboundWindow is the inbound sequence sliding window (spec §6.11.11 Option 1).
const v15InboundWindow = 8

// DefaultV15AuthTypes is the default set of v1.5 auth types the reference BMC
// advertises and accepts.
var DefaultV15AuthTypes = []V15AuthType{V15AuthTypeMD5}

// V15AuthTypeToCapsBit maps an auth type to the corresponding bit in Get
// Channel Authentication Capabilities response byte 3 (bits [5:0]).
func V15AuthTypeToCapsBit(t V15AuthType) uint8 {
	switch t {
	case V15AuthTypeNone:
		return 1 << 0
	case V15AuthTypeMD2:
		return 1 << 1
	case V15AuthTypeMD5:
		return 1 << 2
	case V15AuthTypePassword:
		return 1 << 4
	case V15AuthTypeOEM:
		return 1 << 5
	default:
		return 0
	}
}

// V15SessionState tracks IPMI v1.5 session negotiation progress.
type V15SessionState uint8

const (
	V15SessionStatePending V15SessionState = iota
	V15SessionStateActive
	V15SessionStateClosed
)

// V15Session holds IPMI v1.5 session state.
type V15Session struct {
	TempSessionID uint32
	SessionID     uint32
	State         V15SessionState

	AuthType  V15AuthType
	Challenge [16]byte

	InboundSeq  uint32
	InboundRcvd uint8 // bitmap: bit i => (InboundSeq - i) received
	OutboundSeq uint32

	User           *User
	PrivilegeLevel PrivilegeLevel
	MaxPrivilege   PrivilegeLevel

	Channel uint8

	CreatedAt    time.Time
	LastActivity time.Time
}

// V15SessionStore is a thread-safe registry of IPMI v1.5 sessions.
type V15SessionStore struct {
	mu       sync.Mutex
	sessions map[uint32]*V15Session
	max      int
	timeout  time.Duration
	clock    clock.Clock
}

// NewV15SessionStore creates a V15SessionStore with the default limits.
func NewV15SessionStore(clk clock.Clock) *V15SessionStore {
	return &V15SessionStore{
		sessions: make(map[uint32]*V15Session, MaxSessions),
		max:      MaxSessions,
		timeout:  DefaultInactivityTimeout,
		clock:    clk,
	}
}

// CreatePending allocates a pending v1.5 session after Get Session Challenge.
func (s *V15SessionStore) CreatePending(authType V15AuthType, user *User, challenge [16]byte, channel uint8) (*V15Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evictExpiredLocked()

	if len(s.sessions) >= s.max {
		if !s.evictOldestPendingLocked() {
			return nil, ErrSessionFull
		}
	}

	tempID, err := randomUint32()
	if err != nil {
		return nil, fmt.Errorf("generate temp session ID: %w", err)
	}
	for s.sessions[tempID] != nil || tempID == 0 {
		tempID, err = randomUint32()
		if err != nil {
			return nil, fmt.Errorf("generate temp session ID: %w", err)
		}
	}

	now := s.clock.Now()
	sess := &V15Session{
		TempSessionID: tempID,
		State:         V15SessionStatePending,
		AuthType:      authType,
		Challenge:     challenge,
		User:          user,
		Channel:       channel,
		CreatedAt:     now,
		LastActivity:  now,
	}
	s.sessions[tempID] = sess
	return sess, nil
}

// Get returns a session by its current lookup ID without updating activity.
func (s *V15SessionStore) Get(id uint32) (*V15Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("v1.5 session 0x%08x: %w", id, ErrNoSession)
	}
	return sess, nil
}

// Touch records valid session activity for inactivity timeout (spec §6.11.13).
func (s *V15SessionStore) Touch(sess *V15Session) {
	if sess == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess.LastActivity = s.clock.Now()
}

// CountActiveSessions returns the number of active v1.5 sessions.
func (s *V15SessionStore) CountActiveSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive {
			n++
		}
	}
	return n
}

// CountActiveSessionsForUser returns active sessions owned by userID.
func (s *V15SessionStore) CountActiveSessionsForUser(userID uint8) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive && sess.User != nil && sess.User.ID == userID {
			n++
		}
	}
	return n
}

// CountActiveSessionsWithMaxPrivilegeAtLeast counts active sessions whose
// negotiated maximum privilege is >= min (for Table 18-17 completion 0x83).
func (s *V15SessionStore) CountActiveSessionsWithMaxPrivilegeAtLeast(min PrivilegeLevel) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive && sess.MaxPrivilege >= min {
			n++
		}
	}
	return n
}

// Activate transitions a pending session to active with a new permanent ID.
// maxPrivilege is the requested ceiling; initial privilege is USER per §18.16
// (Callback when max is Callback).
func (s *V15SessionStore) Activate(pending *V15Session, permanentID, inboundSeq, outboundSeq uint32, maxPrivilege PrivilegeLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pending.State != V15SessionStatePending {
		return errors.New("session is not pending")
	}
	delete(s.sessions, pending.TempSessionID)

	for s.sessions[permanentID] != nil || permanentID == 0 {
		var err error
		permanentID, err = randomUint32()
		if err != nil {
			return fmt.Errorf("generate permanent session ID: %w", err)
		}
	}

	initialPriv := PrivilegeLevelUser
	if maxPrivilege == PrivilegeLevelCallback {
		initialPriv = PrivilegeLevelCallback
	}

	pending.SessionID = permanentID
	pending.State = V15SessionStateActive
	pending.InboundSeq = inboundSeq
	pending.InboundRcvd = 0
	pending.OutboundSeq = outboundSeq
	pending.PrivilegeLevel = initialPriv
	pending.MaxPrivilege = maxPrivilege
	pending.LastActivity = s.clock.Now()

	s.sessions[permanentID] = pending
	return nil
}

// Close removes a session by permanent or temp ID.
func (s *V15SessionStore) Close(id uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("v1.5 session 0x%08x: %w", id, ErrNoSession)
	}
	delete(s.sessions, id)
	if sess.State == V15SessionStateActive && sess.TempSessionID != id && sess.TempSessionID != 0 {
		delete(s.sessions, sess.TempSessionID)
	}
	return nil
}

// EvictExpired removes inactive v1.5 sessions past the timeout.
func (s *V15SessionStore) EvictExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.evictExpiredLocked()
}

func (s *V15SessionStore) evictExpiredLocked() int {
	now := s.clock.Now()
	limit := s.timeout + DefaultInactivityTimeoutTolerance
	n := 0
	for id, sess := range s.sessions {
		if now.Sub(sess.LastActivity) > limit {
			delete(s.sessions, id)
			n++
		}
	}
	return n
}

func (s *V15SessionStore) evictOldestPendingLocked() bool {
	var oldest *V15Session
	var oldestID uint32
	for id, sess := range s.sessions {
		if sess.State == V15SessionStatePending {
			if oldest == nil || sess.CreatedAt.Before(oldest.CreatedAt) {
				oldest = sess
				oldestID = id
			}
		}
	}
	if oldest == nil {
		return false
	}
	delete(s.sessions, oldestID)
	return true
}

// v15SeqDiff returns seq-high as a signed delta with uint32 wrap-around.
func v15SeqDiff(high, seq uint32) int64 {
	diff := int64(seq) - int64(high)
	if diff > 1<<31 {
		diff -= 1 << 32
	} else if diff < -(1 << 31) {
		diff += 1 << 32
	}
	return diff
}

// TryAcceptInboundSeq implements spec §6.11.11 Option 1 (+/-8 window, no dupes).
func (sess *V15Session) TryAcceptInboundSeq(seq uint32) bool {
	if seq == 0 {
		return false
	}
	high := sess.InboundSeq
	diff := v15SeqDiff(high, seq)

	if diff == 0 {
		return false
	}
	if diff > v15InboundWindow || diff < -v15InboundWindow {
		return false
	}

	if diff > 0 {
		shift := uint(diff)
		if shift > v15InboundWindow {
			return false
		}
		sess.InboundRcvd <<= shift
		sess.InboundRcvd |= 1
		sess.InboundSeq = seq
		return true
	}

	behind := uint(-diff)
	bit := uint8(1) << (behind - 1)
	if sess.InboundRcvd&bit != 0 {
		return false
	}
	sess.InboundRcvd |= bit
	return true
}

// V15InboundSeqValid reports whether seq is acceptable under Option 1 without
// mutating session state (for tests).
func V15InboundSeqValid(sess *V15Session, seq uint32) bool {
	if sess == nil || seq == 0 {
		return false
	}
	high := sess.InboundSeq
	diff := v15SeqDiff(high, seq)
	if diff == 0 {
		return false
	}
	if diff > v15InboundWindow || diff < -v15InboundWindow {
		return false
	}
	if diff < 0 {
		bit := uint8(1) << (uint(-diff) - 1)
		return sess.InboundRcvd&bit == 0
	}
	return true
}

// NextOutboundSeq returns the sequence number for the current outbound message
// and advances the counter for the next one.
func (sess *V15Session) NextOutboundSeq() uint32 {
	seq := sess.OutboundSeq
	sess.OutboundSeq++
	return seq
}

// GenerateChallenge fills dst with random bytes for Get Session Challenge.
func GenerateChallenge(dst *[16]byte) error {
	_, err := rand.Read(dst[:])
	return err
}

// GenerateInboundSeq returns a non-zero initial inbound sequence number.
func GenerateInboundSeq() (uint32, error) {
	seq, err := randomUint32()
	if err != nil {
		return 0, err
	}
	if seq == 0 {
		seq = 1
	}
	return seq, nil
}

// PackSessionIDLE is a helper for auth code input construction.
func PackSessionIDLE(id uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, id)
	return b
}

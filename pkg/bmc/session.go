package bmc

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Session inactivity timeout per IPMI spec:
//   - v1.5 §6.11.13 Session Inactivity Timeout
//   - v2.0 §6.12.15 Session Inactivity Timeout
const DefaultInactivityTimeout = 60 * time.Second

// DefaultSessionEvictInterval is how often the server scans for idle sessions.
// The spec defines the 60-second inactivity limit, not the scan period.
const DefaultSessionEvictInterval = 3 * time.Second

// DefaultInactivityTimeoutTolerance is the LAN inactivity tolerance per
// IPMI v1.5 Table 6-7 (+/- 3 seconds).
const DefaultInactivityTimeoutTolerance = 3 * time.Second

// MaxSessions is the minimum number of concurrent sessions required by the spec.
const MaxSessions = 4

// ErrNoSession is returned when the session ID is not in the store.
var ErrNoSession = errors.New("session not found")

// ErrSessionFull is returned when the store has reached capacity.
var ErrSessionFull = errors.New("no session slots available")

// SessionState tracks which phase of session negotiation has been reached.
type SessionState uint8

const (
	// SessionStatePending means Open Session was received but RAKP is incomplete.
	SessionStatePending SessionState = iota
	// SessionStateActive means RAKP completed and commands may flow.
	SessionStateActive
	// SessionStateClosed means the session was explicitly closed or timed out.
	SessionStateClosed
)

// Session holds all state for one active or pending IPMI session.
type Session struct {
	// BMCID is the session ID assigned by the BMC (sent in Open Session Response).
	BMCID uint32
	// ConsoleID is the session ID chosen by the remote console.
	ConsoleID uint32

	State SessionState

	// Negotiated algorithms
	AuthAlg      types.AuthAlg
	IntegrityAlg types.IntegrityAlg
	CryptAlg     types.CryptAlg

	// Sequence tracking.
	// InboundSeq is the last accepted sequence number from the console.
	// OutboundSeq is the next sequence number the BMC will use.
	InboundSeq  uint32
	OutboundSeq uint32

	// Session keys derived during RAKP.
	SIK []byte
	K1  []byte
	K2  []byte

	// RAKP exchange state (zeroed once session is active).
	ConsoleRand [16]byte
	BMCRand     [16]byte
	Role        uint8 // whole byte from RAKP1, used in HMAC input

	// User and privilege
	User           *User
	PrivilegeLevel PrivilegeLevel
	MaxPrivilege   PrivilegeLevel

	// Channel this session arrived on.
	Channel uint8

	// Timing
	CreatedAt    time.Time
	LastActivity time.Time
}

// SessionStore is a thread-safe registry of active and pending sessions.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[uint32]*Session
	max      int
	timeout  time.Duration
	clock    clock.Clock
}

// NewSessionStore creates a SessionStore limited to [MaxSessions] concurrent sessions
// with the default inactivity timeout.
func NewSessionStore(clk clock.Clock) *SessionStore {
	return &SessionStore{
		sessions: make(map[uint32]*Session, MaxSessions),
		max:      MaxSessions,
		timeout:  DefaultInactivityTimeout,
		clock:    clk,
	}
}

// Option configures a [SessionStore].
type SessionStoreOption func(*SessionStore)

// WithMaxSessions overrides the default session limit.
func WithMaxSessions(n int) SessionStoreOption {
	return func(s *SessionStore) { s.max = n }
}

// WithInactivityTimeout overrides the default 60-second inactivity timeout.
func WithInactivityTimeout(d time.Duration) SessionStoreOption {
	return func(s *SessionStore) { s.timeout = d }
}

// NewSessionStoreWithOptions creates a SessionStore with custom options.
func NewSessionStoreWithOptions(clk clock.Clock, opts ...SessionStoreOption) *SessionStore {
	s := NewSessionStore(clk)
	for _, o := range opts {
		o(s)
	}
	return s
}

// Allocate creates a new pending session and returns it.
// If capacity is reached, it evicts the oldest pending session (LRU per spec).
// Returns [ErrSessionFull] only when all slots are occupied by active sessions.
func (s *SessionStore) Allocate(consoleID uint32, authAlg types.AuthAlg, integrityAlg types.IntegrityAlg, cryptAlg types.CryptAlg) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evictExpiredLocked()

	if len(s.sessions) >= s.max {
		// Evict oldest pending session if any exist.
		if !s.evictOldestPendingLocked() {
			return nil, ErrSessionFull
		}
	}

	bmcID, err := randomUint32()
	if err != nil {
		return nil, fmt.Errorf("generate session ID: %w", err)
	}
	// Avoid collision with existing IDs.
	for s.sessions[bmcID] != nil || bmcID == 0 {
		bmcID, err = randomUint32()
		if err != nil {
			return nil, fmt.Errorf("generate session ID: %w", err)
		}
	}

	now := s.clock.Now()
	sess := &Session{
		BMCID:        bmcID,
		ConsoleID:    consoleID,
		State:        SessionStatePending,
		AuthAlg:      authAlg,
		IntegrityAlg: integrityAlg,
		CryptAlg:     cryptAlg,
		CreatedAt:    now,
		LastActivity: now,
	}
	s.sessions[bmcID] = sess
	return sess, nil
}

// Get returns the session for bmcID, or [ErrNoSession].
// It also updates [Session.LastActivity].
func (s *SessionStore) Get(bmcID uint32) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[bmcID]
	if !ok {
		return nil, fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	sess.LastActivity = s.clock.Now()
	return sess, nil
}

// Close marks a session as closed and removes it from the store.
func (s *SessionStore) Close(bmcID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[bmcID]; !ok {
		return fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	delete(s.sessions, bmcID)
	return nil
}

// EvictExpired removes all sessions that have been inactive beyond the timeout.
// Called periodically by the server.
func (s *SessionStore) EvictExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.evictExpiredLocked()
}

func (s *SessionStore) evictExpiredLocked() int {
	now := s.clock.Now()
	n := 0
	for id, sess := range s.sessions {
		if now.Sub(sess.LastActivity) > s.timeout+DefaultInactivityTimeoutTolerance {
			delete(s.sessions, id)
			n++
		}
	}
	return n
}

// evictOldestPendingLocked removes the oldest pending session.
// Returns false if no pending sessions exist.
func (s *SessionStore) evictOldestPendingLocked() bool {
	var oldest *Session
	for _, sess := range s.sessions {
		if sess.State == SessionStatePending {
			if oldest == nil || sess.CreatedAt.Before(oldest.CreatedAt) {
				oldest = sess
			}
		}
	}
	if oldest == nil {
		return false
	}
	delete(s.sessions, oldest.BMCID)
	return true
}

// Count returns the number of sessions currently in the store.
func (s *SessionStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

// InboundSeqValid checks whether seq is within the acceptable sliding window
// defined by the IPMI spec (section 6.12.13):  +15 / -16 of the last accepted value.
// Session sequence numbers start at 1; 0 is reserved for pre-session packets.
func InboundSeqValid(last, seq uint32) bool {
	if seq == 0 {
		return false
	}
	diff := int64(seq) - int64(last)
	return diff >= -16 && diff <= 15
}

func randomUint32() (uint32, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

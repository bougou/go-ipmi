package bmc

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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
	// Handle is the one-byte session handle Get Session Info reports
	// (spec v2.0§22.20). Assigned at allocation, unique among the store's live
	// sessions, never 0x00 ("no session") or the reserved 0xFF.
	Handle uint8

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

	// Addr is the remote console's transport address, refreshed on every
	// inbound packet. Asynchronous traffic (SOL data push, §15.3) targets it.
	// Guarded by addrMu; use [Session.SetAddr] / [Session.GetAddr] rather
	// than touching the field directly. The SOL pump reads it from a
	// goroutine that must not take ProcMu (a packet handler holding ProcMu
	// can wait on the pump during deactivation — the reverse order would
	// deadlock), while packet handling writes it under ProcMu.
	addrMu sync.Mutex
	Addr   net.Addr

	// seqMu serializes outbound sequence assignment: command responses and
	// asynchronous SOL packets share one space (§15.5) and are sent from
	// different goroutines.
	seqMu sync.Mutex

	// ProcMu serializes per-session packet processing. The server spawns a
	// goroutine per inbound packet; without this lock, a burst of packets
	// from one session is processed in scheduler order — scrambling SOL
	// keystroke bytes and racing the inbound-seq check-then-set. The RAKP
	// handlers take it too: duplicate handshake messages for one pending
	// session otherwise race on the nonces, derived keys, and state below.
	// ProcMu may be held while taking the store lock, never the reverse.
	ProcMu sync.Mutex

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

	// Timing. LastActivity is guarded by the store lock: it is refreshed via
	// [SessionStore.Touch] when a validated packet is processed and read by
	// eviction, both under that lock. CreatedAt is set before the session is
	// published into the store and never written again.
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
	// nextHandle seeds session handle assignment; see allocHandleLocked.
	nextHandle uint8

	// onRemove fires whenever a session leaves the store (Close or
	// eviction). The BMC wires this to payload deactivation: when a session
	// terminates, payloads active under it are automatically deactivated
	// (spec v2.0 §24.2). It is invoked without the store lock: payload
	// teardown can block (an SOL instance close waits for its pump
	// goroutine, which may be mid reconnect dial), and holding the lock
	// across that would freeze every session lookup and allocation.
	onRemove func(bmcID uint32)
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
//
// maxPriv and channel are stored before the session is inserted into the map so
// the struct is fully initialized before it becomes reachable to other
// goroutines; callers must not write session fields after Allocate returns
// without holding [Session.ProcMu].
func (s *SessionStore) Allocate(consoleID uint32, authAlg types.AuthAlg, integrityAlg types.IntegrityAlg, cryptAlg types.CryptAlg, maxPriv PrivilegeLevel, channel uint8) (*Session, error) {
	s.mu.Lock()

	// Collect the evicted IDs so their removal hooks (payload deactivation,
	// which can block) run after the lock is released — even on the error
	// paths below, the evicted sessions are gone and must be cleaned up.
	removed := s.evictExpiredLocked()
	defer func() {
		s.mu.Unlock()
		s.fireRemoveAll(removed)
	}()

	if len(s.sessions) >= s.max {
		// Evict oldest pending session if any exist.
		if id, ok := s.evictOldestPendingLocked(); ok {
			removed = append(removed, id)
		} else {
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
		Handle:       s.allocHandleLocked(),
		State:        SessionStatePending,
		AuthAlg:      authAlg,
		IntegrityAlg: integrityAlg,
		CryptAlg:     cryptAlg,
		MaxPrivilege: maxPriv,
		Channel:      channel,
		CreatedAt:    now,
		LastActivity: now,
	}
	s.sessions[bmcID] = sess
	return sess, nil
}

// allocHandleLocked returns the next free one-byte session handle: a rotating
// counter skipping 0x00 ("no session"), the reserved 0xFF, and any handle a
// live session holds. Two colliding handles would make Get Session Info
// ambiguous about which session it describes. s.mu must be held; the store
// capacity is far below 254, so a free value always exists.
func (s *SessionStore) allocHandleLocked() uint8 {
	for {
		s.nextHandle++
		if s.nextHandle == 0 || s.nextHandle == 0xFF {
			s.nextHandle = 1
		}
		taken := false
		for _, sess := range s.sessions {
			if sess.Handle == s.nextHandle {
				taken = true
				break
			}
		}
		if !taken {
			return s.nextHandle
		}
	}
}

// Get returns the session for bmcID, or [ErrNoSession]. It is a pure lookup:
// activity is refreshed separately via [SessionStore.Touch], only once a packet
// has passed integrity and sequence validation. Refreshing on lookup would let
// packets that fail validation, or that merely name a session ID, keep the
// session alive forever.
func (s *SessionStore) Get(bmcID uint32) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[bmcID]
	if !ok {
		return nil, fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	return sess, nil
}

// Touch refreshes the session's inactivity clock. The server calls it for every
// packet that passed integrity and sequence validation. Handshake packets never
// touch: RAKP messages carry no authenticator, so the inactivity budget stamped
// at allocation bounds the whole handshake instead.
func (s *SessionStore) Touch(bmcID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[bmcID]; ok {
		sess.LastActivity = s.clock.Now()
	}
}

// Activate marks the session active after RAKP completes, or returns
// [ErrNoSession] when the session was evicted in the meantime (capacity
// pressure evicts the oldest pending session, which can be one whose RAKP3 is
// in flight); activating the orphaned struct would hand the console a
// successful RAKP4 for a session that no longer exists. It takes the store
// lock because the pending-session eviction scan reads State under it, while
// the caller holds [Session.ProcMu], so a reader under either lock observes a
// consistent value.
func (s *SessionStore) Activate(bmcID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[bmcID]
	if !ok {
		return fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	sess.State = SessionStateActive
	return nil
}

// Close marks a session as closed and removes it from the store.
func (s *SessionStore) Close(bmcID uint32) error {
	s.mu.Lock()
	if _, ok := s.sessions[bmcID]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	delete(s.sessions, bmcID)
	s.mu.Unlock()
	s.fireRemove(bmcID)
	return nil
}

// SetOnRemove registers the hook fired when a session leaves the store.
func (s *SessionStore) SetOnRemove(fn func(bmcID uint32)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRemove = fn
}

// fireRemove invokes the removal hook for one session ID outside the store
// lock; the hook may block (SOL teardown waits on the pump), so it must never
// run under s.mu.
func (s *SessionStore) fireRemove(bmcID uint32) {
	if s.onRemove != nil {
		s.onRemove(bmcID)
	}
}

// fireRemoveAll invokes the removal hook for every ID collected by a locked
// eviction pass.
func (s *SessionStore) fireRemoveAll(ids []uint32) {
	for _, id := range ids {
		s.fireRemove(id)
	}
}

// EvictExpired removes all sessions that have been inactive beyond the timeout.
// Called periodically by the server.
func (s *SessionStore) EvictExpired() int {
	s.mu.Lock()
	removed := s.evictExpiredLocked()
	s.mu.Unlock()
	s.fireRemoveAll(removed)
	return len(removed)
}

// evictExpiredLocked removes sessions inactive beyond the timeout and returns
// their IDs. s.mu must be held.
func (s *SessionStore) evictExpiredLocked() []uint32 {
	now := s.clock.Now()
	var removed []uint32
	for id, sess := range s.sessions {
		if now.Sub(sess.LastActivity) > s.timeout+DefaultInactivityTimeoutTolerance {
			delete(s.sessions, id)
			removed = append(removed, id)
		}
	}
	return removed
}

// evictOldestPendingLocked removes the oldest pending session and returns its
// ID. ok=false when no pending sessions exist. s.mu must be held.
func (s *SessionStore) evictOldestPendingLocked() (uint32, bool) {
	var oldest *Session
	for _, sess := range s.sessions {
		if sess.State == SessionStatePending {
			if oldest == nil || sess.CreatedAt.Before(oldest.CreatedAt) {
				oldest = sess
			}
		}
	}
	if oldest == nil {
		return 0, false
	}
	delete(s.sessions, oldest.BMCID)
	return oldest.BMCID, true
}

// Count returns the number of sessions currently in the store.
func (s *SessionStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

// Cap returns the maximum number of concurrent sessions the store can hold,
// i.e. the number of slots in the session table.
func (s *SessionStore) Cap() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max
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

// SetAddr records the console's transport address under the guard. The
// server calls it on every accepted inbound packet.
func (sess *Session) SetAddr(addr net.Addr) {
	sess.addrMu.Lock()
	sess.Addr = addr
	sess.addrMu.Unlock()
}

// GetAddr returns the console's transport address, safe to call from any
// goroutine (the SOL pump targets asynchronous data at it).
func (sess *Session) GetAddr() net.Addr {
	sess.addrMu.Lock()
	defer sess.addrMu.Unlock()
	return sess.Addr
}

// NextOutboundSeq returns the session sequence number for the next outbound
// packet, advancing the counter shared by command responses and async SOL
// packets (§15.5).
func (sess *Session) NextOutboundSeq() uint32 {
	sess.seqMu.Lock()
	defer sess.seqMu.Unlock()
	sess.OutboundSeq++
	return sess.OutboundSeq
}

func randomUint32() (uint32, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

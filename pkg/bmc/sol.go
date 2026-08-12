package bmc

// SOL (Serial over LAN) state: per-channel configuration parameters
// (spec v2.0 Table 26-5) and the payload instance state machine
// (spec v2.0 §15.9 payload data format, §15.11 acknowledge and retries).

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	// SOLMaxInstances is the number of simultaneously activatable SOL payload
	// instances (spec v2.0 Table 24-6, 1-based). One instance matches the
	// single shared serial port of the reference hardware model (§15.3).
	SOLMaxInstances = 1

	// SOLMaxPayloadChars bounds character data per SOL packet. The Accepted
	// Character Count field is one byte (Table 15-2) and the 4-byte SOL
	// packet header shares the 255-byte reported payload size.
	SOLMaxPayloadChars = 251

	// SOLRXBufferCap bounds console output buffered between remote-console
	// polls. A full buffer applies backpressure (draining pauses until the
	// console catches up) — the in-tree equivalent of the BMC deasserting
	// CTS (§15.6) — so buffered data is never silently dropped.
	SOLRXBufferCap = 4096

	// SOLPayloadUDPPortDefault is the primary RMCP port reported when the
	// server transport does not expose its bound address.
	SOLPayloadUDPPortDefault = 623
)

// SOL activation failure reasons, mapped by the Activate Payload handler to
// the command-specific completion codes named in pkg/types (spec v2.0
// Table 24-2 / Table 24-3 / Table 24-5).
var (
	// ErrSOLAlreadyActive → CodeActivatePayloadAlreadyActive (Table 24-2).
	ErrSOLAlreadyActive = errors.New("SOL payload already active")
	// ErrSOLDisabled → CodeActivatePayloadTypeDisabled (Table 24-2).
	ErrSOLDisabled = errors.New("SOL payload type is disabled")
	// ErrSOLPrivilege → CodeInsufficientPrivilege: session privilege below the
	// configured SOL level.
	ErrSOLPrivilege = errors.New("insufficient privilege to activate SOL")
	// ErrSOLEncryptionUnavailable → CodeActivatePayloadCannotActivateWithEncryption
	// (Table 24-2): the session negotiated no encryption algorithm.
	ErrSOLEncryptionUnavailable = errors.New("cannot activate SOL with encryption")
	// ErrSOLEncryptionRequired → CodeActivatePayloadCannotActivateWithoutEncryption
	// (Table 24-2): policy forces encryption the console declined.
	ErrSOLEncryptionRequired = errors.New("cannot activate SOL without encryption")
	// ErrSOLAuthenticationUnavailable → CodeRequestDataFieldInvalid: Table 24-2
	// defines no authentication-specific completion code, so an unsatisfiable
	// authentication request/policy falls back to the generic invalid-data-field
	// code.
	ErrSOLAuthenticationUnavailable = errors.New("cannot activate SOL with authentication")

	// ErrSOLInstanceNotActive → CodeSuspendResumePayloadEncryptionNotActive
	// (Table 24-5): the session owns no active SOL payload instance.
	ErrSOLInstanceNotActive = errors.New("SOL payload instance not active")
	// ErrSOLEncryptionForced → CodeSuspendResumePayloadEncryptionNotAllowed
	// (Table 24-5): SOL configuration parameter #2 forces encryption, so
	// suspending it is not allowed.
	ErrSOLEncryptionForced = errors.New("SOL encryption forced by configuration")
	// ErrSOLEncryptionUnavailableForSession →
	// CodeSuspendResumePayloadEncryptionNotAvailable (Table 24-5): the session
	// negotiated no encryption algorithm at open time.
	ErrSOLEncryptionUnavailableForSession = errors.New("encryption not available for session")
	// ErrSOLOperationUnsupported → CodeSuspendResumePayloadEncryptionNotSupported
	// (Table 24-5): IV regeneration is xRC4-specific; AES-CBC payloads draw a
	// fresh IV per packet already.
	ErrSOLOperationUnsupported = errors.New("operation not supported for SOL payload")
	// ErrSOLNotActive → CodeDeactivatePayloadAlreadyDeactivated (Table 24-3).
	ErrSOLNotActive = errors.New("SOL payload not active")
	// ErrSOLNotOwner → CodeInsufficientPrivilege (Deactivate): session owns no
	// instance and lacks the privilege to force-deactivate another session's
	// payload.
	ErrSOLNotOwner = errors.New("SOL payload owned by another session")
)

// SOLConfig holds the SOL configuration parameters of spec v2.0 Table 26-5.
// Only Set In Progress (#0) and the volatile bit rate (#6) are volatile; the
// volatile bit rate is reloaded from the non-volatile one on every payload
// activation (§15.8). Defaults are manufacturer choices (the spec leaves them
// open): SOL enabled, ADMINISTRATOR privilege, 115.2 kbps.
type SOLConfig struct {
	mu sync.Mutex

	setInProgress uint8 // #0: 0 = set complete, 1 = set in progress (rollback not implemented)

	enabled             bool           // #1
	forceEncryption     bool           // #2 [7]
	forceAuthentication bool           // #2 [6]
	privilegeLevel      PrivilegeLevel // #2 [3:0]

	accumulateInterval5ms uint8 // #3 byte 1, 5 ms units
	sendThreshold         uint8 // #3 byte 2

	retryCount        uint8 // #4 byte 1 [2:0]
	retryInterval10ms uint8 // #4 byte 2

	nvBitRate uint8 // #5
	vBitRate  uint8 // #6 (volatile copy)

	payloadChannel uint8 // #7: channel the activation arrived on (1 until first activation)

	// PayloadPort is the RMCP port carrying the SOL payload (Table 26-5 #8).
	// Server-internal, written once at construction and never settable via
	// IPMI (#7/#8 are read-only per SetParam), so a plain field suffices.
	PayloadPort uint16
}

// NewSOLConfig returns a SOLConfig with manufacturer defaults.
func NewSOLConfig() *SOLConfig {
	return &SOLConfig{
		enabled:               true,
		privilegeLevel:        PrivilegeLevelAdministrator,
		accumulateInterval5ms: 10, // 50 ms
		sendThreshold:         SOLMaxPayloadChars,
		retryCount:            3,
		retryInterval10ms:     5, // 50 ms
		nvBitRate:             0x0a,
		vBitRate:              0x0a,
		payloadChannel:        1,
		PayloadPort:           SOLPayloadUDPPortDefault,
	}
}

// solBitRateSupported reports whether rate is a settable bit rate value
// (Table 26-5 #5: 6h-Ah; 0h selects the serial-channel setting and is
// reserved here because there is no IPMI-over-serial channel).
func solBitRateSupported(rate uint8) bool {
	return rate >= 0x06 && rate <= 0x0a
}

// GetParam returns the parameter data bytes for selector (Table 26-5),
// or false when the selector is not a supported parameter.
func (c *SOLConfig) GetParam(selector uint8) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch selector {
	case 0:
		return []byte{c.setInProgress}, true
	case 1:
		var b uint8
		if c.enabled {
			b = 0x01
		}
		return []byte{b}, true
	case 2:
		var b uint8
		if c.forceEncryption {
			b |= 0x80
		}
		if c.forceAuthentication {
			b |= 0x40
		}
		return []byte{b | uint8(c.privilegeLevel)}, true
	case 3:
		return []byte{c.accumulateInterval5ms, c.sendThreshold}, true
	case 4:
		return []byte{c.retryCount & 0x07, c.retryInterval10ms}, true
	case 5:
		return []byte{c.nvBitRate}, true
	case 6:
		return []byte{c.vBitRate}, true
	case 7:
		return []byte{c.payloadChannel}, true
	case 8:
		return []byte{uint8(c.PayloadPort), uint8(c.PayloadPort >> 8)}, true
	default:
		return nil, false
	}
}

// SetParam validates and applies one parameter write (Table 26-3/26-5),
// returning the command-specific completion code on failure.
func (c *SOLConfig) SetParam(selector uint8, data []byte) types.CompletionCode {
	c.mu.Lock()
	defer c.mu.Unlock()

	if selector == 7 || selector == 8 {
		// Read-only here: the payload channel is derived from the activation,
		// and the port is fixed by the transport the server listens on.
		return types.CodeParamConfigSetReadOnly
	}
	if selector > 8 {
		return types.CodeParameterNotSupported // 80h
	}
	if len(data) == 0 {
		return types.CodeRequestDataLengthInvalid
	}

	switch selector {
	case 0:
		// Commit write (10b) requires rollback support, which is not
		// implemented; 11b is reserved (Table 26-5 #0).
		if data[0] > 1 {
			return types.CodeRequestDataFieldInvalid
		}
		if data[0] == 1 && c.setInProgress != 0 {
			// Table 26-3: 81h recognizes that another party has already
			// 'claimed' the parameters (set in progress is a notification
			// flag, not a lock — later writers see this code).
			return types.CodeParamConfigSetInProgressConflict
		}
		c.setInProgress = data[0]
	case 1:
		c.enabled = data[0]&0x01 != 0
	case 2:
		priv := PrivilegeLevel(data[0] & 0x0f)
		switch priv {
		case PrivilegeLevelUser, PrivilegeLevelOperator, PrivilegeLevelAdministrator, PrivilegeLevelOEM:
		default:
			return types.CodeRequestDataFieldInvalid
		}
		c.forceEncryption = data[0]&0x80 != 0
		c.forceAuthentication = data[0]&0x40 != 0
		c.privilegeLevel = priv
	case 3:
		if len(data) < 2 {
			return types.CodeRequestDataLengthInvalid
		}
		if data[0] == 0 {
			return types.CodeRequestDataFieldInvalid // accumulate interval is 1-based (00h reserved)
		}
		if data[1] > SOLMaxPayloadChars {
			return types.CodeParameterOutOfRange
		}
		c.accumulateInterval5ms = data[0]
		c.sendThreshold = data[1]
	case 4:
		if len(data) < 2 {
			return types.CodeRequestDataLengthInvalid
		}
		c.retryCount = data[0] & 0x07
		c.retryInterval10ms = data[1]
	case 5, 6:
		if !solBitRateSupported(data[0]) {
			return types.CodeParameterOutOfRange // C9h, per Table 26-5 #5 note
		}
		if selector == 5 {
			c.nvBitRate = data[0]
		} else {
			c.vBitRate = data[0]
		}
	}
	return types.CodeOK
}

// snapshot returns the activation-relevant settings under one lock.
func (c *SOLConfig) snapshot() (enabled, forceEnc, forceAuth bool, priv PrivilegeLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enabled, c.forceEncryption, c.forceAuthentication, c.privilegeLevel
}

// timing returns the data-plane tuning of Table 26-5 #3/#4 as durations.
// The SOL payload format has no room for a zero retry interval: 00h means
// back-to-back retries, clamped here to a 10 ms floor.
func (c *SOLConfig) timing() (accumulate time.Duration, threshold int, retryCount int, retryInterval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	accumulate = time.Duration(c.accumulateInterval5ms) * 5 * time.Millisecond
	retryInterval = max(time.Duration(c.retryInterval10ms)*10*time.Millisecond, 10*time.Millisecond)
	return accumulate, int(c.sendThreshold), int(c.retryCount), retryInterval
}

// markActivated implements §15.8: on payload activation the volatile bit
// rate is reloaded from the non-volatile one, and Table 26-5 #7 records the
// channel the activation arrived on.
func (c *SOLConfig) markActivated(ch uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vBitRate = c.nvBitRate
	c.payloadChannel = ch
}

// ResetVolatile clears the volatile SOL configuration — only #0 set in
// progress and #6 volatile bit rate are volatile (Table 26-5) — as after a
// BMC cold reset / power cycle, which aborts any parameter set in progress
// (Table 26-3 "set complete" rule).
func (c *SOLConfig) ResetVolatile() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setInProgress = 0
	c.vBitRate = c.nvBitRate
}

// consoleReconnectOpenTimeout bounds each reconnect dial attempt. The pump
// is the only driver of recovery, so a hung dial must not stall console
// output indefinitely — deactivation waits on this tick via stopCh/stopped.
const consoleReconnectOpenTimeout = 5 * time.Second

// ReconnectPolicy controls when the pump retries attaching to a failed
// console, after reconnection has been enabled via [SOLStore.SetReconnectPolicy].
// By default reconnection is disabled: the payload stays active, reports
// status bit [5] (Table 15-2), and recovers only via deactivate/reactivate —
// the spec's own behavior.
//
// Field semantics follow k8s.io/apimachinery/pkg/util/wait.Backoff. The
// delay for attempt n (n ≥ 1, counting from the first failure) is
//
//	Initial * Factor^(n-1), capped at Cap, plus a jitter of up to
//	Jitter × the capped value.
//
// Console failures and reconnection are invisible to the remote console:
// ipmitool keeps its SOL session alive with a Get Device ID keepalive every
// 15s (ipmi_sol.c, SOL_KEEPALIVE_TIMEOUT) and only exits after 6 missed
// keepalives (~90s), so the server may back off as long as its RMCP+ path
// stays up. Keystrokes typed during the outage are lost either way:
// ipmitool treats a NACK as delivered and never retransmits.
type ReconnectPolicy struct {
	// Initial is the delay after the first failure; 0 retries immediately.
	Initial time.Duration
	// Factor multiplies the delay on each failure (>= 1).
	Factor float64
	// Jitter adds a random delay of up to Jitter × the computed delay
	// (0..1) to desynchronize concurrent reconnection attempts; 0 = none.
	Jitter float64
	// Cap bounds the delay; <= 0 means unbounded.
	Cap time.Duration
	// Steps bounds the number of reconnect attempts (1-based); once
	// exhausted the instance gives up and behaves as if reconnection were
	// disabled. 0 means retry forever.
	Steps int
}

// DefaultReconnectPolicy is the built-in strategy: retry after 1s, then
// double to a 30s cap, forever. VM-migration-scale outages (minutes) are
// covered by the cap holding: attempts keep coming, just no more often
// than every 30s.
var DefaultReconnectPolicy = ReconnectPolicy{Initial: time.Second, Factor: 2, Cap: 30 * time.Second}

// Delay returns the wait before attempt failures+1 after failures
// consecutive failed attempts (failures ≥ 1), and whether reconnection
// should be given up (Steps exhausted).
func (p *ReconnectPolicy) Delay(failures int) (wait time.Duration, giveUp bool) {
	if p == nil {
		return 0, true
	}
	// failures counts the console failure that started the cycle plus every
	// failed dial (reconnectLocked seeds it at 1 on first entry), so a dial
	// is allowed while failures <= Steps — Steps bounds the number of dials,
	// not the seed.
	if p.Steps > 0 && failures > p.Steps {
		return 0, true
	}
	f := 1.0
	if p.Factor > 1 {
		f = math.Pow(p.Factor, float64(failures-1))
	}
	// Compute in float seconds and clamp to Cap before converting to
	// Duration: the raw Initial * Factor^(n-1) multiply overflows int64
	// nanoseconds once the factor passes ~2^34, and a wrapped (negative)
	// wait is already in the past — the Cap clamp never applies, turning
	// the backoff into a busy retry loop. Float math also avoids the
	// integer truncation that freezes non-integer Factor values (e.g. 1.5)
	// on the first doubling steps.
	secs := p.Initial.Seconds() * f
	if p.Cap > 0 && secs > p.Cap.Seconds() {
		secs = p.Cap.Seconds()
	}
	wait = time.Duration(secs * float64(time.Second))
	if p.Jitter > 0 {
		wait += time.Duration(p.Jitter * float64(wait) * rand.Float64())
	}
	return wait, false
}

// SOLInstance is one activated SOL payload: the binding between an RMCP+
// session and the system console, plus the payload-level sequence state of
// Table 15-2. All methods are safe for concurrent use.
type SOLInstance struct {
	SessionID uint32 // BMC session ID owning the activation

	// encryptOutbound records the outbound protection negotiated at activation
	// (Table 24-2 auxiliary data + Table 26-5 #2 forcing bits) — the spec
	// constrains BMC→console data only, so outbound SOL packets are protected
	// per this setting, never per inbound packet flags (Table 24-4) — unless
	// the console toggles BMC→console encryption at run time via
	// Suspend/Resume Payload Encryption (Table 24-5). The send path reads it
	// from the pump goroutine without inst.mu, hence the atomic.
	encryptOutbound atomic.Bool

	conn hal.ConsoleConn
	// open re-attaches to the console after a failure; the HAL Open call that
	// created the initial conn. The reconnect state machine is the only user.
	open func(context.Context) (hal.ConsoleConn, error)
	send SOLSendFunc // nil when the server injected no sender (unit tests)

	// tracef is the optional diagnostic sink for console lifecycle events
	// (reconnect); nil (the default) disables it. Written from the pump only.
	tracef func(format string, args ...any)

	clock   clock.Clock
	config  *SOLConfig
	stopCh  chan struct{}
	stopped chan struct{}

	mu sync.Mutex

	inSeq        uint8     // last accepted console→BMC packet sequence number (0 = none yet)
	lastAccepted uint8     // accepted count reported for that packet (replayed on console retries)
	lastNACK     bool      // NACK state reported for that packet (replayed with lastAccepted)
	outSeq       uint8     // last assigned BMC→console packet sequence number (advances 1..0x0f)
	pending      []byte    // outbound character data awaiting acknowledge (§15.11)
	pendingSince time.Time // when pending was last (re)sent, for retry timing
	retries      int       // resend count for the pending packet
	suspended    bool      // Suspend NACK received; pending is withheld until resumed (Table 15-3)
	overrun      bool      // characters were dropped since the last outbound packet (status [3])
	rx           []byte    // drained console output awaiting transmission
	rxSince      time.Time // when the oldest rx byte arrived (accumulate interval)
	broken       bool      // console conn failed; reported via status bit [5]

	// Reconnect state, driven by the pump (the only goroutine that touches
	// conn while broken). policy is the activation-time snapshot of the
	// store's ReconnectPolicy (nil disables reconnection); failures counts
	// consecutive failed attempts; reconnectAt is when the next attempt may
	// run, zero until the first failure starts a cycle; giveUp latches once
	// the policy's Steps are exhausted.
	policy      *ReconnectPolicy
	failures    int
	reconnectAt time.Time
	giveUp      bool
}

// nextSOLSeq advances a payload sequence number in 1..0x0f; 0h is reserved
// for ACK-only packets and never assigned to data packets (Table 15-2).
func nextSOLSeq(seq uint8) uint8 {
	return (seq % 0x0f) + 1
}

// SOLSendFunc transmits one BMC→console SOL payload packet. The server
// supplies it; it owns session-level encryption, sequencing, and transport.
type SOLSendFunc func(pkt *types.SOLPayloadPacket) error

// SOLSenderFactory builds the send function for an activation. sess.Addr
// must already hold the console's transport address.
type SOLSenderFactory func(sess *Session, inst *SOLInstance) SOLSendFunc

// SOLStore owns the SOL configuration and active instances. Packet
// sequencing lives in [SOLInstance]; asynchronous output is pushed by the
// instance pump (spec v2.0 §15.3: the BMC sends SOL data unrequested).
type SOLStore struct {
	console hal.ConsoleHAL // nil when the target has no redirectable console
	config  *SOLConfig
	clock   clock.Clock

	senderFactory SOLSenderFactory // set by the server at construction

	// policy decides console reconnect timing; nil (the default) disables
	// reconnection. Snapshot per activation (see Activate), so the store's
	// value may be changed between activations without locking the data
	// plane.
	policy *ReconnectPolicy

	mu   sync.Mutex
	inst *SOLInstance // nil when deactivated (SOLMaxInstances == 1)
}

// NewSOLStore creates a SOLStore. h may be nil or return a nil ConsoleHAL,
// in which case SOL stays inactive and unadvertised. Console reconnection
// is disabled by default; enable it with [SOLStore.SetReconnectPolicy].
func NewSOLStore(h hal.HAL, clk clock.Clock) *SOLStore {
	s := &SOLStore{config: NewSOLConfig(), clock: clk}
	if h != nil {
		s.console = h.Console()
	}
	return s
}

// SetReconnectPolicy enables and configures console reconnection: after a
// console failure the pump retries the HAL Open on the policy's schedule
// (see [ReconnectPolicy]). A nil policy (the default) disables reconnection:
// the payload reports status bit [5] and recovers only via
// deactivate/reactivate. Takes effect on the next activation.
func (s *SOLStore) SetReconnectPolicy(p *ReconnectPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = p
}

// SetSenderFactory installs the factory used to build per-activation senders.
// Called by the server once the transport exists.
func (s *SOLStore) SetSenderFactory(f SOLSenderFactory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.senderFactory = f
}

// Config returns the SOL configuration parameter store.
func (s *SOLStore) Config() *SOLConfig { return s.config }

// Supported reports whether the SOL payload type can be activated at all,
// i.e. a console exists and the type is enabled (Table 26-5 #1).
func (s *SOLStore) Supported() bool {
	if s.console == nil {
		return false
	}
	enabled, _, _, _ := s.config.snapshot()
	return enabled
}

// Activate attaches the system console to sess (spec v2.0 §24.1). wantEnc /
// wantAuth are the remote console's Encryption/Authentication Activation
// bits from the Activate Payload auxiliary request data. The returned error
// maps to a Table 24-2 completion code in the handler.
func (s *SOLStore) Activate(ctx context.Context, sess *Session, wantEnc, wantAuth bool) (*SOLInstance, error) {
	if s.console == nil {
		return nil, ErrSOLDisabled
	}
	enabled, forceEnc, forceAuth, solPriv := s.config.snapshot()
	if !enabled {
		return nil, ErrSOLDisabled
	}

	s.mu.Lock()
	if s.inst != nil {
		s.mu.Unlock()
		return nil, ErrSOLAlreadyActive
	}
	s.mu.Unlock()

	if sess.PrivilegeLevel < solPriv {
		return nil, ErrSOLPrivilege
	}
	if sess.User != nil && !sess.User.PayloadAccessFor(sess.Channel).SOLEnabled() {
		return nil, ErrSOLPrivilege
	}

	// Encryption requires authentication: the spec algorithms have no
	// encrypt-only mode (Table 24-2 aux data note).
	if wantEnc && !wantAuth {
		return nil, ErrSOLEncryptionUnavailable
	}
	sessionCanCrypt := sess.CryptAlg != types.CryptAlg_None && len(sess.K2) >= 16
	sessionCanAuth := sess.IntegrityAlg != types.IntegrityAlg_None
	switch {
	case forceEnc && !wantEnc:
		// The console declined encryption while policy mandates it
		// (Table 24-2: 84h "BMC requires encryption for all payloads").
		return nil, ErrSOLEncryptionRequired
	case wantEnc && !sessionCanCrypt:
		// The console asked for encryption the session never negotiated.
		return nil, ErrSOLEncryptionUnavailable
	case (wantAuth || forceAuth) && !sessionCanAuth:
		return nil, ErrSOLAuthenticationUnavailable
	}

	conn, err := s.console.Open(ctx)
	if err != nil {
		return nil, err
	}

	// Outbound SOL packets are protected per the activation-negotiated
	// setting; inbound packets keep their own packet flags (Table 24-4).
	encrypted := sessionCanCrypt && (wantEnc || forceEnc)
	inst := &SOLInstance{
		SessionID: sess.BMCID,
		conn:      conn,
		open:      s.console.Open,
		clock:     s.clock,
		config:    s.config,
		stopCh:    make(chan struct{}),
		stopped:   make(chan struct{}),
	}
	inst.encryptOutbound.Store(encrypted)
	if s.senderFactory != nil && sess.GetAddr() != nil {
		inst.send = s.senderFactory(sess, inst)
	}

	s.mu.Lock()
	if s.inst != nil {
		// Lost the race against a concurrent activation on another session.
		s.mu.Unlock()
		_ = conn.Close()
		return nil, ErrSOLAlreadyActive
	}
	// Snapshot the reconnect policy under the store lock; the pump reads it
	// on this instance only, so later SetReconnectPolicy calls do not need
	// to synchronize with the data plane.
	inst.policy = s.policy
	s.inst = inst
	s.mu.Unlock()

	s.config.markActivated(sess.Channel)
	if inst.send != nil {
		go inst.pump()
	} else {
		close(inst.stopped)
	}
	return inst, nil
}

// Deactivate detaches the console (spec v2.0 §24.2). The owning session may
// always deactivate; another session may force-deactivate when its privilege
// meets the configured SOL level (Table 26-5 #2) — this is the recovery path
// for payloads orphaned by a crashed console.
func (s *SOLStore) Deactivate(sess *Session) error {
	_, _, _, solPriv := s.config.snapshot()

	s.mu.Lock()
	inst := s.inst
	switch {
	case inst == nil:
		s.mu.Unlock()
		return ErrSOLNotActive
	case inst.SessionID != sess.BMCID && sess.PrivilegeLevel < solPriv:
		s.mu.Unlock()
		return ErrSOLNotOwner
	}
	s.inst = nil
	s.mu.Unlock()

	return s.teardownInst(inst)
}

// DeactivateBySession drops the instance owned by bmcID, if any. Wired to
// SessionStore removals: session termination automatically deactivates its
// payloads (spec v2.0 §24.2 note).
func (s *SOLStore) DeactivateBySession(bmcID uint32) {
	s.mu.Lock()
	inst := s.inst
	if inst == nil || inst.SessionID != bmcID {
		s.mu.Unlock()
		return
	}
	s.inst = nil
	s.mu.Unlock()

	_ = s.teardownInst(inst)
}

// CloseAll deactivates every instance; called when the server shuts down.
func (s *SOLStore) CloseAll() {
	s.mu.Lock()
	inst := s.inst
	s.inst = nil
	s.mu.Unlock()
	if inst != nil {
		_ = s.teardownInst(inst)
	}
}

// teardownInst notifies the remote console and releases an instance that has
// already been unregistered from the store. The notify packet must go out
// before close, which stops the pump.
func (s *SOLStore) teardownInst(inst *SOLInstance) error {
	inst.sendDeactivating()
	return inst.close()
}

// ActivationStatus reports the Table 24-6 instance bitmask: bit (n-1) set
// when instance n is active.
func (s *SOLStore) ActivationStatus() (capacity uint8, active1to8, active9to16 uint8) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inst != nil {
		active1to8 = 0x01
	}
	return SOLMaxInstances, active1to8, 0
}

// ActiveSessionID returns the owning session ID for instance (1-based),
// or 0 when not activated (Table 24-7).
func (s *SOLStore) ActiveSessionID(instance uint8) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inst == nil || instance != 1 {
		return 0
	}
	return s.inst.SessionID
}

// InstanceBySession returns the instance owned by bmcID, or nil. The server
// reads the negotiated protection flags from it before processing packets.
func (s *SOLStore) InstanceBySession(bmcID uint32) *SOLInstance {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inst == nil || s.inst.SessionID != bmcID {
		return nil
	}
	return s.inst
}

// OutboundEncrypted reports whether BMC→console SOL data is currently
// encrypted: the activation-negotiated setting unless the console toggled it
// via Suspend/Resume Payload Encryption (Table 24-5). The command only
// governs data from the BMC ("encryption on all transfers of specified
// payload data from the BMC"); inbound packets keep their own packet flags.
func (inst *SOLInstance) OutboundEncrypted() bool {
	return inst.encryptOutbound.Load()
}

// Table 24-5 operations for Suspend/Resume Payload Encryption.
const (
	SOLEncryptionOpSuspend = 0
	SOLEncryptionOpResume  = 1
	SOLEncryptionOpRegenIV = 2
)

// SuspendResumeEncryption applies a Table 24-5 operation to the instance
// owned by sess. Only BMC→console encryption is affected; authentication is
// untouched and inbound packets keep their activation-negotiated protection.
func (s *SOLStore) SuspendResumeEncryption(sess *Session, op uint8) error {
	s.mu.Lock()
	inst := s.inst
	s.mu.Unlock()
	if inst == nil || inst.SessionID != sess.BMCID {
		return ErrSOLInstanceNotActive
	}
	switch op {
	case SOLEncryptionOpSuspend:
		if _, forceEnc, _, _ := s.config.snapshot(); forceEnc {
			return ErrSOLEncryptionForced
		}
		inst.encryptOutbound.Store(false)
	case SOLEncryptionOpResume:
		// "Resume/Start encryption": valid even when the payload was
		// activated without encryption, provided the session can encrypt.
		if sess.CryptAlg == types.CryptAlg_None || len(sess.K2) < 16 {
			return ErrSOLEncryptionUnavailableForSession
		}
		inst.encryptOutbound.Store(true)
	default:
		return ErrSOLOperationUnsupported
	}
	return nil
}

// sendDeactivating issues the one-time packet with status bit [4] (SOL
// deactivated/deactivating, Table 15-2 footnote 2) just before the payload
// goes down, so the remote console can tell an administrative deactivation
// (Deactivate Payload / Close Session, possibly from another session) apart
// from a communication failure.
func (inst *SOLInstance) sendDeactivating() {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.send == nil {
		return
	}
	_ = inst.send(&types.SOLPayloadPacket{
		AckedSequenceNumber:    inst.inSeq,
		AcceptedCharacterCount: inst.lastAccepted,
		ControlByte:            0x10,
	})
}

func (inst *SOLInstance) close() error {
	if inst.stopCh != nil {
		close(inst.stopCh)
		<-inst.stopped // pump outlives conn use; wait before closing
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.conn != nil {
		err := inst.conn.Close()
		inst.conn = nil
		return err
	}
	return nil
}

// pump asynchronously pushes console output to the remote console and drives
// the §15.11 retry engine. Consoles such as ipmitool never poll with SOL
// packets (their keepalive is Get Device ID), so response-piggybacking alone
// would starve the BMC→console direction.
//
// Tick granularity is fixed at 10 ms; configured intervals (5 ms / 10 ms
// units) are compared against timestamps, so the tick only bounds precision.
func (inst *SOLInstance) pump() {
	defer close(inst.stopped)
	ticker := inst.clock.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-inst.stopCh:
			return
		case <-ticker.C():
			inst.pumpTick()
		}
	}
}

func (inst *SOLInstance) pumpTick() {
	accumulate, threshold, retryCount, retryInterval := inst.config.timing()

	inst.mu.Lock()
	now := inst.clock.Now()

	// A failed console conn enters the reconnect state machine before any
	// other pump work: recovery is the pump's job alone (ProcessPacket never
	// touches conn while broken), so the data plane must not skip it behind
	// a nil-conn check. reconnectLocked temporarily drops the lock while the
	// HAL dials (a dial may block), then re-acquires it to install state.
	if inst.broken {
		inst.reconnectLocked(now)
		inst.mu.Unlock()
		return
	}
	defer inst.mu.Unlock()
	if inst.conn == nil || inst.send == nil {
		return
	}

	if inst.pending != nil {
		if inst.suspended {
			// Suspend NACK (Table 15-3): stop sending/retrying the pending
			// packet until a Partial/Completion ACK, Resume ACK, or Flush
			// Outbound arrives. Console output keeps buffering meanwhile.
			inst.drainConsoleLocked()
			return
		}
		// Unacknowledged packet: resend on the retry interval, drop when
		// retries are exhausted (§15.11: "the data will be lost") and flag
		// the loss on the next outbound packet (Table 15-2 status [3]).
		if now.Sub(inst.pendingSince) >= retryInterval {
			if inst.retries >= retryCount {
				inst.pending = nil
				inst.retries = 0
				inst.overrun = true
			} else {
				inst.sendLocked()
				inst.retries++
				inst.pendingSince = now
			}
		}
		return
	}

	inst.drainConsoleLocked()
	full := threshold > 0 && len(inst.rx) >= threshold
	aged := len(inst.rx) > 0 && now.Sub(inst.rxSince) >= accumulate
	if full || aged {
		inst.makeOutboundLocked()
		inst.sendLocked()
	}
}

// reconnectLocked drives the console recovery state machine. Called by the
// pump on every tick while broken, entered with inst.mu held:
//
//   - first entry: release the dead conn. ProcessPacket never touches conn
//     while broken (it answers with NACK + status bit [5] instead), so
//     replacing the conn here cannot race a keystroke write.
//   - on each policy-determined expiry: ask the console HAL for a fresh
//     conn. A successful attach clears broken and drops the pre-failure RX
//     residue — bytes read before the failure belong to the old console
//     session (e.g. a rebooted system) and must not reach the remote
//     console. Pending outbound data is kept: the remote console still owns
//     its ACK (Table 15-3), and the retry engine already decides its fate.
//
// A nil policy (reconnection disabled) releases the dead conn once and then
// does nothing: the payload stays active with status bit [5] until the
// remote console deactivates it.
//
// The HAL dial happens with the lock dropped: it may block for up to
// consoleReconnectOpenTimeout, and holding inst.mu across it would stall
// keystroke processing and status responses. State is only ever mutated
// under the lock, so the unlock window is safe — broken stays true (only
// this goroutine clears it) and ProcessPacket only inspects it.
func (inst *SOLInstance) reconnectLocked(now time.Time) {
	if inst.conn != nil {
		_ = inst.conn.Close()
		inst.conn = nil
	}
	if inst.policy == nil || inst.giveUp {
		return // reconnection disabled, or Steps exhausted
	}
	if inst.reconnectAt.IsZero() {
		// First failure: the policy decides the first retry delay.
		inst.failures = 1
		inst.reconnectAt = now.Add(inst.policyDelay())
	}
	if now.Before(inst.reconnectAt) {
		return // still backing off
	}
	inst.mu.Unlock() // dial without the lock; re-acquire to install state
	ctx, cancel := context.WithTimeout(context.Background(), consoleReconnectOpenTimeout)
	conn, err := inst.open(ctx)
	cancel()
	inst.mu.Lock()
	if err != nil {
		inst.failures++
		// Re-arm from when the attempt finished, not the tick's pre-dial
		// time: a dial that burned most of the wait would otherwise leave
		// reconnectAt already in the past, collapsing the backoff into
		// back-to-back attempts no matter how large the delay has grown.
		inst.reconnectAt = inst.clock.Now().Add(inst.policyDelay())
		return
	}
	inst.conn = conn
	inst.broken = false
	if inst.tracef != nil {
		inst.tracef("sol! sess=%x console reconnected after %d failed attempt(s)\n", inst.SessionID, inst.failures)
	}
	inst.failures = 0
	inst.reconnectAt = time.Time{}
	inst.giveUp = false
	inst.rx = nil
	inst.rxSince = time.Time{}
}

// SetTracef installs the diagnostic sink for console lifecycle events. The
// server wires it to its solDebug-gated SOL trace; a nil sink (the default)
// keeps the library silent. Callable from any goroutine; events are emitted
// from the pump.
//
// The sink is printf-shaped, so printf-style loggers plug in directly —
// logrus.Printf and zap's SugaredLogger.Infof match this signature as-is.
func (inst *SOLInstance) SetTracef(f func(format string, args ...any)) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.tracef = f
}

// policyDelay asks the policy for the wait before the next attempt,
// latching giveUp once its Steps are exhausted. inst.mu must be held.
func (inst *SOLInstance) policyDelay() time.Duration {
	wait, giveUp := inst.policy.Delay(inst.failures)
	if giveUp {
		inst.giveUp = true
	}
	return wait
}

// sendLocked transmits the pending packet. inst.mu must be held.
func (inst *SOLInstance) sendLocked() {
	pkt := &types.SOLPayloadPacket{
		SequenceNumber:         inst.outSeq,
		AckedSequenceNumber:    inst.inSeq,
		AcceptedCharacterCount: inst.lastAccepted,
		CharacterData:          inst.pending,
	}
	inst.applyStatusLocked(pkt)
	_ = inst.send(pkt) // send failure keeps pending; the retry path owns it
}

// applyStatusLocked stamps the BMC→console status bits on out and consumes
// the overrun latch. inst.mu must be held.
func (inst *SOLInstance) applyStatusLocked(out *types.SOLPayloadPacket) {
	if inst.broken {
		out.ControlByte |= 0x20 // Table 15-2 status [5]: character transfer unavailable
	}
	if inst.overrun {
		out.ControlByte |= 0x08 // Table 15-2 status [3]: characters dropped since the previous packet
		inst.overrun = false
	}
}

// drainConsoleLocked moves immediately-available console output into rx.
// inst.mu must be held.
func (inst *SOLInstance) drainConsoleLocked() {
	if inst.broken || inst.conn == nil || len(inst.rx) >= SOLRXBufferCap {
		return
	}
	chunk := make([]byte, SOLRXBufferCap-len(inst.rx))
	n, err := inst.conn.ReadAvailable(chunk)
	if err != nil {
		inst.broken = true
		return
	}
	if n > 0 {
		if len(inst.rx) == 0 {
			inst.rxSince = inst.clock.Now()
		}
		inst.rx = append(inst.rx, chunk[:n]...)
	}
}

// makeOutboundLocked promotes buffered console output to the pending packet.
// One outstanding packet at a time (Table 15-2): a new sequence number is
// assigned only when nothing is unacknowledged. inst.mu must be held.
func (inst *SOLInstance) makeOutboundLocked() {
	if inst.pending != nil || len(inst.rx) == 0 {
		return
	}
	take := min(len(inst.rx), SOLMaxPayloadChars)
	inst.pending = append([]byte{}, inst.rx[:take]...)
	inst.rx = inst.rx[take:]
	inst.outSeq = nextSOLSeq(inst.outSeq)
	inst.retries = 0
	inst.pendingSince = inst.clock.Now()
}

// ProcessPacket handles one inbound SOL payload packet from the owning
// session and produces the response packet (spec v2.0 §15.9/§15.11,
// Table 15-3). It returns nil when the session owns no active instance.
func (s *SOLStore) ProcessPacket(ctx context.Context, sessionID uint32, in *types.SOLPayloadPacket) *types.SOLPayloadPacket {
	s.mu.Lock()
	inst := s.inst
	s.mu.Unlock()
	if inst == nil || inst.SessionID != sessionID {
		return nil
	}
	return inst.ProcessPacket(ctx, in)
}

// ProcessPacket handles one inbound SOL payload packet (spec v2.0
// §15.9/§15.11, Table 15-3) and produces the response packet.
func (inst *SOLInstance) ProcessPacket(ctx context.Context, in *types.SOLPayloadPacket) *types.SOLPayloadPacket {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.conn == nil && !inst.broken {
		return nil
	}

	out := &types.SOLPayloadPacket{}

	// The console is acknowledging our pending outbound packet (Table 15-3).
	if in.AckedSequenceNumber != 0 && inst.pending != nil && in.AckedSequenceNumber == inst.outSeq {
		switch {
		case in.NACK && in.AcceptedCharacterCount == 0:
			// Suspend NACK (Table 15-3): the console cannot accept SOL data
			// right now. Stop sending/retrying the pending packet until a
			// Partial/Completion ACK, Resume ACK, or Flush Outbound arrives.
			inst.suspended = true
		case in.AcceptedCharacterCount == 0:
			inst.suspended = false // Resume ACK: pending rides this response
		case int(in.AcceptedCharacterCount) >= len(inst.pending):
			inst.pending = nil // Completion ACK
			inst.suspended = false
		default:
			// Partial ACK — a NACK with a non-zero count reads the same
			// (rev 1.1: NACK means "some or all data could not be accepted",
			// count meaningful): retransmit the unaccepted remainder as a
			// fresh packet. A same-sequence resend would look like the
			// original packet to consoles deduping by sequence number, and
			// ipmitool's character-offset logic would see zero new bytes in
			// a shorter retransmission (§15.11), so the remainder advances
			// the sequence number.
			inst.pending = inst.pending[in.AcceptedCharacterCount:]
			inst.outSeq = nextSOLSeq(inst.outSeq)
			inst.retries = 0
			inst.suspended = false
		}
	}

	// Console→BMC operations (Table 15-2) execute before character data.
	// Flush Outbound doubles as the recovery path from a Suspend NACK
	// (Table 15-3), so it is honored on any packet, including ACK-only ones.
	if in.ControlByte&0x01 != 0 {
		inst.pending = nil
		inst.retries = 0
		inst.rx = nil
		inst.suspended = false
	}
	// Flush Inbound (bit [1]) needs no instance-side action: keystrokes are
	// written to the system console synchronously, and the server layer
	// drops the session's not-yet-processed queued packets on flush.

	// Inbound operations and character data. ACK-only packets (seq 0h) carry
	// no data and are never acknowledged (§15.11).
	var accepted uint8
	var nack bool
	if in.SequenceNumber != 0 {
		if in.SequenceNumber == inst.inSeq {
			// Console retry: content is unchanged (§15.9), so replay the
			// response the original packet got — reporting len(CharacterData)
			// here would ACK bytes a partial accept or NACK never wrote.
			accepted = inst.lastAccepted
			nack = inst.lastNACK
		} else {
			if inst.broken {
				nack = true
			} else {
				if in.ControlByte&0x10 != 0 { // bit [4]: generate BREAK
					_ = inst.conn.SendBreak(ctx) // best effort; BREAK is advisory
				}
				n, err := inst.conn.Write(in.CharacterData)
				accepted = uint8(n)
				if err != nil {
					inst.broken = true
					nack = true
				} else if n < len(in.CharacterData) {
					nack = true
				}
			}
			inst.inSeq = in.SequenceNumber
			inst.lastAccepted = accepted
			inst.lastNACK = nack
		}
	} else {
		// ACK-only packet (seq 0h, §15.11): the acked echo names the last
		// accepted data packet, so replay its accepted count — reporting 0
		// would read as "packet N accepted nothing" (Table 15-3) and make
		// a strict console resend packet N's already-delivered keystrokes.
		accepted = inst.lastAccepted
	}

	// Outbound character data rides the response: polling-style consoles
	// (which send empty SOL packets continuously) get output with the lowest
	// latency this way; the pump covers consoles that never poll. A suspended
	// packet (Suspend NACK, Table 15-3) is withheld from the response.
	if !inst.broken {
		inst.drainConsoleLocked()
	}
	if !inst.suspended {
		inst.makeOutboundLocked()
	}

	out.SequenceNumber = 0 // ACK-only packet unless data is pending (Table 15-2)
	if inst.pending != nil && !inst.suspended {
		out.SequenceNumber = inst.outSeq
		out.CharacterData = inst.pending
		inst.pendingSince = inst.clock.Now() // the response is a (re)send
	}
	out.AckedSequenceNumber = inst.inSeq
	out.AcceptedCharacterCount = accepted
	out.NACK = nack
	inst.applyStatusLocked(out)
	return out
}

// Package bmc holds the runtime state of a Baseboard Management Controller.
//
// Nothing in this package does I/O; it is pure in-memory state backed by the
// abstractions in pkg/hal.  The server layer (server.go) wires transport,
// clock, and HAL together with this state to produce a working BMC.
package bmc

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal"
)

// DeviceInfo contains the identification data returned by Get Device ID.
type DeviceInfo struct {
	DeviceID       uint8
	DeviceRevision uint8
	FirmwareMajor  uint8 // major revision (bits 6:0)
	FirmwareMinor  uint8 // minor revision, BCD
	IPMIVersion    uint8 // 0x20 for IPMI 2.0
	ManufacturerID uint32
	ProductID      uint16
	AuxFirmwareRev [4]byte
	// AdditionalDeviceSupport bitfield per Table 20-2.
	AdditionalDeviceSupport uint8
}

// BMC is the central state object for an IPMI server.
//
// Callers create a BMC via [New] and pass it to the server together with a
// transport and HAL.  The BMC does not own any goroutines; lifecycle management
// belongs to the server.
type BMC struct {
	Info DeviceInfo
	GUID [16]byte
	KG   []byte // BMC key (Kg); nil means "one-key" mode using Kuid only

	// CipherSuites is the set of RMCP+ cipher suites the server advertises and
	// accepts during the Open Session handshake. Defaults to
	// [DefaultCipherSuites] when nil. Each suite must be supported by the
	// reference server (see [SupportedCipherSuite]); this is validated in
	// [WithCipherSuites].
	CipherSuites []CipherSuiteID

	Users    *UserStore
	Channels *ChannelStore
	Sessions *SessionStore

	hal   hal.HAL
	clock clock.Clock
}

// Option configures a [BMC].
type Option func(*BMC)

// WithKG sets the BMC-level key (Kg) used for two-key RAKP authentication.
// Leave unset (or pass nil) to use one-key mode (Kuid only).
func WithKG(kg []byte) Option {
	return func(b *BMC) {
		if len(kg) > 0 {
			b.KG = kg
		}
	}
}

// WithClock injects a custom [clock.Clock].  Defaults to [clock.Real].
func WithClock(c clock.Clock) Option {
	return func(b *BMC) { b.clock = c }
}

// WithCipherSuites sets the RMCP+ cipher suites the server advertises and
// accepts. Each ID must be a suite the reference server implements
// ([SupportedCipherSuite]); otherwise an error is returned by New and the
// default suite list is kept. Pass nil/empty to restore [DefaultCipherSuites].
func WithCipherSuites(ids []CipherSuiteID) Option {
	return func(b *BMC) {
		b.setCipherSuites(ids)
	}
}

// ResolvedCipherSuites returns the cipher suite list to use for advertisement,
// falling back to [DefaultCipherSuites] when none was configured.
func (b *BMC) ResolvedCipherSuites() []CipherSuiteID {
	if len(b.CipherSuites) > 0 {
		return b.CipherSuites
	}
	return DefaultCipherSuites
}

// SetCipherSuites replaces the configured cipher suite list. Each ID must be
// supported by the reference server ([SupportedCipherSuite]); an unsupported
// ID panics, failing at configuration time rather than at handshake time.
func (b *BMC) SetCipherSuites(ids []CipherSuiteID) {
	b.setCipherSuites(ids)
}

func (b *BMC) setCipherSuites(ids []CipherSuiteID) {
	if len(ids) == 0 {
		b.CipherSuites = nil
		return
	}
	validateCipherSuites(ids)
	b.CipherSuites = append(b.CipherSuites[:0:0], ids...)
}

// New creates a BMC with sane defaults.
//
// h is required; it provides hardware access.  opts are applied in order.
func New(info DeviceInfo, guid [16]byte, h hal.HAL, opts ...Option) *BMC {
	b := &BMC{
		Info:  info,
		GUID:  guid,
		hal:   h,
		clock: clock.Real,

		Users:    NewUserStore(),
		Channels: NewChannelStore(),
		Sessions: NewSessionStore(clock.Real),
	}
	for _, o := range opts {
		o(b)
	}
	// Re-apply clock to SessionStore after options in case WithClock was used.
	b.Sessions.clock = b.clock
	return b
}

// validateCipherSuites panics if any configured cipher suite is not implemented
// by the reference server. Failing at construction avoids runtime handshake
// failures from advertising suites we cannot negotiate.
func validateCipherSuites(ids []CipherSuiteID) {
	for _, id := range ids {
		if !SupportedCipherSuite(id) {
			panic(fmt.Sprintf("bmc: cipher suite %d is not implemented by the reference server", id))
		}
	}
}

// HAL returns the underlying hardware abstraction.
func (b *BMC) HAL() hal.HAL { return b.hal }

// Clock returns the time source used by this BMC.
func (b *BMC) Clock() clock.Clock { return b.clock }

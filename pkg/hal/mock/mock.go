// Package mock provides in-memory [hal.HAL] implementations for use in tests
// and simulation environments.
//
// All fields are exported so tests can set initial state and inspect results
// directly.  Methods are safe for concurrent use only if the test sets fields
// before calling [NewHAL] and does not mutate them afterwards; for race-free
// concurrent tests, protect shared state with a mutex in a custom sub-struct.
package mock

import (
	"context"
	"sync"

	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// HAL is a fully in-memory [hal.HAL].
type HAL struct {
	chassis *Chassis
	sensors *Sensors
	storage *Storage
	network *Network
	gpio    *GPIO
	i2c     *I2C
	console hal.ConsoleHAL
}

// New returns a [HAL] with all sub-interfaces initialised.
// Each sub-type is also exported individually for targeted embedding.
//
// The console sub-interface is left unset: most simulated targets have no
// redirectable serial port, so [HAL.Console] returns nil until [HAL.SetConsole]
// installs one.
func New() *HAL {
	return &HAL{
		chassis: &Chassis{},
		sensors: &Sensors{},
		storage: &Storage{},
		network: &Network{},
		gpio:    &GPIO{levels: map[string]bool{}, watchers: map[string][]func(bool){}},
		i2c:     &I2C{},
	}
}

func (h *HAL) Chassis() hal.ChassisHAL { return h.chassis }
func (h *HAL) Sensors() hal.SensorHAL  { return h.sensors }
func (h *HAL) Storage() hal.StorageHAL { return h.storage }
func (h *HAL) Network() hal.NetworkHAL { return h.network }
func (h *HAL) GPIO() hal.GPIOHAL       { return h.gpio }
func (h *HAL) I2C() hal.I2CHAL         { return h.i2c }

// Console returns the console installed by [HAL.SetConsole], or nil when the
// simulated target has no redirectable serial port.
func (h *HAL) Console() hal.ConsoleHAL { return h.console }

// SetConsole installs the console sub-interface returned by [HAL.Console].
// Tests typically pass [*Console]; production wiring may pass any
// [hal.ConsoleHAL] implementation.
func (h *HAL) SetConsole(c hal.ConsoleHAL) { h.console = c }

func (h *HAL) Close() error { return nil }

// --- Chassis ---

// Chassis is the mock [hal.ChassisHAL].
type Chassis struct {
	mu              sync.Mutex
	On              bool
	Intruded        bool
	ColdResets      int
	WarmResets      int
	PowerCycles     int
	LastIdentifySec uint8
	BootFlags       *types.BootOptionParam_BootFlags
	BootInfoAck     *types.BootOptionParam_BootInfoAcknowledge

	// Hook allows tests to inject custom behaviour.
	SetPowerHook func(on bool) error
}

func (c *Chassis) PowerState(_ context.Context) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.On, nil
}

func (c *Chassis) SetPower(_ context.Context, on bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.SetPowerHook != nil {
		return c.SetPowerHook(on)
	}
	c.On = on
	return nil
}

func (c *Chassis) ColdReset(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ColdResets++
	return nil
}

// PowerCycle performs a power cycle. The mock counts calls independently so
// tests can assert that Chassis Control action 0x02 dispatched here rather than
// to ColdReset. Real noop HALs that lack a distinct power-cycle operation may
// delegate to ColdReset; the mock keeps the counters separate for clarity.
func (c *Chassis) PowerCycle(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PowerCycles++
	return nil
}

func (c *Chassis) WarmReset(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.WarmResets++
	return nil
}

func (c *Chassis) Identify(_ context.Context, seconds uint8) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastIdentifySec = seconds
	return nil
}

func (c *Chassis) IntrusionState(_ context.Context) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Intruded, nil
}

// SetBootFlags stores the full boot flags structure so tests and the
// reference server can round-trip Set/Get System Boot Options.
func (c *Chassis) SetBootFlags(_ context.Context, flags *types.BootOptionParam_BootFlags) error {
	if flags == nil {
		return hal.ErrNotSupported
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := *flags
	c.BootFlags = &cp
	return nil
}

// GetBootFlags returns the last stored boot flags, or [hal.ErrNotSupported]
// when none have been set.
func (c *Chassis) GetBootFlags(_ context.Context) (*types.BootOptionParam_BootFlags, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.BootFlags == nil {
		return nil, hal.ErrNotSupported
	}
	cp := *c.BootFlags
	return &cp, nil
}

// SetBootInfoAcknowledge stores the boot info acknowledge data.
func (c *Chassis) SetBootInfoAcknowledge(_ context.Context, ack *types.BootOptionParam_BootInfoAcknowledge) error {
	if ack == nil {
		return hal.ErrNotSupported
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := *ack
	c.BootInfoAck = &cp
	return nil
}

// GetBootInfoAcknowledge returns the last stored acknowledge data,
// or a default value when none have been stored.
func (c *Chassis) GetBootInfoAcknowledge(_ context.Context) (*types.BootOptionParam_BootInfoAcknowledge, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.BootInfoAck == nil {
		return &types.BootOptionParam_BootInfoAcknowledge{}, nil
	}
	cp := *c.BootInfoAck
	return &cp, nil
}

// --- Sensors ---

// Sensors is the mock [hal.SensorHAL].
type Sensors struct {
	mu     sync.Mutex
	Values map[uint8]uint8
	Descs  []hal.SensorDescriptor
}

func (s *Sensors) ReadRaw(_ context.Context, id uint8) (uint8, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Values == nil {
		return 0, hal.ErrNotSupported
	}
	v, ok := s.Values[id]
	if !ok {
		return 0, hal.ErrNotSupported
	}
	return v, nil
}

func (s *Sensors) List(_ context.Context) ([]hal.SensorDescriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Descs, nil
}

// --- Network ---

// Network is the mock [hal.NetworkHAL].
type Network struct {
	mu  sync.Mutex
	Cfg hal.IPConfig
}

func (n *Network) GetConfig(_ context.Context) (*hal.IPConfig, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	cp := n.Cfg
	return &cp, nil
}

func (n *Network) SetConfig(_ context.Context, cfg *hal.IPConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Cfg = *cfg
	return nil
}

// --- GPIO ---

// GPIO is the mock [hal.GPIOHAL].
type GPIO struct {
	mu       sync.Mutex
	levels   map[string]bool
	watchers map[string][]func(bool)
}

func (g *GPIO) Set(_ context.Context, pin string, high bool) error {
	g.mu.Lock()
	watchers := append([]func(bool){}, g.watchers[pin]...)
	g.levels[pin] = high
	g.mu.Unlock()

	for _, fn := range watchers {
		fn(high)
	}
	return nil
}

func (g *GPIO) Get(_ context.Context, pin string) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.levels[pin], nil
}

func (g *GPIO) Watch(_ context.Context, pin string, cb func(bool)) (func(), error) {
	g.mu.Lock()
	g.watchers[pin] = append(g.watchers[pin], cb)
	idx := len(g.watchers[pin]) - 1
	g.mu.Unlock()

	cancel := func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		ws := g.watchers[pin]
		if idx < len(ws) {
			ws[idx] = ws[len(ws)-1]
			g.watchers[pin] = ws[:len(ws)-1]
		}
	}
	return cancel, nil
}

// --- Console ---

// Console is the mock [hal.ConsoleHAL].
type Console struct {
	mu sync.Mutex

	// Conn is returned by Open when OpenHook is nil. Nil means "no console
	// hardware": Open then returns hal.ErrNotSupported.
	Conn hal.ConsoleConn

	// OpenHook allows tests to inject custom behaviour (e.g. fail the open).
	OpenHook func(ctx context.Context) (hal.ConsoleConn, error)

	// OpenCount records how many activations attached to the console.
	OpenCount int
}

func (c *Console) Open(ctx context.Context) (hal.ConsoleConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OpenCount++
	if c.OpenHook != nil {
		return c.OpenHook(ctx)
	}
	if c.Conn == nil {
		return nil, hal.ErrNotSupported
	}
	return c.Conn, nil
}

// FakeConsoleConn is an in-memory [hal.ConsoleConn] for tests.
type FakeConsoleConn struct {
	mu sync.Mutex

	// RX holds console output waiting to be drained by ReadAvailable.
	RX []byte
	// TX records everything written toward the system console.
	TX []byte

	Closed   bool
	Breaks   int
	WriteErr error
	ReadErr  error

	// WriteLimit, when > 0, caps how many bytes a single Write accepts,
	// forcing the partial accepts (and NACKs) a flow-controlled serial
	// controller produces.
	WriteLimit int
}

// SetReadErr arms a read failure for the next ReadAvailable, safe for
// concurrent use with a running SOL session (unlike direct ReadErr writes).
func (f *FakeConsoleConn) SetReadErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ReadErr = err
}

// SetWriteErr arms a write failure for the next Write, safe for concurrent
// use with a running SOL session (unlike direct WriteErr writes).
func (f *FakeConsoleConn) SetWriteErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.WriteErr = err
}

func (f *FakeConsoleConn) ReadAvailable(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ReadErr != nil {
		return 0, f.ReadErr
	}
	n := copy(p, f.RX)
	f.RX = f.RX[n:]
	return n, nil
}

func (f *FakeConsoleConn) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.WriteErr != nil {
		return 0, f.WriteErr
	}
	if f.WriteLimit > 0 && len(p) > f.WriteLimit {
		p = p[:f.WriteLimit]
	}
	f.TX = append(f.TX, p...)
	return len(p), nil
}

func (f *FakeConsoleConn) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Closed = true
	return nil
}

func (f *FakeConsoleConn) SendBreak(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Breaks++
	return nil
}

// FeedRX appends bytes to the pending console output, as if the system had
// written them to its serial port. Safe for concurrent use with a running
// SOL session, unlike direct RX mutation.
func (f *FakeConsoleConn) FeedRX(p []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.RX = append(f.RX, p...)
}

// TXString returns everything written toward the system console so far.
func (f *FakeConsoleConn) TXString() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.TX)
}

// IsClosed reports whether Close was called. Safe for concurrent use with a
// running SOL session, unlike direct Closed reads.
func (f *FakeConsoleConn) IsClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Closed
}

// --- I2C ---

// I2C is the mock [hal.I2CHAL].
// Reads return data from ReadFunc if set, otherwise ErrNotSupported.
type I2C struct {
	mu       sync.Mutex
	ReadFunc func(bus int, addr, reg uint8, length int) ([]byte, error)
	Writes   []I2CWrite
}

// I2CWrite records a single I2C write for test assertions.
type I2CWrite struct {
	Bus  int
	Addr uint8
	Reg  uint8
	Data []byte
}

func (i *I2C) Read(_ context.Context, bus int, addr, reg uint8, length int) ([]byte, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.ReadFunc != nil {
		return i.ReadFunc(bus, addr, reg, length)
	}
	return nil, hal.ErrNotSupported
}

func (i *I2C) Write(_ context.Context, bus int, addr, reg uint8, data []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	i.Writes = append(i.Writes, I2CWrite{Bus: bus, Addr: addr, Reg: reg, Data: cp})
	return nil
}

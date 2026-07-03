package client

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
)

// TestServerSessionInactivityEviction verifies that idle v1.5 and RMCP+ sessions
// are removed after the spec-mandated 60s inactivity timeout. The test injects a
// manual clock so CI does not need to sleep a full minute.
func TestServerSessionInactivityEviction(t *testing.T) {
	const (
		username = "ADMIN"
		password = "ADMIN"
	)

	t.Run("lan v1.5", func(t *testing.T) {
		testSessionInactivityEviction(t, username, password, InterfaceLan)
	})
	t.Run("lanplus v2.0", func(t *testing.T) {
		testSessionInactivityEviction(t, username, password, InterfaceLanplus)
	})
}

func testSessionInactivityEviction(t *testing.T, username, password string, intf Interface) {
	t.Helper()

	clk := newManualClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	b := newTestBMC(t, clk, username, password)

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	conn := udp.Wrap(pc)
	addr := pc.LocalAddr().(*net.UDPAddr)
	srv := server.NewServer(b, conn)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = srv.Serve(ctx)
	}()

	c, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(intf).
		WithTimeout(2 * time.Second).
		WithRetry(0)

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	if _, err := c.GetDeviceID(context.Background()); err != nil {
		t.Fatalf("GetDeviceID before idle: %v", err)
	}

	// Advance past inactivity timeout, tolerance, and one eviction scan.
	clk.Advance(bmc.DefaultInactivityTimeout + bmc.DefaultInactivityTimeoutTolerance + bmc.DefaultSessionEvictInterval + time.Second)
	time.Sleep(50 * time.Millisecond)

	if _, err := c.GetDeviceID(context.Background()); err == nil {
		t.Fatal("expected GetDeviceID to fail on evicted session")
	}

	c2, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient reconnect: %v", err)
	}
	c2.WithInterface(intf).WithTimeout(2 * time.Second).WithRetry(0)
	if err := c2.Connect(context.Background()); err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	defer func() { _ = c2.Close(context.Background()) }()
	if _, err := c2.GetDeviceID(context.Background()); err != nil {
		t.Fatalf("GetDeviceID after reconnect: %v", err)
	}
}

func newTestBMC(t *testing.T, clk clock.Clock, username, password string) *bmc.BMC {
	t.Helper()
	info := bmc.DeviceInfo{
		DeviceID:                32,
		DeviceRevision:          1,
		FirmwareMajor:           1,
		FirmwareMinor:           0,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x3D,
	}
	var guid [16]byte
	copy(guid[:], "go-ipmi-test\x00\x00\x00\x00\x00")
	b := bmc.New(info, guid, mock.New(), bmc.WithClock(clk))

	user, err := b.Users.Add(2, username)
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte(password))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	return b
}

// manualClock drives BMC/session timestamps and the server's eviction ticker.
type manualClock struct {
	mu      sync.Mutex
	now     time.Time
	tickers []*manualTicker
}

type manualTicker struct {
	clock    *manualClock
	interval time.Duration
	next     time.Time
	ch       chan time.Time
	stopped  bool
}

func newManualClock(start time.Time) *manualClock {
	return &manualClock{now: start}
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *manualClock) NewTimer(d time.Duration) clock.Timer {
	return clock.Real.NewTimer(d)
}

func (c *manualClock) NewTicker(d time.Duration) clock.Ticker {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &manualTicker{
		clock:    c,
		interval: d,
		next:     c.now.Add(d),
		ch:       make(chan time.Time, 4),
	}
	c.tickers = append(c.tickers, t)
	return t
}

func (t *manualTicker) C() <-chan time.Time { return t.ch }

func (t *manualTicker) Stop() {
	t.clock.mu.Lock()
	t.stopped = true
	t.clock.mu.Unlock()
}

func (c *manualClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	now := c.now
	for _, t := range c.tickers {
		if t.stopped {
			continue
		}
		for !now.Before(t.next) {
			select {
			case t.ch <- now:
			default:
			}
			t.next = t.next.Add(t.interval)
		}
	}
	c.mu.Unlock()
}

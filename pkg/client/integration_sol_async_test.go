package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/command/transport"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/types"
)

// filteredPacketConn sits under the test server and inspects inbound
// (client→server) datagrams, dropping those the callback marks. It emulates
// BMC behaviors the reference server does not exhibit — e.g. Supermicro BMCs
// silently discard SOL packets that carry no character data.
type filteredPacketConn struct {
	conn           *net.UDPConn
	inspectInbound func(p []byte) (drop bool)
}

func (f *filteredPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	for {
		// Mirror udp.Conn's read deadline so the server keeps waking up for
		// context checks even when every packet gets dropped.
		_ = f.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, addr, err := f.conn.ReadFromUDP(p)
		if err != nil {
			return n, addr, err
		}
		if f.inspectInbound != nil && f.inspectInbound(p[:n]) {
			continue
		}
		return n, addr, nil
	}
}

func (f *filteredPacketConn) WriteTo(data []byte, addr net.Addr) (int, error) {
	ua, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, fmt.Errorf("expected *net.UDPAddr, got %T", addr)
	}
	return f.conn.WriteToUDP(data, ua)
}

func (f *filteredPacketConn) Close() error { return f.conn.Close() }

// solWireStats decodes just enough of the RMCP+ session header (always
// cleartext) and — for unencrypted sessions — the SOL header, to classify
// client→server packets.
type solWireStats struct {
	ipmiCommands  atomic.Int32
	solData       atomic.Int32
	solAckOnly    atomic.Int32
	solEmptyPolls atomic.Int32
}

// classify returns true when the packet is an empty SOL data-phase poll
// (non-zero sequence number, no character data) — the packets a
// Supermicro-class BMC never answers.
func (s *solWireStats) classify(p []byte) (emptyPoll bool) {
	if len(p) < 16 || p[0] != types.RmcpVersion || p[4] != byte(types.AuthTypeRMCPPlus) {
		return false
	}
	_, _, payloadType, flags, payload, ok := protocol.ParseRMCPPlusHeader(p)
	if !ok {
		return false
	}
	switch types.PayloadType(payloadType) {
	case types.PayloadTypeIPMI:
		s.ipmiCommands.Add(1)
		return false
	case types.PayloadTypeSOL:
	default:
		return false
	}
	if flags&types.PayloadFlagEncrypted != 0 || len(payload) < 4 {
		return false // encrypted payload: cannot inspect; never an empty poll
	}
	if len(payload) > 4 {
		s.solData.Add(1)
		return false
	}
	if payload[0]&0x0f == 0 {
		s.solAckOnly.Add(1)
		return false
	}
	s.solEmptyPolls.Add(1)
	return true
}

// newSOLTestRig starts the reference server behind a filteredPacketConn and
// returns a connected client, after applying preConnect tweaks (e.g. forcing
// a cleartext cipher suite so the filter can read SOL headers).
func newSOLTestRig(t *testing.T, inspect func(p []byte) (drop bool), preConnect func(*Client), bmcOpts ...bmc.Option) (*Client, *bmc.BMC, *mock.FakeConsoleConn) {
	t.Helper()
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})

	opts := append([]bmc.Option{bmc.WithClock(clock.Real)}, bmcOpts...)
	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, opts...)
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	srv := server.NewServer(b, &filteredPacketConn{conn: pc, inspectInbound: inspect})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = srv.Close() })
	go func() { _ = srv.Serve(ctx) }()

	addr := pc.LocalAddr().(*net.UDPAddr)
	c, err := NewClient(addr.IP.String(), addr.Port, "ADMIN", "ADMIN")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLanplus)
	if preConnect != nil {
		preConnect(c)
	}
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("client connect: %v", err)
	}
	return c, b, fake
}

// TestSOLStreamToleratesUnansweredEmptyPackets reproduces Supermicro X10
// behavior verified on real hardware (2026-10): that BMC
// never acknowledges SOL packets carrying no character data, so a client that
// polls with empty packets dies on its first unanswered exchange. The correct
// model is asynchronous receive (§15.3) with ACK-only packets (sequence 0h,
// §15.11). This test fails if empty-packet polling ever returns.
func TestSOLStreamToleratesUnansweredEmptyPackets(t *testing.T) {
	stats := &solWireStats{}
	// Suite 2 (HMAC-SHA1-96, no confidentiality) keeps the SOL header
	// readable so the filter can tell empty polls from ACK-only packets.
	c, _, fake := newSOLTestRig(t, stats.classify, func(c *Client) {
		c.session.v20.customSuiteIDs = []types.CipherSuiteID{types.CipherSuiteID2}
	}, bmc.WithCipherSuites([]types.CipherSuiteID{
		types.CipherSuiteID2, types.CipherSuiteID3, types.CipherSuiteID17,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inR, inW := io.Pipe()
	defer inW.Close()
	out := &lockedBuffer{}

	if _, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: 1,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c.DeactivatePayload(context.Background(), &transport.DeactivatePayloadRequest{
			PayloadType:     types.PayloadTypeSOL,
			PayloadInstance: 1,
		})
	})

	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLStream(ctx, inR, out)
	}()

	// BMC→console: async push must be delivered without any polling.
	fake.FeedRX([]byte("async-output\r\n"))
	waitForCondition(t, "async console output", func() bool {
		return strings.Contains(out.String(), "async-output")
	})

	// Console→BMC: keystrokes still flow.
	if _, err := inW.Write([]byte("k")); err != nil {
		t.Fatalf("write input: %v", err)
	}
	keystrokeDeadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(keystrokeDeadline) && fake.TXString() != "k" {
		time.Sleep(10 * time.Millisecond)
	}
	if fake.TXString() != "k" {
		t.Fatalf("keystroke did not reach console; wire stats: ipmi=%d solData=%d ackOnly=%d emptyPolls=%d",
			stats.ipmiCommands.Load(), stats.solData.Load(), stats.solAckOnly.Load(), stats.solEmptyPolls.Load())
	}

	// Stay idle well past a poll interval: a polling client would have sent
	// dozens of empty packets and died on the first one.
	select {
	case err := <-solErr:
		t.Fatalf("SOLStream died during idle period: %v", err)
	case <-time.After(3 * time.Second):
	}

	// The stream is still functional afterwards, both ways.
	fake.FeedRX([]byte("still-alive\r\n"))
	waitForCondition(t, "output after idle", func() bool {
		return strings.Contains(out.String(), "still-alive")
	})

	if got := stats.solEmptyPolls.Load(); got != 0 {
		t.Fatalf("client sent %d empty SOL poll packets (dropped by the filter); polling must not be used", got)
	}
	if stats.solAckOnly.Load() == 0 {
		t.Fatal("client never acknowledged console output with an ACK-only packet")
	}

	cancel()
	select {
	case <-solErr:
	case <-time.After(5 * time.Second):
		t.Fatal("SOLStream did not return after cancel")
	}
}

// BMCs that do not count SOL payload packets as session
// activity (OpenIPMI's ipmi_sim; spec §6.12.15 is ambiguous about what counts)
// expire the session at their inactivity timeout — 30s by default for
// ipmi_sim, 60s per Table 6-7. The client's only IPMI-command traffic during
// an idle SOL session is the keepalive, so the keepalive must run well under
// 30 seconds. This test watches the wire and fails if no command traffic
// appears within 12 seconds of an idle SOL session.
func TestSOLStreamKeepsSessionAliveDuringIdle(t *testing.T) {
	stats := &solWireStats{}
	c, _, _ := newSOLTestRig(t, func(p []byte) bool {
		stats.classify(p)
		return false
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inR, inW := io.Pipe()
	defer inW.Close()

	if _, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: 1,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c.DeactivatePayload(context.Background(), &transport.DeactivatePayloadRequest{
			PayloadType:     types.PayloadTypeSOL,
			PayloadInstance: 1,
		})
	})

	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLStream(ctx, inR, &lockedBuffer{})
	}()

	before := stats.ipmiCommands.Load()
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-solErr:
			t.Fatalf("SOLStream died while idle: %v", err)
		default:
		}
		if stats.ipmiCommands.Load() > before {
			return // keepalive observed on the wire
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no IPMI command traffic during 12s idle SOL session; keepalive interval %ds cannot beat aggressive inactivity timeouts (ipmi_sim defaults to 30s)",
		DefaultKeepAliveIntervalSec)
}

package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/command/transport"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

// lockedBuffer is an io.Writer safe for concurrent reads from the test.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// newSOLTestServer starts the reference server on a UDP loopback port with a
// fake console attached, and connects a lanplus client to it.
func newSOLTestServer(t *testing.T) (c *Client, b *bmc.BMC, fake *mock.FakeConsoleConn) {
	t.Helper()
	fake = &mock.FakeConsoleConn{}
	c, b = newSOLTestServerWithConsole(t, &mock.Console{Conn: fake})
	return c, b, fake
}

// newSOLTestServerWithConsole is newSOLTestServer with a caller-supplied
// console HAL (e.g. one that swaps the attached console on reconnect).
func newSOLTestServerWithConsole(t *testing.T, ch *mock.Console) (c *Client, b *bmc.BMC) {
	t.Helper()

	m := mock.New()
	m.SetConsole(ch)

	b = bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, bmc.WithClock(clock.Real))
	// Reconnection is opt-in (disabled by default); the loopback tests that
	// go through this helper all exercise the reconnect path.
	b.SOL.SetReconnectPolicy(&bmc.DefaultReconnectPolicy)
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

	srv := server.NewServer(b, udp.Wrap(pc))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = srv.Close() })
	go func() { _ = srv.Serve(ctx) }()

	addr := pc.LocalAddr().(*net.UDPAddr)
	c, err = NewClient(addr.IP.String(), addr.Port, "ADMIN", "ADMIN")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLanplus)
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("client connect: %v", err)
	}
	return c, b
}

func waitForCondition(t *testing.T, what string, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// TestSOLSessionLoopback runs a full SOL session against the reference
// server over UDP loopback: activation, console output to the remote side,
// keystrokes to the console, and clean deactivation on input EOF.
func TestSOLSessionLoopback(t *testing.T) {
	c, b, fake := newSOLTestServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inR, inW := io.Pipe()
	out := &lockedBuffer{}

	// Poll slower than the server's 50ms retry interval (§15.11): the retry
	// engine must fire between polls so async retransmissions exist — a fast
	// poller would ack every packet before a retry could be sent.
	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLActivate(ctx, inR, out, &SOLActivateOptions{PollInterval: 100 * time.Millisecond})
	}()

	waitForCondition(t, "SOL activation", func() bool {
		return b.SOL.ActiveSessionID(1) != 0
	})

	// BMC→console: bytes appearing on the system serial port must reach the
	// remote console's output stream.
	fake.FeedRX([]byte("login: "))
	waitForCondition(t, "console output at remote", func() bool {
		return strings.Contains(out.String(), "login: ")
	})

	// Console→BMC: keystrokes from the remote console must land on the
	// system serial port.
	if _, err := inW.Write([]byte("root\r")); err != nil {
		t.Fatalf("write input: %v", err)
	}
	waitForCondition(t, "keystrokes at console", func() bool {
		return strings.Contains(fake.TXString(), "root\r")
	})

	// §15.9/§15.11: a retransmission reuses its sequence number, so the
	// receiver must not re-deliver it. The server pushes console data both
	// async and piggybacked on poll responses, and retries unacked packets;
	// the client must consume only the response to its current poll and
	// display the data exactly once. The keystroke round-trip above spans
	// several poll cycles, so any duplicate delivery has settled by now.
	if got := strings.Count(out.String(), "login: "); got != 1 {
		t.Fatalf("console output delivered %d times, want exactly once: %q", got, out.String())
	}

	// Input EOF ends the session and deactivates the payload (client defer),
	// which closes the console conn.
	_ = inW.Close()
	select {
	case err := <-solErr:
		if err != nil {
			t.Fatalf("SOLActivate: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SOLActivate did not return after input EOF")
	}
	waitForCondition(t, "payload deactivation", func() bool {
		return b.SOL.ActiveSessionID(1) == 0
	})
	waitForCondition(t, "console conn close", fake.IsClosed)
}

// TestSOLActivateConflict verifies that a second activation while SOL is
// active fails with 80h (payload already active on another session,
// Table 24-2).
func TestSOLActivateConflict(t *testing.T) {
	c1, _, _ := newSOLTestServer(t)
	ctx := context.Background()

	if _, err := c1.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: 1,
	}); err != nil {
		t.Fatalf("first activate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c1.DeactivatePayload(ctx, &transport.DeactivatePayloadRequest{
			PayloadType:     types.PayloadTypeSOL,
			PayloadInstance: 1,
		})
	})

	// Second client on the same server.
	c2, _, _ := newSOLTestServerFromExisting(t, c1)
	_, err := c2.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: 1,
	})
	if err == nil {
		t.Fatal("second activate must fail")
	}
	respErr, ok := types.IsResponseError(err)
	if !ok {
		t.Fatalf("second activate error %v is not a ResponseError", err)
	}
	if respErr.CompletionCode() != types.CompletionCode(0x80) {
		t.Fatalf("second activate CC = %#02x, want 0x80", respErr.CompletionCode())
	}
}

// newSOLTestServerFromExisting connects a second client to the server an
// existing client is already talking to.
func newSOLTestServerFromExisting(t *testing.T, c1 *Client) (c2 *Client, b *bmc.BMC, fake *mock.FakeConsoleConn) {
	t.Helper()
	c2, err := NewClient(c1.Host, c1.Port, "ADMIN", "ADMIN")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c2.WithInterface(InterfaceLanplus)
	if err := c2.Connect(context.Background()); err != nil {
		t.Fatalf("client2 connect: %v", err)
	}
	return c2, nil, nil
}

// TestSOLReconnectLoopback verifies the remote console experiences a
// transparent recovery: when the system console fails and the server
// re-attaches to a fresh console, the SOL session keeps running with no
// client action — output resumes, keystrokes flow, the payload stays active.
func TestSOLReconnectLoopback(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	fake2 := &mock.FakeConsoleConn{}
	opens := 0
	ch := &mock.Console{OpenHook: func(context.Context) (hal.ConsoleConn, error) {
		opens++
		if opens == 1 {
			return fake1, nil // activation
		}
		return fake2, nil // reconnects
	}}
	c, b := newSOLTestServerWithConsole(t, ch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inR, inW := io.Pipe()
	out := &lockedBuffer{}

	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLActivate(ctx, inR, out, &SOLActivateOptions{PollInterval: 20 * time.Millisecond})
	}()
	waitForCondition(t, "SOL activation", func() bool {
		return b.SOL.ActiveSessionID(1) != 0
	})

	// Normal round before the outage.
	fake1.FeedRX([]byte("login: "))
	waitForCondition(t, "console output at remote", func() bool {
		return strings.Contains(out.String(), "login: ")
	})
	if _, err := inW.Write([]byte("root\r")); err != nil {
		t.Fatalf("write input: %v", err)
	}
	waitForCondition(t, "keystrokes at console", func() bool {
		return strings.Contains(fake1.TXString(), "root\r")
	})

	// Console dies mid-session; anything queued on it afterwards is stale
	// residue that must not reach the remote console.
	fake1.SetReadErr(errors.New("console died"))
	fake1.FeedRX([]byte("stale"))

	// Recovery is server-driven: the dead conn is released, a fresh one
	// attached, and output resumes — with no client-side action.
	waitForCondition(t, "console reconnect", fake1.IsClosed)
	fake2.FeedRX([]byte("recovered: "))
	waitForCondition(t, "console output after reconnect", func() bool {
		return strings.Contains(out.String(), "recovered: ")
	})
	if strings.Contains(out.String(), "stale") {
		t.Fatalf("stale pre-outage residue leaked to the remote console: %q", out.String())
	}

	// Keystrokes now land on the fresh console; the session is untouched.
	if _, err := inW.Write([]byte("who\r")); err != nil {
		t.Fatalf("write input after reconnect: %v", err)
	}
	waitForCondition(t, "keystrokes at reconnected console", func() bool {
		return strings.Contains(fake2.TXString(), "who\r")
	})
	if b.SOL.ActiveSessionID(1) == 0 {
		t.Fatal("payload deactivated by the reconnect")
	}
	select {
	case err := <-solErr:
		t.Fatalf("SOLActivate returned during reconnect: %v", err)
	default:
	}

	// Clean deactivation still works at the end.
	_ = inW.Close()
	select {
	case err := <-solErr:
		if err != nil {
			t.Fatalf("SOLActivate: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SOLActivate did not return after input EOF")
	}
}

// TestSOLPayloadAckOnly verifies an ACK-only SOL packet (sequence 0h,
// §15.11) still exchanges successfully after data has flowed: the BMC
// echoes the last accepted data sequence number, which the response
// matcher must not demand to equal the request's (zero) sequence number.
func TestSOLPayloadAckOnly(t *testing.T) {
	c, _, _ := newSOLTestServer(t)
	ctx := context.Background()
	if _, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: 1,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c.DeactivatePayload(ctx, &transport.DeactivatePayloadRequest{
			PayloadType:     types.PayloadTypeSOL,
			PayloadInstance: 1,
		})
	})

	// Data packet first, so the server's acked echo becomes non-zero.
	if _, err := c.SOLPayload(ctx, &types.SOLPayloadRequest{
		SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("k")},
	}); err != nil {
		t.Fatalf("data exchange: %v", err)
	}
	// ACK-only (seq 0h) must still get a reply.
	if _, err := c.SOLPayload(ctx, &types.SOLPayloadRequest{}); err != nil {
		t.Fatalf("ack-only exchange: %v", err)
	}
}

// TestGetChannelPayloadSupport verifies the §24.8 reply is spec-length (8
// bytes): the client rejects shorter replies, so this is what keeps payload
// discovery working.
func TestGetChannelPayloadSupport(t *testing.T) {
	c, _, _ := newSOLTestServer(t)
	res, err := c.GetChannelPayloadSupport(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetChannelPayloadSupport: %v", err)
	}
	if !res.PayloadTypeIPMI || !res.PayloadTypeSOL {
		t.Fatalf("standard payloads: IPMI=%v SOL=%v, want both supported", res.PayloadTypeIPMI, res.PayloadTypeSOL)
	}
	if !res.PayloadTypeRmcpOpenSessionRequest || !res.PayloadTypeRAKPMessage4 {
		t.Fatalf("session-setup payloads not fully supported: %+v", res)
	}
}

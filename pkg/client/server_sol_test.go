package client

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/command/transport"
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
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})

	b = bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, bmc.WithClock(clock.Real))
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
	return c, b, fake
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

	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLActivate(ctx, inR, out, &SOLActivateOptions{PollInterval: 20 * time.Millisecond})
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

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
	"testing/iotest"
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

// A removed session makes the probe time out. One timeout is a UDP blip and
// must not close the client; a second consecutive miss does.
func TestSessionProbeLostOnlyForDeadSessionStatus(t *testing.T) {
	if !sessionProbeLost(types.NewRmcpStatusError(types.RmcpStatusCodeInvalidSessionID)) {
		t.Fatal("invalid session id should close")
	}
	if !sessionProbeLost(types.NewRmcpStatusError(types.RmcpStatusCodeInactiveSessionID)) {
		t.Fatal("inactive session id should close")
	}
	if sessionProbeLost(types.NewRmcpStatusError(types.RmcpStatusCodeInvalidPayloadType)) {
		t.Fatal("unrelated RMCP status should not close")
	}
	if sessionProbeLost(types.NewResponseError(0x80, "answered")) {
		t.Fatal("a completion code means the BMC answered")
	}
	if sessionProbeLost(errors.New("udp timeout")) {
		t.Fatal("a timeout is inconclusive")
	}
}

func TestKeepaliveClosesWhenSessionDies(t *testing.T) {
	c, b, _ := newSOLTestServer(t)
	if err := b.Sessions.Close(c.session.v20.bmcSessionID); err != nil {
		t.Fatalf("close session: %v", err)
	}
	c.WithRetry(0).WithTimeout(200 * time.Millisecond)
	c.keepaliveCancel()
	<-c.keepaliveDone

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.keepSessionAlive(context.Background(), 1)
	}()
	time.Sleep(1500 * time.Millisecond)
	select {
	case <-c.lanClosed:
		t.Fatal("closed on the first inconclusive probe")
	default:
	}
	select {
	case <-c.lanClosed:
	case <-time.After(3 * time.Second):
		t.Fatal("keepalive left a dead session open after repeated probes")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("keepalive did not return")
	}
}

// piggyOnWrite makes the keystroke's own SOL response carry console output,
// instead of letting the pump push that output as an earlier packet.
type piggyOnWrite struct {
	*mock.FakeConsoleConn
	msg []byte
}

func (p *piggyOnWrite) Write(b []byte) (int, error) {
	n, err := p.FakeConsoleConn.Write(b)
	if err == nil && len(b) > 0 && len(p.msg) > 0 {
		p.FeedRX(p.msg)
		p.msg = nil
	}
	return n, err
}

func TestSOLStreamWritesPiggybackedOutputBeforeEOF(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	c, _ := newSOLTestServerWithConsole(t, &mock.Console{Conn: &piggyOnWrite{FakeConsoleConn: fake, msg: []byte("login: ")}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	inR, inW := io.Pipe()
	out := &lockedBuffer{}
	solErr := make(chan error, 1)
	go func() { solErr <- c.SOLStream(ctx, inR, out) }()

	if _, err := inW.Write([]byte("k")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := inW.Close(); err != nil {
		t.Fatalf("close input: %v", err)
	}
	select {
	case err := <-solErr:
		if err != nil {
			t.Fatalf("SOLStream: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SOLStream did not return on EOF")
	}
	if !strings.Contains(out.String(), "login: ") {
		t.Fatalf("output = %q, want piggybacked login:", out.String())
	}
	if fake.TXString() != "k" {
		t.Fatalf("console got %q, want k", fake.TXString())
	}
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
		solErr <- c.SOLActivate(ctx, inR, out, nil)
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
	if _, err := inW.Write([]byte("r")); err != nil {
		t.Fatalf("write input: %v", err)
	}
	waitForCondition(t, "single keystroke before more input", func() bool {
		return fake.TXString() == "r"
	})
	if _, err := inW.Write([]byte("oot\r")); err != nil {
		t.Fatalf("write input: %v", err)
	}
	waitForCondition(t, "keystrokes at console", func() bool {
		return strings.Contains(fake.TXString(), "root\r")
	})

	// §15.9/§15.11: a retransmission reuses its sequence number, so the
	// receiver must not re-deliver it. The server pushes console data
	// asynchronously and retries unacked packets; the client acknowledges
	// each packet once and must display the data exactly once. The round
	// trips above gave any retry a chance to arrive, so duplicate delivery
	// would be visible by now.
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

func TestSOLStreamImmediatelyAcknowledgesOutput(t *testing.T) {
	c, _, fake := newSOLTestServer(t)
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

	// Output buffered before the stream starts sits in the socket's kernel
	// buffer; the receive loop must pick it up without any client polling.
	want := bytes.Repeat([]byte("x"), 2*bmc.SOLMaxPayloadChars+1)
	fake.FeedRX(want)

	inR, inW := io.Pipe()
	out := &lockedBuffer{}
	solErr := make(chan error, 1)
	go func() {
		solErr <- c.SOLStream(ctx, inR, out)
	}()

	waitForCondition(t, "buffered console output", func() bool {
		return out.String() == string(want)
	})

	// Multiple BMC packets were delivered, each acknowledged (§15.11): the
	// server retransmits unacknowledged output, so exactly-once delivery here
	// proves the ACK path.
	_ = inW.Close()
	select {
	case err := <-solErr:
		if err != nil {
			t.Fatalf("SOLStream: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SOLStream did not return after input EOF")
	}
}

type rejectingConsole struct {
	mock.FakeConsoleConn

	mu sync.Mutex

	rejects       int
	writes        int
	acceptLimit   int
	maxWrite      int
	partialWrites int
}

func (c *rejectingConsole) Write(p []byte) (int, error) {
	c.mu.Lock()
	if len(p) == 0 {
		c.mu.Unlock()
		return 0, nil
	}
	c.writes++
	c.maxWrite = max(c.maxWrite, len(p))
	if c.rejects != 0 {
		if c.rejects > 0 {
			c.rejects--
		}
		c.mu.Unlock()
		return 0, nil
	}
	if c.acceptLimit > 0 && len(p) > c.acceptLimit {
		p = p[:c.acceptLimit]
		c.partialWrites++
	}
	c.mu.Unlock()
	return c.FakeConsoleConn.Write(p)
}

func (c *rejectingConsole) state() (writes int, tx string) {
	c.mu.Lock()
	writes = c.writes
	c.mu.Unlock()
	return writes, c.TXString()
}

func TestSOLStreamBatchesInput(t *testing.T) {
	for _, tt := range []struct {
		name        string
		payloadSize uint16
		maxChars    int
		acceptLimit int
		readErr     error
	}{
		{name: "negotiated size", payloadSize: 255, maxChars: 251},
		{name: "small payload", payloadSize: 12, maxChars: 8},
		{name: "acknowledgement limit", payloadSize: 1024, maxChars: 255},
		{name: "unknown size", maxChars: 1},
		{name: "header only", payloadSize: 4, maxChars: 1},
		{name: "partial acceptance", payloadSize: 255, maxChars: 251, acceptLimit: 100},
		{name: "reader error after data", payloadSize: 255, maxChars: 251, readErr: io.ErrUnexpectedEOF},
	} {
		t.Run(tt.name, func(t *testing.T) {
			console := &rejectingConsole{acceptLimit: tt.acceptLimit}
			c, _ := newSOLTestServerWithConsole(t, &mock.Console{Conn: console})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			t.Cleanup(func() { _ = c.Close(context.Background()) })
			activation, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
				PayloadType: types.PayloadTypeSOL, PayloadInstance: 1,
			})
			if err != nil {
				t.Fatalf("ActivatePayload() error = %v", err)
			}
			if c.session.v20.solInboundPayloadSize != activation.InboundPayloadSize {
				t.Fatal("ActivatePayload did not retain the inbound payload size")
			}
			// Exercise other advertised limits without changing the reference BMC.
			c.session.v20.solInboundPayloadSize = tt.payloadSize

			want := strings.Repeat("a~X\n", 140)
			var input io.Reader = iotest.DataErrReader(strings.NewReader(want))
			if tt.readErr != nil {
				input = io.MultiReader(strings.NewReader(want), iotest.ErrReader(tt.readErr))
			}
			err = c.SOLStream(ctx, input, io.Discard)
			if !errors.Is(err, tt.readErr) {
				t.Fatalf("SOLStream() error = %v, want %v", err, tt.readErr)
			}
			if got := console.TXString(); got != want {
				t.Fatalf("console received %q, want %q", got, want)
			}
			console.mu.Lock()
			maxWrite, partialWrites := console.maxWrite, console.partialWrites
			console.mu.Unlock()
			if maxWrite > tt.maxChars || (tt.maxChars > 1 && maxWrite <= 1) {
				t.Fatalf("largest input packet = %d characters, want batching bounded by %d", maxWrite, tt.maxChars)
			}
			if tt.acceptLimit > 0 && partialWrites == 0 {
				t.Fatal("no packet exercised partial acceptance")
			}
		})
	}
}

func TestSOLStreamRetriesUnacceptedData(t *testing.T) {
	console := &rejectingConsole{rejects: 1}
	c, _ := newSOLTestServerWithConsole(t, &mock.Console{Conn: console})

	err := c.SOLActivate(context.Background(), strings.NewReader("x"), io.Discard, nil)
	if err != nil {
		t.Fatalf("SOLActivate() error = %v", err)
	}

	writes, tx := console.state()
	if writes != 2 || tx != "x" {
		t.Fatalf("console writes = %d, data = %q, want 2 writes and %q", writes, tx, "x")
	}
}

func TestSOLStreamBoundsUnacceptedDataRetries(t *testing.T) {
	console := &rejectingConsole{rejects: -1}
	c, _ := newSOLTestServerWithConsole(t, &mock.Console{Conn: console})

	err := c.SOLActivate(context.Background(), strings.NewReader("x"), io.Discard, nil)
	if err == nil || !strings.Contains(err.Error(), "did not accept") {
		t.Fatalf("SOLActivate() error = %v, want bounded data transmission error", err)
	}

	writes, tx := console.state()
	if writes != defaultSOLTransmitRetryCount+1 || tx != "" {
		t.Fatalf("console writes = %d, data = %q, want %d writes and no data",
			writes, tx, defaultSOLTransmitRetryCount+1)
	}
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
		solErr <- c.SOLActivate(ctx, inR, out, nil)
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

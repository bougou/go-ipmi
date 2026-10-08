package client

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/types"
)

func waitLANSignal(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("lifecycle operation did not reach barrier")
	}
}

func TestLANRepeatedConnectJoinsPreviousKeepalive(t *testing.T) {
	c, closeClient := newLANConcurrencyClient(t, InterfaceLanplus)
	previousCancel, previousDone := c.keepaliveCancel, c.keepaliveDone
	defer func() { previousCancel(); waitLANSignal(t, previousDone) }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-previousDone:
	default:
		t.Fatal("reconnect returned before the previous keepalive stopped")
	}
	closeClient()
	select {
	case <-c.keepaliveDone:
	default:
		t.Fatal("Close returned before the replacement keepalive stopped")
	}
}

func TestLANSetupEntryPoints(t *testing.T) {
	for _, name := range []string{"Connect15", "Connect20", "Auto15", "Auto20"} {
		t.Run(name, func(t *testing.T) {
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					res, cc, err := next.Handle(ctx, hctx, data)
					if name == "Auto15" && hctx.Command == types.CommandGetChannelAuthCapabilities && len(res) >= 4 {
						res[3] &^= 2 // Advertise IPMI 1.5 only.
					}
					return res, cc, err
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, closeClient := newLANSetupClient(t, InterfaceLanplus, server.WithHandlerRegistry(registry))
			connect := c.ConnectAuto
			switch name {
			case "Connect15":
				connect = c.Connect15
			case "Connect20":
				connect = c.Connect20
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := connect(ctx); err != nil {
				t.Fatal(err)
			}
			wantV20 := name == "Connect20" || name == "Auto20"
			if c.v20 != wantV20 {
				t.Fatalf("v20 = %v, want %v", c.v20, wantV20)
			}
			closeClient()
			// Every entry point must reject setup before mutating the established mode.
			for _, setup := range []func(context.Context) error{c.Connect, c.Connect15, c.Connect20, c.ConnectAuto} {
				if err := setup(ctx); !errors.Is(err, net.ErrClosed) {
					t.Fatalf("setup after Close: %v", err)
				}
				if c.v20 != wantV20 {
					t.Fatal("closed setup changed session mode")
				}
			}
		})
	}
}

func TestLANCloseJoinsSetupBeforeSessionCleanup(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			var closes atomic.Int32
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					if hctx.Command == types.CommandCloseSession {
						closes.Add(1)
					}
					return next.Handle(ctx, hctx, data)
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, _ := newLANSetupClient(t, intf, server.WithHandlerRegistry(registry))
			setup := c.connect15
			if intf == InterfaceLanplus {
				setup = c.connect20
			}
			entered, release := make(chan struct{}), make(chan struct{})
			defer func() {
				select {
				case <-release:
				default:
					close(release)
				}
			}()
			connected := make(chan error, 1)
			go func() {
				connected <- c.connectLAN(context.Background(), func(ctx context.Context) error {
					err := setup(ctx) // Real wire setup, paused before keepalive publication.
					close(entered)
					<-release
					return err
				})
			}()
			waitLANSignal(t, entered)
			closed := make(chan error, 1)
			go func() { closed <- c.Close(context.Background()) }()
			waitLANSignal(t, c.lanClosed)
			select {
			case err := <-closed:
				t.Fatalf("Close returned during setup: %v", err)
			case <-time.After(20 * time.Millisecond):
			}
			if got := closes.Load(); got != 0 {
				t.Fatalf("CloseSession ran before setup finished: %d", got)
			}
			close(release)
			if err := awaitUDPResult(t, connected); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("setup: %v", err)
			}
			if err := awaitUDPResult(t, closed); err != nil {
				t.Fatal(err)
			}
			if got := closes.Load(); got != 1 {
				t.Fatalf("CloseSession calls = %d, want 1", got)
			}
			if c.keepaliveDone != nil {
				t.Fatal("late setup published keepalive after Close")
			}
		})
	}
}

func TestLANSetupQueueCancellationAndClose(t *testing.T) {
	c, _ := newLANSetupClient(t, InterfaceLanplus)
	entered, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	first := make(chan error, 1)
	go func() {
		first <- c.connectLAN(context.Background(), func(ctx context.Context) error {
			close(entered)
			<-release
			return nil
		})
	}()
	waitLANSignal(t, entered)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := c.Connect(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("queued setup: %v", err)
	}
	queued := make(chan error, 1)
	go func() { queued <- c.Connect(context.Background()) }()
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer closeCancel()
	if err := c.Close(closeCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Close with paused setup: %v", err)
	}
	if err := awaitUDPResult(t, queued); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("queued setup after Close: %v", err)
	}
	close(release)
	if err := awaitUDPResult(t, first); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("active setup after Close: %v", err)
	}
	if c.keepaliveDone != nil {
		t.Fatal("setup published keepalive after Close deadline")
	}
}

func TestLANSetupCancellationPreventsKeepalive(t *testing.T) {
	c, _ := newLANSetupClient(t, InterfaceLanplus)
	ctx, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("setup canceled")
	err := c.connectLAN(ctx, func(ctx context.Context) error {
		if err := c.connect20(ctx); err != nil {
			return err
		}
		cancel(cause)
		return nil
	})
	if !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("setup cancellation: %v", err)
	}
	if c.keepaliveDone != nil {
		t.Fatal("canceled setup published keepalive")
	}
}

// Let cancellation interrupt real socket I/O, but hold the keepalive before
// Read returns. CloseSession cannot mask a missing join: the session is inactive.
type pausedKeepaliveConn struct {
	net.Conn
	reading, interrupted, release chan struct{}
}

func (c *pausedKeepaliveConn) Write(b []byte) (int, error) { return len(b), nil }
func (c *pausedKeepaliveConn) Read(b []byte) (int, error) {
	close(c.reading)
	n, err := c.Conn.Read(b)
	close(c.interrupted)
	<-c.release
	return n, err
}

func newPausedKeepalive(t *testing.T) (*Client, *pausedKeepaliveConn) {
	t.Helper()
	c, err := NewClient("127.0.0.1", 623, "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	conn, peer := net.Pipe()
	paused := &pausedKeepaliveConn{conn, make(chan struct{}), make(chan struct{}), make(chan struct{})}
	c.udpClient.conn = paused
	if err := c.startSessionKeepalive(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		select {
		case <-paused.release:
		default:
			close(paused.release)
		}
		c.keepaliveCancel()
		waitLANSignal(t, c.keepaliveDone)
		_ = c.Close(context.Background())
		_ = peer.Close()
	})
	waitLANSignal(t, paused.reading)
	return c, paused
}

func TestLANCloseJoinsInactiveKeepalive(t *testing.T) {
	c, paused := newPausedKeepalive(t)
	closed := make(chan error, 1)
	go func() { closed <- c.Close(context.Background()) }()
	waitLANSignal(t, paused.interrupted)
	select {
	case err := <-closed:
		t.Fatalf("Close returned before keepalive exited: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(paused.release)
	if err := awaitUDPResult(t, closed); err != nil {
		t.Fatal(err)
	}
	select {
	case <-c.keepaliveDone:
	default:
		t.Fatal("Close did not join keepalive")
	}
}

func TestLANReconnectWaitsForKeepalive(t *testing.T) {
	for _, cancelSetup := range []bool{false, true} {
		t.Run(map[bool]string{false: "join", true: "canceled"}[cancelSetup], func(t *testing.T) {
			c, paused := newPausedKeepalive(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			connected := make(chan error, 1)
			entered := make(chan struct{})
			setupErr := errors.New("setup reached")
			go func() {
				connected <- c.connectLAN(ctx, func(context.Context) error { close(entered); return setupErr })
			}()
			waitLANSignal(t, paused.interrupted)
			select {
			case <-entered:
				t.Fatal("setup started before previous keepalive exited")
			case <-time.After(20 * time.Millisecond):
			}
			if cancelSetup {
				cancel()
				if err := awaitUDPResult(t, connected); !errors.Is(err, context.Canceled) {
					t.Fatalf("setup: %v", err)
				}
				closed := make(chan error, 1)
				go func() { closed <- c.Close(context.Background()) }()
				waitLANSignal(t, c.lanClosed)
				select {
				case err := <-closed:
					t.Fatalf("Close lost the previous keepalive: %v", err)
				case <-time.After(20 * time.Millisecond):
				}
				close(paused.release)
				if err := awaitUDPResult(t, closed); err != nil {
					t.Fatal(err)
				}
				select {
				case <-entered:
					t.Fatal("canceled setup ran")
				default:
				}
			} else {
				close(paused.release)
				if err := awaitUDPResult(t, connected); !errors.Is(err, setupErr) {
					t.Fatalf("setup: %v", err)
				}
			}
		})
	}
}

func TestLANConcurrentSetup(t *testing.T) {
	c, _ := newLANSetupClient(t, InterfaceLanplus)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 2)
	for range 2 {
		go func() { done <- c.Connect(ctx) }()
	}
	for range 2 {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("concurrent setup blocked")
		}
	}
	if _, err := c.GetDeviceID(ctx); err != nil {
		t.Fatalf("command after setup: %v", err)
	}
}

func TestLANCloseDeadlineDuringSetupClosesUDP(t *testing.T) {
	c, _ := newLANSetupClient(t, InterfaceLanplus)
	entered, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	connected := make(chan error, 1)
	go func() {
		connected <- c.connectLAN(context.Background(), func(ctx context.Context) error {
			if err := c.connect20(ctx); err != nil {
				return err
			}
			close(entered)
			<-release
			return nil
		})
	}()
	waitLANSignal(t, entered)
	conn := c.udpClient.conn
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := c.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Close during setup: %v", err)
	}
	if _, err := conn.Write([]byte("closed")); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("UDP socket after Close: %v", err)
	}
	close(release)
	if err := awaitUDPResult(t, connected); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("setup after Close: %v", err)
	}
	if c.keepaliveDone != nil {
		t.Fatal("late keepalive after Close deadline")
	}
}

// Exercise the small window between keepalive publication and Connect returning
// through real LAN setup, without adding a production hook to that boundary.
func TestLANCallerCancellationAtConnectCompletion(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			type cancellation struct {
				cancel context.CancelFunc
				delay  time.Duration
				done   chan struct{}
			}
			var current atomic.Pointer[cancellation]
			var hookAt atomic.Int64
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					res, cc, err := next.Handle(ctx, hctx, data)
					if hctx.Command == types.CommandSetSessionPrivilegeLevel {
						now := time.Now()
						hookAt.Store(now.UnixNano())
						attempt := current.Load()
						go func() {
							defer close(attempt.done)
							if attempt.delay < 0 {
								return
							}
							// A timer's scheduling granularity can miss the entire window.
							for time.Since(now) < attempt.delay {
							}
							attempt.cancel()
						}()
					}
					return res, cc, err
				})
			})
			handlers.RegisterAllHandlers(registry)
			peer, _ := newLANSetupClient(t, intf, server.WithHandlerRegistry(registry))
			run := func(delay time.Duration) (time.Duration, bool) {
				c, err := NewClient(peer.Host, peer.Port, peer.Username, peer.Password)
				if err != nil {
					t.Fatal(err)
				}
				c.WithInterface(intf).WithTimeout(time.Second).WithRetry(0)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				defer func() {
					closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
					defer closeCancel()
					if err := c.Close(closeCtx); err != nil {
						t.Errorf("Close: %v", err)
					}
				}()
				attempt := &cancellation{cancel: cancel, delay: delay, done: make(chan struct{})}
				current.Store(attempt)
				err = c.Connect(ctx)
				elapsed := time.Duration(time.Now().UnixNano() - hookAt.Load())
				if err != nil && (delay < 0 || !errors.Is(err, context.Canceled)) {
					t.Fatalf("Connect: %v", err)
				}
				waitLANSignal(t, attempt.done)
				if err != nil && c.keepaliveDone != nil {
					select {
					case <-c.keepaliveDone:
					default:
						return elapsed, true
					}
				}
				return elapsed, false
			}
			var total time.Duration
			for range 50 {
				elapsed, _ := run(-1)
				total += elapsed
			}
			mean := total / 50
			inconsistent := 0
			for i := range 500 {
				delay := max(0, mean-40*time.Microsecond+time.Duration(i%51)*time.Microsecond)
				_, live := run(delay)
				if live {
					inconsistent++
				}
			}
			if inconsistent != 0 {
				t.Fatalf("%d Connect calls reported cancellation while leaving keepalive running", inconsistent)
			}
		})
	}
}

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

// Drop the keepalive reply after the BMC has finished handling it. Blocking a
// BMC handler instead would also block CloseSession on the server's session lock.
type dropKeepaliveReplyConn struct {
	net.Conn
	drop    atomic.Bool
	dropped chan struct{}
}

func (c *dropKeepaliveReplyConn) Read(b []byte) (int, error) {
	for {
		n, err := c.Conn.Read(b)
		if err == nil && c.drop.Swap(false) {
			close(c.dropped)
			continue
		}
		return n, err
	}
}

func TestLANCloseCancelsKeepalive(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			t.Parallel()
			conn := &dropKeepaliveReplyConn{dropped: make(chan struct{})}
			var closes atomic.Int32
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					if hctx.Command == types.CommandGetSessionInfo {
						conn.drop.Store(true)
					}
					if hctx.Command == types.CommandCloseSession {
						closes.Add(1)
					}
					return next.Handle(ctx, hctx, data)
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, closeClient := newLANConcurrencyClient(t, intf, server.WithHandlerRegistry(registry))
			c.udpClient.lock.Lock()
			conn.Conn = c.udpClient.conn
			c.udpClient.conn = conn
			c.udpClient.lock.Unlock()
			select {
			case <-conn.dropped:
			case <-time.After(35 * time.Second):
				t.Fatal("keepalive did not start")
			}
			started := time.Now()
			closeClient()
			if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
				t.Errorf("Close waited %s for keepalive", elapsed)
			}
			if got := closes.Load(); got != 1 {
				t.Errorf("CloseSession calls = %d, want 1", got)
			}
			select {
			case <-c.keepaliveDone:
			default:
				t.Error("Close returned before keepalive stopped")
			}
		})
	}
}

func TestLANCloseRepeated(t *testing.T) {
	c, err := NewClient("127.0.0.1", 623, "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("repeated Close panicked: %v", p)
		}
	}()
	for range 2 {
		if err := c.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLANConcurrentClose(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			entered, release := make(chan struct{}, 1), make(chan struct{})
			defer func() {
				select {
				case <-release:
				default:
					close(release)
				}
			}()
			var closes atomic.Int32
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					if hctx.Command == types.CommandCloseSession {
						closes.Add(1)
						entered <- struct{}{}
						<-release
					}
					return next.Handle(ctx, hctx, data)
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, _ := newLANConcurrencyClient(t, intf, server.WithHandlerRegistry(registry))
			first := make(chan error, 1)
			go func() { first <- c.Close(context.Background()) }()
			<-entered
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			if err := c.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("waiting Close: %v", err)
			}
			if _, err := c.GetDeviceID(context.Background()); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("command during Close: %v", err)
			}
			second := make(chan error, 1)
			go func() { second <- c.Close(context.Background()) }()
			close(release)
			if err := awaitUDPResult(t, first); err != nil {
				t.Fatal(err)
			}
			if err := awaitUDPResult(t, second); err != nil {
				t.Fatal(err)
			}
			if err := c.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			if got := closes.Load(); got != 1 {
				t.Fatalf("CloseSession calls: %d", got)
			}
			if _, err := c.GetDeviceID(context.Background()); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("command after Close: %v", err)
			}
		})
	}
}

func TestLANCloseDeadlineInterruptsCommand(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			entered, release := make(chan struct{}, 1), make(chan struct{})
			defer close(release)
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					if hctx.Command == types.CommandGetDeviceID {
						entered <- struct{}{}
						<-release
					}
					return next.Handle(ctx, hctx, data)
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, _ := newLANConcurrencyClient(t, intf, server.WithHandlerRegistry(registry))
			first := make(chan error, 1)
			go func() { _, err := c.GetDeviceID(context.Background()); first <- err }()
			<-entered
			queued := make(chan error, 1)
			go func() { _, err := c.GetDeviceID(context.Background()); queued <- err }()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			if err := c.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Close: %v", err)
			}
			if err := awaitUDPResult(t, first); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("active command: %v", err)
			}
			if err := awaitUDPResult(t, queued); !errors.Is(err, net.ErrClosed) {
				t.Fatalf("queued command: %v", err)
			}
		})
	}
}

type pausedLANTimeout struct{ entered, release chan struct{} }

func (*pausedLANTimeout) Error() string   { return "test timeout" }
func (*pausedLANTimeout) Temporary() bool { return true }
func (e *pausedLANTimeout) Timeout() bool { close(e.entered); <-e.release; return true }

type retryLANConn struct {
	closeTrackingConn
	timeout *pausedLANTimeout
}

func (c *retryLANConn) Write(b []byte) (int, error) { return len(b), nil }
func (c *retryLANConn) Read([]byte) (int, error) {
	return 0, &net.OpError{Op: "read", Net: "udp", Err: c.timeout}
}

type retryLANDialer struct {
	conn  net.Conn
	calls atomic.Int32
}

func (d *retryLANDialer) Dial(string, string) (net.Conn, error) {
	if d.calls.Add(1) != 1 {
		return nil, errors.New("unexpected redial")
	}
	return d.conn, nil
}
func TestLANClosePreventsRetryReopeningUDP(t *testing.T) {
	timeout := &pausedLANTimeout{make(chan struct{}), make(chan struct{})}
	defer func() {
		select {
		case <-timeout.release:
		default:
			close(timeout.release)
		}
	}()
	dialer := &retryLANDialer{conn: &retryLANConn{timeout: timeout}}
	c, err := NewClient("127.0.0.1", 623, "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	c.WithUDPProxy(dialer).WithRetry(1)
	done := make(chan error, 1)
	go func() { _, err := c.RmcpPing(context.Background()); done <- err }()
	<-timeout.entered
	if err := c.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	close(timeout.release)
	if err := awaitUDPResult(t, done); err == nil {
		t.Fatal("closed command succeeded")
	}
	if got := dialer.calls.Load(); got != 1 {
		t.Fatalf("Close was followed by another dial; calls = %d", got)
	}
}

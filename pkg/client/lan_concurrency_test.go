package client

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

func newLANConcurrencyClient(t *testing.T, intf Interface, options ...server.ServerOption) (*Client, func()) {
	t.Helper()
	b := newV15TestBMC(t, "test", "test-password")
	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	srv := server.NewServer(b, udp.Wrap(pc), options...)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); _ = srv.Serve(ctx) }()
	t.Cleanup(func() { cancel(); _ = srv.Close(); <-done })
	addr := pc.LocalAddr().(*net.UDPAddr)
	c, err := NewClient(addr.IP.String(), addr.Port, "test", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	c.WithInterface(intf).WithTimeout(time.Second).WithRetry(0)
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer connectCancel()
	if err := c.Connect(connectCtx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	var closeOnce sync.Once
	closeClient := func() {
		closeOnce.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = c.Close(ctx)
		})
	}
	t.Cleanup(closeClient)
	return c, closeClient
}

type pausedLANResponse struct {
	app.GetDeviceIDResponse
	entered chan struct{}
	release chan struct{}
}

func (r *pausedLANResponse) Unpack(data []byte) error {
	close(r.entered)
	<-r.release
	return r.GetDeviceIDResponse.Unpack(data)
}

type observedLANRequest struct {
	types.Request
	packed chan struct{}
}

func (r *observedLANRequest) Pack() []byte {
	close(r.packed)
	return r.Request.Pack()
}

func TestLANExchangeSerializesConstructionAndParsing(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			c, _ := newLANConcurrencyClient(t, intf)
			first := &pausedLANResponse{entered: make(chan struct{}), release: make(chan struct{})}
			firstDone := make(chan error, 1)
			go func() { firstDone <- c.Exchange(context.Background(), &app.GetDeviceIDRequest{}, first) }()
			defer func() {
				close(first.release)
				if err := <-firstDone; err != nil {
					t.Errorf("first exchange: %v", err)
				}
			}()
			select {
			case <-first.entered:
			case <-time.After(5 * time.Second):
				t.Fatal("first response did not arrive")
			}

			second := &observedLANRequest{Request: &app.GetDeviceIDRequest{}, packed: make(chan struct{})}
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			if err := c.Exchange(ctx, second, &app.GetDeviceIDResponse{}); !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("waiting exchange error = %v, want context deadline", err)
			}
			select {
			case <-second.packed:
				t.Error("waiting request consumed sequence numbers before the previous response finished parsing")
			default:
			}
		})
	}
}

func TestLANConcurrentCommandsAndKeepalive(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			var keepalives atomic.Int32
			registry := handlers.NewRegistry()
			registry.Use(func(next handlers.Handler) handlers.Handler {
				return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
					response, code, err := next.Handle(ctx, hctx, data)
					if hctx.Command == types.CommandGetSessionInfo && code == types.CodeOK && err == nil {
						keepalives.Add(1)
					}
					return response, code, err
				})
			})
			handlers.RegisterAllHandlers(registry)
			c, closeClient := newLANConcurrencyClient(t, intf, server.WithHandlerRegistry(registry))
			// Exercise the real keepalive loop without waiting for its 30-second default.
			keepaliveDone := make(chan struct{})
			go func() { defer close(keepaliveDone); c.keepSessionAlive(context.Background(), 1) }()
			defer func() { closeClient(); <-keepaliveDone }()
			var wg sync.WaitGroup
			errs := make(chan error, 8)
			start := make(chan struct{})
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					for range 80 {
						response, err := c.GetDeviceID(context.Background())
						if err != nil {
							errs <- err
							return
						}
						if response.DeviceID != 32 {
							errs <- errors.New("response belongs to the wrong command")
							return
						}
						time.Sleep(15 * time.Millisecond)
					}
				}()
			}
			close(start)
			wg.Wait()
			close(errs)
			for err := range errs {
				t.Error(err)
			}
			if keepalives.Load() == 0 {
				t.Error("no keepalive was handled during concurrent commands")
			}
		})
	}
}

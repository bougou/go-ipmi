package client

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
)

// a hostname resolving to both ::1 and 127.0.0.1 with an IPv4-only BMC.
func TestUDPDualStackHostnameIPv4OnlyPeer(t *testing.T) {
	peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	go func() {
		buf := make([]byte, 64)
		for {
			n, addr, err := peer.ReadFromUDP(buf)
			if err != nil {
				return
			}
			_, _ = peer.WriteToUDP(buf[:n], addr)
		}
	}()
	port := peer.LocalAddr().(*net.UDPAddr).Port
	c := NewUDPClient("localhost", port).SetTimeout(300 * time.Millisecond)
	defer c.Close()
	data, err := c.Exchange(context.Background(), strings.NewReader("ping"))
	if err != nil || string(data) != "ping" {
		t.Fatalf("Exchange via localhost: %q, %v", data, err)
	}
}

// reference BMC on 127.0.0.1, client configured with "localhost".
func TestLANConnectDualStackHostname(t *testing.T) {
	for _, intf := range []Interface{InterfaceLan, InterfaceLanplus} {
		t.Run(string(intf), func(t *testing.T) {
			b := newV15TestBMC(t, "test", "test-password")
			pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			srv := server.NewServer(b, udp.Wrap(pc))
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { defer close(done); _ = srv.Serve(ctx) }()
			defer func() { cancel(); _ = srv.Close(); <-done }()
			c, err := NewClient("localhost", pc.LocalAddr().(*net.UDPAddr).Port, "test", "test-password")
			if err != nil {
				t.Fatal(err)
			}
			c.WithInterface(intf).WithTimeout(300 * time.Millisecond).WithRetry(0)
			cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer ccancel()
			if err := c.Connect(cctx); err != nil {
				t.Fatalf("Connect via localhost: %v", err)
			}
			_ = c.Close(context.Background())
		})
	}
}

func TestUDPAddressLiterals(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:623", "[::1]:623", "[fe80::1%zone-test]:623", ":623"} {
		t.Run(addr, func(t *testing.T) {
			got, err := resolveUDPAddress(context.Background(), addr)
			if err != nil || got != addr {
				t.Fatalf("address = %q, %v; want %q", got, err, addr)
			}
		})
	}
}

// These tests are deliberately non-parallel because they replace DefaultResolver.
func TestUDPHostnameResolution(t *testing.T) {
	peer, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, remote, err := peer.ReadFrom(buf)
			if err != nil {
				return
			}
			var req dnsmessage.Message
			if err := req.Unpack(buf[:n]); err != nil {
				continue
			}
			res := dnsmessage.Message{Header: dnsmessage.Header{ID: req.ID, Response: true, RecursionAvailable: true}, Questions: req.Questions}
			for _, q := range req.Questions {
				hdr := dnsmessage.ResourceHeader{Name: q.Name, Type: q.Type, Class: dnsmessage.ClassINET, TTL: 60}
				switch q.Type {
				case dnsmessage.TypeAAAA:
					res.Answers = append(res.Answers, dnsmessage.Resource{Header: hdr, Body: &dnsmessage.AAAAResource{AAAA: [16]byte{15: 1}}})
				case dnsmessage.TypeA:
					if q.Name.String() == "dual.example." {
						res.Answers = append(res.Answers, dnsmessage.Resource{Header: hdr, Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}}})
					}
				}
			}
			wire, err := res.Pack()
			if err == nil {
				_, _ = peer.WriteTo(wire, remote)
			}
		}
	}()
	defer func() { _ = peer.Close(); <-done }()
	previous := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "udp4", peer.LocalAddr().String())
	}}
	defer func() { net.DefaultResolver = previous }()
	for _, tt := range []struct{ host, want string }{
		{"dual.example.", "127.0.0.1:623"},
		{"v6.example.", "[::1]:623"},
	} {
		t.Run(tt.host, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			got, err := resolveUDPAddress(ctx, net.JoinHostPort(tt.host, "623"))
			if err != nil || got != tt.want {
				t.Fatalf("resolved = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestUDPHostnameResolutionCancellation(t *testing.T) {
	previous := net.DefaultResolver
	entered := make(chan struct{}, 1)
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	defer func() { net.DefaultResolver = previous }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		conn, err := dialUDP(ctx, nil, "canceled.example.:623")
		if conn != nil {
			_ = conn.Close()
		}
		result <- err
	}()
	waitLANSignal(t, entered)
	cancel()
	if err := awaitUDPResult(t, result); !errors.Is(err, context.Canceled) {
		t.Fatalf("DNS cancellation: %v", err)
	}
}

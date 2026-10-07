package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

var udpExchanges = map[string]func(*UDPClient, context.Context, io.Reader) ([]byte, error){
	"Exchange": (*UDPClient).Exchange,
	"ExchangeUntilMatch": func(c *UDPClient, ctx context.Context, r io.Reader) ([]byte, error) {
		return c.ExchangeUntilMatch(ctx, r, func(b []byte) (bool, error) { return string(b) == "reply", nil })
	},
}

func newUDPTestPeer(t *testing.T) (*UDPClient, <-chan string) {
	t.Helper()
	peer, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	received := make(chan string, 32)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, addr, err := peer.ReadFromUDP(buf)
			if err != nil {
				return
			}
			received <- string(buf[:n])
			if string(buf[:n]) == "reply" {
				_, _ = peer.WriteToUDP(buf[:n], addr)
			}
		}
	}()
	t.Cleanup(func() { _ = peer.Close(); <-done })
	addr := peer.LocalAddr().(*net.UDPAddr)
	c := NewUDPClient(addr.IP.String(), addr.Port).SetTimeout(time.Second)
	t.Cleanup(func() { _ = c.Close() })
	return c, received
}

func awaitUDPResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(250 * time.Millisecond):
		t.Fatal("operation did not stop promptly")
		return nil
	}
}

func TestUDPExchangeCancellationAndReuse(t *testing.T) {
	for name, exchange := range udpExchanges {
		t.Run(name, func(t *testing.T) {
			c, received := newUDPTestPeer(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := exchange(c, ctx, strings.NewReader("hold")); done <- err }()
			<-received
			queuedCtx, queuedCancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer queuedCancel()
			queued := make(chan error, 1)
			go func() { _, err := exchange(c, queuedCtx, strings.NewReader("queued")); queued <- err }()
			if err := awaitUDPResult(t, queued); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("queued error: %v", err)
			}
			cancel()
			if err := awaitUDPResult(t, done); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancel error: %v", err)
			}
			reuseCtx, reuseCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer reuseCancel()
			data, err := exchange(c, reuseCtx, strings.NewReader("reply"))
			if err != nil || string(data) != "reply" {
				t.Fatalf("reuse: %q, %v", data, err)
			}
			if got := <-received; got != "reply" {
				t.Fatalf("canceled queued call wrote %q", got)
			}
		})
	}
}

func TestUDPExchangeCloseInterruptsRead(t *testing.T) {
	for name, exchange := range udpExchanges {
		t.Run(name, func(t *testing.T) {
			c, received := newUDPTestPeer(t)
			done := make(chan error, 1)
			go func() { _, err := exchange(c, context.Background(), strings.NewReader("hold")); done <- err }()
			<-received
			closed := make(chan error, 1)
			go func() { closed <- c.Close() }()
			if err := awaitUDPResult(t, closed); err != nil {
				t.Fatal(err)
			}
			if err := awaitUDPResult(t, done); err == nil {
				t.Fatal("exchange succeeded after close")
			}
			for range 2 {
				data, err := exchange(c, context.Background(), strings.NewReader("reply"))
				if err != nil || string(data) != "reply" {
					t.Fatalf("reuse after close: %q, %v", data, err)
				}
			}
		})
	}
}

type writeSignalConn struct {
	net.Conn
	started chan struct{}
}

func (c *writeSignalConn) Write(b []byte) (int, error) { close(c.started); return c.Conn.Write(b) }
func TestUDPExchangeCancelBlockedWrite(t *testing.T) {
	for name, exchange := range udpExchanges {
		t.Run(name, func(t *testing.T) {
			conn, peer := net.Pipe()
			defer conn.Close()
			defer peer.Close()
			started := make(chan struct{})
			c := NewUDPClient("unused", 623)
			c.conn = &writeSignalConn{conn, started}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := exchange(c, ctx, strings.NewReader("request")); done <- err }()
			<-started
			cancel()
			if err := awaitUDPResult(t, done); !errors.Is(err, context.Canceled) {
				t.Fatalf("error: %v", err)
			}
			closed := make(chan error, 1)
			go func() { closed <- c.Close() }()
			if err := awaitUDPResult(t, closed); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type delayedUDPDialer struct {
	started, release chan struct{}
	conn             net.Conn
}

func (d *delayedUDPDialer) Dial(string, string) (net.Conn, error) {
	close(d.started)
	<-d.release
	return d.conn, nil
}
func TestUDPExchangeCancelLegacyDial(t *testing.T) {
	conn, peer := net.Pipe()
	defer conn.Close()
	defer peer.Close()
	dialer := &delayedUDPDialer{make(chan struct{}), make(chan struct{}), conn}
	c := NewUDPClient("unused", 623).SetProxy(dialer)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.Exchange(ctx, strings.NewReader("request")); done <- err }()
	<-dialer.started
	t.Cleanup(func() {
		select {
		case <-dialer.release:
		default:
			close(dialer.release)
		}
	})
	cancel()
	if err := awaitUDPResult(t, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("error: %v", err)
	}
	closed := make(chan error, 1)
	go func() { closed <- c.Close() }()
	if err := awaitUDPResult(t, closed); err != nil {
		t.Fatal(err)
	}
	close(dialer.release)
	_ = peer.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
	if _, err := peer.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("late dial connection not closed: %v", err)
	}
}

type contextUDPDialer struct{ started chan struct{} }

func (d *contextUDPDialer) Dial(string, string) (net.Conn, error) {
	return nil, errors.New("legacy Dial used")
}
func (d *contextUDPDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	close(d.started)
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestUDPExchangeCloseInterruptsDial(t *testing.T) {
	d := &contextUDPDialer{make(chan struct{})}
	c := NewUDPClient("unused", 623).SetProxy(d)
	done := make(chan error, 1)
	go func() { _, err := c.Exchange(context.Background(), strings.NewReader("request")); done <- err }()
	<-d.started
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := awaitUDPResult(t, done); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("close during dial: %v", err)
	}
}

func TestUDPExchangeUntilMatchOverallDeadline(t *testing.T) {
	conn, peer := net.Pipe()
	defer conn.Close()
	defer peer.Close()
	c := NewUDPClient("unused", 623).SetTimeout(60 * time.Millisecond)
	c.conn = conn
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 128)
		if _, err := peer.Read(buf); err != nil {
			return
		}
		for {
			if _, err := peer.Write([]byte("unrelated")); err != nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	defer func() { _ = peer.Close(); <-done }()
	started := time.Now()
	_, err := c.ExchangeUntilMatch(context.Background(), strings.NewReader("request"), func([]byte) (bool, error) { return false, nil })
	if !errors.Is(err, errNoDatagramMatched) {
		t.Fatalf("error: %v", err)
	}
	if time.Since(started) > 250*time.Millisecond {
		t.Fatal("rejected datagrams extended the deadline")
	}
}

type delayedDeadlineConn struct {
	net.Conn
	entered, release chan struct{}
}

func (c *delayedDeadlineConn) SetDeadline(d time.Time) error {
	err := c.Conn.SetDeadline(d)
	if !d.After(time.Now()) {
		close(c.entered)
		<-c.release
	}
	return err
}
func TestUDPExchangeJoinsCancellationCallback(t *testing.T) {
	c, received := newUDPTestPeer(t)
	addr := net.JoinHostPort(c.Host, fmt.Sprint(c.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	wrapped := &delayedDeadlineConn{conn, make(chan struct{}), make(chan struct{})}
	defer func() {
		select {
		case <-wrapped.release:
		default:
			close(wrapped.release)
		}
	}()
	c.conn = wrapped
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.Exchange(ctx, strings.NewReader("hold")); done <- err }()
	<-received
	cancel()
	<-wrapped.entered
	select {
	case err := <-done:
		t.Fatalf("exchange returned before its deadline callback: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(wrapped.release)
	if err := awaitUDPResult(t, done); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := c.Exchange(context.Background(), strings.NewReader("reply")); err != nil {
		t.Fatalf("next exchange: %v", err)
	}
}

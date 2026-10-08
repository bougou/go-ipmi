package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var udpRecvBufferSize = 4096
var udpReadTimeoutSeconds = 10

// errNoDatagramMatched is returned by ExchangeUntilMatch when the overall read deadline
// elapses without any inbound datagram for which the match callback returns true.
var errNoDatagramMatched = errors.New("udp exchange: no datagram matched before deadline")

// UDPClient exposes some common methods for communicating with UDP target addr.
type UDPClient struct {
	// Target Host
	Host string
	// Target Port
	Port int

	proxy      proxy.Dialer
	timeout    time.Duration
	bufferSize int

	// lock protects connection ownership, never network I/O.
	lock         sync.Mutex
	conn         net.Conn
	exchange     chan struct{}
	generation   uint64
	activeCancel context.CancelCauseFunc
}

func NewUDPClient(host string, port int) *UDPClient {
	udpClient := &UDPClient{
		Host:       host,
		Port:       port,
		bufferSize: udpRecvBufferSize,
		timeout:    time.Duration(udpReadTimeoutSeconds) * time.Second,
	}
	return udpClient
}

// dialUDP honors ContextDialer when available. A legacy Dialer cannot be
// interrupted, but a connection returned after cancellation must not leak.
func dialUDP(ctx context.Context, dialer proxy.Dialer, addr string) (net.Conn, error) {
	if dialer == nil {
		resolved, err := resolveUDPAddress(ctx, addr)
		if err != nil {
			return nil, err
		}
		return (&net.Dialer{}).DialContext(ctx, "udp", resolved)
	}
	if d, ok := dialer.(proxy.ContextDialer); ok {
		return d.DialContext(ctx, "udp", addr)
	}
	type result struct {
		conn net.Conn
		err  error
	}
	done := make(chan result)
	go func() {
		conn, err := dialer.Dial("udp", addr)
		select {
		case done <- result{conn, err}:
		case <-ctx.Done():
			if conn != nil {
				_ = conn.Close()
			}
		}
	}()
	select {
	case r := <-done:
		return r.conn, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// resolveUDPAddress preserves ResolveUDPAddr's IPv4 preference while allowing
// hostname resolution to be canceled. Literal IPv6 addresses retain their zone.
func resolveUDPAddress(ctx context.Context, addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	if _, err := netip.ParseAddr(host); err == nil || host == "" {
		return addr, nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", &net.DNSError{Err: "no suitable address found", Name: host, IsNotFound: true}
	}
	selected := addrs[0]
	for _, ip := range addrs {
		if ip.IP.To4() != nil {
			selected = ip
			break
		}
	}
	return net.JoinHostPort(selected.String(), port), nil
}

func (c *UDPClient) initConn(ctx context.Context, generation uint64) (net.Conn, error) {
	c.lock.Lock()
	conn, dialer := c.conn, c.proxy
	addr := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	c.lock.Unlock()
	if conn != nil {
		return conn, nil
	}
	conn, err := dialUDP(ctx, dialer, addr)
	if err != nil {
		return nil, fmt.Errorf("udp dial failed: %w", err)
	}
	c.lock.Lock()
	// Close advances generation before invoking cancellation outside the lock.
	if c.generation != generation || ctx.Err() != nil {
		c.lock.Unlock()
		_ = conn.Close()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, net.ErrClosed
	}
	c.conn = conn
	c.lock.Unlock()
	return conn, nil
}

func (c *UDPClient) SetProxy(proxy proxy.Dialer) *UDPClient {
	c.lock.Lock()
	c.proxy = proxy
	c.lock.Unlock()
	return c
}

func (c *UDPClient) SetTimeout(timeout time.Duration) *UDPClient {
	c.lock.Lock()
	c.timeout = timeout
	c.lock.Unlock()
	return c
}

func (c *UDPClient) SetBufferSize(bufferSize int) *UDPClient {
	c.lock.Lock()
	c.bufferSize = bufferSize
	c.lock.Unlock()
	return c
}

// RemoteIP returns the parsed ip address of the target.
func (c *UDPClient) RemoteIP() string {
	if net.ParseIP(c.Host) == nil {
		addrs, err := net.LookupHost(c.Host)
		if err == nil && len(addrs) > 0 {
			return addrs[0]
		}
	}
	return c.Host
}

func (c *UDPClient) LocalIPPort() (string, int) {
	conn, err := net.Dial("udp", net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	if err != nil {
		return "", 0
	}
	defer conn.Close()
	host, port, _ := net.SplitHostPort(conn.LocalAddr().String())
	p, _ := strconv.Atoi(port)
	return host, p
}

// Close interrupts the active exchange and invalidates calls already waiting
// for it. A later exchange may open a new connection.
func (c *UDPClient) Close() error {
	c.lock.Lock()
	conn, cancel := c.conn, c.activeCancel
	c.conn, c.activeCancel = nil, nil
	c.generation++
	c.lock.Unlock()
	if cancel != nil {
		cancel(net.ErrClosed)
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("close udp conn failed: %w", err)
		}
	}
	return nil
}

// Exchange sends the request read from reader and returns the first reply.
// It does not retry. Context cancellation interrupts socket I/O and waiting
// for another exchange. The reader must not block indefinitely.
func (c *UDPClient) Exchange(ctx context.Context, reader io.Reader) ([]byte, error) {
	return c.exchangeDatagrams(ctx, reader, nil)
}

// ExchangeUntilMatch sends the request, then discards datagrams until match
// returns true or an error. One deadline covers the entire operation, including
// writes and rejected datagrams: the earlier of ctx's deadline and the client
// timeout. With no matching reply it returns errNoDatagramMatched; caller
// cancellation preserves the context error. The reader and match callback run
// synchronously and must return promptly.
func (c *UDPClient) ExchangeUntilMatch(ctx context.Context, reader io.Reader, match func([]byte) (bool, error)) ([]byte, error) {
	return c.exchangeDatagrams(ctx, reader, match)
}

func (c *UDPClient) exchangeDatagrams(ctx context.Context, reader io.Reader, match func([]byte) (bool, error)) ([]byte, error) {
	c.lock.Lock()
	if c.exchange == nil {
		c.exchange = make(chan struct{}, 1)
	}
	gate, generation := c.exchange, c.generation
	c.lock.Unlock()
	select {
	case gate <- struct{}{}:
		defer func() { <-gate }()
	case <-ctx.Done():
		return nil, udpExchangeError(ctx, ctx.Err())
	}
	if err := ctx.Err(); err != nil {
		return nil, udpExchangeError(ctx, err)
	}
	opCtx, cancel := context.WithCancelCause(ctx)
	c.lock.Lock()
	if generation != c.generation {
		c.lock.Unlock()
		cancel(net.ErrClosed)
		return nil, net.ErrClosed
	}
	c.activeCancel = cancel
	timeout, bufferSize := c.timeout, c.bufferSize
	c.lock.Unlock()
	defer func() {
		cancel(nil)
		c.lock.Lock()
		if generation == c.generation {
			c.activeCancel = nil
		}
		c.lock.Unlock()
	}()
	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	dialCtx, dialCancel := context.WithDeadline(opCtx, deadline)
	conn, err := c.initConn(dialCtx, generation)
	dialCancel()
	if err != nil {
		return nil, udpExchangeError(opCtx, err)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, udpExchangeError(opCtx, err)
	}
	// Join an already-running callback before releasing the exchange slot: a
	// late deadline update must never interrupt the next exchange on this socket.
	interrupted := make(chan struct{})
	stop := context.AfterFunc(opCtx, func() {
		defer close(interrupted)
		_ = conn.SetDeadline(time.Now())
	})
	defer func() {
		if !stop() {
			<-interrupted
		}
	}()
	if err := opCtx.Err(); err != nil {
		return nil, udpExchangeError(opCtx, err)
	}
	if _, err := io.Copy(conn, reader); err != nil {
		return nil, fmt.Errorf("write to conn failed: %w", udpExchangeError(opCtx, err))
	}
	recvBuffer := make([]byte, bufferSize)
	for {
		if err := opCtx.Err(); err != nil {
			return nil, udpExchangeError(opCtx, err)
		}
		n, err := conn.Read(recvBuffer)
		if err != nil {
			err = udpExchangeError(opCtx, err)
			var ne net.Error
			if match != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) && errors.As(err, &ne) && ne.Timeout() {
				return nil, errNoDatagramMatched
			}
			return nil, fmt.Errorf("read from conn failed: %w", err)
		}
		if err := opCtx.Err(); err != nil {
			return nil, udpExchangeError(opCtx, err)
		}
		if match == nil {
			return recvBuffer[:n], nil
		}
		recv := append([]byte(nil), recvBuffer[:n]...)
		ok, err := match(recv)
		if err != nil {
			return nil, err
		}
		if ok {
			return recv, nil
		}
		if !time.Now().Before(deadline) {
			return nil, udpExchangeError(opCtx, errNoDatagramMatched)
		}
	}
}

func udpExchangeError(ctx context.Context, err error) error {
	if cause := context.Cause(ctx); cause != nil {
		if errors.Is(cause, ctx.Err()) {
			return cause
		}
		return errors.Join(ctx.Err(), cause)
	}
	// A socket deadline can fire just before the context timer is scheduled.
	if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}
	return err
}

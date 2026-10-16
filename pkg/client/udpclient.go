package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
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

	// Receive-loop state (SOL sessions). While a loop runs, it owns socket
	// reads: exchanges register recvWaiter instead of reading, and datagrams
	// the waiter does not claim go to recvFn. recvMu serializes waiter
	// registration against loop delivery.
	recvMu     sync.Mutex
	recvWaiter *udpWaiter
	recvFn     func([]byte)
	recvStop   chan struct{}
	recvDone   chan struct{}
}

// udpWaiter receives the datagram an in-flight exchange is waiting for while
// the receive loop owns socket reads.
type udpWaiter struct {
	match func([]byte) (bool, error)
	ch    chan udpDelivery
}

type udpDelivery struct {
	data     []byte
	matchErr error
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
	// The receive loop exits when its read on the closed connection fails;
	// a fresh StartRecvLoop must not see its state.
	c.recvMu.Lock()
	c.recvFn = nil
	c.recvMu.Unlock()
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

// StartRecvLoop starts a goroutine that owns all socket reads until
// [UDPClient.StopRecvLoop]: datagrams claimed by an in-flight exchange's
// waiter go to that exchange, everything else goes to fn. While the loop runs,
// [UDPClient.Exchange]/[UDPClient.ExchangeUntilMatch] never read the socket
// directly, so asynchronous traffic (SOL output pushes, §15.3) is consumed
// immediately instead of only while an exchange happens to be in flight.
//
// Starting the loop acquires the exchange gate, so it never races an exchange
// already blocked in Read: datagrams buffered by the kernel stay readable and
// are delivered to fn once the loop takes over.
func (c *UDPClient) StartRecvLoop(ctx context.Context, fn func([]byte)) error {
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
		return udpExchangeError(ctx, ctx.Err())
	}
	c.recvMu.Lock()
	if c.recvFn != nil {
		c.recvMu.Unlock()
		return errors.New("udp recv loop already running")
	}
	c.recvMu.Unlock()
	conn, err := c.initConn(ctx, generation)
	if err != nil {
		return err
	}
	stop, done := make(chan struct{}), make(chan struct{})
	c.recvMu.Lock()
	c.recvFn, c.recvStop, c.recvDone = fn, stop, done
	c.recvMu.Unlock()
	go c.recvLoop(conn, stop, done, fn)
	return nil
}

// StopRecvLoop stops the receive loop and returns the socket to
// exchange-driven reads. Safe to call when no loop is running.
func (c *UDPClient) StopRecvLoop() {
	c.lock.Lock()
	gate := c.exchange
	c.lock.Unlock()
	if gate == nil {
		return
	}
	// Holding the gate waits out any in-flight waiter exchange: after it
	// returns, no waiter or loop reader remains on the socket.
	gate <- struct{}{}
	defer func() { <-gate }()
	c.recvMu.Lock()
	if c.recvFn == nil {
		c.recvMu.Unlock()
		return
	}
	stop, done := c.recvStop, c.recvDone
	close(stop)
	c.recvMu.Unlock()
	<-done
	c.recvMu.Lock()
	c.recvFn, c.recvStop, c.recvDone, c.recvWaiter = nil, nil, nil, nil
	c.recvMu.Unlock()
}

func (c *UDPClient) recvLoop(conn net.Conn, stop chan struct{}, done chan<- struct{}, fn func([]byte)) {
	defer close(done)
	c.lock.Lock()
	buf := make([]byte, c.bufferSize)
	c.lock.Unlock()
	for {
		// Poll the stop channel on a short read tick; the connection stays
		// open so exchanges can resume direct reads after StopRecvLoop.
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := conn.Read(buf)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				select {
				case <-stop:
					return
				default:
					continue
				}
			}
			return
		}
		data := append([]byte(nil), buf[:n]...)
		var claimed chan udpDelivery
		var delivery udpDelivery
		c.recvMu.Lock()
		w := c.recvWaiter
		if w != nil {
			ok, merr := true, error(nil)
			if w.match != nil {
				ok, merr = w.match(data)
			}
			if ok || merr != nil {
				c.recvWaiter = nil
				claimed, delivery = w.ch, udpDelivery{data: data, matchErr: merr}
			}
		}
		c.recvMu.Unlock()
		// Dispatch before waking the exchange. A claimed datagram can also
		// carry asynchronous SOL output (§15.11); the exchange must not
		// return until that output is queued, or an EOF right after the
		// keystroke drops it.
		fn(data)
		if claimed != nil {
			select {
			case claimed <- delivery:
			default: // the exchange already timed out; drop
			}
		}
	}
}

// exchangeViaLoop is the receive-loop counterpart of the direct read loop:
// register the waiter, write the request, wait for the loop to deliver a
// matching datagram. The socket itself is never read or deadline-interrupted
// here — the loop owns reads.
func (c *UDPClient) exchangeViaLoop(ctx context.Context, conn net.Conn, reader io.Reader, match func([]byte) (bool, error), deadline time.Time) ([]byte, error) {
	w := &udpWaiter{match: match, ch: make(chan udpDelivery, 1)}
	c.recvMu.Lock()
	c.recvWaiter = w
	c.recvMu.Unlock()
	defer func() {
		c.recvMu.Lock()
		if c.recvWaiter == w {
			c.recvWaiter = nil
		}
		c.recvMu.Unlock()
	}()
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return nil, udpExchangeError(ctx, err)
	}
	if _, err := io.Copy(conn, reader); err != nil {
		return nil, fmt.Errorf("write to conn failed: %w", udpExchangeError(ctx, err))
	}
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case d := <-w.ch:
		if d.matchErr != nil {
			return nil, d.matchErr
		}
		return d.data, nil
	case <-ctx.Done():
		return nil, udpExchangeError(ctx, ctx.Err())
	case <-timer.C:
		if match != nil {
			return nil, udpExchangeError(ctx, errNoDatagramMatched)
		}
		return nil, fmt.Errorf("read from conn failed: %w", udpExchangeError(ctx, os.ErrDeadlineExceeded))
	}
}

// Send writes a single datagram without waiting for a response. Used for
// ACK-only SOL packets (§15.11: sequence 0h packets are not acknowledged).
// Sending shares the socket with the receive loop and with a possible
// in-flight exchange: UDP writes are datagram-atomic and Send never reads,
// so no gate is needed.
func (c *UDPClient) Send(ctx context.Context, reader io.Reader) error {
	c.lock.Lock()
	generation, timeout := c.generation, c.timeout
	c.lock.Unlock()
	conn, err := c.initConn(ctx, generation)
	if err != nil {
		return udpExchangeError(ctx, err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return udpExchangeError(ctx, err)
	}
	if _, err := io.Copy(conn, reader); err != nil {
		return fmt.Errorf("write to conn failed: %w", udpExchangeError(ctx, err))
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
	// While a receive loop owns socket reads, hand the response matching to
	// the loop's waiter instead of reading here.
	c.recvMu.Lock()
	loopActive := c.recvFn != nil
	c.recvMu.Unlock()
	if loopActive {
		return c.exchangeViaLoop(opCtx, conn, reader, match, deadline)
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

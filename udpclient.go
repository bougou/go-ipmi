package ipmi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
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

	conn   net.Conn
	closed bool

	// lock is used to protect udp Exchange method to prevent another
	// send/receive operation from occurring while one is in progress.
	lock sync.Mutex
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

func (c *UDPClient) initConn() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn != nil && !c.closed {
		return nil
	}

	if c.proxy != nil {
		conn, err := c.proxy.Dial("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
		if err != nil {
			return fmt.Errorf("udp proxy dial failed, err: %w", err)
		}
		c.conn = conn
		return nil
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return fmt.Errorf("resolve addr failed, err: %w", err)
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return fmt.Errorf("udp dial failed, err: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *UDPClient) SetProxy(proxy proxy.Dialer) *UDPClient {
	c.proxy = proxy
	return c
}

func (c *UDPClient) SetTimeout(timeout time.Duration) *UDPClient {
	c.timeout = timeout
	return c
}

func (c *UDPClient) SetBufferSize(bufferSize int) *UDPClient {
	c.bufferSize = bufferSize
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

func (c *UDPClient) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn == nil {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close udp conn failed, err: %w", err)
	}

	c.conn = nil
	c.closed = true
	return nil
}

// Exchange sends the payload read from reader as a single UDP datagram to the client's host:port,
// then waits for exactly one inbound datagram and returns its payload. The read uses the client's
// configured timeout and, if set, the earlier of ctx’s deadline and that timeout. It does not
// retry on failure. Cancellation is honored via ctx.
func (c *UDPClient) Exchange(ctx context.Context, reader io.Reader) ([]byte, error) {
	if err := c.initConn(); err != nil {
		return nil, fmt.Errorf("init udp connection failed, err: %w", err)
	}

	recvBuffer := make([]byte, c.bufferSize)

	// Use a single goroutine to handle the entire exchange operation
	// This ensures proper context cancellation and resource cleanup
	resultChan := make(chan struct {
		data []byte
		err  error
	}, 1)

	go func() {
		defer close(resultChan)

		c.lock.Lock()
		defer c.lock.Unlock()

		// Step 1: Check if context is already cancelled
		if ctx.Err() != nil {
			return
		}

		// Step 2: Send the request
		_, err := io.Copy(c.conn, reader)
		if err != nil {
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("write to conn failed, err: %w", err)}
			return
		}

		// Step 3: Check context after write
		if ctx.Err() != nil {
			return
		}

		// Step 4: Set read deadline
		// Use context deadline if available, otherwise use configured timeout
		deadline := time.Now().Add(c.timeout)
		if ctxDeadline, ok := ctx.Deadline(); ok {
			// Use the earlier deadline between context and configured timeout
			if ctxDeadline.Before(deadline) {
				deadline = ctxDeadline
			}
		}
		err = c.conn.SetReadDeadline(deadline)
		if err != nil {
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("set conn read deadline failed, err: %w", err)}
			return
		}

		// Step 5: Read the response
		nRead, err := c.conn.Read(recvBuffer)

		// Step 6: Check context after read (in case context was cancelled during read)
		if ctx.Err() != nil {
			return
		}

		if err != nil {
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("read from conn failed, err: %w", err)}
			return
		}

		// Step 7: Return the response data
		resultChan <- struct {
			data []byte
			err  error
		}{recvBuffer[:nRead], nil}
	}()

	// Wait for the result or context cancellation
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("canceled from caller: %w", ctx.Err())
	case result, ok := <-resultChan:
		if ok {
			return result.data, result.err
		}

		if ctx.Err() != nil {
			return nil, fmt.Errorf("canceled from caller: %w", ctx.Err())
		}
		return nil, fmt.Errorf("result channel closed")
	}
}

// ExchangeUntilMatch sends the payload read from reader as a single UDP datagram, then reads
// repeatedly until match returns true for some received datagram, or until the overall deadline
// elapses (the same deadline rules as Exchange: min of context deadline and client timeout).
// Datagrams for which match returns (false, nil) are discarded and reading continues.
//
// Unlike Exchange, which returns the first datagram unconditionally, ExchangeUntilMatch filters
// inbound traffic: stray packets, unrelated replies, or multiplexed traffic on the same socket can
// be skipped by returning (false, nil) from match. If the deadline passes with no matching
// datagram, the returned error indicates that no datagram satisfied match before the deadline.
//
// If match returns a non-nil error, that error is returned immediately. When match returns
// (true, nil), the corresponding datagram's payload is returned.
func (c *UDPClient) ExchangeUntilMatch(ctx context.Context, reader io.Reader, match func(recv []byte) (ok bool, err error)) ([]byte, error) {
	if err := c.initConn(); err != nil {
		return nil, fmt.Errorf("init udp connection failed, err: %w", err)
	}

	recvBuffer := make([]byte, c.bufferSize)
	resultChan := make(chan struct {
		data []byte
		err  error
	}, 1)

	go func() {
		defer close(resultChan)

		c.lock.Lock()
		defer c.lock.Unlock()

		if ctx.Err() != nil {
			return
		}

		_, err := io.Copy(c.conn, reader)
		if err != nil {
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("write to conn failed, err: %w", err)}
			return
		}

		if ctx.Err() != nil {
			return
		}

		deadline := time.Now().Add(c.timeout)
		if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
			deadline = ctxDeadline
		}

		for {
			if ctx.Err() != nil {
				return
			}

			remaining := time.Until(deadline)
			if remaining <= 0 {
				resultChan <- struct {
					data []byte
					err  error
				}{nil, errNoDatagramMatched}
				return
			}

			readDur := remaining
			if c.timeout > 0 && c.timeout < remaining {
				readDur = c.timeout
			}

			err := c.conn.SetReadDeadline(time.Now().Add(readDur))
			if err != nil {
				resultChan <- struct {
					data []byte
					err  error
				}{nil, fmt.Errorf("set conn read deadline failed, err: %w", err)}
				return
			}

			nRead, err := c.conn.Read(recvBuffer)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				resultChan <- struct {
					data []byte
					err  error
				}{nil, fmt.Errorf("read from conn failed, err: %w", err)}
				return
			}

			recv := append([]byte(nil), recvBuffer[:nRead]...)
			ok, mErr := match(recv)
			if mErr != nil {
				resultChan <- struct {
					data []byte
					err  error
				}{nil, mErr}
				return
			}
			if ok {
				resultChan <- struct {
					data []byte
					err  error
				}{recv, nil}
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("canceled from caller: %w", ctx.Err())
	case result, ok := <-resultChan:
		if ok {
			return result.data, result.err
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("canceled from caller: %w", ctx.Err())
		}
		return nil, fmt.Errorf("result channel closed")
	}
}

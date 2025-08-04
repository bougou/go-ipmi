package ipmi

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// UDPClient exposes some common methods for communicating with UDP target addr.
type UDPClient struct {
	// Target Host
	Host string
	// Target Port
	Port int

	proxy      proxy.Dialer
	timeout    time.Duration
	bufferSize int

	conn net.Conn

	// lock is used to protect udp Exchange method to prevent another
	// send/receive operation from occurring while one is in progress.
	lock sync.Mutex
}

func NewUDPClient(host string, port int) *UDPClient {
	udpClient := &UDPClient{
		Host: host,
		Port: port,
	}
	return udpClient
}

func (c *UDPClient) initConn() error {
	if c.conn != nil {
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
	if c.conn == nil {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close udp conn failed, err: %w", err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.conn = nil
	return nil
}

// Exchange performs a synchronous UDP query.
// It sends the request, and waits for a reply.
// Exchange does not retry a failed query.
// The sent content is read from reader.
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
		c.lock.Lock()
		defer c.lock.Unlock()

		// Step 1: Check if context is already cancelled
		select {
		case <-ctx.Done():
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("context cancelled before starting operations: %w", ctx.Err())}
			return
		default:
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
		select {
		case <-ctx.Done():
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("context cancelled after write: %w", ctx.Err())}
			return
		default:
		}

		// Step 4: Set read deadline if timeout is configured
		if c.timeout > 0 {
			deadline := time.Now().Add(c.timeout)
			err = c.conn.SetReadDeadline(deadline)
			if err != nil {
				resultChan <- struct {
					data []byte
					err  error
				}{nil, fmt.Errorf("set conn read deadline failed with read timeout %v, err: %w", c.timeout, err)}
				return
			}
		}

		// Step 5: Read the response
		nRead, err := c.conn.Read(recvBuffer)
		if err != nil {
			resultChan <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("read from conn failed with read timeout %v, err: %w", c.timeout, err)}
			return
		}

		// Step 6: Return the response data
		resultChan <- struct {
			data []byte
			err  error
		}{recvBuffer[:nRead], nil}
	}()

	// Wait for the result or context cancellation
	select {
	case <-ctx.Done():
		// Try to close the connection to unblock any ongoing operations
		if c.conn != nil {
			c.conn.Close()
		}
		return nil, fmt.Errorf("canceled from caller: %w", ctx.Err())
	case result := <-resultChan:
		return result.data, result.err
	}
}

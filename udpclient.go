package ipmi

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
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
			return fmt.Errorf("udp proxy dial failed, err: %s", err)
		}
		c.conn = conn
		return nil
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return fmt.Errorf("resolve addr failed, err: %s", err)
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return fmt.Errorf("udp dial failed, err: %s", err)
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
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
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
		return fmt.Errorf("close udp conn failed, err: %s", err)
	}

	c.conn = nil
	return nil
}

// Exchange performs a synchronous UDP query.
// It sends the request, and waits for a reply.
// Exchange does not retry a failed query.
// The sent content is read from reader.
func (c *UDPClient) Exchange(ctx context.Context, reader io.Reader) ([]byte, error) {
	if err := c.initConn(); err != nil {
		return nil, fmt.Errorf("init udp connection failed, err: %s", err)
	}

	recvBuffer := make([]byte, c.bufferSize)

	doneChan := make(chan error, 1)
	// recvChan stores the integer number which indicates how many bytes
	recvChan := make(chan int, 1)
	go func() {
		// It is possible that this action blocks, although this
		// should only occur in very resource-intensive situations:
		// - when you've filled up the socket buffer and the OS
		//   can't dequeue the queue fast enough.
		_, err := io.Copy(c.conn, reader)
		if err != nil {
			doneChan <- fmt.Errorf("write to conn failed, err: %w", err)
			return
		}

		// Set a deadline for the Read operation so that we don't
		// wait forever for a server that might not respond on
		// a reasonable amount of time.
		deadline := time.Now().Add(c.timeout)
		err = c.conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- fmt.Errorf("set conn read deadline failed, err: %w", err)
			return
		}

		nRead, err := c.conn.Read(recvBuffer)
		if err != nil {
			doneChan <- fmt.Errorf("read from conn failed, err: %w", err)
			return
		}

		doneChan <- nil
		recvChan <- nRead
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("canceled from caller")
	case err := <-doneChan:
		if err != nil {
			return nil, err
		}
		recvCount := <-recvChan
		return recvBuffer[:recvCount], nil
	}
}

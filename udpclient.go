package ipmi

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// UDPClient exposes some common methods for communicating with UDP target addr.
type UDPClient struct {
	Host string
	Port int

	timeout    time.Duration
	bufferSize int
}

func (c *UDPClient) RemoteIP() string {
	if net.ParseIP(c.Host) == nil {
		addrs, err := net.LookupHost(c.Host)
		if err == nil && len(addrs) > 0 {
			return addrs[0]
		}
	}
	return c.Host
}

func (c *UDPClient) LocalIP() string {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return c.Host
	}
	defer conn.Close()
	host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	return host
}

func (c *UDPClient) Write(data []byte) error {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return fmt.Errorf("dial failed, err: %s", err)
	}
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write failed, err: %s", err)
	}
	return nil
}

// Exchange performs a synchronous UDP query.
// It sends the request, and waits for a reply.
// Exchange does not retry a failed query.
// The sent content is read from reader.
func (c *UDPClient) Exchanged(ctx context.Context, reader io.Reader) ([]byte, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, fmt.Errorf("resolve addr failed, err: %s", err)
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("dial failed, err: %s", err)
	}
	defer conn.Close()

	buffer := make([]byte, c.bufferSize)
	doneChan := make(chan error, 1)
	recvChan := make(chan int, 1)
	go func() {
		// It is possible that this action blocks, although this
		// should only occur in very resource-intensive situations:
		// - when you've filled up the socket buffer and the OS
		//   can't dequeue the queue fast enough.
		_, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- fmt.Errorf("write to conn failed, err: %s", err)
			return
		}

		// Set a deadline for the ReadOperation so that we don't
		// wait forever for a server that might not respond on
		// a resonable amount of time.
		deadline := time.Now().Add(c.timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- fmt.Errorf("set conn read deadline failed, err: %s", err)
			return
		}

		nRead, _, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- fmt.Errorf("read from conn failed, err: %s", err)
			return
		}

		doneChan <- nil
		recvChan <- nRead
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("canceled from caller")
	case err = <-doneChan:
		if err != nil {
			return nil, err
		}
		recvCount := <-recvChan
		return buffer[:recvCount], nil
	}
}

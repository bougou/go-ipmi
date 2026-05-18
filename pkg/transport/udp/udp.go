// Package udp provides a [transport.PacketConn] backed by a standard UDP socket.
//
// It is the reference transport for LAN-attached BMCs.  Embedded targets or
// test code should implement [transport.PacketConn] directly rather than
// depending on this package.
package udp

import (
	"fmt"
	"net"
	"time"
)

const (
	DefaultPort        = 623
	DefaultBufferSize  = 4096
	DefaultReadTimeout = 30 * time.Second
)

// Conn wraps a [net.UDPConn] and satisfies [transport.PacketConn].
type Conn struct {
	conn        *net.UDPConn
	readTimeout time.Duration
}

// Option configures a [Conn].
type Option func(*Conn)

// WithReadTimeout overrides the per-read deadline (default 30 s).
func WithReadTimeout(d time.Duration) Option {
	return func(c *Conn) { c.readTimeout = d }
}

// Listen binds a UDP socket on addr (e.g. ":623") and returns a [Conn].
func Listen(addr string, opts ...Option) (*Conn, error) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", addr, err)
	}
	conn, err := net.ListenUDP("udp", ua)
	if err != nil {
		return nil, fmt.Errorf("listen udp %q: %w", addr, err)
	}
	c := &Conn{conn: conn, readTimeout: DefaultReadTimeout}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Wrap adapts an existing [net.UDPConn] so it satisfies [transport.PacketConn].
// This lets callers attach the server to a socket they already own.
func Wrap(conn *net.UDPConn, opts ...Option) *Conn {
	c := &Conn{conn: conn, readTimeout: DefaultReadTimeout}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Conn) ReadFrom(buf []byte) (int, net.Addr, error) {
	if c.readTimeout > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
			return 0, nil, fmt.Errorf("set read deadline: %w", err)
		}
	}
	n, addr, err := c.conn.ReadFromUDP(buf)
	return n, addr, err
}

func (c *Conn) WriteTo(data []byte, addr net.Addr) (int, error) {
	ua, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, fmt.Errorf("expected *net.UDPAddr, got %T", addr)
	}
	return c.conn.WriteToUDP(data, ua)
}

func (c *Conn) Close() error { return c.conn.Close() }

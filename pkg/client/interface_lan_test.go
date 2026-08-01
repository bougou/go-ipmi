package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

// closeTrackingConn provides the minimal net.Conn behavior needed by closeLAN tests.
type closeTrackingConn struct {
	writeErr error
	closed   bool
}

func (*closeTrackingConn) Read([]byte) (int, error)         { return 0, errors.New("unexpected read") }
func (c *closeTrackingConn) Write([]byte) (int, error)      { return 0, c.writeErr }
func (c *closeTrackingConn) Close() error                   { c.closed = true; return nil }
func (*closeTrackingConn) LocalAddr() net.Addr              { return &net.UDPAddr{} }
func (*closeTrackingConn) RemoteAddr() net.Addr             { return &net.UDPAddr{} }
func (*closeTrackingConn) SetDeadline(time.Time) error      { return nil }
func (*closeTrackingConn) SetReadDeadline(time.Time) error  { return nil }
func (*closeTrackingConn) SetWriteDeadline(time.Time) error { return nil }

func TestCloseConnectionWhenCloseSessionFails(t *testing.T) {
	writeErr := errors.New("BMC is unreachable")
	conn := &closeTrackingConn{writeErr: writeErr}
	client, err := NewClient("127.0.0.1", 623, "user", "password")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.WithInterface(InterfaceLan)
	client.v20 = false
	client.udpClient.conn = conn

	err = client.Close(context.Background())
	if !errors.Is(err, writeErr) {
		t.Fatalf("Close() error = %v, want wrapped CloseSession error", err)
	}
	if !conn.closed {
		t.Fatal("UDP connection remained open after CloseSession failure")
	}
}

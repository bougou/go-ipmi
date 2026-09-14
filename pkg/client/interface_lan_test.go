package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
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

func TestCloseLAN(t *testing.T) {
	for _, tt := range []struct {
		name      string
		lanplus   bool
		session   session
		wantClose bool
	}{
		{name: "LAN before activation"},
		{name: "LAN temporary session", session: session{v15: v15{preSession: true, sessionID: 1}}},
		{name: "LAN active", session: session{v15: v15{active: true, sessionID: 1}}, wantClose: true},
		{name: "LANplus before activation", lanplus: true},
		{name: "LANplus authentication incomplete", lanplus: true,
			session: session{v20: v20{state: types.SessionStateRakp2Received, bmcSessionID: 1}}},
		{name: "LANplus active", lanplus: true,
			session: session{v20: v20{state: types.SessionStateActive, bmcSessionID: 1}}, wantClose: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			writeErr := errors.New("BMC is unreachable")
			conn := &closeTrackingConn{writeErr: writeErr}
			client, err := NewClient("127.0.0.1", 623, "user", "password")
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if !tt.lanplus {
				client.WithInterface(InterfaceLan)
			}
			client.v20 = tt.lanplus
			client.session = &tt.session
			client.udpClient.conn = conn

			err = client.Close(context.Background())
			if tt.wantClose {
				if !errors.Is(err, writeErr) {
					t.Fatalf("Close() error = %v, want wrapped CloseSession error", err)
				}
			} else if err != nil {
				t.Fatalf("Close() error = %v, want no CloseSession attempt", err)
			}
			if !conn.closed {
				t.Fatal("UDP connection remained open after Close")
			}
		})
	}
}

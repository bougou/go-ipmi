package client

// End-to-end Get Session Info test over a real IPMI v1.5 (-A MD5) session,
// proving the keepalive the client fires periodically also succeeds on the
// v1.5 path and describes the caller's own session.

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestGetCurrentSessionInfoV15(t *testing.T) {
	const (
		username = "ADMIN"
		password = "ADMIN"
	)

	b := newV15TestBMC(t, username, password)

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	conn := udp.Wrap(pc)
	addr := pc.LocalAddr().(*net.UDPAddr) //nolint:forcetypeassert
	srv := server.NewServer(b, conn)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()

	c, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLan).
		WithTimeout(2 * time.Second).
		WithRetry(0)
	c.session.authType = types.AuthTypeMD5

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect (v1.5): %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	resp, err := c.GetCurrentSessionInfo(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentSessionInfo: %v", err)
	}

	if resp.SessionHandle == 0 {
		t.Errorf("session handle must be non-zero")
	}
	// newV15TestBMC seeds the user in slot 2 on the LAN channel (1); the session
	// negotiates as IPMI v1.5, so the auxiliary-data nibble is 0.
	if resp.UserID != 2 {
		t.Errorf("user id = %d, want 2", resp.UserID)
	}
	if resp.AuxiliaryData != 0 {
		t.Errorf("auxiliary data = %d, want 0 (IPMI v1.5)", resp.AuxiliaryData)
	}
	if resp.ChannelNumber != 1 {
		t.Errorf("channel number = %d, want 1", resp.ChannelNumber)
	}
	if want := uint8(b.V15Sessions.Cap()); resp.PossibleActiveSessions != want {
		t.Errorf("possible active sessions = %d, want %d", resp.PossibleActiveSessions, want)
	}
	if resp.CurrentActiveSessions < 1 {
		t.Errorf("current active sessions = %d, want >= 1", resp.CurrentActiveSessions)
	}
}

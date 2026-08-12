package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

// TestV15Reject20BytePasswordUnauthenticated is a raw-packet regression test
// for the security-critical property of the 20-byte-password rejection: the
// refusal response must NOT be authenticated, because an authenticated v1.5
// response computes its AuthCode from the 16-byte truncation of the secret
// (the straight-password AuthCode IS those bytes; MD2/MD5 make it an offline
// verifier). It drives a real Activate Session for a 20-byte-password account
// and asserts the reply is AuthType=None with no AuthCode and completion code
// 0x85. Flipping the rejection back to an authenticated response would restore
// the leak and fail here.
func TestV15Reject20BytePasswordUnauthenticated(t *testing.T) {
	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, mock.New(), bmc.WithClock(clock.Real))
	user, err := b.Users.Add(2, "op20")
	if err != nil {
		t.Fatal(err)
	}
	user.SetPassword([]byte("passwordlongerthan16")) // 20 bytes -> Password20
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	// Pre-create the pending session directly, standing in for a completed Get
	// Session Challenge (which now accepts the valid name). The gate fires
	// before AuthCode verification, so the Activate packet needs no valid
	// crypto.
	var challenge [16]byte
	sess, err := b.V15Sessions.CreatePending(bmc.V15AuthTypeMD5, user, challenge, 1)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	srv := NewServer(b, udp.Wrap(pc))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()

	// Activate Session IPMI request (App 0x3A). IPMB request framing:
	// rsAddr, netFn<<2, csum1, rqAddr, rqSeq<<2, cmd, data..., csum2.
	req := []byte{0x20, handlers.NetFnAppRequest << 2, 0x00, 0x81, 0x00, handlers.CmdActivateSession}
	activateData := make([]byte, 22)
	activateData[0] = uint8(bmc.V15AuthTypeMD5)
	activateData[1] = uint8(bmc.PrivilegeLevelAdministrator)
	req = append(req, activateData...)
	req = append(req, 0x00) // trailing checksum (server does not verify it)

	hdr := types.SessionHeader15{
		AuthType:      types.AuthTypeMD5,
		Sequence:      0,
		SessionID:     sess.TempSessionID,
		AuthCode:      make([]byte, 16), // bogus: rejected before it is checked
		PayloadLength: uint8(len(req)),
	}
	v15 := append(hdr.Pack(), req...)
	pkt := append([]byte{0x06, 0x00, 0xff, 0x07}, v15...) // RMCP header + v1.5 session

	client, err := net.DialUDP("udp", nil, pc.LocalAddr().(*net.UDPAddr)) //nolint:forcetypeassert
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if _, err := client.Write(pkt); err != nil {
		t.Fatal(err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 512)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("no rejection response (must reply, unauthenticated): %v", err)
	}

	var resp types.Session15
	if err := resp.Unpack(buf[4:n]); err != nil { // strip RMCP header
		t.Fatalf("unpack v1.5 response: %v", err)
	}
	if resp.SessionHeader15.AuthType != types.AuthTypeNone {
		t.Errorf("response AuthType = 0x%02x, want None (0x00) — an authenticated reply leaks the truncated secret", uint8(resp.SessionHeader15.AuthType))
	}
	if len(resp.SessionHeader15.AuthCode) != 0 {
		t.Errorf("response carries a %d-byte AuthCode, want none", len(resp.SessionHeader15.AuthCode))
	}
	if _, _, data, _, ok := protocol.ParseIPMIRequest(resp.Payload); !ok || len(data) < 0 {
		t.Fatalf("unparseable response payload")
	}
	// Completion code is the byte after rsAddr/netfn/csum/rqAddr/rqSeq/cmd.
	if len(resp.Payload) < 7 || resp.Payload[6] != uint8(handlers.CCV15InvalidSessionID) {
		t.Errorf("completion code = 0x%02x, want 0x85", resp.Payload[6])
	}
}

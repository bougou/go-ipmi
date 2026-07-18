package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// TestServerV15SessionActivation verifies the reference BMC server accepts a
// full IPMI v1.5 session handshake (Get Channel Auth Caps → Get Session
// Challenge → Activate Session) and executes commands over the authenticated
// session. This mirrors ipmitool -I lan -A MD5.
func TestServerV15SessionActivation(t *testing.T) {
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
	addr := pc.LocalAddr().(*net.UDPAddr)
	srv := server.NewServer(b, conn)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = srv.Serve(ctx)
	}()

	c, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLan).
		WithTimeout(2 * time.Second).
		WithRetry(0)
	c.session.authType = ipmi.AuthTypeMD5

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect (v1.5): %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	resp, err := c.GetDeviceID(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceID: %v", err)
	}
	if resp.DeviceID == 0 {
		t.Fatalf("unexpected device ID: %d", resp.DeviceID)
	}
}

func TestServerV15SessionActivationMD2(t *testing.T) {
	const (
		username = "ADMIN"
		password = "ADMIN"
	)

	b := newV15TestBMC(t, username, password)
	bmc.WithV15AuthTypes([]bmc.V15AuthType{bmc.V15AuthTypeMD2})(b)

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	conn := udp.Wrap(pc)
	addr := pc.LocalAddr().(*net.UDPAddr)
	srv := server.NewServer(b, conn)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = srv.Serve(ctx)
	}()

	c, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLan).
		WithTimeout(2 * time.Second).
		WithRetry(0)

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect (v1.5 MD2): %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	if c.session.authType != ipmi.AuthTypeMD2 {
		t.Fatalf("want MD2 auth type, got %v", c.session.authType)
	}

	resp, err := c.GetDeviceID(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceID: %v", err)
	}
	if resp.DeviceID == 0 {
		t.Fatalf("unexpected device ID: %d", resp.DeviceID)
	}
}

// TestV15ClientActivateSeedsInboundSeq locks the client half of the Activate
// Session inbound-seq contract (spec §18.15 / §6.12.9): after Activate returns
// starting seq N, inSeq must be N-1 so the pre-increment on the next send
// emits N. Against a correct BMC the sliding window often still accepts N+1,
// so client-e2e alone may not catch a regression — this assertion does.
func TestV15ClientActivateSeedsInboundSeq(t *testing.T) {
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
	addr := pc.LocalAddr().(*net.UDPAddr)
	srv := server.NewServer(b, conn)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = srv.Serve(ctx)
	}()

	c, err := NewClient(addr.IP.String(), addr.Port, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLan).
		WithTimeout(2 * time.Second).
		WithRetry(0)
	c.v20 = false
	c.session.authType = ipmi.AuthTypeMD5
	c.maxPrivilegeLevel = ipmi.PrivilegeLevelAdministrator

	bg := context.Background()
	if _, err := c.GetChannelAuthenticationCapabilities(bg, ipmi.ChannelNumberSelf, c.maxPrivilegeLevel); err != nil {
		t.Fatalf("GetChannelAuthenticationCapabilities: %v", err)
	}
	if _, err := c.GetSessionChallenge(bg); err != nil {
		t.Fatalf("GetSessionChallenge: %v", err)
	}
	c.session.v15.preSession = true

	activate, err := c.ActivateSession(bg)
	if err != nil {
		t.Fatalf("ActivateSession: %v", err)
	}
	wantSeed := activate.InitialInboundSequenceNumber - 1
	if c.session.v15.inSeq != wantSeed {
		t.Fatalf("after Activate: inSeq=0x%08x, want 0x%08x (returned N-1; N=0x%08x)",
			c.session.v15.inSeq, wantSeed, activate.InitialInboundSequenceNumber)
	}

	if _, err := c.SetSessionPrivilegeLevel(bg, c.maxPrivilegeLevel); err != nil {
		t.Fatalf("SetSessionPrivilegeLevel with starting inbound seq: %v", err)
	}
	if c.session.v15.inSeq != activate.InitialInboundSequenceNumber {
		t.Fatalf("after first post-Activate cmd: inSeq=0x%08x, want returned N 0x%08x",
			c.session.v15.inSeq, activate.InitialInboundSequenceNumber)
	}
}

func newV15TestBMC(t *testing.T, username, password string) *bmc.BMC {
	t.Helper()
	info := bmc.DeviceInfo{
		DeviceID:                32,
		DeviceRevision:          1,
		FirmwareMajor:           1,
		FirmwareMinor:           0,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x3D,
	}
	var guid [16]byte
	copy(guid[:], "go-ipmi-v15\x00\x00\x00\x00\x00")
	b := bmc.New(info, guid, mock.New(), bmc.WithClock(clock.Real))

	user, err := b.Users.Add(2, username)
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte(password))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	return b
}

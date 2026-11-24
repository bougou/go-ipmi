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
	"github.com/bougou/go-ipmi/pkg/types"
)

// newCipherSuiteTestServer starts the reference BMC offering suites, with
// user 2 ADMIN/ADMIN, and returns its address.
func newCipherSuiteTestServer(t *testing.T, suites []types.CipherSuiteID) *net.UDPAddr {
	t.Helper()
	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, mock.New(), bmc.WithClock(clock.Real), bmc.WithCipherSuites(suites))
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	srv := server.NewServer(b, udp.Wrap(pc))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = srv.Close() })
	go func() { _ = srv.Serve(ctx) }()
	return pc.LocalAddr().(*net.UDPAddr)
}

func connectWith(t *testing.T, addr *net.UDPAddr, password string, suite types.CipherSuiteID) (*Client, error) {
	t.Helper()
	c, err := NewClient(addr.IP.String(), addr.Port, "ADMIN", password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLanplus).WithTimeout(2 * time.Second).WithRetry(0).WithCipherSuiteID(suite)
	err = c.Connect(context.Background())
	if err == nil {
		t.Cleanup(func() { _ = c.Close(context.Background()) })
	}
	return c, err
}

// TestCipherSuite0Session checks that a cipher suite 0 (RAKP-none) session
// can be set up and used. RAKP 2 and RAKP 4 carry no authentication code
// with RAKP-none; the client used to compare its nil code with the BMC's
// and reject every such session with "RAKP2 authentication code mismatch".
func TestCipherSuite0Session(t *testing.T) {
	addr := newCipherSuiteTestServer(t, []types.CipherSuiteID{types.CipherSuiteID0, types.CipherSuiteID3})

	c, err := connectWith(t, addr, "ADMIN", types.CipherSuiteID0)
	if err != nil {
		t.Fatalf("suite 0 Connect: %v", err)
	}
	if _, err := c.GetDeviceID(context.Background()); err != nil {
		t.Fatalf("suite 0 GetDeviceID: %v", err)
	}

	// Suite 0 authenticates nothing, so a wrong password is accepted too:
	// that is the weakness behind CVE-2013-4782 on BMCs that enable it.
	if _, err := connectWith(t, addr, "wrong", types.CipherSuiteID0); err != nil {
		t.Errorf("suite 0 with a wrong password: %v (RAKP-none checks no password)", err)
	}

	// Authenticated suites still reject a wrong password.
	if _, err := connectWith(t, addr, "wrong", types.CipherSuiteID3); err == nil {
		t.Error("suite 3 accepted a wrong password")
	}
}

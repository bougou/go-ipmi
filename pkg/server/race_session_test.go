package server

// RMCP+ per-session data-race regression tests driven over raw UDP.

import (
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

func raceMustWrite(t *testing.T, c *net.UDPConn, b []byte) {
	t.Helper()
	if _, err := c.Write(b); err != nil {
		t.Fatal(err)
	}
}

func raceMustReadPayload(t *testing.T, c *net.UDPConn) []byte {
	t.Helper()
	buf := make([]byte, 4096)
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := c.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, _, payload, ok := protocol.ParseRMCPPlusHeader(buf[:n])
	if !ok {
		t.Fatal("bad rmcp+ response")
	}
	return payload
}

// raceOpenSessionSuite0 opens a suite-0 (no auth/integrity/confidentiality)
// session so the session can go active and carry plaintext IPMI without
// reimplementing RAKP crypto.
func raceOpenSessionSuite0(t *testing.T, c *net.UDPConn, consoleID uint32) uint32 {
	t.Helper()
	authP, integP, cryptP := rmcpplus.NewAlgorithmPayloads(
		types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)
	req := &rmcpplus.OpenSessionRequest{
		RequestedMaximumPrivilegeLevel: types.PrivilegeLevelAdministrator,
		RemoteConsoleSessionID:         consoleID,
		AuthenticationPayload:          authP,
		IntegrityPayload:               integP,
		ConfidentialityPayload:         cryptP,
	}
	raceMustWrite(t, c, protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRmcpOpenSessionRequest), 0, 0, 0, req.Pack()))

	var resp rmcpplus.OpenSessionResponse
	if err := resp.Unpack(raceMustReadPayload(t, c)); err != nil {
		t.Fatal(err)
	}
	if resp.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		t.Fatalf("open session status %v", resp.RmcpStatusCode)
	}
	return resp.ManagedSystemSessionID
}

// raceDoRAKPNone completes the RAKP handshake for an AuthAlg_None session.
func raceDoRAKPNone(t *testing.T, c *net.UDPConn, bmcID uint32) {
	t.Helper()
	m1 := &rmcpplus.RAKPMessage1{
		ManagedSystemSessionID:         bmcID,
		RequestedMaximumPrivilegeLevel: types.PrivilegeLevelAdministrator,
		UsernameLength:                 uint8(len(raceUser)),
		Username:                       []byte(raceUser),
	}
	raceMustWrite(t, c, protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRAKPMessage1), 0, 0, 0, m1.Pack()))
	var msg2 rmcpplus.RAKPMessage2
	msg2.AuthAlg = types.AuthAlg_None
	if err := msg2.Unpack(raceMustReadPayload(t, c)); err != nil {
		t.Fatalf("rakp2 unpack: %v", err)
	}
	if msg2.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		t.Fatalf("rakp2 status %v", msg2.RmcpStatusCode)
	}

	m3 := &rmcpplus.RAKPMessage3{
		RmcpStatusCode:         types.RmcpStatusCodeNoErrors,
		ManagedSystemSessionID: bmcID,
	}
	raceMustWrite(t, c, protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRAKPMessage3), 0, 0, 0, m3.Pack()))
	var msg4 rmcpplus.RAKPMessage4
	msg4.AuthAlg = types.AuthAlg_None
	if err := msg4.Unpack(raceMustReadPayload(t, c)); err != nil {
		t.Fatalf("rakp4 unpack: %v", err)
	}
	if msg4.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		t.Fatalf("rakp4 status %v (session not active)", msg4.RmcpStatusCode)
	}
}

// raceGetDeviceIDPayload builds a minimal IPMB-format Get Device ID request.
func raceGetDeviceIDPayload() []byte {
	return []byte{
		0x20,      // rsAddr
		0x06 << 2, // netFn/lun (App)
		0x00,      // csum1
		0x81,      // rqAddr
		0x01 << 2, // rqSeq/lun
		0x01,      // cmd = Get Device ID
		0x00,      // csum2 (stripped)
	}
}

// TestActiveSessionSeqRace establishes an active suite-0 session, then fires
// many plaintext IPMI packets on the SAME session from several sockets. Each is
// dispatched in its own server goroutine; before the per-session lock, the
// shared inbound/outbound sequence writes raced.
func TestActiveSessionSeqRace(t *testing.T) {
	b := raceNewBMC(t, bmc.WithCipherSuites([]types.CipherSuiteID{types.CipherSuiteID0}))
	port, _, stop := raceStartServer(t, b)
	defer stop()

	hs, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	bmcID := raceOpenSessionSuite0(t, hs, 0x11223344)
	raceDoRAKPNone(t, hs, bmcID)
	_ = hs.Close()

	sessPkt := protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeIPMI), 0, bmcID, 1, raceGetDeviceIDPayload())

	var wg sync.WaitGroup
	var responded atomic.Bool
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
			if err != nil {
				return
			}
			defer c.Close()
			buf := make([]byte, 4096)
			for range 60 {
				_, _ = c.Write(sessPkt)
				_ = c.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
				if _, err := c.Read(buf); err == nil {
					responded.Store(true)
				}
			}
		}()
	}
	wg.Wait()
	// The race detector is the real assertion; this only guards against the
	// packets silently never reaching the session dispatch path at all.
	if !responded.Load() {
		t.Fatal("no session packet got a response; the burst never exercised the dispatch path")
	}
	if n := b.Sessions.Count(); n != 1 {
		t.Fatalf("want the 1 established session to survive the burst, got %d", n)
	}
}

func raceRAKP1Packet(bmcSessionID uint32, tag uint8) []byte {
	msg := &rmcpplus.RAKPMessage1{
		MessageTag:                     tag,
		ManagedSystemSessionID:         bmcSessionID,
		RequestedMaximumPrivilegeLevel: types.PrivilegeLevelAdministrator,
		UsernameLength:                 uint8(len(raceUser)),
		Username:                       []byte(raceUser),
	}
	return protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRAKPMessage1), 0, 0, 0, msg.Pack())
}

// raceOpenSessionSHA1 opens a suite-3 session (used only to obtain a pending
// session ID for the RAKP1 race).
func raceOpenSessionSHA1(t *testing.T, cl *net.UDPConn, consoleID uint32) uint32 {
	t.Helper()
	authP, integP, cryptP := rmcpplus.NewAlgorithmPayloads(
		types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128)
	req := &rmcpplus.OpenSessionRequest{
		RequestedMaximumPrivilegeLevel: types.PrivilegeLevelAdministrator,
		RemoteConsoleSessionID:         consoleID,
		AuthenticationPayload:          authP,
		IntegrityPayload:               integP,
		ConfidentialityPayload:         cryptP,
	}
	raceMustWrite(t, cl, protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRmcpOpenSessionRequest), 0, 0, 0, req.Pack()))
	var resp rmcpplus.OpenSessionResponse
	if err := resp.Unpack(raceMustReadPayload(t, cl)); err != nil {
		t.Fatal(err)
	}
	if resp.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		t.Fatalf("open session error status %v", resp.RmcpStatusCode)
	}
	return resp.ManagedSystemSessionID
}

// TestRAKP1SamePendingSessionRace fires many RAKP Message 1 packets targeting
// the SAME pending session. Each is dispatched in its own server goroutine;
// before the per-session lock, HandleRAKP1's writes to the shared session
// (ConsoleRand, Role, User, BMCRand) raced.
func TestRAKP1SamePendingSessionRace(t *testing.T) {
	b := raceNewBMC(t)
	port, _, stop := raceStartServer(t, b)
	defer stop()

	cl, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	bmcID := raceOpenSessionSHA1(t, cl, 0xaabbccdd)

	var wg sync.WaitGroup
	var responded atomic.Bool
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
			if err != nil {
				return
			}
			defer c.Close()
			buf := make([]byte, 4096)
			for range 40 {
				_, _ = c.Write(raceRAKP1Packet(bmcID, uint8(i)))
				_ = c.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
				if _, err := c.Read(buf); err == nil {
					responded.Store(true)
				}
			}
		}(i)
	}
	wg.Wait()
	// The race detector is the real assertion; this only guards against the
	// RAKP1 burst silently never reaching the handshake path at all.
	if !responded.Load() {
		t.Fatal("no RAKP1 packet got a response; the burst never exercised the handshake path")
	}
}

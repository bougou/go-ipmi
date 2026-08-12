package vmproto

// Tests for the OpenIPMI VM protocol frontend, driven by a fake QEMU client
// that speaks the ipmi-bmc-extern codec: single-byte control commands on
// connect, then IPMI messages framed with the 0xA0/0xA1/0xAA bytes and an IPMB
// checksum. They prove Get Device ID succeeds, an unknown command answers 0xC1
// without dropping the connection, the escape path round-trips, and a fresh
// QEMU process can reconnect.

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
)

const (
	vmTestUser = "admin"
	vmTestPass = "adminpass1234567"
)

// newTestBMC builds a BMC with an enabled admin user on the LAN channel, the
// same shape the RMCP+ server tests use.
func newTestBMC(t *testing.T) *bmc.BMC {
	t.Helper()

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, mock.New(), bmc.WithClock(clock.Real))

	admin, err := b.Users.Add(2, vmTestUser)
	if err != nil {
		t.Fatal(err)
	}
	admin.SetPassword([]byte(vmTestPass))
	admin.Enabled = true
	admin.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	return b
}

// vmClient speaks the QEMU side of the OpenIPMI VM protocol, the way
// ipmi-bmc-extern forwards guest KCS transactions.
type vmClient struct {
	conn  net.Conn
	msgID byte
}

func dialVM(t *testing.T, addr string) *vmClient {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial VM listener: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return &vmClient{conn: conn}
}

// sendControl sends a single-byte control command (version, capabilities),
// terminated by the end-of-command marker instead of end-of-message.
// Control command bytes QEMU sends on connect; the server consumes and ignores
// them, and these tests prove that leaves the connection usable.
const (
	vmTestCmdVersion      = 0xFF
	vmTestCmdCapabilities = 0x08
)

func (c *vmClient) sendControl(t *testing.T, data ...byte) {
	t.Helper()

	out := vmEncodeTest(data)
	out[len(out)-1] = vmCmdChar

	if _, err := c.conn.Write(out); err != nil {
		t.Fatalf("write control: %v", err)
	}
}

// request performs one IPMI transaction and returns (completionCode, data).
func (c *vmClient) request(t *testing.T, netFn, cmd byte, data ...byte) (byte, []byte) {
	t.Helper()

	c.msgID++

	msg := append([]byte{c.msgID, netFn << 2, cmd}, data...)
	msg = append(msg, -sum8(msg))

	if _, err := c.conn.Write(vmEncodeTest(msg)); err != nil {
		t.Fatalf("write request: %v", err)
	}

	resp := c.readMessage(t)

	if resp[0] != c.msgID {
		t.Fatalf("response msgID mismatch: got %d want %d", resp[0], c.msgID)
	}
	if got := resp[1] >> 2; got != netFn|1 {
		t.Fatalf("response netfn mismatch: got %#x want %#x", got, netFn|1)
	}
	if resp[2] != cmd {
		t.Fatalf("response cmd mismatch: got %#x want %#x", resp[2], cmd)
	}

	return resp[3], resp[4 : len(resp)-1]
}

// readMessage reads one unescaped, checksum-verified message from the stream.
func (c *vmClient) readMessage(t *testing.T) []byte {
	t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	var (
		acc      []byte
		inEscape bool
		buf      = make([]byte, 1)
	)

	for {
		if _, err := c.conn.Read(buf); err != nil {
			t.Fatalf("read: %v", err)
		}

		switch ch := buf[0]; ch {
		case vmMsgChar:
			if len(acc) < 5 {
				t.Fatalf("message too short: % x", acc)
			}
			if sum8(acc) != 0 {
				t.Fatalf("checksum mismatch: % x", acc)
			}
			return acc
		case vmEscapeChar:
			inEscape = true
		default:
			if inEscape {
				ch &^= 0x10
				inEscape = false
			}
			acc = append(acc, ch)
		}
	}
}

func sum8(bs []byte) byte {
	var s byte
	for _, b := range bs {
		s += b
	}
	return s
}

// vmEncodeTest is the client-side encoder: escape framing bytes, append the
// end-of-message marker.
func vmEncodeTest(bs []byte) []byte {
	out := make([]byte, 0, len(bs)+8)
	for _, b := range bs {
		if b == vmMsgChar || b == vmCmdChar || b == vmEscapeChar {
			out = append(out, vmEscapeChar, b|0x10)
			continue
		}
		out = append(out, b)
	}
	return append(out, vmMsgChar)
}

// startVM starts s on a loopback TCP listener and returns its address.
func startVM(t *testing.T, s *VMServer) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		if serveErr := s.Serve(ctx, ln); serveErr != nil {
			t.Errorf("serve VM: %v", serveErr)
		}
	}()

	t.Cleanup(func() { cancel(); <-done })

	return ln.Addr().String()
}

// TestVMGetDeviceID replays the QEMU connect handshake and proves the kernel's
// Get Device ID probe succeeds over the VM protocol.
func TestVMGetDeviceID(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))
	vm := dialVM(t, addr)

	// QEMU connect handshake: version + capabilities announcements.
	vm.sendControl(t, vmTestCmdVersion, 1)
	vm.sendControl(t, vmTestCmdCapabilities, 0x3F)

	cc, data := vm.request(t, 0x06, 0x01)
	if cc != 0 {
		t.Fatalf("get device ID: cc=%#x", cc)
	}
	if len(data) < 11 {
		t.Fatalf("get device ID: short response len=%d", len(data))
	}
}

// TestVMGetChannelInfoSystemInterface proves an in-band Get Channel Info for
// 0x0E ("this channel") over the VM frontend reports the system interface
// (channel 0x0F, medium system-interface), not LAN. The VM dispatch must
// attribute the request to the system-interface channel for 0x0E to resolve
// correctly.
func TestVMGetChannelInfoSystemInterface(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))
	vm := dialVM(t, addr)

	vm.sendControl(t, vmTestCmdVersion, 1)

	// Get Channel Info (App 0x42) for channel 0x0E.
	cc, data := vm.request(t, 0x06, 0x42, 0x0E)
	if cc != 0 {
		t.Fatalf("get channel info: cc=%#x", cc)
	}
	if len(data) < 2 {
		t.Fatalf("get channel info: short response len=%d", len(data))
	}
	if data[0] != 0x0F {
		t.Errorf("channel number = %#x, want 0x0F (system interface)", data[0])
	}
	if data[1] != uint8(bmc.ChannelMediumSystemIF) {
		t.Errorf("medium = %#x, want %#x (system interface)", data[1], uint8(bmc.ChannelMediumSystemIF))
	}
}

// TestVMUnknownCommand proves an unimplemented command answers 0xC1 (Invalid
// Command) and the connection survives to serve the next request.
func TestVMUnknownCommand(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))
	vm := dialVM(t, addr)

	vm.sendControl(t, vmTestCmdVersion, 1)

	// DCMI Get Capabilities: not implemented by the default registry.
	if cc, _ := vm.request(t, 0x2c, 0x01, 0x00); cc != 0xC1 {
		t.Fatalf("unknown command: cc=%#x, want 0xC1", cc)
	}

	// The connection must still work.
	if cc, _ := vm.request(t, 0x06, 0x01); cc != 0 {
		t.Fatalf("get device ID after unknown command: cc=%#x", cc)
	}
}

// TestVMEscapeRoundTrip forces a framing byte into the request wire form so the
// server's decoder must reverse an escape sequence to validate the checksum and
// parse the command. msgID 0x3D makes the Get Device ID request checksum 0xAA
// (an escape byte); if the server mis-decoded the escape, the checksum would
// fail and the message would be dropped, timing out the read below.
func TestVMEscapeRoundTrip(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))
	vm := dialVM(t, addr)

	vm.sendControl(t, vmTestCmdVersion, 1)

	vm.msgID = 0x3C // next request uses 0x3D
	cc, _ := vm.request(t, 0x06, 0x01)
	if cc != 0 {
		t.Fatalf("get device ID over escaped request: cc=%#x", cc)
	}
}

// TestVMEncodeUnit checks the server encoder escapes every framing byte and
// that the codec round-trips a payload containing all three of them, covering
// the response-side escape path deterministically.
func TestVMEncodeUnit(t *testing.T) {
	payload := []byte{0x01, vmMsgChar, 0x02, vmCmdChar, 0x03, vmEscapeChar, 0x04}

	encoded := vmEncode(payload)
	if encoded[len(encoded)-1] != vmMsgChar {
		t.Fatalf("encoded frame not terminated by 0x%02x", vmMsgChar)
	}

	// Decode by the same rules the client and server read loops use.
	var (
		got      []byte
		inEscape bool
	)
	for _, ch := range encoded[:len(encoded)-1] {
		switch ch {
		case vmEscapeChar:
			inEscape = true
		default:
			if inEscape {
				ch &^= 0x10
				inEscape = false
			}
			got = append(got, ch)
		}
	}

	if string(got) != string(payload) {
		t.Fatalf("round-trip mismatch: got % x want % x", got, payload)
	}
}

// TestVMReconnect proves a fresh QEMU process (every guest power cycle is one)
// can reconnect to the same listener and keep working.
func TestVMReconnect(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))

	for i := range 3 {
		vm := dialVM(t, addr)
		vm.sendControl(t, vmTestCmdVersion, 1)

		if cc, _ := vm.request(t, 0x06, 0x01); cc != 0 {
			t.Fatalf("connection %d: get device ID cc=%#x", i, cc)
		}

		_ = vm.conn.Close()
		// Give the accept loop a moment to observe the close and loop back to
		// Accept before the next dial.
		time.Sleep(50 * time.Millisecond)
	}
}

// TestVMGetSessionChallengeRejected proves the system interface does not serve
// LAN session establishment: an in-band Get Session Challenge must be rejected
// without allocating a pending v1.5 session, because the pending-session table
// evicts its oldest entry under pressure and in-band software could otherwise
// evict a remote console's in-flight handshake at will.
func TestVMGetSessionChallengeRejected(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))
	vm := dialVM(t, addr)

	vm.sendControl(t, vmTestCmdVersion, 1)

	// Get Session Challenge (App 0x39): auth type MD5 + 16-byte username.
	data := make([]byte, 17)
	data[0] = 0x02 // MD5
	copy(data[1:], vmTestUser)
	cc, _ := vm.request(t, 0x06, 0x39, data...)
	if cc == 0 {
		t.Fatal("in-band Get Session Challenge succeeded, want rejection")
	}
	if got := b.V15Sessions.Count(); got != 0 {
		t.Fatalf("in-band Get Session Challenge allocated %d pending v1.5 session(s), want 0", got)
	}
}

// TestClientRoundTrip drives the exported [Client] against a real VMServer,
// proving the client/server pair round-trips a Get Device ID over the VM
// protocol. This is the same exchange the goipmi-server VM socket demo and its
// e2e use.
func TestClientRoundTrip(t *testing.T) {
	b := newTestBMC(t)
	addr := startVM(t, NewVMServer(b))

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	c := NewClient(conn, 5*time.Second)
	cc, data, err := c.Command(0x06, 0x01) // Get Device ID
	if err != nil {
		t.Fatalf("Command: %v", err)
	}
	if cc != 0 {
		t.Fatalf("completion code = %#x, want 0", cc)
	}
	if len(data) < 11 {
		t.Fatalf("Get Device ID short response: %d bytes", len(data))
	}
}

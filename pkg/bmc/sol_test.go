package bmc

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// newSOLTestSession returns an active-looking session for SOL activation
// tests: admin privilege, AES + HMAC-SHA1 negotiated, user with default
// payload access on channel 1.
func newSOLTestSession(t *testing.T, b *BMC) *Session {
	t.Helper()
	sess, err := b.Sessions.Allocate(0xA5A5A5A5, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	sess.State = SessionStateActive
	sess.Channel = 1
	sess.PrivilegeLevel = PrivilegeLevelAdministrator
	sess.K2 = make([]byte, 16)
	user, err := b.Users.GetByName("sol")
	if err != nil {
		user, err = b.Users.Add(2, "sol")
		if err != nil {
			t.Fatalf("add user: %v", err)
		}
	}
	sess.User = user
	return sess
}

func TestSOLConfigDefaults(t *testing.T) {
	c := NewSOLConfig()
	cases := []struct {
		selector uint8
		want     []byte
	}{
		{0, []byte{0x00}}, // set complete
		{1, []byte{0x01}}, // enabled
		{2, []byte{0x04}}, // no forcing, ADMINISTRATOR privilege
		{3, []byte{10, SOLMaxPayloadChars}},
		{4, []byte{3, 5}},
		{5, []byte{0x0a}},       // 115.2k non-volatile
		{6, []byte{0x0a}},       // 115.2k volatile
		{7, []byte{0x01}},       // channel 1
		{8, []byte{0x6f, 0x02}}, // 623 LE
	}
	for _, tc := range cases {
		got, ok := c.GetParam(tc.selector)
		if !ok {
			t.Fatalf("param %d: not supported", tc.selector)
		}
		if string(got) != string(tc.want) {
			t.Errorf("param %d: got %x, want %x", tc.selector, got, tc.want)
		}
	}
	if _, ok := c.GetParam(9); ok {
		t.Errorf("param 9 must not be supported")
	}
}

func TestSOLConfigSetValidation(t *testing.T) {
	c := NewSOLConfig()
	cases := []struct {
		name     string
		selector uint8
		data     []byte
		wantCC   types.CompletionCode
	}{
		{"read-only channel", 7, []byte{2}, 0x82},
		{"read-only port", 8, []byte{0, 0}, 0x82},
		{"unknown param", 9, []byte{0}, types.CodeParameterNotSupported},
		{"commit write unsupported", 0, []byte{2}, types.CodeRequestDataFieldInvalid},
		{"set in progress", 0, []byte{1}, types.CodeOK},
		// The table runs in order on one config: a second claim while the
		// first is outstanding gets 81h (Table 26-3); set complete releases.
		{"set in progress already claimed", 0, []byte{1}, 0x81},
		{"set complete", 0, []byte{0}, types.CodeOK},
		{"set in progress after complete", 0, []byte{1}, types.CodeOK},
		{"set complete again", 0, []byte{0}, types.CodeOK},
		{"privilege reserved value", 2, []byte{0x01}, types.CodeRequestDataFieldInvalid},
		{"privilege operator", 2, []byte{0x03}, types.CodeOK},
		{"zero accumulate interval", 3, []byte{0, 10}, types.CodeRequestDataFieldInvalid},
		{"threshold too large", 3, []byte{10, 0xff}, types.CodeParameterOutOfRange},
		{"char params ok", 3, []byte{5, 100}, types.CodeOK},
		{"retry masks reserved bits", 4, []byte{0xff, 7}, types.CodeOK},
		{"bit rate too low", 5, []byte{0x05}, types.CodeParameterOutOfRange},
		{"bit rate serial-channel reserved", 5, []byte{0x00}, types.CodeParameterOutOfRange},
		{"bit rate ok", 6, []byte{0x07}, types.CodeOK},
	}
	for _, tc := range cases {
		if cc := c.SetParam(tc.selector, tc.data); cc != tc.wantCC {
			t.Errorf("%s: got CC %#02x, want %#02x", tc.name, cc, tc.wantCC)
		}
	}
	// retry count reserved bits [7:3] must not leak into the stored value
	if got, _ := c.GetParam(4); got[0] != 0x07 {
		t.Errorf("retry count: got %#02x, want 0x07 (reserved bits masked)", got[0])
	}
}

func TestSOLActivateRules(t *testing.T) {
	ctx := context.Background()

	t.Run("no console hardware", func(t *testing.T) {
		m := mock.New() // Console.Conn nil → no console
		b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, false, false); !errors.Is(err, ErrSOLDisabled) {
			t.Fatalf("got %v, want ErrSOLDisabled", err)
		}
	})

	t.Run("config disabled", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		if cc := b.SOL.Config().SetParam(1, []byte{0}); cc != types.CodeOK {
			t.Fatalf("disable SOL: CC %#02x", cc)
		}
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, false, false); !errors.Is(err, ErrSOLDisabled) {
			t.Fatalf("got %v, want ErrSOLDisabled", err)
		}
	})

	t.Run("privilege below configured level", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		sess.PrivilegeLevel = PrivilegeLevelOperator // config default is ADMINISTRATOR
		if _, err := b.SOL.Activate(ctx, sess, false, false); !errors.Is(err, ErrSOLPrivilege) {
			t.Fatalf("got %v, want ErrSOLPrivilege", err)
		}
	})

	t.Run("user payload access revoked", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		sess.User.PayloadAccessFor(1).Standard1 = 0 // SOL bit cleared
		if _, err := b.SOL.Activate(ctx, sess, false, false); !errors.Is(err, ErrSOLPrivilege) {
			t.Fatalf("got %v, want ErrSOLPrivilege", err)
		}
	})

	t.Run("encryption without authentication rejected", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, true, false); !errors.Is(err, ErrSOLEncryptionUnavailable) {
			t.Fatalf("got %v, want ErrSOLEncryptionUnavailable", err)
		}
	})

	t.Run("forced encryption refused when console declines", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		if cc := b.SOL.Config().SetParam(2, []byte{0x80 | 0x04}); cc != types.CodeOK {
			t.Fatalf("force encryption: CC %#02x", cc)
		}
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, false, false); !errors.Is(err, ErrSOLEncryptionRequired) {
			t.Fatalf("got %v, want ErrSOLEncryptionRequired", err)
		}
	})

	t.Run("second activation conflicts", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess1 := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess1, false, false); err != nil {
			t.Fatalf("first activate: %v", err)
		}
		sess2 := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess2, false, false); !errors.Is(err, ErrSOLAlreadyActive) {
			t.Fatalf("got %v, want ErrSOLAlreadyActive", err)
		}
		if cap, a1, _ := b.SOL.ActivationStatus(); cap != 1 || a1 != 0x01 {
			t.Fatalf("activation status: cap=%d active=%#02x, want 1/0x01", cap, a1)
		}
	})

	t.Run("volatile bit rate reloaded on activation", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		if cc := b.SOL.Config().SetParam(6, []byte{0x07}); cc != types.CodeOK {
			t.Fatalf("set volatile bit rate: CC %#02x", cc)
		}
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, false, false); err != nil {
			t.Fatalf("activate: %v", err)
		}
		got, _ := b.SOL.Config().GetParam(6)
		if got[0] != 0x0a { // §15.8: volatile copied from non-volatile at activation
			t.Fatalf("volatile bit rate after activate: %#02x, want 0x0a", got[0])
		}
	})
}

// feed is a test shorthand for processing one inbound packet.
func feed(t *testing.T, s *SOLStore, sess *Session, in *types.SOLPayloadPacket) *types.SOLPayloadPacket {
	t.Helper()
	out := s.ProcessPacket(context.Background(), sess.BMCID, in)
	if out == nil {
		t.Fatalf("ProcessPacket returned nil")
	}
	return out
}

func TestSOLProcessInboundData(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// seq 1 carries "ab": fully accepted.
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("ab")})
	if out.AcceptedCharacterCount != 2 || out.AckedSequenceNumber != 1 || out.NACK {
		t.Fatalf("ack: %+v, want accepted=2 ackedSeq=1 ACK", out)
	}
	if string(fake.TX) != "ab" {
		t.Fatalf("console TX %q, want %q", fake.TX, "ab")
	}

	// Retried seq 1 (same content per §15.9): re-ACK, no double write.
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("ab")})
	if out.AcceptedCharacterCount != 2 || string(fake.TX) != "ab" {
		t.Fatalf("dup: accepted=%d TX=%q, want 2/%q", out.AcceptedCharacterCount, fake.TX, "ab")
	}

	// seq 2 with the BREAK operation bit set (Table 15-2 bit [4]).
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 2, ControlByte: 0x10, CharacterData: []byte("c")})
	if fake.Breaks != 1 {
		t.Fatalf("breaks=%d, want 1", fake.Breaks)
	}
	if out.AcceptedCharacterCount != 1 {
		t.Fatalf("accepted=%d, want 1", out.AcceptedCharacterCount)
	}

	// A stranger session owns nothing.
	if other := newSOLTestSession(t, b); b.SOL.ProcessPacket(context.Background(), other.BMCID, &types.SOLPayloadPacket{SequenceNumber: 9}) != nil {
		t.Fatalf("packet from non-owner session must be ignored")
	}
}

func TestSOLProcessInboundWriteFailure(t *testing.T) {
	fake := &mock.FakeConsoleConn{WriteErr: errors.New("console gone")}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("x")})
	if !out.NACK || out.AcceptedCharacterCount != 0 {
		t.Fatalf("got NACK=%v accepted=%d, want NACK accepted=0", out.NACK, out.AcceptedCharacterCount)
	}
	// Table 15-2 status bit [5]: character transfer unavailable.
	if out.ControlByte&0x20 == 0 {
		t.Fatalf("status byte %#02x missing transfer-unavailable bit", out.ControlByte)
	}
}

func TestSOLProcessOutboundData(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}
	fake.RX = []byte("hello")

	// An empty poll (ACK-only, seq 0h) collects the pending console output.
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.SequenceNumber != 1 || string(out.CharacterData) != "hello" {
		t.Fatalf("outbound: seq=%d data=%q, want 1/%q", out.SequenceNumber, out.CharacterData, "hello")
	}

	// A poll that does not acknowledge seq 1 gets the same packet again
	// (§15.11 retry semantics; the BMC resends with the same sequence number).
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.SequenceNumber != 1 || string(out.CharacterData) != "hello" {
		t.Fatalf("resend: seq=%d data=%q, want 1/%q", out.SequenceNumber, out.CharacterData, "hello")
	}

	// Suspend NACK for seq 1 (Table 15-3): the BMC stops sending the pending
	// packet — the response is ACK-only — until the console resumes.
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 1, NACK: true})
	if out.SequenceNumber != 0 || len(out.CharacterData) != 0 {
		t.Fatalf("suspend: seq=%d data=%q, want ACK-only response", out.SequenceNumber, out.CharacterData)
	}

	// While suspended even a plain poll must not resurrect the packet.
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.SequenceNumber != 0 || len(out.CharacterData) != 0 {
		t.Fatalf("suspended poll: seq=%d data=%q, want ACK-only response", out.SequenceNumber, out.CharacterData)
	}

	// Resume ACK (ACK with count 0, Table 15-3): retransmit immediately.
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 1})
	if out.SequenceNumber != 1 || string(out.CharacterData) != "hello" {
		t.Fatalf("resume: seq=%d data=%q, want 1/%q", out.SequenceNumber, out.CharacterData, "hello")
	}

	// Partial ACK of 2 chars (Table 15-3): retransmit the remainder as a
	// fresh packet — a same-sequence resend would yield zero new bytes to
	// consoles computing new data by character offset (§15.11).
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 1, AcceptedCharacterCount: 2})
	if out.SequenceNumber != 2 || string(out.CharacterData) != "llo" {
		t.Fatalf("partial resend: seq=%d data=%q, want 2/%q", out.SequenceNumber, out.CharacterData, "llo")
	}

	// Completion ACK frees the sequence; fresh output gets the next number.
	fake.RX = []byte("world")
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 2, AcceptedCharacterCount: 3})
	if out.SequenceNumber != 3 || string(out.CharacterData) != "world" {
		t.Fatalf("next: seq=%d data=%q, want 3/%q", out.SequenceNumber, out.CharacterData, "world")
	}
}

func TestSOLDeactivateLifecycle(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// Deactivate by a non-owner session without the SOL privilege level
	// (default ADMINISTRATOR) is refused.
	other := newSOLTestSession(t, b)
	other.PrivilegeLevel = PrivilegeLevelOperator
	if err := b.SOL.Deactivate(other); !errors.Is(err, ErrSOLNotOwner) {
		t.Fatalf("non-owner deactivate: got %v, want ErrSOLNotOwner", err)
	}

	// A non-owner session at the SOL privilege level may force-deactivate:
	// the recovery path for consoles that crashed without deactivating.
	other.PrivilegeLevel = PrivilegeLevelAdministrator
	if err := b.SOL.Deactivate(other); err != nil {
		t.Fatalf("admin force-deactivate: %v", err)
	}
	if _, a1, _ := b.SOL.ActivationStatus(); a1 != 0 {
		t.Fatalf("instance still active after force-deactivate")
	}
	if !fake.Closed {
		t.Fatalf("console conn not closed on force-deactivate")
	}

	// Re-activate, then §24.2: terminating the session auto-deactivates its
	// payloads.
	fake2 := &mock.FakeConsoleConn{}
	m.Console().(*mock.Console).Conn = fake2
	sess = newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("re-activate: %v", err)
	}
	if err := b.Sessions.Close(sess.BMCID); err != nil {
		t.Fatalf("close session: %v", err)
	}
	if _, a1, _ := b.SOL.ActivationStatus(); a1 != 0 {
		t.Fatalf("instance still active after session close")
	}
	if !fake2.Closed {
		t.Fatalf("console conn not closed on session termination")
	}

	// Double deactivate reports "already deactivated" (Table 24-3, 80h).
	if err := b.SOL.Deactivate(sess); !errors.Is(err, ErrSOLNotActive) {
		t.Fatalf("double deactivate: got %v, want ErrSOLNotActive", err)
	}
}

func TestSOLProcessRetryReplaysOriginalAccept(t *testing.T) {
	fake := &mock.FakeConsoleConn{WriteLimit: 1}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// seq 1 carries "ab" but the console only takes 1 byte: NACK + accepted=1.
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("ab")})
	if !out.NACK || out.AcceptedCharacterCount != 1 {
		t.Fatalf("partial: NACK=%v accepted=%d, want NACK/1", out.NACK, out.AcceptedCharacterCount)
	}

	// A lost response makes the console retry the packet unchanged (§15.9).
	// The replay must report the original partial accept — ACKing both bytes
	// here would lose "b", which was never written to the console.
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("ab")})
	if !out.NACK || out.AcceptedCharacterCount != 1 {
		t.Fatalf("retry: NACK=%v accepted=%d, want NACK/1 (replayed)", out.NACK, out.AcceptedCharacterCount)
	}
	if string(fake.TX) != "a" {
		t.Fatalf("TX %q, want %q (no re-write on retry)", fake.TX, "a")
	}

	// The console resends the rejected remainder under a fresh sequence.
	fake.WriteLimit = 0
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 2, CharacterData: []byte("b")})
	if out.NACK || out.AcceptedCharacterCount != 1 || string(fake.TX) != "ab" {
		t.Fatalf("remainder: NACK=%v accepted=%d TX=%q", out.NACK, out.AcceptedCharacterCount, fake.TX)
	}
}

func TestSOLFlushOutbound(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}
	fake.RX = []byte("hello")

	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.SequenceNumber != 1 || string(out.CharacterData) != "hello" {
		t.Fatalf("outbound: seq=%d data=%q, want 1/%q", out.SequenceNumber, out.CharacterData, "hello")
	}

	// Suspend NACK, then Flush Outbound on an ACK-only packet (Table 15-3's
	// recovery path): the pending packet is dropped, not resumed.
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 1, NACK: true})
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{ControlByte: 0x01})
	if out.SequenceNumber != 0 || len(out.CharacterData) != 0 {
		t.Fatalf("flush: seq=%d data=%q, want ACK-only response", out.SequenceNumber, out.CharacterData)
	}

	// Output produced after the flush flows under a fresh sequence number.
	fake.RX = []byte("world")
	out = feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.SequenceNumber != 2 || string(out.CharacterData) != "world" {
		t.Fatalf("post-flush: seq=%d data=%q, want 2/%q", out.SequenceNumber, out.CharacterData, "world")
	}
}

func TestSOLHALErrorPropagation(t *testing.T) {
	m := mock.New()
	m.SetConsole(&mock.Console{})
	m.Console().(*mock.Console).OpenHook = func(context.Context) (hal.ConsoleConn, error) {
		return nil, hal.ErrNotSupported
	}
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	sess := newSOLTestSession(t, b)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); !errors.Is(err, hal.ErrNotSupported) {
		t.Fatalf("got %v, want hal.ErrNotSupported", err)
	}
}

// pumpCapture records asynchronously sent SOL packets.
type pumpCapture struct {
	mu  sync.Mutex
	pkt []*types.SOLPayloadPacket
}

func (p *pumpCapture) send(pkt *types.SOLPayloadPacket) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := *pkt
	cp.CharacterData = append([]byte{}, pkt.CharacterData...)
	p.pkt = append(p.pkt, &cp)
	return nil
}

func (p *pumpCapture) all() []*types.SOLPayloadPacket {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]*types.SOLPayloadPacket{}, p.pkt...)
}

// newConsoleBMC builds a BMC with a fake console attached.
func newConsoleBMC(t *testing.T) (*BMC, *mock.FakeConsoleConn) {
	t.Helper()
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	return New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real)), fake
}

func newPumpBMC(t *testing.T, fake *mock.FakeConsoleConn, cap *pumpCapture) (*BMC, *Session) {
	t.Helper()
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	b.SOL.SetSenderFactory(func(sess *Session, inst *SOLInstance) SOLSendFunc {
		return cap.send
	})
	sess := newSOLTestSession(t, b)
	sess.Addr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5000} // non-nil Addr enables the sender
	return b, sess
}

func waitForPkt(t *testing.T, cap *pumpCapture, n int) []*types.SOLPayloadPacket {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if got := cap.all(); len(got) >= n {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d async SOL packets", n)
	return nil
}

func TestSOLPumpAsyncPush(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	cap := &pumpCapture{}
	b, sess := newPumpBMC(t, fake, cap)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	// Console output must be pushed without any inbound packet.
	fake.FeedRX([]byte("async-output"))
	pkts := waitForPkt(t, cap, 1)
	if string(pkts[0].CharacterData) != "async-output" {
		t.Fatalf("async data %q, want %q", pkts[0].CharacterData, "async-output")
	}
	if pkts[0].SequenceNumber != 1 {
		t.Fatalf("async seq=%d, want 1", pkts[0].SequenceNumber)
	}

	// ACK it: the pump must then send the next chunk with a fresh sequence.
	fake.FeedRX([]byte("second"))
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: 1, AcceptedCharacterCount: 12})
	pkts = waitForPkt(t, cap, 2)
	if string(pkts[1].CharacterData) != "second" || pkts[1].SequenceNumber != 2 {
		t.Fatalf("second: seq=%d data=%q, want 2/%q", pkts[1].SequenceNumber, pkts[1].CharacterData, "second")
	}
}

// TestSOLDeactivateSendsStatus verifies the one-time status packet with
// status bit [4] (Table 15-2 footnote 2) issued on deactivation paths.
func TestSOLDeactivateSendsStatus(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	cap := &pumpCapture{}
	b, sess := newPumpBMC(t, fake, cap)
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	if err := b.SOL.Deactivate(sess); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	pkts := cap.all()
	if len(pkts) != 1 || pkts[0].NACK || pkts[0].ControlByte&0x10 == 0 || pkts[0].SequenceNumber != 0 {
		t.Fatalf("deactivation packet: %+v, want ACK-only with status [4]", pkts)
	}

	// Session termination (Close Session path, §24.2) issues the same status.
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("re-activate: %v", err)
	}
	if err := b.Sessions.Close(sess.BMCID); err != nil {
		t.Fatalf("close session: %v", err)
	}
	pkts = cap.all()
	if len(pkts) != 2 || pkts[1].NACK || pkts[1].ControlByte&0x10 == 0 {
		t.Fatalf("session-close packet: %+v, want status [4]", pkts[1:])
	}
}

func TestSOLSuspendResumeEncryption(t *testing.T) {
	ctx := context.Background()

	t.Run("toggle on encrypted activation", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		inst, err := b.SOL.Activate(ctx, sess, true, true)
		if err != nil {
			t.Fatalf("activate: %v", err)
		}
		if !inst.OutboundEncrypted() {
			t.Fatalf("activated with encryption but OutboundEncrypted is false")
		}
		if err := b.SOL.SuspendResumeEncryption(sess, SOLEncryptionOpSuspend); err != nil {
			t.Fatalf("suspend: %v", err)
		}
		if inst.OutboundEncrypted() {
			t.Fatalf("suspended but OutboundEncrypted is true")
		}
		if err := b.SOL.SuspendResumeEncryption(sess, SOLEncryptionOpResume); err != nil {
			t.Fatalf("resume: %v", err)
		}
		if !inst.OutboundEncrypted() {
			t.Fatalf("resumed but OutboundEncrypted is false")
		}
	})

	t.Run("suspend refused when configuration forces encryption", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		if cc := b.SOL.Config().SetParam(2, []byte{0x80 | 0x04}); cc != types.CodeOK {
			t.Fatalf("force encryption: CC %#02x", cc)
		}
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, true, true); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if err := b.SOL.SuspendResumeEncryption(sess, SOLEncryptionOpSuspend); !errors.Is(err, ErrSOLEncryptionForced) {
			t.Fatalf("suspend: got %v, want ErrSOLEncryptionForced", err)
		}
	})

	t.Run("resume needs a session with encryption", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, false, false); err != nil {
			t.Fatalf("activate: %v", err)
		}
		sess.CryptAlg = types.CryptAlg_None
		sess.K2 = nil
		if err := b.SOL.SuspendResumeEncryption(sess, SOLEncryptionOpResume); !errors.Is(err, ErrSOLEncryptionUnavailableForSession) {
			t.Fatalf("resume: got %v, want ErrSOLEncryptionUnavailableForSession", err)
		}
	})

	t.Run("regenerate IV unsupported and stranger session rejected", func(t *testing.T) {
		b, _ := newConsoleBMC(t)
		sess := newSOLTestSession(t, b)
		if _, err := b.SOL.Activate(ctx, sess, true, true); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if err := b.SOL.SuspendResumeEncryption(sess, SOLEncryptionOpRegenIV); !errors.Is(err, ErrSOLOperationUnsupported) {
			t.Fatalf("regen IV: got %v, want ErrSOLOperationUnsupported", err)
		}
		other := newSOLTestSession(t, b)
		if err := b.SOL.SuspendResumeEncryption(other, SOLEncryptionOpSuspend); !errors.Is(err, ErrSOLInstanceNotActive) {
			t.Fatalf("stranger: got %v, want ErrSOLInstanceNotActive", err)
		}
	})
}

func TestSOLPumpRetryAndDrop(t *testing.T) {
	fake := &mock.FakeConsoleConn{}
	cap := &pumpCapture{}
	b, sess := newPumpBMC(t, fake, cap)
	// Fast retries for the test: 2 retries at 10 ms (the floor).
	if cc := b.SOL.Config().SetParam(4, []byte{2, 1}); cc != types.CodeOK {
		t.Fatalf("set retry: CC %#02x", cc)
	}
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake.FeedRX([]byte("lost-cause"))
	pkts := waitForPkt(t, cap, 3) // initial send + 2 retries, same seq
	for i, p := range pkts {
		if p.SequenceNumber != 1 || string(p.CharacterData) != "lost-cause" {
			t.Fatalf("pkt %d: seq=%d data=%q, want 1/lost-cause", i, p.SequenceNumber, p.CharacterData)
		}
	}

	// Retries exhausted: the packet is dropped (§15.11) and the pump moves on.
	// The next packet reports the loss (Table 15-2 status [3], overrun).
	fake.FeedRX([]byte("aftermath"))
	pkts = waitForPkt(t, cap, 4)
	if string(pkts[3].CharacterData) != "aftermath" || pkts[3].SequenceNumber != 2 {
		t.Fatalf("after drop: seq=%d data=%q, want 2/%q", pkts[3].SequenceNumber, pkts[3].CharacterData, "aftermath")
	}
	if pkts[3].ControlByte&0x08 == 0 {
		t.Fatalf("after drop: control %#02x missing transmit-overrun bit", pkts[3].ControlByte)
	}
}

// --- console reconnect (recovery from a failed console conn) ---

// newReconnectBMC builds a pump BMC whose console Open is driven by hook,
// returning the BMC, an active session, and the async send capture.
func newReconnectBMC(t *testing.T, hook func(context.Context) (hal.ConsoleConn, error)) (*BMC, *Session, *pumpCapture) {
	t.Helper()
	m := mock.New()
	m.SetConsole(&mock.Console{OpenHook: hook})
	b := New(DeviceInfo{}, [16]byte{}, m, WithClock(clock.Real))
	// Reconnection is opt-in; the helper enables the default policy so the
	// reconnect tests exercise the feature rather than the disabled default.
	b.SOL.SetReconnectPolicy(&DefaultReconnectPolicy)
	cap := &pumpCapture{}
	b.SOL.SetSenderFactory(func(sess *Session, inst *SOLInstance) SOLSendFunc { return cap.send })
	sess := newSOLTestSession(t, b)
	sess.Addr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5000}
	return b, sess, cap
}

func waitForCondition(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}

func instBroken(inst *SOLInstance) bool {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	return inst.broken
}

// TestSOLReconnectRecovers verifies the full recovery cycle: a failing
// console conn flips the instance broken (NACK + status bit [5]), the pump
// re-attaches via the HAL Open, and the data plane resumes — with the
// pre-failure RX residue dropped (it belongs to the dead console session).
func TestSOLReconnectRecovers(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	fake2 := &mock.FakeConsoleConn{}
	opens := 0
	// release gates the first reconnect: the outage window is only one pump
	// tick wide, so the test holds the reconnect until the broken-state
	// behavior (NACK + status bit [5]) has been observed.
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseReconnect := func() { releaseOnce.Do(func() { close(release) }) }
	// Any failed assertion must not strand the pump (which CloseAll waits
	// for), so the gate is always released before cleanup runs.
	t.Cleanup(releaseReconnect)
	b, sess, cap := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens++
		if opens == 1 {
			return fake1, nil // activation
		}
		<-release // hold the reconnect while the outage is verified
		return fake2, nil
	})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	// One normal output round, acknowledged, so pending is non-empty when
	// the console dies: the reconnect must keep that pending state.
	fake1.FeedRX([]byte("hello"))
	pkts := waitForPkt(t, cap, 1)
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: pkts[0].SequenceNumber, AcceptedCharacterCount: 5})

	// Console dies; anything queued on it afterwards is stale residue.
	fake1.SetReadErr(errors.New("console died"))
	fake1.FeedRX([]byte("stale"))

	waitForCondition(t, func() bool { return instBroken(inst) })
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 2, CharacterData: []byte("k")})
	if !out.NACK {
		t.Fatalf("keystroke during outage: NACK=%v, want true", out.NACK)
	}
	if out.ControlByte&0x20 == 0 {
		t.Fatalf("keystroke during outage: control %#02x missing status bit [5]", out.ControlByte)
	}

	// Release the reconnect: it fires immediately, releasing the dead conn.
	releaseReconnect()
	waitForCondition(t, func() bool { return !instBroken(inst) })
	if !fake1.IsClosed() {
		t.Fatalf("dead console conn not closed")
	}

	// Keystrokes reach the fresh console, and its output flows on — the
	// stale residue from the dead console is not forwarded.
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 3, CharacterData: []byte("x")})
	if got := fake2.TXString(); got != "x" {
		t.Fatalf("keystroke after reconnect: TX %q, want %q", got, "x")
	}
	fake2.FeedRX([]byte("fresh"))
	pkts = waitForPkt(t, cap, 2)
	if string(pkts[1].CharacterData) != "fresh" {
		t.Fatalf("after reconnect: data %q, want %q (stale residue must be dropped)", pkts[1].CharacterData, "fresh")
	}

	// Output continues after the reconnect. ACK the "fresh" packet first so
	// the retry engine does not re-send it while the next chunk accumulates.
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: pkts[1].SequenceNumber, AcceptedCharacterCount: 5})
	fake2.FeedRX([]byte("second"))
	waitForCondition(t, func() bool {
		for _, p := range cap.all() {
			if string(p.CharacterData) == "second" {
				return true
			}
		}
		return false
	})
}

// TestSOLReconnectKeepsPending verifies the unacknowledged outbound packet
// survives the reconnect: the remote console still owes its ACK
// (Table 15-3), so the pump must keep it rather than treat the new console
// as a fresh start.
func TestSOLReconnectKeepsPending(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	fake2 := &mock.FakeConsoleConn{}
	opens := 0
	b, sess, cap := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens++
		if opens == 1 {
			return fake1, nil
		}
		return fake2, nil
	})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	// Keep the unacknowledged packet alive across the outage: the retry
	// engine would drop it after its retries are exhausted (§15.11), which
	// is unrelated to the reconnect — widen the window (200 × 10ms) and
	// retry immediately so only the reconnect could be responsible for
	// losing it.
	if cc := b.SOL.Config().SetParam(4, []byte{200, 1}); cc != types.CodeOK {
		t.Fatalf("set retry: CC %#02x", cc)
	}
	b.SOL.SetReconnectPolicy(&ReconnectPolicy{}) // Initial 0: reconnect immediately

	fake1.FeedRX([]byte("abc"))
	pkts := waitForPkt(t, cap, 1) // seq 1, unacknowledged
	// Break the console via the keystroke path: while pending != nil the
	// pump never reads the console, so a read failure could not flip broken
	// before the retry engine dropped the pending packet — a write failure
	// in ProcessPacket can, leaving pending untouched.
	fake1.SetWriteErr(errors.New("console died"))
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 2, CharacterData: []byte("k")})
	if !out.NACK {
		t.Fatalf("keystroke on failing console: NACK=%v, want true", out.NACK)
	}
	waitForCondition(t, func() bool { return instBroken(inst) })
	waitForCondition(t, func() bool { return !instBroken(inst) })

	inst.mu.Lock()
	pending := inst.pending != nil
	inst.mu.Unlock()
	if !pending {
		t.Fatalf("reconnect dropped the unacknowledged pending packet")
	}

	// The pre-outage ACK is still honored after the reconnect.
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{AckedSequenceNumber: pkts[0].SequenceNumber, AcceptedCharacterCount: 3})
	inst.mu.Lock()
	pending = inst.pending != nil
	inst.mu.Unlock()
	if pending {
		t.Fatalf("pending not cleared by ACK after reconnect")
	}
}

// TestSOLReconnectWriteErrorTriggers verifies that a failing console Write
// (keystroke path) also starts the reconnect cycle.
func TestSOLReconnectWriteErrorTriggers(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	fake2 := &mock.FakeConsoleConn{}
	opens := 0
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens++
		if opens == 1 {
			return fake1, nil
		}
		return fake2, nil
	})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake1.SetWriteErr(errors.New("console write failed"))
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 5, CharacterData: []byte("k")})
	if !out.NACK {
		t.Fatalf("keystroke on failing console: NACK=%v, want true", out.NACK)
	}

	waitForCondition(t, func() bool { return !instBroken(inst) })
	if !fake1.IsClosed() {
		t.Fatalf("dead console conn not closed")
	}
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 6, CharacterData: []byte("y")})
	if got := fake2.TXString(); got != "y" {
		t.Fatalf("keystroke after reconnect: TX %q, want %q", got, "y")
	}
}

// TestSOLReconnectBackoff verifies the exponential backoff between failed
// reconnect attempts: the first fires immediately, later failures double
// the delay (bounded by consoleReconnectMax), and no attempt fires early.
func TestSOLReconnectBackoff(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	var opens atomic.Int32
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens.Add(1)
		if opens.Load() == 1 {
			return fake1, nil
		}
		return nil, errors.New("console unreachable")
	})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake1.SetReadErr(errors.New("console died"))

	// First attempt is immediate; its failure arms the 1s backoff.
	waitForCondition(t, func() bool { return opens.Load() >= 2 })
	inst.mu.Lock()
	armed := inst.reconnectAt.Sub(inst.clock.Now())
	inst.mu.Unlock()
	if armed < 900*time.Millisecond || armed > 2*time.Second {
		t.Fatalf("first backoff armed at %v, want ~1s", armed)
	}

	// No early retry while backing off.
	time.Sleep(300 * time.Millisecond)
	if opens.Load() > 2 {
		t.Fatalf("reconnect fired early (opens=%d)", opens.Load())
	}

	// After the 1s backoff the second attempt fires, and its failure arms
	// the next delay from the policy (1s, 2s, 4s, ... sequence).
	waitForCondition(t, func() bool { return opens.Load() >= 3 })
	inst.mu.Lock()
	failures := inst.failures
	armed = inst.reconnectAt.Sub(inst.clock.Now())
	inst.mu.Unlock()
	if failures != 3 {
		t.Fatalf("failures after 2nd failure: %d, want 3", failures)
	}
	if armed < 1900*time.Millisecond {
		t.Fatalf("second backoff armed at %v, want ~2s", armed)
	}
}

// TestSOLReconnectDeactivateWhileBroken verifies deactivation completes
// while the reconnect cycle is backing off (no hang, conn released).
func TestSOLReconnectDeactivateWhileBroken(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	var opens atomic.Int32
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens.Add(1)
		if opens.Load() == 1 {
			return fake1, nil
		}
		return nil, errors.New("console unreachable")
	})
	if _, err := b.SOL.Activate(context.Background(), sess, false, false); err != nil {
		t.Fatalf("activate: %v", err)
	}

	fake1.SetReadErr(errors.New("console died"))
	waitForCondition(t, func() bool { return opens.Load() >= 2 }) // backing off

	done := make(chan error, 1)
	go func() { done <- b.SOL.Deactivate(sess) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("deactivate: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("deactivate hung during reconnect backoff")
	}
	if !fake1.IsClosed() {
		t.Fatalf("dead console conn not closed on deactivate")
	}
}

func TestDefaultReconnectPolicy(t *testing.T) {
	// Interval sequence: 1s initial, doubling capped at 30s; retries
	// forever (Steps 0).
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 30 * time.Second, 30 * time.Second}
	for i, w := range want {
		got, giveUp := DefaultReconnectPolicy.Delay(i + 1)
		if got != w || giveUp {
			t.Fatalf("Delay(%d): got %v giveUp=%v, want %v/false", i+1, got, giveUp, w)
		}
	}
}

// TestReconnectPolicySteps verifies Steps bounds the attempts and Jitter
// stays within the declared range.
func TestReconnectPolicySteps(t *testing.T) {
	p := ReconnectPolicy{Initial: 100 * time.Millisecond, Factor: 2, Jitter: 0.5, Cap: time.Second, Steps: 3}
	// 3 reconnect attempts: failures 1..2 keep going, failure 3 gives up.
	for i := 1; i <= 2; i++ {
		if _, giveUp := p.Delay(i); giveUp {
			t.Fatalf("Delay(%d): gave up early", i)
		}
	}
	if _, giveUp := p.Delay(3); !giveUp {
		t.Fatal("Delay(3): want give-up past Steps")
	}
	// Jitter keeps the wait within [base, base*(1+Jitter)).
	for i := 1; i <= 2; i++ {
		base := p.Initial * time.Duration(1<<(i-1))
		got, _ := p.Delay(i)
		if got < base || got >= base+time.Duration(float64(base)*p.Jitter) {
			t.Fatalf("Delay(%d) jittered to %v, want within [%v, %v)", i, got, base, base+time.Duration(float64(base)*p.Jitter))
		}
	}
}

// TestSOLReconnectDisabled verifies the opt-out (nil ReconnectPolicy):
// the dead conn is released once, no further Open attempts happen, and the
// payload keeps reporting status bit [5] until deactivation — the
// spec-defined behavior without the reconnect enhancement.
func TestSOLReconnectDisabled(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	var opens atomic.Int32
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens.Add(1)
		return fake1, nil
	})
	b.SOL.SetReconnectPolicy(nil)
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake1.SetReadErr(errors.New("console died"))
	waitForCondition(t, func() bool { return instBroken(inst) })

	time.Sleep(300 * time.Millisecond)
	if opens.Load() != 1 {
		t.Fatalf("reconnect attempted despite disabled policy (opens=%d)", opens.Load())
	}
	if !fake1.IsClosed() {
		t.Fatalf("dead console conn not released")
	}

	// The payload keeps answering with status bit [5] so the remote console
	// can tell the outage apart from a dead BMC.
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.ControlByte&0x20 == 0 {
		t.Fatalf("disabled reconnect: control %#02x missing status bit [5]", out.ControlByte)
	}
}

// TestSOLReconnectCustomPolicy verifies a caller-supplied ReconnectPolicy
// drives both the first-retry decision and the later backoff, receiving the
// consecutive failure count (1-based).
func TestSOLReconnectCustomPolicy(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	fake2 := &mock.FakeConsoleConn{}
	var opens atomic.Int32
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens.Add(1)
		switch opens.Load() {
		case 1:
			return fake1, nil // activation
		case 2:
			return nil, errors.New("still down") // first retry fails
		}
		return fake2, nil // second retry succeeds
	})
	// Fixed 100ms retry schedule: the first retry waits 100ms (Initial),
	// the second (which succeeds) another 100ms (Factor 1).
	b.SOL.SetReconnectPolicy(&ReconnectPolicy{Initial: 100 * time.Millisecond, Factor: 1, Cap: 100 * time.Millisecond})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake1.SetReadErr(errors.New("console died"))
	// opens reaching 3 anchors both retries: the first failed, the second
	// (100ms later) succeeded.
	waitForCondition(t, func() bool { return opens.Load() >= 3 })
	waitForCondition(t, func() bool { return !instBroken(inst) })
	if !fake1.IsClosed() {
		t.Fatalf("dead console conn not closed")
	}
	feed(t, b.SOL, sess, &types.SOLPayloadPacket{SequenceNumber: 7, CharacterData: []byte("z")})
	if got := fake2.TXString(); got != "z" {
		t.Fatalf("keystroke after custom-policy reconnect: TX %q, want %q", got, "z")
	}
}

// TestSOLReconnectGiveUpAfterSteps verifies a policy with bounded Steps:
// once the attempts are exhausted the instance stops reconnecting and
// behaves like the disabled default (status bit [5], no further Opens).
func TestSOLReconnectGiveUpAfterSteps(t *testing.T) {
	fake1 := &mock.FakeConsoleConn{}
	var opens atomic.Int32
	b, sess, _ := newReconnectBMC(t, func(context.Context) (hal.ConsoleConn, error) {
		opens.Add(1)
		if opens.Load() == 1 {
			return fake1, nil
		}
		return nil, errors.New("console unreachable")
	})
	// One immediate retry (Steps 1), then give up.
	b.SOL.SetReconnectPolicy(&ReconnectPolicy{Steps: 1})
	inst, err := b.SOL.Activate(context.Background(), sess, false, false)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(b.SOL.CloseAll)

	fake1.SetReadErr(errors.New("console died"))
	// First retry fires immediately and fails; after that, no more.
	waitForCondition(t, func() bool { return opens.Load() >= 2 })
	time.Sleep(300 * time.Millisecond)
	if opens.Load() > 2 {
		t.Fatalf("reconnect continued past Steps (opens=%d)", opens.Load())
	}
	if !instBroken(inst) {
		t.Fatal("instance recovered despite Steps exhaustion")
	}
	out := feed(t, b.SOL, sess, &types.SOLPayloadPacket{})
	if out.ControlByte&0x20 == 0 {
		t.Fatalf("after give-up: control %#02x missing status bit [5]", out.ControlByte)
	}
}

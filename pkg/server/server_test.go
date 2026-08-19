package server

import (
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/types"
)

// TestSOLQueueRetiresForUnknownSession verifies a forged SOL packet for a
// session ID that does not exist retires its queue and worker immediately
// instead of holding them until solWorkerIdleTimeout: unauthenticated
// packets must not accumulate one goroutine + buffered channel per claimed
// ID for 90 s each.
func TestSOLQueueRetiresForUnknownSession(t *testing.T) {
	b := bmc.New(bmc.DeviceInfo{}, [16]byte{}, mock.New())
	s := &Server{
		bmc:       b,
		clk:       b.Clock(),
		solQueues: make(map[uint32]chan solJob),
		solDone:   make(chan struct{}),
	}
	s.enqueueSOL(0xdeadbeef, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5000}, []byte{0})

	deadline := time.Now().Add(3 * time.Second)
	for {
		s.solMu.Lock()
		n := len(s.solQueues)
		s.solMu.Unlock()
		if n == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("queue for unknown session retained; forged packets would accumulate workers")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestSOLSessionPacketRejectsV15 verifies a v1.5 LAN packet is never
// classified as SOL: its sequence/session bytes alias the RMCP+ header
// layout (payload type 1 = SOL) and would otherwise be queued and dropped
// instead of executed.
func TestSOLSessionPacketRejectsV15(t *testing.T) {
	// v1.5 MD5-authenticated packet whose session-sequence LSB (byte 5) is
	// 0x01 — ParseRMCPPlusHeader would read it as a SOL payload; the session
	// ID bytes parse non-zero and the fake payload length (session-ID bytes
	// 2-3) is small enough to pass the bounds check.
	pkt := make([]byte, 24)
	pkt[3] = 0x07 // class IPMI
	pkt[4] = 0x01 // auth type MD5 (v1.5)
	pkt[5] = 0x01 // session sequence LSB, aliases payload type SOL
	pkt[7] = 0x10 // session ID (LE), non-zero
	if _, ok := solSessionPacket(pkt); ok {
		t.Fatal("v1.5 packet classified as SOL")
	}

	// A genuine in-session RMCP+ SOL packet still classifies.
	sol := protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeSOL), 0, 0x1234, 1, []byte{0x01})
	id, ok := solSessionPacket(sol)
	if !ok || id != 0x1234 {
		t.Fatalf("RMCP+ SOL packet: ok=%v id=%#x, want true/0x1234", ok, id)
	}
}

package bmc

import (
	"testing"
	"time"
)

func TestV15SessionStore_EvictExpired(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	s := NewV15SessionStore(clk)

	user := &User{ID: 2, Enabled: true}
	var challenge [16]byte
	sess, err := s.CreatePending(V15AuthTypeMD5, user, challenge, 1)
	if err != nil {
		t.Fatalf("CreatePending: %v", err)
	}
	if err := s.Activate(sess, 0xAABBCCDD, 1, 0x01020304, PrivilegeLevelAdministrator); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	clk.now = clk.now.Add(2 * DefaultInactivityTimeout)
	if n := s.EvictExpired(); n != 1 {
		t.Fatalf("expected 1 eviction, got %d", n)
	}
	if _, err := s.Get(0xAABBCCDD); err == nil {
		t.Fatal("session should have been evicted")
	}
}

// TestV15ActivateAcceptsStartingInboundSeq reproduces the ipmitool -I lan stall:
// Activate returns starting inbound seq N, and the console's first authenticated
// packet uses sequence N. That packet must be accepted immediately.
func TestV15ActivateAcceptsStartingInboundSeq(t *testing.T) {
	s := NewV15SessionStore(&mockClock{now: time.Now()})
	user := &User{ID: 2, Enabled: true}
	var challenge [16]byte
	sess, err := s.CreatePending(V15AuthTypeMD5, user, challenge, 1)
	if err != nil {
		t.Fatalf("CreatePending: %v", err)
	}

	const startInbound uint32 = 0xf5f6fd46
	if err := s.Activate(sess, 0x7c5267b7, startInbound, 0x91e741d3, PrivilegeLevelAdministrator); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	got, err := s.Get(0x7c5267b7)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.InboundSeq != startInbound-1 {
		t.Fatalf("InboundSeq high-water: want 0x%08x, got 0x%08x", startInbound-1, got.InboundSeq)
	}
	if !V15InboundSeqValid(got, startInbound) {
		t.Fatalf("starting inbound seq 0x%08x should be valid before first packet", startInbound)
	}
	if !got.TryAcceptInboundSeq(startInbound) {
		t.Fatal("first post-Activate packet with starting inbound seq must be accepted")
	}
	if got.TryAcceptInboundSeq(startInbound) {
		t.Fatal("duplicate starting inbound seq must be rejected")
	}
}

func TestV15ActivateRemapsReservedInboundSeqZero(t *testing.T) {
	s := NewV15SessionStore(&mockClock{now: time.Now()})
	user := &User{ID: 2, Enabled: true}
	var challenge [16]byte
	sess, err := s.CreatePending(V15AuthTypeMD5, user, challenge, 1)
	if err != nil {
		t.Fatalf("CreatePending: %v", err)
	}
	if err := s.Activate(sess, 0x11111111, 0, 0x01020304, PrivilegeLevelAdministrator); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	got, err := s.Get(0x11111111)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	// inboundSeq 0 remapped to 1 → high-water 0; first packet must use 1.
	if got.InboundSeq != 0 {
		t.Fatalf("InboundSeq high-water: want 0, got 0x%08x", got.InboundSeq)
	}
	if !got.TryAcceptInboundSeq(1) {
		t.Fatal("first packet with remapped starting seq 1 must be accepted")
	}
}

func TestV15NextOutboundSeqSkipsZero(t *testing.T) {
	sess := &V15Session{OutboundSeq: 0xffffffff}
	if got := sess.NextOutboundSeq(); got != 0xffffffff {
		t.Fatalf("first: want 0xffffffff, got 0x%08x", got)
	}
	if got := sess.NextOutboundSeq(); got != 1 {
		t.Fatalf("wrap: want 1 (skip 0), got 0x%08x", got)
	}
	if sess.OutboundSeq != 2 {
		t.Fatalf("after wrap: next OutboundSeq want 2, got 0x%08x", sess.OutboundSeq)
	}
}

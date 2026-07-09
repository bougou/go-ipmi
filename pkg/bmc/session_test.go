package bmc

import (
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// mockClock is a simple fixed-time clock for session tests.
type mockClock struct{ now time.Time }

func (m *mockClock) Now() time.Time                         { return m.now }
func (m *mockClock) NewTimer(d time.Duration) clock.Timer   { return clock.Real.NewTimer(d) }
func (m *mockClock) NewTicker(d time.Duration) clock.Ticker { return clock.Real.NewTicker(d) }

func TestSessionStore_AllocateAndGet(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	s := NewSessionStore(clk)

	sess, err := s.Allocate(0xABCD1234, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if sess.BMCID == 0 {
		t.Fatal("expected non-zero BMCID")
	}
	if sess.State != SessionStatePending {
		t.Fatalf("want Pending, got %v", sess.State)
	}

	got, err := s.Get(sess.BMCID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.BMCID != sess.BMCID {
		t.Fatalf("BMCID mismatch: want %d, got %d", sess.BMCID, got.BMCID)
	}
}

func TestSessionStore_Close(t *testing.T) {
	s := NewSessionStore(clock.Real)
	sess, _ := s.Allocate(1, types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)

	if err := s.Close(sess.BMCID); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := s.Get(sess.BMCID); err == nil {
		t.Fatal("expected error after close, got nil")
	}
}

func TestSessionStore_EvictExpired(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	s := NewSessionStoreWithOptions(clk, WithInactivityTimeout(10*time.Second))

	sess, _ := s.Allocate(1, types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)

	// Advance time past the timeout.
	clk.now = clk.now.Add(20 * time.Second)
	n := s.EvictExpired()
	if n != 1 {
		t.Fatalf("expected 1 eviction, got %d", n)
	}
	if _, err := s.Get(sess.BMCID); err == nil {
		t.Fatal("expected session evicted, but Get succeeded")
	}
}

func TestSessionStore_FullEvictsOldestPending(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	s := NewSessionStoreWithOptions(clk, WithMaxSessions(2))

	s1, _ := s.Allocate(1, types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)
	clk.now = clk.now.Add(time.Second)
	_, _ = s.Allocate(2, types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)

	// Third allocation should evict s1 (oldest pending).
	clk.now = clk.now.Add(time.Second)
	_, err := s.Allocate(3, types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None)
	if err != nil {
		t.Fatalf("expected successful allocation after eviction, got: %v", err)
	}
	if _, err := s.Get(s1.BMCID); err == nil {
		t.Fatal("oldest pending session should have been evicted")
	}
}

func TestInboundSeqValid(t *testing.T) {
	tests := []struct {
		last, seq uint32
		want      bool
	}{
		{100, 100, true},
		{100, 115, true},
		{100, 116, false},
		{100, 84, true},
		{100, 83, false},
		{100, 0, false}, // 0 is reserved
	}
	for _, tc := range tests {
		got := InboundSeqValid(tc.last, tc.seq)
		if got != tc.want {
			t.Errorf("InboundSeqValid(%d, %d) = %v, want %v", tc.last, tc.seq, got, tc.want)
		}
	}
}

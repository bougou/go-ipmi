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

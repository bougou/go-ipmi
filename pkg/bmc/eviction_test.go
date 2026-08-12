package bmc

import (
	"testing"
	"time"
)

// TestV15EvictExpiredNoDeadlockUnderHeldLock guards the eviction lock-order
// invariant: eviction never takes a session's ProcMu, so a goroutine holding
// one (as Get Session Challenge -> CreatePending -> eviction does when
// dispatched under the per-session lock) can trigger eviction safely. The
// clock is advanced past the timeout so the held-ProcMu session is the one
// actually being evicted, not merely skipped as fresh.
func TestV15EvictExpiredNoDeadlockUnderHeldLock(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	store := NewV15SessionStore(clk)
	us := NewUserStore()
	u, err := us.Add(2, "admin")
	if err != nil {
		t.Fatal(err)
	}

	sess, err := store.CreatePending(V15AuthTypeMD5, u, [16]byte{}, 1)
	if err != nil {
		t.Fatal(err)
	}

	clk.now = clk.now.Add(DefaultInactivityTimeout + DefaultInactivityTimeoutTolerance + time.Second)

	sess.ProcMu.Lock()
	defer sess.ProcMu.Unlock()

	done := make(chan int)
	go func() {
		done <- store.EvictExpired()
	}()

	select {
	case n := <-done:
		if n != 1 {
			t.Fatalf("want 1 evicted session, got %d", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("DEADLOCK: EvictExpired blocked on a session lock held by the caller")
	}
}

// TestRMCPPlusEvictExpiredNoDeadlockUnderHeldLock is the RMCP+ counterpart:
// the same self-deadlock would exist if the v2.0 store's eviction ever waited
// on ProcMu.
func TestRMCPPlusEvictExpiredNoDeadlockUnderHeldLock(t *testing.T) {
	clk := &mockClock{now: time.Now()}
	store := NewSessionStore(clk)

	sess, err := store.Allocate(0xABCD1234, 0, 0, 0, PrivilegeLevelAdministrator, 1)
	if err != nil {
		t.Fatal(err)
	}

	clk.now = clk.now.Add(DefaultInactivityTimeout + DefaultInactivityTimeoutTolerance + time.Second)

	sess.ProcMu.Lock()
	defer sess.ProcMu.Unlock()

	done := make(chan int)
	go func() {
		done <- store.EvictExpired()
	}()

	select {
	case n := <-done:
		if n != 1 {
			t.Fatalf("want 1 evicted session, got %d", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("DEADLOCK: EvictExpired blocked on a session lock held by the caller")
	}
}

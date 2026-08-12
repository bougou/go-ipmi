package bmc

import (
	"sync"
	"testing"
)

func TestUserStore(t *testing.T) {
	tests := []struct {
		name    string
		run     func(s *UserStore) error
		wantErr bool
	}{
		{
			name: "anonymous user always present",
			run: func(s *UserStore) error {
				_, err := s.Get(1)
				return err
			},
		},
		{
			name: "add user succeeds",
			run: func(s *UserStore) error {
				_, err := s.Add(2, "admin")
				return err
			},
		},
		{
			name: "duplicate name rejected",
			run: func(s *UserStore) error {
				_, _ = s.Add(2, "admin")
				_, err := s.Add(3, "admin")
				return err
			},
			wantErr: true,
		},
		{
			name: "invalid ID rejected",
			run: func(s *UserStore) error {
				_, err := s.Add(0, "x")
				return err
			},
			wantErr: true,
		},
		{
			name: "get by name",
			run: func(s *UserStore) error {
				_, _ = s.Add(2, "alice")
				_, err := s.GetByName("alice")
				return err
			},
		},
		{
			name: "delete user 1 blocked",
			run: func(s *UserStore) error {
				return s.Delete(1)
			},
			wantErr: true,
		},
		{
			name: "delete non-existent returns error",
			run: func(s *UserStore) error {
				return s.Delete(55)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewUserStore()
			err := tc.run(s)
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUserVerifyPassword(t *testing.T) {
	tests := []struct {
		name      string
		stored    string
		candidate string
		want      bool
	}{
		{"matching", "secret", "secret", true},
		{"wrong password", "secret", "wrong", false},
		{"empty vs set", "", "x", false},
		{"both empty", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{}
			u.SetPassword([]byte(tc.stored))
			got := u.VerifyPassword([]byte(tc.candidate))
			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

// TestUserStoreUpsertConcurrent proves two concurrent upserts creating the same
// empty slot do not lose an acknowledged field: one sets the name, the other
// the password, and both survive because each read-modify-write happens under a
// single lock hold. Run under -race. A non-atomic Add-then-Update sequence would
// let the second create overwrite the first's field.
func TestUserStoreUpsertConcurrent(t *testing.T) {
	const id = 5

	for iter := 0; iter < 200; iter++ {
		s := NewUserStore()

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = s.Upsert(id, func(u *User) error { u.Name = "alice"; return nil })
		}()
		go func() {
			defer wg.Done()
			_ = s.Upsert(id, func(u *User) error { u.SetPassword([]byte("secret")); return nil })
		}()
		wg.Wait()

		u, err := s.Get(id)
		if err != nil {
			t.Fatalf("iter %d: slot %d missing after concurrent upsert: %v", iter, id, err)
		}
		if u.Name != "alice" {
			t.Fatalf("iter %d: name lost: got %q, want alice", iter, u.Name)
		}
		if u.Password == [MaxPasswordLen]byte{} {
			t.Fatalf("iter %d: password lost: slot has no password", iter)
		}
	}
}

// TestUserPayloadAccessConcurrent hammers one account's payload access table
// from many goroutines through the store: two sessions authenticated as the
// same user (or a session plus the SOL activation path) can issue Set/Get
// Payload Access concurrently. Writers go through [UserStore.Update] and
// readers through the snapshot copies [UserStore.Get] hands out, which is the
// concurrency contract the server relies on.
func TestUserPayloadAccessConcurrent(t *testing.T) {
	s := NewUserStore()
	if _, err := s.Add(2, "payload"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				u, err := s.Get(2)
				if err != nil {
					t.Error(err)
					return
				}
				u.PayloadAccessFor(1).SOLEnabled()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if err := s.Update(2, func(u *User) error {
					u.SetPayloadAccess(1, true, 2, 0)
					return nil
				}); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	u, err := s.Get(2)
	if err != nil {
		t.Fatal(err)
	}
	if !u.PayloadAccessFor(1).SOLEnabled() {
		t.Fatal("SOL access bit lost under concurrency")
	}
}

// TestGetByNameDeterministicWithEmptyNamedSlots proves the null-user lookup
// always resolves to slot 1 even when runtime user management has created
// additional empty-named slots: an ID-ordered scan, unlike map iteration,
// cannot resolve to a random slot.
func TestGetByNameDeterministicWithEmptyNamedSlots(t *testing.T) {
	s := NewUserStore()
	// Set User Access on an unset slot creates it with an empty name.
	if err := s.Upsert(5, func(u *User) error { return nil }); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		u, err := s.GetByName("")
		if err != nil {
			t.Fatal(err)
		}
		if u.ID != 1 {
			t.Fatalf("iteration %d: empty name resolved to slot %d, want the anonymous user (1)", i, u.ID)
		}
	}
}

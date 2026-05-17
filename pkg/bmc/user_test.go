package bmc

import "testing"

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

package bmc

import (
	"errors"
	"fmt"
	"sync"
)

const (
	MaxUsers       = 63 // IPMI spec allows user IDs 1-63
	MaxUserNameLen = 16
	MaxPasswordLen = 20 // 20 bytes for IPMI 2.0 passwords
)

// ErrUserNotFound is returned when a user ID or name does not exist.
var ErrUserNotFound = errors.New("user not found")

// ErrUsernameTaken is returned when trying to create a user with an already-used name.
var ErrUsernameTaken = errors.New("username already taken")

// ErrInvalidUserID is returned for user IDs outside the valid range 1-63.
var ErrInvalidUserID = errors.New("user ID must be between 1 and 63")

// UserChannelAccess records per-channel privilege settings for a user.
type UserChannelAccess struct {
	// MaxPrivilege is the highest privilege the user may request on this channel.
	MaxPrivilege PrivilegeLevel
	// CallbackOnly restricts the user to callback sessions only.
	CallbackOnly bool
	// Enabled controls whether the user is allowed on this channel at all.
	Enabled bool
}

// User represents a single BMC user account.
type User struct {
	// ID is the IPMI user slot (1-63).  Slot 1 is the anonymous/null user.
	ID   uint8
	Name string
	// Password is stored as a 20-byte padded value per the IPMI 2.0 spec.
	// Index 0 is valid; a zero-length slice means no password is set.
	Password [MaxPasswordLen]byte
	Enabled  bool

	// ChannelAccess holds per-channel access settings keyed by channel number.
	ChannelAccess map[uint8]UserChannelAccess
}

// SetPassword copies up to MaxPasswordLen bytes from raw into the User's password field.
func (u *User) SetPassword(raw []byte) {
	var p [MaxPasswordLen]byte
	copy(p[:], raw)
	u.Password = p
}

// PasswordV15Padded returns the user's password zero-padded to 16 bytes per
// IPMI v1.5 AuthCode algorithms (spec §18.15.1).
func (u *User) PasswordV15Padded() []byte {
	var p [16]byte
	copy(p[:], u.Password[:])
	return p[:]
}

// VerifyPassword returns true when the supplied raw bytes match the stored password.
// Uses constant-time comparison to avoid timing attacks.
func (u *User) VerifyPassword(raw []byte) bool {
	var candidate [MaxPasswordLen]byte
	copy(candidate[:], raw)
	// subtle.ConstantTimeCompare is in crypto/subtle – stdlib.
	return constantTimeEqual(u.Password[:], candidate[:])
}

// UserStore is a thread-safe registry of BMC users.
type UserStore struct {
	mu    sync.RWMutex
	users map[uint8]*User
}

// NewUserStore creates a UserStore with the mandatory anonymous user (ID 1).
func NewUserStore() *UserStore {
	s := &UserStore{users: make(map[uint8]*User, 4)}
	// Slot 1 is always the anonymous/null user per spec section 6.9.1.
	s.users[1] = &User{
		ID:            1,
		Name:          "",
		Enabled:       true,
		ChannelAccess: make(map[uint8]UserChannelAccess),
	}
	return s
}

// Add creates a new user at the given ID.
// Returns [ErrInvalidUserID] for IDs outside 1-63, or [ErrUsernameTaken] if
// name is non-empty and already in use.
func (s *UserStore) Add(id uint8, name string) (*User, error) {
	if id < 1 || id > MaxUsers {
		return nil, ErrInvalidUserID
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if name != "" {
		for _, u := range s.users {
			if u.Name == name {
				return nil, ErrUsernameTaken
			}
		}
	}
	u := &User{
		ID:            id,
		Name:          name,
		Enabled:       false,
		ChannelAccess: make(map[uint8]UserChannelAccess),
	}
	s.users[id] = u
	return u, nil
}

// Get returns the user at the given ID, or [ErrUserNotFound].
func (s *UserStore) Get(id uint8) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d: %w", id, ErrUserNotFound)
	}
	return u, nil
}

// GetByName returns the user with the given name, or [ErrUserNotFound].
// An empty name matches the anonymous user (ID 1).
func (s *UserStore) GetByName(name string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Name == name {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user %q: %w", name, ErrUserNotFound)
}

// FindEnabledByNameOnChannel scans user IDs 1..MaxUsers in order and returns
// the first enabled user with a matching name and channel access (spec §18.14).
func (s *UserStore) FindEnabledByNameOnChannel(name string, channel uint8) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := uint8(1); id <= MaxUsers; id++ {
		u, ok := s.users[id]
		if !ok || !u.Enabled {
			continue
		}
		if u.Name != name {
			continue
		}
		access, ok := u.ChannelAccess[channel]
		if !ok || !access.Enabled {
			continue
		}
		return u, nil
	}
	return nil, fmt.Errorf("user %q on channel %d: %w", name, channel, ErrUserNotFound)
}

// Delete removes a user by ID.  User 1 (anonymous) cannot be deleted.
func (s *UserStore) Delete(id uint8) error {
	if id == 1 {
		return fmt.Errorf("cannot delete anonymous user (ID 1)")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return fmt.Errorf("user %d: %w", id, ErrUserNotFound)
	}
	delete(s.users, id)
	return nil
}

// Count returns the number of configured users.
func (s *UserStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// PrivilegeLevel mirrors ipmi.PrivilegeLevel to avoid an import cycle.
// The server maps these to the main package type before sending responses.
type PrivilegeLevel uint8

const (
	PrivilegeLevelCallback      PrivilegeLevel = 0x01
	PrivilegeLevelUser          PrivilegeLevel = 0x02
	PrivilegeLevelOperator      PrivilegeLevel = 0x03
	PrivilegeLevelAdministrator PrivilegeLevel = 0x04
	PrivilegeLevelOEM           PrivilegeLevel = 0x05
	PrivilegeLevelNoAccess      PrivilegeLevel = 0x0F
)

// constantTimeEqual performs a constant-time comparison of two equal-length slices.
// We avoid importing crypto/subtle here to stay inside stdlib.
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

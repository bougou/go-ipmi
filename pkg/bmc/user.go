package bmc

import (
	"errors"
	"fmt"
	"maps"
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

	// PayloadAccess holds per-channel payload activation rights keyed by
	// channel number (spec v2.0 §24.6/§24.7). Like every other mutable User
	// field it follows the store's snapshot discipline: lookups hand out
	// deep copies, and a runtime change goes through [UserStore.Update].
	PayloadAccess map[uint8]UserPayloadAccess
}

// UserPayloadAccess records a user's payload activation rights on one
// channel, mirroring the bitfields of spec v2.0 Table 24-8/24-9.
type UserPayloadAccess struct {
	// Standard1 mirrors "Standard Payload enables 1": bit [1] = SOL.
	Standard1 uint8
	// OEM1 mirrors "OEM Payload Enables 1".
	OEM1 uint8
}

// defaultStandardPayload1 enables SOL for new entries: payload access
// defaults to allowed, matching how ChannelAccess defaults to enabled.
const defaultStandardPayload1 = 0x02

// SOLEnabled reports whether the user may activate the SOL payload.
func (a UserPayloadAccess) SOLEnabled() bool { return a.Standard1&0x02 != 0 }

// PayloadAccessFor returns the user's payload access entry for channel, or the
// default rights (SOL enabled) when none was ever set.
func (u *User) PayloadAccessFor(channel uint8) UserPayloadAccess {
	if a, ok := u.PayloadAccess[channel]; ok {
		return a
	}
	return UserPayloadAccess{Standard1: defaultStandardPayload1}
}

// SetPayloadAccess applies an enable/disable update to the user's payload
// access entry for channel (spec v2.0 Table 24-8: on enable, 1-bits set and
// 0-bits leave unchanged; on disable, 1-bits clear). Like any other User
// mutation, call it at construction time or on the live user inside a
// [UserStore.Update] callback, never on a store-returned snapshot.
func (u *User) SetPayloadAccess(channel uint8, enable bool, standard1, oem1 uint8) {
	a := u.PayloadAccessFor(channel)
	if enable {
		a.Standard1 |= standard1
		a.OEM1 |= oem1
	} else {
		a.Standard1 &^= standard1
		a.OEM1 &^= oem1
	}
	if u.PayloadAccess == nil {
		u.PayloadAccess = make(map[uint8]UserPayloadAccess)
	}
	u.PayloadAccess[channel] = a
}

// SetPassword copies up to MaxPasswordLen bytes from raw into the User's password field.
func (u *User) SetPassword(raw []byte) {
	var p [MaxPasswordLen]byte
	copy(p[:], raw)
	u.Password = p
}

// PasswordV15Padded returns the user's password zero-padded to 16 bytes per
// IPMI v1.5 AuthCode algorithms (spec v1.5§18.15.1 / v2.0§22.17.1).
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

// copyUser returns a deep copy of u: the struct value (Password is a value
// array, Name a string, both safe to copy) plus fresh access maps. The copy
// shares no mutable state with the stored user, so callers may read it without
// holding any lock.
func copyUser(u *User) *User {
	if u == nil {
		return nil
	}
	cp := *u
	cp.ChannelAccess = maps.Clone(u.ChannelAccess)
	cp.PayloadAccess = maps.Clone(u.PayloadAccess)
	return &cp
}

// Add creates a new user at the given ID and returns the live [*User] for
// construction-time seeding. Mutating the returned pointer is only safe before
// the server starts serving; use [UserStore.Update] for runtime changes.
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

// Get returns a snapshot copy of the user at the given ID, or [ErrUserNotFound].
// The returned [*User] is a private copy; mutate the store via [UserStore.Update].
func (s *UserStore) Get(id uint8) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d: %w", id, ErrUserNotFound)
	}
	return copyUser(u), nil
}

// GetByName returns a snapshot copy of the user with the given name, or
// [ErrUserNotFound]. An empty name matches the anonymous user (ID 1).
func (s *UserStore) GetByName(name string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Name == name {
			return copyUser(u), nil
		}
	}
	return nil, fmt.Errorf("user %q: %w", name, ErrUserNotFound)
}

// FindEnabledByNameOnChannel scans user IDs 1..MaxUsers in order and returns a
// snapshot copy of the first enabled user with a matching name and channel
// access (spec v1.5§18.24 / v2.0§22.27).
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
		return copyUser(u), nil
	}
	return nil, fmt.Errorf("user %q on channel %d: %w", name, channel, ErrUserNotFound)
}

// Update runs fn against the live [*User] for id under the store write lock, the
// race-free way to mutate a user at runtime (e.g. from a Set User Password
// handler). Returns [ErrUserNotFound] if id is not present.
//
// fn runs with the store lock held: it must not call back into the store (that
// self-deadlocks on the non-reentrant lock) and must not retain the [*User]
// past its return, since using the live pointer outside the lock recreates the
// race the snapshot lookups exist to prevent.
func (s *UserStore) Update(id uint8, fn func(*User) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %d: %w", id, ErrUserNotFound)
	}
	return fn(u)
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

// PrivilegeLevel mirrors [types.PrivilegeLevel] so bmc stays free of wire-type
// conversions in session state; handlers map to types before sending responses.
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
// Implemented locally to avoid pulling in crypto/subtle from this package.
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

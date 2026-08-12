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
	// LinkAuth records the link-authentication enable bit (spec v2.0§22.26).
	// Nothing enforces it on a LAN channel; it is stored so Get User Access
	// round-trips what Set User Access accepted.
	LinkAuth bool
}

// User represents a single BMC user account.
type User struct {
	// ID is the IPMI user slot (1-63).  Slot 1 is the anonymous/null user.
	ID   uint8
	Name string
	// Password is stored as a 20-byte padded value per the IPMI 2.0 spec.
	// Index 0 is valid; a zero-length slice means no password is set.
	Password [MaxPasswordLen]byte
	// Password20 records whether the password was stored with the 20-byte size
	// tag (spec v2.0§22.30). The size is part of the credential: a test with the
	// other size must fail, and a 20-byte password exists only in IPMI 2.0, so
	// it must not authenticate a v1.5 session.
	Password20 bool
	Enabled    bool

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

// SetPassword copies up to MaxPasswordLen bytes from raw into the User's
// password field. The size class is inferred from the input length: anything
// longer than 16 bytes is a 20-byte (IPMI 2.0 only) password. Pass the
// wire-tagged length unmodified so the class survives trailing zero bytes.
func (u *User) SetPassword(raw []byte) {
	var p [MaxPasswordLen]byte
	copy(p[:], raw)
	u.Password = p
	u.Password20 = len(raw) > 16
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
	mu sync.RWMutex
	// maxUsers is the highest user ID the store advertises (Get User Access
	// byte 1, spec v2.0§22.27). It bounds the slot range enumerators walk and
	// is immutable after construction, so it needs no locking.
	maxUsers uint8
	users    map[uint8]*User
}

// UserStoreOption configures a [UserStore] at construction time.
type UserStoreOption func(*UserStore)

// WithMaxUsers sets the highest user ID the store advertises via Get User
// Access. n is clamped to the spec range 1..[MaxUsers]; the default is
// [MaxUsers] (63).
func WithMaxUsers(n uint8) UserStoreOption {
	return func(s *UserStore) {
		if n < 1 {
			n = 1
		}
		if n > MaxUsers {
			n = MaxUsers
		}
		s.maxUsers = n
	}
}

// NewUserStore creates a UserStore with the mandatory anonymous user (ID 1).
func NewUserStore(opts ...UserStoreOption) *UserStore {
	s := &UserStore{
		maxUsers: MaxUsers,
		users:    make(map[uint8]*User, 4),
	}
	for _, o := range opts {
		o(s)
	}
	// Slot 1 is always the anonymous/null user per spec section 6.9.1.
	s.users[1] = &User{
		ID:            1,
		Name:          "",
		Enabled:       true,
		ChannelAccess: make(map[uint8]UserChannelAccess),
	}
	return s
}

// MaxUserCount returns the highest user ID the store advertises (default
// [MaxUsers]). Get User Access reports this so in-band enumerators know how many
// slots to walk.
func (s *UserStore) MaxUserCount() uint8 {
	return s.maxUsers
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
// Returns [ErrInvalidUserID] for IDs outside the store's advertised range, or
// [ErrUsernameTaken] if name is non-empty and already in use. Bounding by the
// advertised maximum keeps the whole store consistent: no slot can exist that
// enumerators cannot see but authentication still finds.
func (s *UserStore) Add(id uint8, name string) (*User, error) {
	if id < 1 || id > s.maxUsers {
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
//
// Slots are scanned in ID order, so the lookup is deterministic even when
// several slots share a name: user-management commands can create additional
// empty-named slots at runtime, and iterating the map directly would make the
// RAKP null-user lookup resolve to a random one of them.
func (s *UserStore) GetByName(name string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := uint8(1); id <= s.maxUsers; id++ {
		if u, ok := s.users[id]; ok && u.Name == name {
			return copyUser(u), nil
		}
	}
	return nil, fmt.Errorf("user %q: %w", name, ErrUserNotFound)
}

// FindEnabledByNameOnChannel scans user IDs in order up to the store maximum
// and returns a
// snapshot copy of the first enabled user with a matching name and channel
// access (spec v1.5§18.24 / v2.0§22.27).
func (s *UserStore) FindEnabledByNameOnChannel(name string, channel uint8) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := uint8(1); id <= s.maxUsers; id++ {
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

// Upsert applies fn to the user in slot id under a single write-lock hold,
// creating the slot first (respecting the store max) when it does not exist.
// This is the atomic create-or-mutate path handlers use so two concurrent
// creates on the same empty slot cannot interleave and lose a field, which a
// separate Add-then-Update sequence would allow.
//
// fn runs against a working copy, so a rejected mutation leaves stored state
// untouched: the slot is committed only when fn returns nil and the resulting
// non-empty name does not collide with a different slot. A colliding name is
// rejected with [ErrUsernameTaken] to keep name-based session lookup
// deterministic. Returns [ErrInvalidUserID] for an id outside 1..max.
func (s *UserStore) Upsert(id uint8, fn func(*User) error) error {
	if id < 1 || id > s.maxUsers {
		return ErrInvalidUserID
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	next := &User{ID: id, ChannelAccess: make(map[uint8]UserChannelAccess)}
	if cur, ok := s.users[id]; ok {
		next = copyUser(cur)
	}
	if err := fn(next); err != nil {
		return err
	}
	if next.Name != "" {
		for otherID, u := range s.users {
			if otherID != id && u.Name == next.Name {
				return ErrUsernameTaken
			}
		}
	}
	next.ID = id
	s.users[id] = next
	return nil
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

// CountEnabled returns the number of enabled users, in one pass under one read
// lock. Get User Access reports this on every query, and counting through
// per-slot snapshot lookups would deep-copy the whole table each time.
func (s *UserStore) CountEnabled() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, u := range s.users {
		if u.Enabled {
			n++
		}
	}
	return n
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

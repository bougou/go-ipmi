package handlers

import (
	"bytes"
	"context"
	"errors"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// User-management command bytes (App netfn, spec v2.0§22.26 through §22.30).
// The privilege table (MinimumPrivilege) references these to gate the Set
// commands at Administrator and the Get commands at Operator.
const (
	CmdSetUserAccess   uint8 = 0x43
	CmdGetUserAccess   uint8 = 0x44
	CmdSetUsername     uint8 = 0x45
	CmdGetUsername     uint8 = 0x46
	CmdSetUserPassword uint8 = 0x47
)

// Set User Password operations (spec v2.0§22.30 Table 22-42).
const (
	passwordOpDisableUser  uint8 = 0x00
	passwordOpEnableUser   uint8 = 0x01
	passwordOpSetPassword  uint8 = 0x02
	passwordOpTestPassword uint8 = 0x03
)

// RegisterUserHandlers adds the user-management command set backed by
// [bmc.UserStore] (spec v2.0§22.26 through §22.30).
//
// The behavioral contract is dictated by real in-band consumers such as the
// bougou client's own GetUsers enumerator:
//   - Get User Access must succeed for every slot from 1 to the reported
//     maximum, populated or not, because enumerators abort on any nonzero
//     completion code.
//   - Get User Name on an empty slot must fail with exactly
//     [types.CodeRequestDataFieldInvalid] (0xCC), the code enumerators
//     interpret as "no such user".
func RegisterUserHandlers(r *Registry) {
	r.RegisterFunc(types.CommandGetUserAccess, handleGetUserAccess)
	r.RegisterFunc(types.CommandSetUserAccess, handleSetUserAccess)
	r.RegisterFunc(types.CommandGetUsername, handleGetUsername)
	r.RegisterFunc(types.CommandSetUsername, handleSetUsername)
	r.RegisterFunc(types.CommandSetUserPassword, handleSetUserPassword)
}

// handleGetUserAccess implements Get User Access (App 0x44, spec v2.0§22.27).
func handleGetUserAccess(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 2 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	store := hctx.BMC.Users
	channel := resolveUserChannel(hctx, req[0]&0x0f)
	userID := req[1] & 0x3f
	if !validUserID(store, userID) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// Enabled-user count is reported globally rather than per-channel. Spec
	// §22.27 byte 3 defines it as the number of enabled users on the queried
	// channel; per-channel counting is deferred as no in-band consumer relies on
	// the distinction.
	enabledCount := clampCount(store.CountEnabled())

	// Byte 1 [7:6]: enable status of the queried user. 01b enabled via Set User
	// Password, 10b disabled, 00b unspecified. Default to disabled for an
	// unpopulated slot.
	enableStatus := uint8(0b10)
	// Byte 3 access flags for the queried user's access on this channel.
	access := uint8(0)
	if u, err := store.Get(userID); err == nil {
		if u.Enabled {
			enableStatus = 0b01
		}
		if ca, ok := u.ChannelAccess[channel]; ok {
			if ca.CallbackOnly {
				access |= 1 << 6
			}
			if ca.LinkAuth {
				access |= 1 << 5
			}
			if ca.Enabled {
				access |= 1 << 4 // IPMI messaging enabled
			}
			access |= uint8(ca.MaxPrivilege) & 0x0f
		} else {
			// No access record on this channel: privilege 0x0F ("no access",
			// Table 22-28). 0x00 is reserved and ipmitool renders it as unknown.
			access |= 0x0f
		}
	} else {
		access |= 0x0f
	}

	return []byte{
		store.MaxUserCount() & 0x3f,
		enableStatus<<6 | enabledCount&0x3f,
		// Fixed-name count: only User 1 (the permanently-associated null user,
		// spec §6.9.1) has a name that cannot be reassigned. 1-based per §22.27.
		0x01,
		access,
	}, types.CodeOK, nil
}

// handleSetUserAccess implements Set User Access (App 0x43, spec v2.0§22.26).
func handleSetUserAccess(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 3 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	store := hctx.BMC.Users

	changeAccess := req[0]&(1<<7) != 0
	callbackOnly := req[0]&(1<<6) != 0
	linkAuth := req[0]&(1<<5) != 0
	ipmiMessaging := req[0]&(1<<4) != 0
	channel := resolveUserChannel(hctx, req[0]&0x0f)
	userID := req[1] & 0x3f
	maxPriv := bmc.PrivilegeLevel(req[2] & 0x0f)

	if !validUserID(store, userID) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// Byte 4 (optional): user session limit. Per-user session limits are not
	// supported, and the spec (§22.26) requires 0xCC for a nonzero request
	// rather than acknowledging a setting that has no effect.
	if len(req) >= 4 && req[3]&0x0f != 0 {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// Byte 1 bit 7 gates only the byte-1 access flags (callback / link-auth /
	// IPMI-messaging); the byte-3 privilege limit always applies (spec §22.26).
	// Read-modify-write through the atomic upsert so a clear change bit preserves
	// the existing flags while still updating the privilege limit.
	err := store.Upsert(userID, func(u *bmc.User) error {
		ca := u.ChannelAccess[channel]
		ca.MaxPrivilege = maxPriv
		if changeAccess {
			ca.CallbackOnly = callbackOnly
			ca.LinkAuth = linkAuth
			ca.Enabled = ipmiMessaging
		}
		u.ChannelAccess[channel] = ca
		return nil
	})
	if err != nil {
		return nil, types.CodeUnspecifiedError, err
	}
	return nil, types.CodeOK, nil
}

// handleGetUsername implements Get User Name (App 0x46, spec v2.0§22.29).
func handleGetUsername(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 1 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	store := hctx.BMC.Users
	userID := req[0] & 0x3f
	if !validUserID(store, userID) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// User 1 is the permanent null user (spec §6.9.1): its name is defined as
	// all zeros and reading it succeeds, unlike an unset regular slot.
	if userID == 1 {
		return make([]byte, bmc.MaxUserNameLen), types.CodeOK, nil
	}

	u, err := store.Get(userID)
	if err != nil || u.Name == "" {
		// Empty/unset slot: exactly 0xCC, the code in-band enumerators read as
		// "no such user".
		return nil, types.CodeRequestDataFieldInvalid, nil //nolint:nilerr // reported as a completion code
	}

	name := make([]byte, bmc.MaxUserNameLen)
	copy(name, u.Name)
	return name, types.CodeOK, nil
}

// handleSetUsername implements Set User Name (App 0x45, spec v2.0§22.28).
func handleSetUsername(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 1+bmc.MaxUserNameLen {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	store := hctx.BMC.Users
	userID := req[0] & 0x3f
	name := string(bytes.TrimRight(req[1:1+bmc.MaxUserNameLen], "\x00"))

	if !validUserID(store, userID) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// User 1 is the permanent null user with a fixed all-zeros name
	// (spec §§6.9.1, 22.28), and Get User Access advertises exactly one
	// fixed-name user; renaming it would contradict both.
	if userID == 1 {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// Real BMCs let Set User Name populate a previously empty slot. A name
	// already held by a different slot is rejected so name-based session lookup
	// (RAKP GetByName) stays deterministic.
	err := store.Upsert(userID, func(u *bmc.User) error {
		u.Name = name
		return nil
	})
	if errors.Is(err, bmc.ErrUsernameTaken) {
		return nil, types.CodeRequestDataFieldInvalid, nil //nolint:nilerr // reported as a completion code
	}
	if err != nil {
		return nil, types.CodeUnspecifiedError, err
	}
	return nil, types.CodeOK, nil
}

// handleSetUserPassword implements Set User Password (App 0x47, spec v2.0§22.30).
func handleSetUserPassword(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 2 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	store := hctx.BMC.Users
	userID := req[0] & 0x3f
	stored20 := req[0]&(1<<7) != 0
	operation := req[1] & 0x03

	if !validUserID(store, userID) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	passwordLen := 16
	if stored20 {
		passwordLen = 20
	}

	var password []byte
	if operation == passwordOpSetPassword || operation == passwordOpTestPassword {
		if len(req) < 2+passwordLen {
			return nil, types.CodeRequestDataTruncated, nil
		}
		password = req[2 : 2+passwordLen]
	}

	switch operation {
	case passwordOpTestPassword:
		// Read-only comparison; must not create a slot.
		u, err := store.Get(userID)
		if err != nil {
			return nil, types.CodeRequestDataFieldInvalid, nil //nolint:nilerr // reported as a completion code
		}
		// The size tag is part of the credential (spec §22.30): testing a
		// password with the other size fails with the dedicated code before any
		// byte comparison.
		if stored20 != u.Password20 {
			return nil, types.CodeSetUserPasswordWrongSize, nil
		}
		// Compare against the full stored 20-byte secret (constant time); the
		// store zero-pads both sides to the same length.
		if !u.VerifyPassword(password) {
			return nil, types.CodeSetUserPasswordDataMismatch, nil
		}
		return nil, types.CodeOK, nil

	case passwordOpSetPassword:
		// Only Set Password may create a slot, matching real BMCs. The password
		// is stored with its wire-tagged length so the size class is retained.
		err := store.Upsert(userID, func(u *bmc.User) error {
			u.SetPassword(password)
			return nil
		})
		if err != nil {
			return nil, types.CodeUnspecifiedError, err
		}
		return nil, types.CodeOK, nil

	default: // enable / disable: must not create a slot.
		err := store.Update(userID, func(u *bmc.User) error {
			u.Enabled = operation == passwordOpEnableUser
			return nil
		})
		if errors.Is(err, bmc.ErrUserNotFound) {
			return nil, types.CodeRequestDataFieldInvalid, nil //nolint:nilerr // reported as a completion code
		}
		if err != nil {
			return nil, types.CodeUnspecifiedError, err
		}
		return nil, types.CodeOK, nil
	}
}

// validUserID reports whether id is a usable slot (1..store max).
func validUserID(store *bmc.UserStore, id uint8) bool {
	return id >= 1 && id <= store.MaxUserCount()
}

// resolveUserChannel resolves the channel nibble of a Get/Set User Access
// request. 0x0E ("this channel") resolves to the request's arrival channel when
// known, else the LAN channel, matching how Get Channel Info resolves it.
func resolveUserChannel(hctx *HandlerContext, nibble uint8) uint8 {
	if nibble == types.ChannelNumberSelf {
		if hctx != nil && hctx.Channel != nil {
			return hctx.Channel.Number
		}
		return lanChannelNumber
	}
	return nibble
}

// clampCount saturates a user count to the 6-bit field of the Get User Access
// response.
func clampCount(n int) uint8 {
	if n > 0x3f {
		return 0x3f
	}
	return uint8(n)
}

package handlers

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
)

// MinimumPrivilege returns the lowest session privilege required for (netFn, cmd)
// on the reference server. Unknown commands default to User level.
func MinimumPrivilege(netFn, cmd uint8) bmc.PrivilegeLevel {
	switch netFn {
	case NetFnChassisRequest:
		switch cmd {
		case CmdChassisControl:
			return bmc.PrivilegeLevelOperator
		default:
			return bmc.PrivilegeLevelUser
		}
	case NetFnAppRequest:
		switch cmd {
		case CmdColdReset, CmdWarmReset:
			return bmc.PrivilegeLevelAdministrator
		default:
			return bmc.PrivilegeLevelUser
		}
	default:
		return bmc.PrivilegeLevelUser
	}
}

// V15AllowsAuthTypeNone reports whether an active v1.5 session may accept a
// post-activation packet with AuthType NONE per spec §6.11.4.
func V15AllowsAuthTypeNone(ch *bmc.Channel, netFn, cmd uint8, sess *bmc.V15Session) bool {
	if ch == nil || sess == nil || sess.State != bmc.V15SessionStateActive {
		return false
	}
	if !ch.PerMessageAuth {
		return true
	}
	minPriv := MinimumPrivilege(netFn, cmd)
	if !ch.UserLevelAuth && minPriv <= bmc.PrivilegeLevelUser {
		return sess.PrivilegeLevel >= minPriv
	}
	return false
}

// fillChannelAuthCapsByte4 sets byte 4 of Get Channel Authentication Capabilities
// response (spec Table 22-15) from channel and user configuration.
func fillChannelAuthCapsByte4(resp []byte, b *bmc.BMC, ch *bmc.Channel) {
	if len(resp) < 3 || ch == nil {
		return
	}
	resp[2] = 0x04 // Non-Null usernames enabled
	if b != nil && len(b.KG) > 0 {
		resp[2] |= 0x20
	}
	if !ch.PerMessageAuth {
		resp[2] |= 0x10 // bit 4: per-message authentication disabled
	}
	if !ch.UserLevelAuth {
		resp[2] |= 0x08 // bit 3: user level authentication disabled
	}
	if b != nil {
		resp[2] |= anonymousLoginCapsBits(b)
	}
}

func anonymousLoginCapsBits(b *bmc.BMC) uint8 {
	var flags uint8
	hasNonNull := false
	hasNullName := false
	hasAnonymous := false

	for id := uint8(1); id <= bmc.MaxUsers; id++ {
		user, err := b.Users.Get(id)
		if err != nil || !user.Enabled {
			continue
		}
		if user.Name == "" {
			hasNullName = true
			if isZeroPassword(user.Password) {
				hasAnonymous = true
			}
		} else {
			hasNonNull = true
		}
	}

	if hasNonNull {
		flags |= 0x04
	}
	if hasNullName {
		flags |= 0x02
	}
	if hasAnonymous {
		flags |= 0x01
	}
	return flags
}

func isZeroPassword(p [20]byte) bool {
	for _, b := range p {
		if b != 0 {
			return false
		}
	}
	return true
}

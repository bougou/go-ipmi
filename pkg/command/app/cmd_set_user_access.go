package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.26 Set User Access Command
)

type SetUserAccessRequest struct {
	EnableChanging bool

	RestrictedToCallback bool

	EnableLinkAuth bool

	EnableIPMIMessaging bool

	ChannelNumber uint8

	UserID uint8

	MaxPrivLevel uint8

	SessionLimit uint8
}

type SetUserAccessResponse struct {
}

func (req *SetUserAccessRequest) Command() types.Command {
	return types.CommandSetUserAccess
}

func (req *SetUserAccessRequest) Pack() []byte {
	out := make([]byte, 4)

	b := req.ChannelNumber & 0x0f
	if req.EnableChanging {
		b = types.SetBit7(b)
	}
	if req.RestrictedToCallback {
		b = types.SetBit6(b)
	}
	if req.EnableLinkAuth {
		b = types.SetBit5(b)
	}
	if req.EnableIPMIMessaging {
		b = types.SetBit4(b)
	}
	types.PackUint8(b, out, 0)
	types.PackUint8(req.UserID&0x3f, out, 1)
	types.PackUint8(req.MaxPrivLevel&0x3f, out, 2)
	types.PackUint8(req.SessionLimit&0x0f, out, 3)

	return out
}

func (res *SetUserAccessResponse) CompletionCodes() map[uint8]string {
	// Note: an implementation will not return an error completion code if the user
	// access level is set higher than the privilege limit for a given channel. If it is
	// desired to bring attention to this condition, it is up to software to check the
	// channel privilege limits set using the Set Channel Access command and
	// provide notification of any mismatch.

	return map[uint8]string{}
}

func (res *SetUserAccessResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetUserAccessResponse) Format() string {
	return ""
}

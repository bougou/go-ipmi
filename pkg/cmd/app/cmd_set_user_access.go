package app

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *SetUserAccessRequest) Command() ipmi.Command {
	return ipmi.CommandSetUserAccess
}

func (req *SetUserAccessRequest) Pack() []byte {
	out := make([]byte, 4)

	b := req.ChannelNumber & 0x0f
	if req.EnableChanging {
		b = ipmi.SetBit7(b)
	}
	if req.RestrictedToCallback {
		b = ipmi.SetBit6(b)
	}
	if req.EnableLinkAuth {
		b = ipmi.SetBit5(b)
	}
	if req.EnableIPMIMessaging {
		b = ipmi.SetBit4(b)
	}
	ipmi.PackUint8(b, out, 0)
	ipmi.PackUint8(req.UserID&0x3f, out, 1)
	ipmi.PackUint8(req.MaxPrivLevel&0x3f, out, 2)
	ipmi.PackUint8(req.SessionLimit&0x0f, out, 3)

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

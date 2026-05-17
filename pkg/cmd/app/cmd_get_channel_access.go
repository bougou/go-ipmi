package app

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 22.23 Get Channel Access Command
type GetChannelAccessRequest struct {
	ChannelNumber uint8

	AccessOption ipmi.ChannelAccessOption
}

type GetChannelAccessResponse struct {
	PEFAlertingDisabled   bool
	PerMsgAuthDisabled    bool
	UserLevelAuthDisabled bool
	AccessMode            ipmi.ChannelAccessMode

	MaxPrivilegeLevel ipmi.PrivilegeLevel
}

func (req *GetChannelAccessRequest) Pack() []byte {
	out := make([]byte, 2)

	ipmi.PackUint8(req.ChannelNumber, out, 0)
	ipmi.PackUint8(uint8(req.AccessOption)<<6, out, 1)

	return out
}

func (req *GetChannelAccessRequest) Command() ipmi.Command {
	return ipmi.CommandGetChannelAccess
}

func (res *GetChannelAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x82: "set not supported on selected channel (e.g. channel is session-less.)",
		0x83: "access mode not supported",
	}
}

func (res *GetChannelAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	b0, _, _ := ipmi.UnpackUint8(msg, 0)
	res.PEFAlertingDisabled = ipmi.IsBit5Set(b0)
	res.PerMsgAuthDisabled = ipmi.IsBit4Set(b0)
	res.UserLevelAuthDisabled = ipmi.IsBit3Set(b0)
	res.AccessMode = ipmi.ChannelAccessMode(b0 & 0x07)

	b1, _, _ := ipmi.UnpackUint8(msg, 1)
	res.MaxPrivilegeLevel = ipmi.PrivilegeLevel(b1 & 0x0f)

	return nil
}

func (res *GetChannelAccessResponse) Format() string {
	return "" +
		fmt.Sprintf("    Alerting            : %s\n", ipmi.FormatBool(res.PEFAlertingDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    Per-message Auth    : %s\n", ipmi.FormatBool(res.PerMsgAuthDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    User Level Auth     : %s\n", ipmi.FormatBool(res.UserLevelAuthDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    Access Mode         : %s\n", res.AccessMode) +
		fmt.Sprintf("    Max Privilege Level : %s\n", res.MaxPrivilegeLevel.String())
}

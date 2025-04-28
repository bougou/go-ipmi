package ipmi

import (
	"context"
	"fmt"
)

// 22.23 Get Channel Access Command
type GetChannelAccessRequest struct {
	ChannelNumber uint8

	AccessOption ChannelAccessOption
}

type GetChannelAccessResponse struct {
	PEFAlertingDisabled   bool
	PerMsgAuthDisabled    bool
	UserLevelAuthDisabled bool
	AccessMode            ChannelAccessMode

	MaxPrivilegeLevel PrivilegeLevel
}

func (req *GetChannelAccessRequest) Pack() []byte {
	out := make([]byte, 2)

	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.AccessOption)<<6, out, 1)

	return out
}

func (req *GetChannelAccessRequest) Command() Command {
	return CommandGetChannelAccess
}

func (res *GetChannelAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x82: "set not supported on selected channel (e.g. channel is session-less.)",
		0x83: "access mode not supported",
	}
}

func (res *GetChannelAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	b0, _, _ := unpackUint8(msg, 0)
	res.PEFAlertingDisabled = isBit5Set(b0)
	res.PerMsgAuthDisabled = isBit4Set(b0)
	res.UserLevelAuthDisabled = isBit3Set(b0)
	res.AccessMode = ChannelAccessMode(b0 & 0x07)

	b1, _, _ := unpackUint8(msg, 1)
	res.MaxPrivilegeLevel = PrivilegeLevel(b1 & 0x0f)

	return nil
}

func (res *GetChannelAccessResponse) Format() string {
	return "" +
		fmt.Sprintf("    Alerting            : %s\n", formatBool(res.PEFAlertingDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    Per-message Auth    : %s\n", formatBool(res.PerMsgAuthDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    User Level Auth     : %s\n", formatBool(res.UserLevelAuthDisabled, "disabled", "enabled")) +
		fmt.Sprintf("    Access Mode         : %s\n", res.AccessMode) +
		fmt.Sprintf("    Max Privilege Level : %s\n", res.MaxPrivilegeLevel.String())
}

func (c *Client) GetChannelAccess(ctx context.Context, channelNumber uint8, accessOption ChannelAccessOption) (response *GetChannelAccessResponse, err error) {
	request := &GetChannelAccessRequest{
		ChannelNumber: channelNumber,
		AccessOption:  accessOption,
	}
	response = &GetChannelAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

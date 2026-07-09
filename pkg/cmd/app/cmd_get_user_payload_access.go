package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 24.7 Get User Payload Access Command
type GetUserPayloadAccessRequest struct {
	ChannelNumber uint8

	UserID uint8
}

type GetUserPayloadAccessResponse struct {
	PayloadTypeSOL  bool
	PayloadTypeOEM0 bool
	PayloadTypeOEM1 bool
	PayloadTypeOEM2 bool
	PayloadTypeOEM3 bool
	PayloadTypeOEM4 bool
	PayloadTypeOEM5 bool
	PayloadTypeOEM6 bool
	PayloadTypeOEM7 bool
}

func (req *GetUserPayloadAccessRequest) Pack() []byte {
	out := make([]byte, 2)
	out[0] = req.ChannelNumber

	var b1 = req.UserID & 0x3f
	out[1] = b1

	return out
}

func (req *GetUserPayloadAccessRequest) Command() types.Command {
	return types.CommandGetUserPayloadAccess
}

func (res *GetUserPayloadAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUserPayloadAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	res.PayloadTypeSOL = types.IsBit1Set(msg[0])
	res.PayloadTypeOEM7 = types.IsBit7Set(msg[2])
	res.PayloadTypeOEM6 = types.IsBit6Set(msg[2])
	res.PayloadTypeOEM5 = types.IsBit5Set(msg[2])
	res.PayloadTypeOEM4 = types.IsBit4Set(msg[2])
	res.PayloadTypeOEM3 = types.IsBit3Set(msg[2])
	res.PayloadTypeOEM2 = types.IsBit2Set(msg[2])
	res.PayloadTypeOEM1 = types.IsBit1Set(msg[2])
	res.PayloadTypeOEM0 = types.IsBit0Set(msg[2])

	return nil
}

func (res *GetUserPayloadAccessResponse) Format() string {
	return "" +
		fmt.Sprintf("PayloadTypeSOL  : %v\n", types.FormatBool(res.PayloadTypeSOL, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM0 : %v\n", types.FormatBool(res.PayloadTypeOEM0, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM1 : %v\n", types.FormatBool(res.PayloadTypeOEM1, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM2 : %v\n", types.FormatBool(res.PayloadTypeOEM2, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM3 : %v\n", types.FormatBool(res.PayloadTypeOEM3, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM4 : %v\n", types.FormatBool(res.PayloadTypeOEM4, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM5 : %v\n", types.FormatBool(res.PayloadTypeOEM5, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM6 : %v\n", types.FormatBool(res.PayloadTypeOEM6, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM7 : %v\n", types.FormatBool(res.PayloadTypeOEM7, "enabled", "disabled"))
}

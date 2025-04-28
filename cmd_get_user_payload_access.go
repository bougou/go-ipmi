package ipmi

import (
	"context"
	"fmt"
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

func (req *GetUserPayloadAccessRequest) Command() Command {
	return CommandGetUserPayloadAccess
}

func (res *GetUserPayloadAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUserPayloadAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	res.PayloadTypeSOL = isBit1Set(msg[0])
	res.PayloadTypeOEM7 = isBit7Set(msg[2])
	res.PayloadTypeOEM6 = isBit6Set(msg[2])
	res.PayloadTypeOEM5 = isBit5Set(msg[2])
	res.PayloadTypeOEM4 = isBit4Set(msg[2])
	res.PayloadTypeOEM3 = isBit3Set(msg[2])
	res.PayloadTypeOEM2 = isBit2Set(msg[2])
	res.PayloadTypeOEM1 = isBit1Set(msg[2])
	res.PayloadTypeOEM0 = isBit0Set(msg[2])

	return nil
}

func (res *GetUserPayloadAccessResponse) Format() string {
	return "" +
		fmt.Sprintf("PayloadTypeSOL  : %v\n", formatBool(res.PayloadTypeSOL, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM0 : %v\n", formatBool(res.PayloadTypeOEM0, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM1 : %v\n", formatBool(res.PayloadTypeOEM1, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM2 : %v\n", formatBool(res.PayloadTypeOEM2, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM3 : %v\n", formatBool(res.PayloadTypeOEM3, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM4 : %v\n", formatBool(res.PayloadTypeOEM4, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM5 : %v\n", formatBool(res.PayloadTypeOEM5, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM6 : %v\n", formatBool(res.PayloadTypeOEM6, "enabled", "disabled")) +
		fmt.Sprintf("PayloadTypeOEM7 : %v\n", formatBool(res.PayloadTypeOEM7, "enabled", "disabled"))
}

func (c *Client) GetUserPayloadAccess(ctx context.Context, channelNumber uint8, userID uint8) (response *GetUserPayloadAccessResponse, err error) {
	request := &GetUserPayloadAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &GetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

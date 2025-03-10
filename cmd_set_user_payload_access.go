package ipmi

import (
	"context"
)

// 24.6 Set User Payload Access Command
type SetUserPayloadAccessRequest struct {
	ChannelNumber uint8

	UserID uint8

	Operation SetUserPayloadAccessOperation

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

type SetUserPayloadAccessOperation uint8

const (
	SetUserPayloadAccessOperationEnable  SetUserPayloadAccessOperation = 0x00
	SetUserPayloadAccessOperationDisable SetUserPayloadAccessOperation = 0x01
)

type SetUserPayloadAccessResponse struct {
}

func (req *SetUserPayloadAccessRequest) Pack() []byte {
	out := make([]byte, 6)
	out[0] = req.ChannelNumber

	var b1 = uint8(req.Operation) << 6
	b1 |= req.UserID & 0x3f
	out[1] = b1

	var b2 uint8
	b2 = setOrClearBit1(b2, req.PayloadTypeSOL)
	out[2] = b2

	out[3] = 0

	var b4 uint8
	b4 = setOrClearBit7(b4, req.PayloadTypeOEM7)
	b4 = setOrClearBit6(b4, req.PayloadTypeOEM6)
	b4 = setOrClearBit5(b4, req.PayloadTypeOEM5)
	b4 = setOrClearBit4(b4, req.PayloadTypeOEM4)
	b4 = setOrClearBit3(b4, req.PayloadTypeOEM3)
	b4 = setOrClearBit2(b4, req.PayloadTypeOEM2)
	b4 = setOrClearBit1(b4, req.PayloadTypeOEM1)
	b4 = setOrClearBit0(b4, req.PayloadTypeOEM0)
	out[4] = b4

	out[5] = 0

	return out
}

func (req *SetUserPayloadAccessRequest) Command() Command {
	return CommandSetUserPayloadAccess
}

func (res *SetUserPayloadAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetUserPayloadAccessResponse) Unpack(msg []byte) error {

	return nil
}

func (res *SetUserPayloadAccessResponse) Format() string {
	return ""
}

func (c *Client) SetUserPayloadAccess(ctx context.Context, payloadType PayloadType, payloadInstance uint8) (response *SetUserPayloadAccessResponse, err error) {
	request := &SetUserPayloadAccessRequest{}
	response = &SetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

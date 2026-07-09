package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
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
	b2 = types.SetOrClearBit1(b2, req.PayloadTypeSOL)
	out[2] = b2

	out[3] = 0

	var b4 uint8
	b4 = types.SetOrClearBit7(b4, req.PayloadTypeOEM7)
	b4 = types.SetOrClearBit6(b4, req.PayloadTypeOEM6)
	b4 = types.SetOrClearBit5(b4, req.PayloadTypeOEM5)
	b4 = types.SetOrClearBit4(b4, req.PayloadTypeOEM4)
	b4 = types.SetOrClearBit3(b4, req.PayloadTypeOEM3)
	b4 = types.SetOrClearBit2(b4, req.PayloadTypeOEM2)
	b4 = types.SetOrClearBit1(b4, req.PayloadTypeOEM1)
	b4 = types.SetOrClearBit0(b4, req.PayloadTypeOEM0)
	out[4] = b4

	out[5] = 0

	return out
}

func (req *SetUserPayloadAccessRequest) Command() types.Command {
	return types.CommandSetUserPayloadAccess
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

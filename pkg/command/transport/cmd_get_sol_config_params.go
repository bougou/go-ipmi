package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 26.3 Get SOL Configuration Parameters Command
type GetSOLConfigParamRequest struct {
	GetParamRevisionOnly bool
	ChannelNumber        uint8
	ParamSelector        types.SOLConfigParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSOLConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSOLConfigParamRequest) Command() types.Command {
	return types.CommandGetSOLConfigParam
}

func (req *GetSOLConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	b := req.ChannelNumber
	if req.GetParamRevisionOnly {
		b = types.SetBit7(b)
	}

	types.PackUint8(b, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackUint8(req.SetSelector, out, 2)
	types.PackUint8(req.BlockSelector, out, 3)
	return out
}

func (res *GetSOLConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSOLConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision = msg[0]
	if len(msg) > 1 {
		res.ParamData, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSOLConfigParamResponse) Format() string {
	return ""
}

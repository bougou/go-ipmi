package transport

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 26.3 Get SOL Configuration Parameters Command
type GetSOLConfigParamRequest struct {
	GetParamRevisionOnly bool
	ChannelNumber        uint8
	ParamSelector        ipmi.SOLConfigParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSOLConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSOLConfigParamRequest) Command() ipmi.Command {
	return ipmi.CommandGetSOLConfigParam
}

func (req *GetSOLConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	b := req.ChannelNumber
	if req.GetParamRevisionOnly {
		b = ipmi.SetBit7(b)
	}

	ipmi.PackUint8(b, out, 0)
	ipmi.PackUint8(uint8(req.ParamSelector), out, 1)
	ipmi.PackUint8(req.SetSelector, out, 2)
	ipmi.PackUint8(req.BlockSelector, out, 3)
	return out
}

func (res *GetSOLConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSOLConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision = msg[0]
	if len(msg) > 1 {
		res.ParamData, _, _ = ipmi.UnpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSOLConfigParamResponse) Format() string {
	return ""
}

package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.14b Get System Info Parameters Command
type GetSystemInfoParamRequest struct {
	GetParamRevisionOnly bool
	ParamSelector        types.SystemInfoParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSystemInfoParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSystemInfoParamRequest) Pack() []byte {
	out := make([]byte, 4)

	var b uint8
	b = types.SetOrClearBit7(b, req.GetParamRevisionOnly)
	out[0] = b

	out[1] = uint8(req.ParamSelector)
	out[2] = req.SetSelector
	out[3] = req.BlockSelector

	return out
}

func (req *GetSystemInfoParamRequest) Command() types.Command {
	return types.CommandGetSystemInfoParam
}

func (res *GetSystemInfoParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemInfoParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision, _, _ = types.UnpackUint8(msg, 0)
	if len(msg) > 1 {
		res.ParamData, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSystemInfoParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Param Revision : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param Data     : %v\n", res.ParamData)
}

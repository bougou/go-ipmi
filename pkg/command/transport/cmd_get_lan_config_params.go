package transport

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 23.2 Get LAN Configuration Parameters Command
type GetLanConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector types.LanConfigParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

type GetLanConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetLanConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	types.PackUint8(req.ChannelNumber, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackUint8(req.SetSelector, out, 2)
	types.PackUint8(req.BlockSelector, out, 3)
	return out
}

func (req *GetLanConfigParamRequest) Command() types.Command {
	return types.CommandGetLanConfigParam
}

func (res *GetLanConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *GetLanConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ParamRevision, _, _ = types.UnpackUint8(msg, 0)
	res.ParamData, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetLanConfigParamResponse) Format() string {

	return "" +
		fmt.Sprintf("Parameter Revision    : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param Data            : %v\n", res.ParamData) +
		fmt.Sprintf("Length of Config Data : %d\n", len(res.ParamData))
}

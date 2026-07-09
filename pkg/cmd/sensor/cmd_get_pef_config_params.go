package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParamRequest struct {
	// [7] - 1b = get parameter revision only. 0b = get parameter
	// [6:0] - Parameter selector
	GetParamRevisionOnly bool
	ParamSelector        types.PEFConfigParamSelector

	SetSelector   uint8 // 00h if parameter does not require a Set Selector
	BlockSelector uint8 // 00h if parameter does not require a block number
}

type GetPEFConfigParamResponse struct {
	// Parameter revision.
	//
	// Format:
	//  - MSN = present revision.
	//  - LSN = oldest revision parameter is backward compatible with.
	//  - 11h for parameters in this specification.
	ParamRevision uint8

	// ParamData not returned when GetParamRevisionOnly is true
	ParamData []byte
}

func (req *GetPEFConfigParamRequest) Command() types.Command {
	return types.CommandGetPEFConfigParam
}

func (req *GetPEFConfigParamRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	b0 := uint8(req.ParamSelector) & 0x7f
	if req.GetParamRevisionOnly {
		b0 = types.SetBit7(b0)
	}
	types.PackUint8(b0, out, 0)
	types.PackUint8(req.SetSelector, out, 1)
	types.PackUint8(req.BlockSelector, out, 2)

	return out
}

func (res *GetPEFConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShort
	}

	res.ParamRevision = msg[0]

	if len(msg) > 1 {
		res.ParamData, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Parameter Revision           : %#02x (%d)\n", res.ParamRevision, res.ParamRevision) +
		fmt.Sprintf("Configuration Parameter Data : %# 02x\n", res.ParamData)
}

// GroupControlsCount:  &PEFConfigParam_GroupControlsCount{},
// GroupControls:       []*PEFConfigParam_GroupControl{},

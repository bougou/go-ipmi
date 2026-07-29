package chassis

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.13 Get System Boot Options Command
type GetSystemBootOptionsParamRequest struct {
	ParamSelector types.BootOptionParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

// Table 28-14, Boot Option Parameters

type GetSystemBootOptionsParamResponse struct {
	ParameterVersion uint8

	// [7] - 1b = mark parameter invalid / locked
	// 0b = mark parameter valid / unlocked
	ParameterInValid bool
	// [6:0] - boot option parameter selector
	ParamSelector types.BootOptionParamSelector

	ParamData []byte // origin parameter data
}

func (req *GetSystemBootOptionsParamRequest) Pack() []byte {
	out := make([]byte, 3)
	types.PackUint8(uint8(req.ParamSelector), out, 0)
	types.PackUint8(req.SetSelector, out, 1)
	types.PackUint8(req.BlockSelector, out, 2)
	return out
}

func (req *GetSystemBootOptionsParamRequest) Command() types.Command {
	return types.CommandGetSystemBootOptions
}

func (res *GetSystemBootOptionsParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemBootOptionsParamResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}
	res.ParameterVersion, _, _ = types.UnpackUint8(msg, 0)
	b, _, _ := types.UnpackUint8(msg, 1)
	res.ParameterInValid = types.IsBit7Set(b)
	res.ParamSelector = types.BootOptionParamSelector(b & 0x7f) // clear bit 7

	res.ParamData, _, _ = types.UnpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSystemBootOptionsParamResponse) Format() string {

	var paramDataFormatted string

	var param types.BootOptionParameter

	switch res.ParamSelector {
	case types.BootOptionParamSelector_SetInProgress:
		param = &types.BootOptionParam_SetInProgress{}
	case types.BootOptionParamSelector_ServicePartitionSelector:
		param = &types.BootOptionParam_ServicePartitionSelector{}
	case types.BootOptionParamSelector_ServicePartitionScan:
		param = &types.BootOptionParam_ServicePartitionScan{}
	case types.BootOptionParamSelector_BMCBootFlagValidBitClear:
		param = &types.BootOptionParam_BMCBootFlagValidBitClear{}
	case types.BootOptionParamSelector_BootInfoAcknowledge:
		param = &types.BootOptionParam_BootInfoAcknowledge{}
	case types.BootOptionParamSelector_BootFlags:
		param = &types.BootOptionParam_BootFlags{}
	case types.BootOptionParamSelector_BootInitiatorInfo:
		param = &types.BootOptionParam_BootInitiatorInfo{}
	case types.BootOptionParamSelector_BootInitiatorMailbox:
		param = &types.BootOptionParam_BootInitiatorMailbox{}
	}

	if param != nil {
		if err := param.Unpack(res.ParamData); err == nil {
			paramDataFormatted = param.Format()
		}
	}

	return "" +
		fmt.Sprintf("Boot parameter version : %d\n", res.ParameterVersion) +
		fmt.Sprintf("Boot parameter %d is %s\n", res.ParamSelector, types.FormatBool(res.ParameterInValid, "invalid/locked", "valid/unlocked")) +
		fmt.Sprintf("Boot parameter data : %02x\n", res.ParamData) +
		fmt.Sprintf("%s : %s\n", res.ParamSelector.String(), paramDataFormatted)
}

// GetSystemBootOptionsParams get all parameters of boot options.

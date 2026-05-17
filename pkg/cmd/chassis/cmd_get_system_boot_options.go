package chassis

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 28.13 Get System Boot Options Command
type GetSystemBootOptionsParamRequest struct {
	ParamSelector ipmi.BootOptionParamSelector
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
	ParamSelector ipmi.BootOptionParamSelector

	ParamData []byte // origin parameter data
}

func (req *GetSystemBootOptionsParamRequest) Pack() []byte {
	out := make([]byte, 3)
	ipmi.PackUint8(uint8(req.ParamSelector), out, 0)
	ipmi.PackUint8(req.SetSelector, out, 1)
	ipmi.PackUint8(req.BlockSelector, out, 2)
	return out
}

func (req *GetSystemBootOptionsParamRequest) Command() ipmi.Command {
	return ipmi.CommandGetSystemBootOptions
}

func (res *GetSystemBootOptionsParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemBootOptionsParamResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 3)
	}
	res.ParameterVersion, _, _ = ipmi.UnpackUint8(msg, 0)
	b, _, _ := ipmi.UnpackUint8(msg, 1)
	res.ParameterInValid = ipmi.IsBit7Set(b)
	res.ParamSelector = ipmi.BootOptionParamSelector(b & 0x7f) // clear bit 7

	res.ParamData, _, _ = ipmi.UnpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSystemBootOptionsParamResponse) Format() string {

	var paramDataFormatted string

	var param ipmi.BootOptionParameter

	switch res.ParamSelector {
	case ipmi.BootOptionParamSelector_SetInProgress:
		param = &ipmi.BootOptionParam_SetInProgress{}
	case ipmi.BootOptionParamSelector_ServicePartitionSelector:
		param = &ipmi.BootOptionParam_ServicePartitionSelector{}
	case ipmi.BootOptionParamSelector_ServicePartitionScan:
		param = &ipmi.BootOptionParam_ServicePartitionScan{}
	case ipmi.BootOptionParamSelector_BMCBootFlagValidBitClear:
		param = &ipmi.BootOptionParam_BMCBootFlagValidBitClear{}
	case ipmi.BootOptionParamSelector_BootInfoAcknowledge:
		param = &ipmi.BootOptionParam_BootInfoAcknowledge{}
	case ipmi.BootOptionParamSelector_BootFlags:
		param = &ipmi.BootOptionParam_BootFlags{}
	case ipmi.BootOptionParamSelector_BootInitiatorInfo:
		param = &ipmi.BootOptionParam_BootInitiatorInfo{}
	case ipmi.BootOptionParamSelector_BootInitiatorMailbox:
		param = &ipmi.BootOptionParam_BootInitiatorMailbox{}
	}

	if param != nil {
		if err := param.Unpack(res.ParamData); err == nil {
			paramDataFormatted = param.Format()
		}
	}

	return "" +
		fmt.Sprintf("Boot parameter version : %d\n", res.ParameterVersion) +
		fmt.Sprintf("Boot parameter %d is %s\n", res.ParamSelector, ipmi.FormatBool(res.ParameterInValid, "invalid/locked", "valid/unlocked")) +
		fmt.Sprintf("Boot parameter data : %02x\n", res.ParamData) +
		fmt.Sprintf("%s : %s\n", res.ParamSelector.String(), paramDataFormatted)
}

// GetSystemBootOptionsParams get all parameters of boot options.

package dcmi

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5] 6.1.3 Get DCMI Configuration Parameters Command
type GetDCMIConfigParamRequest struct {
	ParamSelector types.DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
}

type GetDCMIConfigParamResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMIConfigParamRequest) Pack() []byte {
	out := make([]byte, 3)

	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackUint8(req.SetSelector, out, 2)

	return out
}

func (req *GetDCMIConfigParamRequest) Command() types.Command {
	return types.CommandGetDCMIConfigParam
}

func (res *GetDCMIConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.MajorVersion, _, _ = types.UnpackUint8(msg, 1)
	res.MinorVersion, _, _ = types.UnpackUint8(msg, 2)
	res.ParamRevision, _, _ = types.UnpackUint8(msg, 3)
	res.ParamData, _, _ = types.UnpackBytes(msg, 4, len(msg)-4)

	return nil
}

func (res *GetDCMIConfigParamResponse) Format() string {
	return ""
}

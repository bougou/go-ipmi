package dcmi

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5] 6.1.2 Set DCMI Configuration Parameters
type SetDCMIConfigParamRequest struct {
	ParamSelector types.DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
	ParamData     []byte
}

type SetDCMIConfigParamResponse struct {
}

func (req *SetDCMIConfigParamRequest) Pack() []byte {
	out := make([]byte, 3+len(req.ParamData))

	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackUint8(req.SetSelector, out, 2)
	types.PackBytes(req.ParamData, out, 3)

	return out

}

func (req *SetDCMIConfigParamRequest) Command() types.Command {
	return types.CommandSetDCMIConfigParam
}

func (res *SetDCMIConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIConfigParamResponse) Format() string {
	return ""
}

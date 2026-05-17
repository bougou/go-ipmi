package dcmi

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5] 6.1.2 Set DCMI Configuration Parameters
type SetDCMIConfigParamRequest struct {
	ParamSelector ipmi.DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
	ParamData     []byte
}

type SetDCMIConfigParamResponse struct {
}

func (req *SetDCMIConfigParamRequest) Pack() []byte {
	out := make([]byte, 3+len(req.ParamData))

	ipmi.PackUint8(ipmi.GroupExtensionDCMI, out, 0)
	ipmi.PackUint8(uint8(req.ParamSelector), out, 1)
	ipmi.PackUint8(req.SetSelector, out, 2)
	ipmi.PackBytes(req.ParamData, out, 3)

	return out

}

func (req *SetDCMIConfigParamRequest) Command() ipmi.Command {
	return ipmi.CommandSetDCMIConfigParam
}

func (res *SetDCMIConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := ipmi.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIConfigParamResponse) Format() string {
	return ""
}

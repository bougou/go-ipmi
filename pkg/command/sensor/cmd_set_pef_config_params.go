package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 30.3 Set PEF Configuration Parameters Command
)

type SetPEFConfigParamRequest struct {
	ParamSelector types.PEFConfigParamSelector
	ParamData     []byte
}

type SetPEFConfigParamResponse struct {
	// empty
}

func (req *SetPEFConfigParamRequest) Command() types.Command {
	return types.CommandSetPEFConfigParam
}

func (req *SetPEFConfigParamRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 1+len(req.ParamData))

	// out[0] = req.ParamSelector
	types.PackUint8(uint8(req.ParamSelector), out, 0)
	if len(req.ParamData) > 0 {
		types.PackBytes(req.ParamData, out, 1)
	}
	return out
}

func (res *SetPEFConfigParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetPEFConfigParamResponse) Format() string {
	return ""
}

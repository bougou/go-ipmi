package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 26.2 Set SOL Configuration Parameters Command
type SetSOLConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector types.SOLConfigParamSelector
	ParamData     []byte
}

type SetSOLConfigParamResponse struct {
}

func (req *SetSOLConfigParamRequest) Command() types.Command {
	return types.CommandSetSOLConfigParam
}

func (req *SetSOLConfigParamRequest) Pack() []byte {
	out := make([]byte, 2+len(req.ParamData))
	types.PackUint8(req.ChannelNumber, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackBytes(req.ParamData, out, 2)
	return out
}

func (res *SetSOLConfigParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSOLConfigParamResponse) Format() string {
	return ""
}

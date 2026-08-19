package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.14a Set System Info Parameters Command
type SetSystemInfoParamRequest struct {
	ParamSelector types.SystemInfoParamSelector
	ParamData     []byte
}

type SetSystemInfoParamResponse struct {
}

func (req *SetSystemInfoParamRequest) Pack() []byte {
	out := make([]byte, 1+len(req.ParamData))
	out[0] = byte(req.ParamSelector)
	types.PackBytes(req.ParamData, out, 1)
	return out
}

func (req *SetSystemInfoParamRequest) Command() types.Command {
	return types.CommandSetSystemInfoParam
}

func (res *SetSystemInfoParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemInfoParamResponse) Format() string {
	return ""
}

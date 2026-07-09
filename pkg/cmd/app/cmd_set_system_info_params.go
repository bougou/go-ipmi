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

func (res *SetSystemInfoParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
	}
}

func (res *SetSystemInfoParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemInfoParamResponse) Format() string {
	return ""
}

package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 23.1 Set LAN Configuration Parameters Command
type SetLanConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector types.LanConfigParamSelector
	ParamData     []byte
}

type SetLanConfigParamResponse struct {
	// empty
}

func (req *SetLanConfigParamRequest) Pack() []byte {
	out := make([]byte, 2+len(req.ParamData))

	types.PackUint8(req.ChannelNumber, out, 0)
	types.PackUint8(uint8(req.ParamSelector), out, 1)
	types.PackBytes(req.ParamData, out, 2)

	return out
}

func (req *SetLanConfigParamRequest) Command() types.Command {
	return types.CommandSetLanConfigParam
}

func (res *SetLanConfigParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetLanConfigParamResponse) Format() string {
	return ""
}

package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.12 Set System Boot Options Command. Sets parameters that direct system boot
// following power up or reset. Boot flags apply for one restart; system BIOS
// should read them from the BMC and then clear the flags.
type SetSystemBootOptionsParamRequest struct {
	// Parameter valid
	//  - 1b = mark parameter invalid / locked
	//  - 0b = mark parameter valid / unlocked
	MarkParameterInvalid bool
	// [6:0] - boot option parameter selector
	ParamSelector types.BootOptionParamSelector

	ParamData []byte
}

// Table 28-14, Boot Option Parameters

type SetSystemBootOptionsParamResponse struct {
}

func (req *SetSystemBootOptionsParamRequest) Pack() []byte {

	out := make([]byte, 1+len(req.ParamData))

	b := uint8(req.ParamSelector)
	if req.MarkParameterInvalid {
		b = types.SetBit7(b)
	} else {
		b = types.ClearBit7(b)
	}
	types.PackUint8(b, out, 0)

	types.PackBytes(req.ParamData, out, 1)

	return out
}

func (req *SetSystemBootOptionsParamRequest) Command() types.Command {
	return types.CommandSetSystemBootOptions
}

func (res *SetSystemBootOptionsParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemBootOptionsParamResponse) Format() string {
	return ""
}

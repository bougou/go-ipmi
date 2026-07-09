package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.12 Set System Boot Options Command
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

func (res *SetSystemBootOptionsParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state. (This completion code provides a way to recognize that another party has already 'claimed' the parameters)",
		0x82: "attempt to write read-only parameter",
	}
}

func (res *SetSystemBootOptionsParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemBootOptionsParamResponse) Format() string {
	return ""
}

// This command is used to set parameters that direct the system boot following a system power up or reset.
// The boot flags only apply for one system restart. It is the responsibility of the system BIOS
// to read these settings from the BMC and then clear the boot flags

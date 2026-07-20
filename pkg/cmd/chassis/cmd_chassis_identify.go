package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.5 Chassis Identify Command. Causes the chassis to physically identify itself
// (blinking lights, beeps, LCD panel, etc.).
type ChassisIdentifyRequest struct {
	IdentifyInterval uint8
	ForceIdentifyOn  bool
}

type ChassisIdentifyResponse struct {
	// empty
}

func (req *ChassisIdentifyRequest) Pack() []byte {
	out := make([]byte, 2)
	types.PackUint8(uint8(req.IdentifyInterval), out, 0)

	var force uint8 = 0
	if req.ForceIdentifyOn {
		force = 1
	}
	types.PackUint8(force, out, 1)
	return out
}

func (req *ChassisIdentifyRequest) Command() types.Command {
	return types.CommandChassisIdentify
}

func (res *ChassisIdentifyResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ChassisIdentifyResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ChassisIdentifyResponse) Format() string {
	return ""
}

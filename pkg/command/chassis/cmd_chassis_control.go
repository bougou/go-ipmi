package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

type ChassisControl uint8

const (
	ChassisControlPowerDown           ChassisControl = 0 // down, off
	ChassisControlPowerUp             ChassisControl = 1
	ChassisControlPowerCycle          ChassisControl = 2
	ChassisControlHardReset           ChassisControl = 3
	ChassisControlDiagnosticInterrupt ChassisControl = 4
	ChassisControlSoftShutdown        ChassisControl = 5
)

// 28.3 Chassis Control Command
type ChassisControlRequest struct {
	ChassisControl ChassisControl
}

type ChassisControlResponse struct {
}

func (req *ChassisControlRequest) Pack() []byte {
	out := make([]byte, 1)
	types.PackUint8(uint8(req.ChassisControl), out, 0)
	return out
}

// Unpack parses a Chassis Control request body (§28.3): a single byte whose
// lower nibble is the action. Upper nibble is reserved and ignored.
func (req *ChassisControlRequest) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	req.ChassisControl = ChassisControl(msg[0] & 0x0F)
	return nil
}

func (req *ChassisControlRequest) Command() types.Command {
	return types.CommandChassisControl
}

func (res *ChassisControlResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ChassisControlResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ChassisControlResponse) Format() string {
	return ""
}

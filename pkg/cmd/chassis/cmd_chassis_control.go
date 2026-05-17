package chassis

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
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
	ipmi.PackUint8(uint8(req.ChassisControl), out, 0)
	return out
}

func (req *ChassisControlRequest) Command() ipmi.Command {
	return ipmi.CommandChassisControl
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

package chassis

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 28.9 Set Power Cycle Interval
)

type SetPowerCycleIntervalRequest struct {
	IntervalInSec uint8
}

type SetPowerCycleIntervalResponse struct {
	// empty
}

func (req *SetPowerCycleIntervalRequest) Pack() []byte {
	out := make([]byte, 1)
	ipmi.PackUint8(req.IntervalInSec, out, 0)
	return out
}

func (req *SetPowerCycleIntervalRequest) Command() ipmi.Command {
	return ipmi.CommandSetPowerCycleInterval
}

func (res *SetPowerCycleIntervalResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetPowerCycleIntervalResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetPowerCycleIntervalResponse) Format() string {
	return ""
}

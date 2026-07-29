package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
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
	types.PackUint8(req.IntervalInSec, out, 0)
	return out
}

func (req *SetPowerCycleIntervalRequest) Command() types.Command {
	return types.CommandSetPowerCycleInterval
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

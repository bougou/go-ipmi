package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 27.6 Set Watchdog Timer Command
)

type SetWatchdogTimerRequest struct {
	DontLog       bool
	DontStopTimer bool
	TimerUse      TimerUse

	PreTimeoutInterrupt   PreTimeoutInterrupt
	TimeoutAction         TimeoutAction
	PreTimeoutIntervalSec uint8

	ExpirationFlags  uint8
	InitialCountdown uint16
}

type SetWatchdogTimerResponse struct {
}

func (req *SetWatchdogTimerRequest) Pack() []byte {
	out := make([]byte, 6)

	b0 := uint8(req.TimerUse)
	if req.DontLog {
		b0 = types.SetBit7(b0)
	}
	if req.DontStopTimer {
		b0 = types.SetBit6(b0)
	}
	types.PackUint8(b0, out, 0)

	b1 := uint8(req.TimeoutAction)
	b1 |= uint8(req.PreTimeoutInterrupt) << 4
	types.PackUint8(b1, out, 1)

	types.PackUint8(req.PreTimeoutIntervalSec, out, 2)
	types.PackUint8(req.ExpirationFlags, out, 3)
	types.PackUint16L(req.InitialCountdown, out, 4)

	return out
}

func (req *SetWatchdogTimerRequest) Command() types.Command {
	return types.CommandSetWatchdogTimer
}

func (res *SetWatchdogTimerResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetWatchdogTimerResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *SetWatchdogTimerResponse) Format() string {
	return ""
}

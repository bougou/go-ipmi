package ipmi

import "context"

// 27.6 Set Watchdog Timer Command
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
		b0 = setBit7(b0)
	}
	if req.DontStopTimer {
		b0 = setBit6(b0)
	}
	packUint8(b0, out, 0)

	b1 := uint8(req.TimeoutAction)
	b1 |= uint8(req.PreTimeoutInterrupt) << 4
	packUint8(b1, out, 1)

	packUint8(req.PreTimeoutIntervalSec, out, 2)
	packUint8(req.ExpirationFlags, out, 3)
	packUint16L(req.InitialCountdown, out, 4)

	return out
}

func (req *SetWatchdogTimerRequest) Command() Command {
	return CommandSetWatchdogTimer
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

func (c *Client) SetWatchdogTimer(ctx context.Context) (response *SetWatchdogTimerResponse, err error) {
	request := &SetWatchdogTimerRequest{}
	response = &SetWatchdogTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

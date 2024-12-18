package ipmi

import "context"

// 28.9 Set Power Cycle Interval
type SetPowerCycleIntervalRequest struct {
	IntervalInSec uint8
}

type SetPowerCycleIntervalResponse struct {
	// empty
}

func (req *SetPowerCycleIntervalRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.IntervalInSec, out, 0)
	return out
}

func (req *SetPowerCycleIntervalRequest) Command() Command {
	return CommandSetPowerCycleInterval
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

func (c *Client) SetPowerCycleInterval(ctx context.Context, intervalInSec uint8) (response *SetPowerCycleIntervalResponse, err error) {
	request := &SetPowerCycleIntervalRequest{
		IntervalInSec: intervalInSec,
	}
	response = &SetPowerCycleIntervalResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package ipmi

import "context"

// 27.5 Reset Watchdog Timer Command
type ResetWatchdogTimerRequest struct {
}

type ResetWatchdogTimerResponse struct {
}

func (req *ResetWatchdogTimerRequest) Pack() []byte {
	return []byte{}
}

func (req *ResetWatchdogTimerRequest) Command() Command {
	return CommandResetWatchdogTimer
}

func (res *ResetWatchdogTimerResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ResetWatchdogTimerResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{
		0x80: "Attempt to start un-initialized watchdog",
	}
}

func (res *ResetWatchdogTimerResponse) Format() string {
	return ""
}

func (c *Client) ResetWatchdogTimer(ctx context.Context) (response *ResetWatchdogTimerResponse, err error) {
	request := &ResetWatchdogTimerRequest{}
	response = &ResetWatchdogTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

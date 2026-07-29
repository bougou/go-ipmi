package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 27.5 Reset Watchdog Timer Command
)

type ResetWatchdogTimerRequest struct {
	// empty
}

type ResetWatchdogTimerResponse struct {
}

func (req *ResetWatchdogTimerRequest) Pack() []byte {
	return []byte{}
}

func (req *ResetWatchdogTimerRequest) Command() types.Command {
	return types.CommandResetWatchdogTimer
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

package ipmi

import "fmt"

// 27.7 Get Watchdog Timer Command
type GetWatchdogTimerRequest struct {
}

func (req *GetWatchdogTimerRequest) Pack() []byte {
	return []byte{}
}

func (req *GetWatchdogTimerRequest) Command() Command {
	return CommandGetWatchdogTimer
}

type GetWatchdogTimerResponse struct {
	DontLog        bool
	TimerIsStarted bool
	TimerUse       TimerUse

	PreTimeoutInterrupt   PreTimeoutInterrupt
	TimeoutAction         TimeoutAction
	PreTimeoutIntervalSec uint8

	ExpirationFlags  uint8
	InitialCountdown uint16
	PresentCountdown uint16
}

func (res *GetWatchdogTimerResponse) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	res.DontLog = isBit7Set(msg[0])
	res.TimerIsStarted = isBit6Set(msg[0])
	res.TimerUse = TimerUse(0x07 & msg[0])

	res.PreTimeoutInterrupt = PreTimeoutInterrupt((0x70 & msg[1]) >> 4)
	res.TimeoutAction = TimeoutAction(0x07 & msg[1])

	res.PreTimeoutIntervalSec = msg[2]
	res.ExpirationFlags = msg[3]
	res.InitialCountdown, _, _ = unpackUint16L(msg, 4)
	res.PresentCountdown, _, _ = unpackUint16L(msg, 6)
	return nil
}

func (res *GetWatchdogTimerResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetWatchdogTimerResponse) Format() string {
	return fmt.Sprintf(`Watchdog Timer Use:     %s (%#02x)
Watchdog Timer Is:      %s
Watchdog Timer Actions: %s (%#02x)
Pre-timeout interval:   %d seconds
Timer Expiration Flags: %#02x
Initial Countdown:      %d sec
Present Countdown:      %d sec`,
		res.TimerUse, uint8(res.TimerUse),
		formatBool(res.TimerIsStarted, "Started", "Stopped"),
		res.TimeoutAction, uint8(res.TimeoutAction),
		res.PreTimeoutIntervalSec,
		res.ExpirationFlags,
		res.InitialCountdown,
		res.PresentCountdown,
	)
}

func (c *Client) GetWatchdogTimer() (response *GetWatchdogTimerResponse, err error) {
	request := &GetWatchdogTimerRequest{}
	response = &GetWatchdogTimerResponse{}
	err = c.Exchange(request, response)
	return
}

type TimerUse uint8

const (
	TimerUseBIOSFRB2 TimerUse = 0x01 // BIOS/FRB2
	TimerUseBIOSPOST TimerUse = 0x02 // BIOS/POST
	TimerUseOSLoad   TimerUse = 0x03
	TimerUseSMSOS    TimerUse = 0x04 // SMS/OS
	TimerUseOEM      TimerUse = 0x05
)

func (t TimerUse) String() string {
	m := map[TimerUse]string{
		0x01: "BIOS FRB2",
		0x02: "BIOS/POST",
		0x03: "OS Load",
		0x04: "SMS/OS",
		0x05: "OEM",
	}
	s, ok := m[t]
	if ok {
		return s
	}
	return ""
}

type PreTimeoutInterrupt uint8

const (
	PreTimeoutInterruptNone      PreTimeoutInterrupt = 0x00
	PreTimeoutInterruptSMI       PreTimeoutInterrupt = 0x01
	PreTimeoutInterruptNMI       PreTimeoutInterrupt = 0x02
	PreTimeoutInterruptMessaging PreTimeoutInterrupt = 0x03
)

func (t PreTimeoutInterrupt) String() string {
	m := map[PreTimeoutInterrupt]string{
		0x00: "None",
		0x01: "SMI",
		0x02: "NMI / Diagnostic Interrupt",
		0x03: "Messaging Interrupt",
	}
	s, ok := m[t]
	if ok {
		return s
	}
	return ""
}

type TimeoutAction uint8

const (
	TimeoutActionNoAction   TimeoutAction = 0x00
	TimeoutActionHardReset  TimeoutAction = 0x01
	TimeoutActionPowerDown  TimeoutAction = 0x02
	TimeoutActionPowerCycle TimeoutAction = 0x03
)

func (t TimeoutAction) String() string {
	m := map[TimeoutAction]string{
		0x00: "No action",
		0x01: "Hard Reset",
		0x02: "Power Down",
		0x03: "Power Cycle",
	}
	s, ok := m[t]
	if ok {
		return s
	}
	return ""
}

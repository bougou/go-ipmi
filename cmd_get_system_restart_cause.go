package ipmi

import "fmt"

// 28.11 Get System Restart Cause Command
type GetSystemRestartCauseRequest struct {
	// no data
}

type GetSystemRestartCauseResponse struct {
	SystemRestartCause SystemRestartCause
	ChannelNumber      uint8
}

type SystemRestartCause uint8

var systemRestartCauses = map[SystemRestartCause]string{
	0x00: "unkown",                                        // unknown (system start/restart detected, but cause unknown)
	0x01: "chassis power control command",                 //
	0x02: "reset via pushbutton",                          //
	0x03: "power-up via pushbutton",                       //
	0x04: "watchdog expired",                              //
	0x05: "OEM",                                           //
	0x06: "power-up due to always-on restore power plicy", // automatic power-up on AC being applied due to 'always restore' power restore policy
	0x07: "power-up due to previous restore power policy", // automatic power-up on AC being applied due to 'restore previous power state' power restore policy
	0x08: "reset via PEF",                                 //
	0x09: "power-cycle via PEF",                           //
	0x0a: "soft reset",                                    // soft reset (e.g. CTRL-ALT-DEL) [optional]
	0x0b: "power-up via RTC wakeup",                       // power-up via RTC (system real time clock) wakeup [optional]
}

func (c SystemRestartCause) String() string {
	s, ok := systemRestartCauses[c]
	if ok {
		return s
	}
	return "invalid"
}

func (req *GetSystemRestartCauseRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSystemRestartCauseRequest) Command() Command {
	return CommandGetSystemRestartCause
}

func (res *GetSystemRestartCauseResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSystemRestartCauseResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}

	b, _, _ := unpackUint8(msg, 0)
	res.SystemRestartCause = SystemRestartCause(b)
	res.ChannelNumber, _, _ = unpackUint8(msg, 1)
	return nil
}

func (res *GetSystemRestartCauseResponse) Format() string {
	return fmt.Sprintf("System restart cause: %s", res.SystemRestartCause.String())
}

func (c *Client) GetSystemRestartCause() (response *GetSystemRestartCauseResponse, err error) {
	request := &GetSystemRestartCauseRequest{}
	response = &GetSystemRestartCauseResponse{}
	err = c.Exchange(request, response)
	return
}

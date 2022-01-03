package ipmi

// 22.3 Clear Message Flags Command
type ClearMessageFlagsRequest struct {
	ClearOEM2                            bool
	ClearOEM1                            bool
	ClearOEM0                            bool
	ClearWatchdogPreTimeoutInterruptFlag bool
	ClearEventMessageBuffer              bool
	ClearReceiveMessageQueue             bool
}

type ClearMessageFlagsResponse struct {
}

func (req *ClearMessageFlagsRequest) Command() Command {
	return CommandClearMessageFlags
}

func (req *ClearMessageFlagsRequest) Pack() []byte {
	var b uint8 = 0
	if req.ClearOEM2 {
		b = setBit7(b)
	}
	if req.ClearOEM1 {
		b = setBit6(b)
	}
	if req.ClearOEM0 {
		b = setBit5(b)
	}
	if req.ClearWatchdogPreTimeoutInterruptFlag {
		b = setBit3(b)
	}
	if req.ClearEventMessageBuffer {
		b = setBit1(b)
	}
	if req.ClearReceiveMessageQueue {
		b = setBit0(b)
	}

	return []byte{b}
}

func (res *ClearMessageFlagsResponse) Unpack(msg []byte) error {
	return nil
}

func (*ClearMessageFlagsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *ClearMessageFlagsResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) ClearMessageFlags() (response *ClearMessageFlagsResponse, err error) {
	request := &ClearMessageFlagsRequest{}
	response = &ClearMessageFlagsResponse{}
	err = c.Exchange(request, response)
	return
}

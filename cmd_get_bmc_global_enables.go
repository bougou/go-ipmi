package ipmi

// 22.2 Get BMC Global Enables Command
type GetBMCGlobalEnablesRequest struct {
	// empty
}

type GetBMCGlobalEnablesResponse struct {
	OEM2Enabled bool
	OEM1Enabled bool
	OEM0Enabled bool

	SystemEventLoggingEnabled              bool
	EventMessageBufferEnabled              bool
	EventMessageBufferFullInterruptEnabled bool
	ReceiveMessageQueueInterruptEnabled    bool
}

func (req *GetBMCGlobalEnablesRequest) Command() Command {
	return CommandGetBMCGlobalEnables
}

func (req *GetBMCGlobalEnablesRequest) Pack() []byte {
	return []byte{}
}

func (res *GetBMCGlobalEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := unpackUint8(msg, 0)
	res.OEM2Enabled = isBit7Set(b)
	res.OEM1Enabled = isBit6Set(b)
	res.OEM0Enabled = isBit5Set(b)
	res.SystemEventLoggingEnabled = isBit3Set(b)
	res.EventMessageBufferEnabled = isBit2Set(b)
	res.EventMessageBufferFullInterruptEnabled = isBit1Set(b)
	res.ReceiveMessageQueueInterruptEnabled = isBit0Set(b)
	return nil
}

func (*GetBMCGlobalEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetBMCGlobalEnablesResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetBMCGlobalEnables() (response *GetBMCGlobalEnablesResponse, err error) {
	request := &GetBMCGlobalEnablesRequest{}
	response = &GetBMCGlobalEnablesResponse{}
	err = c.Exchange(request, response)
	return
}

package ipmi

// 22.4 Get Message Flags Command
type GetMessageFlagsRequest struct {
}

type GetMessageFlagsResponse struct {
	OEM2Avaiable                        bool
	OEM1Avaiable                        bool
	OEM0Avaiable                        bool
	WatchdogPreTimeoutInterruptOccurred bool
	EventMessageBufferFull              bool
	ReceiveMessageQueueAvaiable         bool // One or more messages ready for reading from Receive Message Queue
}

func (req *GetMessageFlagsRequest) Command() Command {
	return CommandGetMessageFlags
}

func (req *GetMessageFlagsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageFlagsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	b, _, _ := unpackUint8(msg, 0)
	res.OEM2Avaiable = isBit7Set(b)
	res.OEM1Avaiable = isBit6Set(b)
	res.OEM0Avaiable = isBit5Set(b)
	res.WatchdogPreTimeoutInterruptOccurred = isBit3Set(b)
	res.EventMessageBufferFull = isBit1Set(b)
	res.ReceiveMessageQueueAvaiable = isBit0Set(b)
	return nil
}

func (*GetMessageFlagsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetMessageFlagsResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetMessageFlags() (response *GetMessageFlagsResponse, err error) {
	request := &GetMessageFlagsRequest{}
	response = &GetMessageFlagsResponse{}
	err = c.Exchange(request, response)
	return
}

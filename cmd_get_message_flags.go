package ipmi

import "context"

// 22.4 Get Message Flags Command
type GetMessageFlagsRequest struct {
	// empty
}

type GetMessageFlagsResponse struct {
	OEM2Available                       bool
	OEM1Available                       bool
	OEM0Available                       bool
	WatchdogPreTimeoutInterruptOccurred bool
	EventMessageBufferFull              bool
	ReceiveMessageQueueAvailable        bool // One or more messages ready for reading from Receive Message Queue
}

func (req *GetMessageFlagsRequest) Command() Command {
	return CommandGetMessageFlags
}

func (req *GetMessageFlagsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageFlagsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := unpackUint8(msg, 0)
	res.OEM2Available = isBit7Set(b)
	res.OEM1Available = isBit6Set(b)
	res.OEM0Available = isBit5Set(b)
	res.WatchdogPreTimeoutInterruptOccurred = isBit3Set(b)
	res.EventMessageBufferFull = isBit1Set(b)
	res.ReceiveMessageQueueAvailable = isBit0Set(b)
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

func (c *Client) GetMessageFlags(ctx context.Context) (response *GetMessageFlagsResponse, err error) {
	request := &GetMessageFlagsRequest{}
	response = &GetMessageFlagsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package ipmi

// 22.6 Get Message Command
type GetMessageRequest struct {
	// empty
}

type GetMessageResponse struct {
	ChannelNumber uint8
	MessageData   []byte
}

func (req *GetMessageRequest) Command() Command {
	return CommandGetMessage
}

func (req *GetMessageRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "data not available (queue / buffer empty)",
	}
}

func (res *GetMessageResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}
	res.ChannelNumber, _, _ = unpackUint8(msg, 0)
	res.MessageData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetMessageResponse) Format() string {
	return ""
}

func (c *Client) GetMessage() (response *GetMessageResponse, err error) {
	request := &GetMessageRequest{}
	response = &GetMessageResponse{}
	err = c.Exchange(request, response)
	return
}

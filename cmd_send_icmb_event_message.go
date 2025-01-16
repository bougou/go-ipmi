package ipmi

import "context"

type SendICMBEventMessageRequest struct {
	// 0 = Online 1 = Attention (a.k.a. 'Sensor')
	EventCode uint8
}

type SendICMBEventMessageResponse struct {
}

func (req *SendICMBEventMessageRequest) Pack() []byte {
	out := make([]byte, 1)
	out[0] = req.EventCode
	return out
}

func (req *SendICMBEventMessageRequest) Command() Command {
	return CommandSendICMBEventMessage
}

func (res *SendICMBEventMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SendICMBEventMessageResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SendICMBEventMessageResponse) Format() string {
	return ""
}

func (c *Client) SendICMBEventMessage(ctx context.Context, eventCode uint8) (response *SendICMBEventMessageResponse, err error) {
	request := &SendICMBEventMessageRequest{
		EventCode: eventCode,
	}
	response = &SendICMBEventMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package ipmi

import "context"

// 24.2 Deactivate Payload Command
type DeactivatePayloadRequest struct {
	PayloadType     PayloadType
	PayloadInstance uint8
}

type DeactivatePayloadResponse struct {
}

func (req DeactivatePayloadRequest) Command() Command {
	return CommandDeactivatePayload
}

func (req *DeactivatePayloadRequest) Pack() []byte {
	out := make([]byte, 6)

	out[0] = byte(req.PayloadType)
	out[1] = req.PayloadInstance

	out[2] = 0
	out[3] = 0
	out[4] = 0
	out[5] = 0

	return out
}

func (res *DeactivatePayloadResponse) Unpack(msg []byte) error {
	return nil
}

func (*DeactivatePayloadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Payload already deactivated",
		0x81: "Payload type is disabled",
	}
}

func (res *DeactivatePayloadResponse) Format() string {
	return ""
}

func (c *Client) DeactivatePayload(ctx context.Context, request *DeactivatePayloadRequest) (response *DeactivatePayloadResponse, err error) {
	response = &DeactivatePayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

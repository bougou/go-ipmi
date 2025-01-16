package ipmi

import "context"

type SetEventDestinationRequest struct {
	// ICMB address of event destination bridge
	ICMBAddr uint16
}

type SetEventDestinationResponse struct {
}

func (req *SetEventDestinationRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint16L(req.ICMBAddr, out, 0)

	return out
}

func (req *SetEventDestinationRequest) Command() Command {
	return CommandSetEventDestination
}

func (res *SetEventDestinationResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetEventDestinationResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetEventDestinationResponse) Format() string {
	return ""
}

func (c *Client) SetEventDestination(ctx context.Context, icmbAddr uint16) (response *SetEventDestinationResponse, err error) {
	request := &SetEventDestinationRequest{
		ICMBAddr: icmbAddr,
	}
	response = &SetEventDestinationResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

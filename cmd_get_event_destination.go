package ipmi

import (
	"context"
	"fmt"
)

type GetEventDestinationRequest struct {
}

type GetEventDestinationResponse struct {
	// ICMB address of event destination bridge
	ICMBAddr uint16
}

func (req *GetEventDestinationRequest) Pack() []byte {
	return []byte{}
}

func (req *GetEventDestinationRequest) Command() Command {
	return CommandGetEventDestination
}

func (res *GetEventDestinationResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetEventDestinationResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ICMBAddr, _, _ = unpackUint16L(msg, 0)
	return nil
}

func (res *GetEventDestinationResponse) Format() string {
	return fmt.Sprintf("ICMB Address : %#04x", res.ICMBAddr)
}

func (c *Client) GetEventDestination(ctx context.Context) (response *GetEventDestinationResponse, err error) {
	request := &GetEventDestinationRequest{}
	response = &GetEventDestinationResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package ipmi

import (
	"context"
	"fmt"
)

// 24.5 Get Payload Instance Info Command
type GetPayloadInstanceInfoRequest struct {
	PayloadType     PayloadType
	PayloadInstance uint8
}

type GetPayloadInstanceInfoResponse struct {
	SessionID uint32

	// For Payload Type = SOL:
	//  Byte 1: Port Number
	//    A number representing the system serial port that is being redirected.
	//    1-based. 0h = unspecified. Used when more than one port can be redirected on a system.
	PortNumber uint8

	PayloadType PayloadType
}

func (req *GetPayloadInstanceInfoRequest) Pack() []byte {
	out := make([]byte, 2)
	out[0] = byte(req.PayloadType)
	out[1] = byte(req.PayloadInstance)
	return out
}

func (req *GetPayloadInstanceInfoRequest) Command() Command {
	return CommandGetPayloadInstanceInfo
}

func (res *GetPayloadInstanceInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPayloadInstanceInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 12 {
		return ErrUnpackedDataTooShortWith(len(msg), 12)
	}

	res.SessionID, _, _ = unpackUint32L(msg, 0)
	res.PortNumber = msg[4]

	return nil
}

func (res *GetPayloadInstanceInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Session ID      : %d\n", res.SessionID) +
		fmt.Sprintf("Payload Type    : %s (%d)\n", res.PayloadType.String(), uint8(res.PayloadType)) +
		fmt.Sprintf("SOL Port Number : %d\n", uint8(res.PortNumber))
}

func (c *Client) GetPayloadInstanceInfo(ctx context.Context, payloadType PayloadType, payloadInstance uint8) (response *GetPayloadInstanceInfoResponse, err error) {
	request := &GetPayloadInstanceInfoRequest{
		PayloadType:     payloadType,
		PayloadInstance: payloadInstance,
	}
	response = &GetPayloadInstanceInfoResponse{}
	response.PayloadType = request.PayloadType
	err = c.Exchange(ctx, request, response)
	return
}

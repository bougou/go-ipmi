package ipmi

import "context"

type GetEventReceptionStateRequest struct {
}

type GetEventReceptionStateResponse struct {
	ReceptionState uint8
	EventSA        uint8
	LUN            uint8
}

func (req *GetEventReceptionStateRequest) Pack() []byte {
	return []byte{}
}

func (req *GetEventReceptionStateRequest) Command() Command {
	return CommandGetEventReceptionState
}

func (res *GetEventReceptionStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetEventReceptionStateResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.ReceptionState = msg[0]
	res.EventSA = msg[1]
	res.LUN = msg[2]
	return nil
}

func (res *GetEventReceptionStateResponse) Format() string {
	return ""
}

func (c *Client) GetEventReceptionState(ctx context.Context, receptionState uint8, eventSA uint8, lun uint8) (response *GetEventReceptionStateResponse, err error) {
	request := &GetEventReceptionStateRequest{}
	response = &GetEventReceptionStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

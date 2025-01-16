package ipmi

import "context"

type SetEventReceptionStateRequest struct {
	ReceptionState uint8
	EventSA        uint8
	LUN            uint8
}

type SetEventReceptionStateResponse struct {
}

func (req *SetEventReceptionStateRequest) Pack() []byte {
	out := make([]byte, 3)
	out[0] = req.ReceptionState
	out[1] = req.EventSA
	out[2] = req.LUN

	return out
}

func (req *SetEventReceptionStateRequest) Command() Command {
	return CommandSetEventReceptionState
}

func (res *SetEventReceptionStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetEventReceptionStateResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetEventReceptionStateResponse) Format() string {
	return ""
}

func (c *Client) SetEventReceptionState(ctx context.Context, receptionState uint8, eventSA uint8, lun uint8) (response *SetEventReceptionStateResponse, err error) {
	request := &SetEventReceptionStateRequest{
		ReceptionState: receptionState,
		EventSA:        eventSA,
		LUN:            lun,
	}
	response = &SetEventReceptionStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

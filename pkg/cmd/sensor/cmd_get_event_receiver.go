package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 29.2 Get Event Receiver Command
)

type GetEventReceiverRequest struct {
}

type GetEventReceiverResponse struct {
	SlaveAddress uint8
	LUN          uint8
}

func (req *GetEventReceiverRequest) Pack() []byte {
	return []byte{}
}

func (req *GetEventReceiverRequest) Command() types.Command {
	return types.CommandGetEventReceiver
}

func (res *GetEventReceiverResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetEventReceiverResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.SlaveAddress = msg[0]
	res.LUN = msg[1]
	return nil
}

func (res *GetEventReceiverResponse) Format() string {
	return ""
}

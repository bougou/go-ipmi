package ipmi

import "context"

// 29.3 Platform Event Message Command
type PlatformEventMessageRequest struct {
	// The Generator ID field is a required element of an Event Request Message.
	// This field identifies the device that has generated the Event Message.
	// This is the 7-bit Requester's Slave Address (RqSA) and 2-bit Requester's LUN (RqLUN)
	// if the message was received from the IPMB, or the 7-bit System Software ID
	// if the message was received from system software.
	//
	// For IPMB messages, this field is equated to the Requester's Slave Address and LUN fields.
	// Thus, the Generator ID information is not carried in the data field of an IPMB request message.
	//
	// For 'system side' interfaces, it is not as useful or appropriate to 'overlay' the Generator ID field
	// with the message source address information, and so it is specified as being carried in the data field of the request.
	GeneratorID  uint8
	EvMRev       uint8
	SensorType   uint8
	SensorNumber uint8
	EventDir     EventDir
	EventType    EventReadingType
	EventData    EventData
}

type PlatformEventMessageResponse struct {
}

func (req *PlatformEventMessageRequest) Pack() []byte {
	out := make([]byte, 8)
	out[0] = req.GeneratorID
	out[1] = req.EvMRev
	out[2] = req.SensorType
	out[3] = req.SensorNumber

	var b4 = uint8(req.EventType)
	if req.EventDir {
		b4 |= 0x80
	}
	out[4] = b4

	out[5] = req.EventData.EventData1
	out[6] = req.EventData.EventData2
	out[7] = req.EventData.EventData3

	return []byte{}
}

func (req *PlatformEventMessageRequest) Command() Command {
	return CommandPlatformEventMessage
}

func (res *PlatformEventMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *PlatformEventMessageResponse) Unpack(msg []byte) error {
	return nil
}

func (res *PlatformEventMessageResponse) Format() string {
	return ""
}

func (c *Client) PlatformEventMessage(ctx context.Context, request *PlatformEventMessageRequest) (response *PlatformEventMessageResponse, err error) {
	// Todo, consider GeneratorID
	response = &PlatformEventMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package ipmi

import "context"

// 30.8 PET Acknowledge Command
// This message is used to acknowledge a Platform Event Trap (PET) alert.
type PETAcknowledgeRequest struct {
	SequenceNumber  uint16
	LocalTimestamp  uint32
	EventSourceType uint8
	SensorDevice    uint8
	SensorNumber    uint8
	EventData       EventData
}

type PETAcknowledgeResponse struct {
	PETAcknowledgeStatus uint8
}

func (req *PETAcknowledgeRequest) Pack() []byte {
	out := make([]byte, 12)

	packUint16L(req.SequenceNumber, out, 0)
	packUint32L(req.LocalTimestamp, out, 2)
	out[6] = req.EventSourceType
	out[7] = req.SensorDevice
	out[8] = req.SensorNumber
	out[9] = uint8(req.EventData.EventData1)
	out[10] = uint8(req.EventData.EventData2)
	out[11] = uint8(req.EventData.EventData3)

	return out
}

func (req *PETAcknowledgeRequest) Unpack(data []byte) error {
	if len(data) < 12 {
		return ErrUnpackedDataTooShortWith(len(data), 12)
	}

	req.SequenceNumber, _, _ = unpackUint16L(data, 0)
	req.LocalTimestamp, _, _ = unpackUint32L(data, 2)
	req.EventSourceType = data[6]
	req.SensorDevice = data[7]
	req.SensorNumber = data[8]
	req.EventData.EventData1 = data[9]
	req.EventData.EventData2 = data[10]
	req.EventData.EventData3 = data[11]

	return nil
}

func (req *PETAcknowledgeRequest) Command() Command {
	return CommandPETAcknowledge
}

func (res *PETAcknowledgeResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "Alert Immediate rejected due to alert already in progress",
		0x82: "Alert Immediate rejected due to IPMI messaging session active on this channel",
		0x83: "Platform Event Parameters (4:11) not supported",
	}
}

func (res *PETAcknowledgeResponse) Unpack(msg []byte) error {
	return nil
}

func (res *PETAcknowledgeResponse) Pack() []byte {
	return []byte{}
}

func (res *PETAcknowledgeResponse) Format() string {
	return ""
}

func (c *Client) PETAcknowledge(ctx context.Context, request *PETAcknowledgeRequest) (response *PETAcknowledgeResponse, err error) {
	response = &PETAcknowledgeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

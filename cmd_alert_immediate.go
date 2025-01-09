package ipmi

import "context"

type AlertImmediateOperation uint8

const (
	AlertImmediateOperationInitiateAlert AlertImmediateOperation = 0b00
	AlertImmediateOperationGetStatus     AlertImmediateOperation = 0b01
	AlertImmediateOperationClearStatus   AlertImmediateOperation = 0b10
	AlertImmediateOperationReserved      AlertImmediateOperation = 0b11
)

type AlertImmediateStatus uint8

const (
	AlertImmediateStatusNoStatus      AlertImmediateStatus = 0x00
	AlertImmediateStatusNormalEnd     AlertImmediateStatus = 0x01
	AlertImmediateStatusFailedRetry   AlertImmediateStatus = 0x02
	AlertImmediateStatusFailedWaitACK AlertImmediateStatus = 0x03
	AlertImmediateStatusInProgress    AlertImmediateStatus = 0xff
)

type AlertImmediateRequest struct {
	ChannelNumber uint8

	DestinationSelector uint8
	Operation           uint8

	SendAlertString     bool
	AlertStringSelector uint8

	GeneratorID  uint8
	EvMRev       uint8
	SensorType   SensorType
	SensorNumber SensorNumber

	EventDir         EventDir
	EventReadingType EventReadingType
	EventData        EventData
}

type AlertImmediateResponse struct {
	AlertImmediateStatus uint8
}

func (req *AlertImmediateRequest) Pack() []byte {
	out := make([]byte, 1)
	return out
}

func (req *AlertImmediateRequest) Command() Command {
	return CommandAlertImmediate
}

func (res *AlertImmediateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "Alert Immediate rejected due to alert already in progress",
		0x82: "Alert Immediate rejected due to IPMI messaging session active on this channel",
		0x83: "Platform Event Parameters (4:11) not supported",
	}
}

func (res *AlertImmediateResponse) Unpack(msg []byte) error {
	return nil
}

func (res *AlertImmediateResponse) Format() string {
	return ""
}

func (c *Client) AlertImmediate(ctx context.Context) (response *AlertImmediateResponse, err error) {
	request := &AlertImmediateRequest{}
	response = &AlertImmediateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

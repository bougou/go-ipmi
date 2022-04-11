package ipmi

// 29.1 Set Event Receiver Command
type SetEventReceiverRequest struct {
	// Event Receiver Slave Address. 0FFh disables Event Message Generation, Otherwise:
	// [7:1] - IPMB (I2C) Slave Address
	// [0] - always 0b when [7:1] hold I2C slave address
	SlaveAddress uint8
	// [1:0] - Event Receiver LUN
	LUN uint8
}

type SetEventReceiverResponse struct {
}

func (req *SetEventReceiverRequest) Pack() []byte {
	return []byte{req.SlaveAddress, req.LUN}
}

func (req *SetEventReceiverRequest) Command() Command {
	return CommandSetEventReceiver
}

func (res *SetEventReceiverResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetEventReceiverResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetEventReceiverResponse) Format() string {
	return ""
}

func (c *Client) SetEventReceiver(slaveAddress uint8, lun uint8) (response *SetEventReceiverResponse, err error) {
	request := &SetEventReceiverRequest{
		SlaveAddress: slaveAddress,
		LUN:          lun,
	}
	response = &SetEventReceiverResponse{}
	err = c.Exchange(request, response)
	return
}

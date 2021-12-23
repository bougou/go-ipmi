package ipmi

import "fmt"

// 22.11 Master Write-Read Command
type MasterWriteReadRequest struct {
	ChannelNumber uint8

	BusID uint8

	BusType uint8

	SlaveAddress uint8

	ReadCount uint8

	Data []byte
}

type MasterWriteReadResponse struct {
	// A management controller shall return an error Completion Code if an attempt is made to access an unsupported bus.
	// generic, plus command specific codes
	CompletionCode

	// Bytes read from specified slave address. This field will be absent if the read count is 0. The controller terminates the I2C transaction with a STOP condition after reading the requested number of bytes.
	Data []byte
}

func (req *MasterWriteReadRequest) Command() Command {
	return CommandMasterWriteRead
}

func (req *MasterWriteReadRequest) Pack() []byte {
	return nil
}

func (res *MasterWriteReadResponse) Unpack(msg []byte) error {
	return nil
}

func (*MasterWriteReadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
		0x84: "Truncated Read",
	}
}

func (res *MasterWriteReadResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) MasterWriteRead() (*MasterWriteReadResponse, error) {
	// var cc = map[CompletionCode]string{
	// 	0x81: "Lost Arbitration",
	// 	0x82: "Bus Error",
	// 	0x83: "NAK on Write",
	// 	0x84: "Truncated Read",
	// }

	request := &MasterWriteReadRequest{}
	response := &MasterWriteReadResponse{}
	err := c.Exchange(request, response)
	return response, err
}

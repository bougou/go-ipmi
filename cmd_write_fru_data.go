package ipmi

import (
	"fmt"
)

// 34.3 Write FRU Data Command
type WriteFRUDataRequest struct {
	FRUDeviceID uint8
	WriteOffset uint16
	WriteData   []byte
}

type WriteFRUDataResponse struct {
	CountWritten uint8
}

func (req *WriteFRUDataRequest) Command() Command {
	return CommandWriteFRUData
}

func (req *WriteFRUDataRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.FRUDeviceID, out, 0)
	packUint16L(req.WriteOffset, out, 1)
	packBytes(req.WriteData, out, 3)
	return out
}

func (res *WriteFRUDataResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.CountWritten, _, _ = unpackUint8(msg, 0)
	return nil
}

func (r *WriteFRUDataResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "write-protected offset",
		0x81: "FRU device busy",
	}
}

func (res *WriteFRUDataResponse) Format() string {
	return fmt.Sprintf(`Count written : %d`, res.CountWritten)
}

// The command writes the specified byte or word to the FRU Inventory Info area. This is a low level direct interface to a non-volatile storage area. This means that the interface does not interpret or check any semantics or formatting for the data being written.
func (c *Client) WriteFRUData(fruDeviceID uint8, writeOffset uint16, writeData []byte) (response *WriteFRUDataResponse, err error) {
	request := &WriteFRUDataRequest{
		FRUDeviceID: fruDeviceID,
		WriteOffset: writeOffset,
		WriteData:   writeData,
	}
	response = &WriteFRUDataResponse{}
	err = c.Exchange(request, response)
	return
}

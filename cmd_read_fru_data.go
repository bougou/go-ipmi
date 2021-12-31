package ipmi

import (
	"fmt"
)

// 34.2 Read FRU Data Command
type ReadFRUDataRequest struct {
	FRUDeviceID uint8
	ReadOffset  uint16
	ReadCount   uint8
}

type ReadFRUDataResponse struct {
	CountReturned uint8
	Data          []byte
}

func (req *ReadFRUDataRequest) Command() Command {
	return CommandReadFRUData
}

func (req *ReadFRUDataRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.FRUDeviceID, out, 0)
	packUint16L(req.ReadOffset, out, 1)
	packUint8(req.ReadCount, out, 3)
	return out
}

func (res *ReadFRUDataResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.CountReturned, _, _ = unpackUint8(msg, 0)
	res.Data, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (r *ReadFRUDataResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "FRU device busy",
	}
}

func (res *ReadFRUDataResponse) Format() string {
	return fmt.Sprintf(`Count returned : %d
Data           : %02x`,
		res.CountReturned,
		res.Data,
	)
}

// The command returns the specified data from the FRU Inventory Info area.
func (c *Client) ReadFRUData(fruDeviceID uint8, readOffset uint16, readCount uint8) (response *ReadFRUDataResponse, err error) {
	request := &ReadFRUDataRequest{
		FRUDeviceID: fruDeviceID,
		ReadOffset:  readOffset,
		ReadCount:   readCount,
	}
	response = &ReadFRUDataResponse{}
	err = c.Exchange(request, response)
	return
}

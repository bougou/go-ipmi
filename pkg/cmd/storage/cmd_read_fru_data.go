package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
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

func (req *ReadFRUDataRequest) Command() types.Command {
	return types.CommandReadFRUData
}

func (req *ReadFRUDataRequest) Pack() []byte {
	out := make([]byte, 4)
	types.PackUint8(req.FRUDeviceID, out, 0)
	types.PackUint16L(req.ReadOffset, out, 1)
	types.PackUint8(req.ReadCount, out, 3)
	return out
}

func (res *ReadFRUDataResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.CountReturned, _, _ = types.UnpackUint8(msg, 0)
	res.Data, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (r *ReadFRUDataResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "FRU device busy",
	}
}

func (res *ReadFRUDataResponse) Format() string {
	return "" +
		fmt.Sprintf("Count returned : %d\n", res.CountReturned) +
		fmt.Sprintf("Data           : %02x\n", res.Data)
}

func ReadFRUDataLength2Big(cc types.CompletionCode) bool {
	return cc == types.CompletionCodeRequestDataLengthInvalid ||
		cc == types.CompletionCodeRequestDataLengthLimitExceeded ||
		cc == types.CompletionCodeCannotReturnRequestedDataBytes
}

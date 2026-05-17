package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *ReadFRUDataRequest) Command() ipmi.Command {
	return ipmi.CommandReadFRUData
}

func (req *ReadFRUDataRequest) Pack() []byte {
	out := make([]byte, 4)
	ipmi.PackUint8(req.FRUDeviceID, out, 0)
	ipmi.PackUint16L(req.ReadOffset, out, 1)
	ipmi.PackUint8(req.ReadCount, out, 3)
	return out
}

func (res *ReadFRUDataResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.CountReturned, _, _ = ipmi.UnpackUint8(msg, 0)
	res.Data, _, _ = ipmi.UnpackBytes(msg, 1, len(msg)-1)
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

// The command returns the specified data from the FRU Inventory Info area.

// readFRUDataByLength reads FRU Data in loop until reaches the specified data length

// update offset

// tryReadFRUData will try to read FRU data with a read count which starts with
// the minimal number of the specified length and the hard-coded 32, if the
// ReadFRUData failed, it try another request with a decreased read count.

func ReadFRUDataLength2Big(cc ipmi.CompletionCode) bool {
	return cc == ipmi.CompletionCodeRequestDataLengthInvalid ||
		cc == ipmi.CompletionCodeRequestDataLengthLimitExceeded ||
		cc == ipmi.CompletionCodeCannotReturnRequestedDataBytes
}

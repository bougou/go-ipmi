package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

type GetSELAllocInfoRequest struct {
	// empty
}

type GetSELAllocInfoResponse struct {
	PossibleAllocUnits uint16
	AllocUnitsSize     uint16 // Allocation unit size in bytes. 0000h indicates unspecified.
	FreeAllocUnits     uint16
	LargestFreeBlock   uint16
	MaximumRecordSize  uint8
}

func (req *GetSELAllocInfoRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSELAllocInfoRequest) Command() ipmi.Command {
	return ipmi.CommandGetSELAllocInfo
}

func (res *GetSELAllocInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 9)
	}
	res.PossibleAllocUnits, _, _ = ipmi.UnpackUint16L(msg, 0)
	res.AllocUnitsSize, _, _ = ipmi.UnpackUint16L(msg, 2)
	res.FreeAllocUnits, _, _ = ipmi.UnpackUint16L(msg, 4)
	res.LargestFreeBlock, _, _ = ipmi.UnpackUint16L(msg, 6)
	res.MaximumRecordSize, _, _ = ipmi.UnpackUint8(msg, 8)
	return nil
}

func (res *GetSELAllocInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSELAllocInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("# of Alloc Units : %d\n", res.PossibleAllocUnits) +
		fmt.Sprintf("Alloc Unit Size  : %d\n", res.AllocUnitsSize) +
		fmt.Sprintf("# Free Units     : %d\n", res.FreeAllocUnits) +
		fmt.Sprintf("Largest Free Blk : %d\n", res.LargestFreeBlock) +
		fmt.Sprintf("Max Record Size  : %d\n", res.MaximumRecordSize)
}

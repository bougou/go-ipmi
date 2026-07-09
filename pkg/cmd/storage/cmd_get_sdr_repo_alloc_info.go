package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 33.10 Get SDR Repository Allocation Info Command
type GetSDRRepoAllocInfoRequest struct {
	// empty
}

type GetSDRRepoAllocInfoResponse struct {
	PossibleAllocUnits uint16
	AllocUnitsSize     uint16 // Allocation unit size in bytes. 0000h indicates unspecified.
	FreeAllocUnits     uint16
	LargestFreeBlock   uint16
	MaximumRecordSize  uint8
}

func (req *GetSDRRepoAllocInfoRequest) Pack() []byte {
	return nil
}

func (req *GetSDRRepoAllocInfoRequest) Command() types.Command {
	return types.CommandGetSDRRepoAllocInfo
}

func (res *GetSDRRepoAllocInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 9)
	}
	res.PossibleAllocUnits, _, _ = types.UnpackUint16L(msg, 0)
	res.AllocUnitsSize, _, _ = types.UnpackUint16L(msg, 2)
	res.FreeAllocUnits, _, _ = types.UnpackUint16L(msg, 4)
	res.LargestFreeBlock, _, _ = types.UnpackUint16L(msg, 6)
	res.MaximumRecordSize, _, _ = types.UnpackUint8(msg, 8)
	return nil
}

func (res *GetSDRRepoAllocInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSDRRepoAllocInfoResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

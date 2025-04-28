package ipmi

import (
	"context"
	"fmt"
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

func (req *GetSELAllocInfoRequest) Command() Command {
	return CommandGetSELAllocInfo
}

func (res *GetSELAllocInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return ErrUnpackedDataTooShortWith(len(msg), 9)
	}
	res.PossibleAllocUnits, _, _ = unpackUint16L(msg, 0)
	res.AllocUnitsSize, _, _ = unpackUint16L(msg, 2)
	res.FreeAllocUnits, _, _ = unpackUint16L(msg, 4)
	res.LargestFreeBlock, _, _ = unpackUint16L(msg, 6)
	res.MaximumRecordSize, _, _ = unpackUint8(msg, 8)
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

func (c *Client) GetSELAllocInfo(ctx context.Context) (response *GetSELAllocInfoResponse, err error) {
	request := &GetSELAllocInfoRequest{}
	response = &GetSELAllocInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

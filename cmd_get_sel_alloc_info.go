package ipmi

import "fmt"

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
	return fmt.Sprintf(`# of Alloc Units             : %d
Alloc Unit Size              : %d
# Free Units                 : %d
Largest Free Blk             : %d
Max Record Size              : %d`,
		res.PossibleAllocUnits,
		res.AllocUnitsSize,
		res.FreeAllocUnits,
		res.LargestFreeBlock,
		res.MaximumRecordSize,
	)
}

func (c *Client) GetSELAllocInfo() (response *GetSELAllocInfoResponse, err error) {
	request := &GetSELAllocInfoRequest{}
	response = &GetSELAllocInfoResponse{}
	err = c.Exchange(request, response)
	return
}

package ipmi

import "fmt"

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

func (req *GetSDRRepoAllocInfoRequest) Command() Command {
	return CommandGetSDRRepoAllocInfo
}

func (res *GetSDRRepoAllocInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return ErrUnpackedDataTooShort
	}
	res.PossibleAllocUnits, _, _ = unpackUint16L(msg, 0)
	res.AllocUnitsSize, _, _ = unpackUint16L(msg, 2)
	res.FreeAllocUnits, _, _ = unpackUint16L(msg, 4)
	res.LargestFreeBlock, _, _ = unpackUint16L(msg, 6)
	res.MaximumRecordSize, _, _ = unpackUint8(msg, 8)
	return nil
}

func (res *GetSDRRepoAllocInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSDRRepoAllocInfoResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) GetSDRRepoAllocInfo() (response *GetSDRRepoAllocInfoResponse, err error) {
	request := &GetSDRRepoAllocInfoRequest{}
	response = &GetSDRRepoAllocInfoResponse{}
	err = c.Exchange(request, response)
	return
}

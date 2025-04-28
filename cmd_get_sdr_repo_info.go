package ipmi

import (
	"context"
	"fmt"
	"time"
)

// 33.9 Get SDR Repository Info Command
type GetSDRRepoInfoRequest struct {
	// empty
}

type GetSDRRepoInfoResponse struct {
	SDRVersion             uint8  // version number of the SDR command set for the SDR Device. 51h for this specification.
	RecordCount            uint16 // LS Byte first
	FreeSpaceBytes         uint16 // LS Byte first
	MostRecentAdditionTime time.Time
	MostRecentEraseTime    time.Time

	SDROperationSupport SDROperationSupport
}

type SDROperationSupport struct {
	Overflow                     bool
	SupportModalSDRRepoUpdate    bool // A modal SDR Repository is only updated when the controller is in an SDR Repository update mode.
	SupportNonModalSDRRepoUpdate bool // A non-modal SDR Repository can be written to at any time
	SupportDeleteSDR             bool
	SupportPartialAddSDR         bool
	SupportReserveSDRRepo        bool
	SupportGetSDRRepoAllocInfo   bool
}

func (req *GetSDRRepoInfoRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSDRRepoInfoRequest) Command() Command {
	return CommandGetSDRRepoInfo
}

func (res *GetSDRRepoInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return ErrUnpackedDataTooShortWith(len(msg), 14)
	}

	res.SDRVersion, _, _ = unpackUint8(msg, 0)
	res.RecordCount, _, _ = unpackUint16L(msg, 1)
	res.FreeSpaceBytes, _, _ = unpackUint16L(msg, 3)

	addTS, _, _ := unpackUint32L(msg, 5)
	res.MostRecentAdditionTime = parseTimestamp(addTS)

	deleteTS, _, _ := unpackUint32L(msg, 9)
	res.MostRecentEraseTime = parseTimestamp(deleteTS)

	b, _, _ := unpackUint8(msg, 13)
	res.SDROperationSupport = SDROperationSupport{
		Overflow:                     isBit7Set(b),
		SupportModalSDRRepoUpdate:    isBit6Set(b),
		SupportNonModalSDRRepoUpdate: isBit5Set(b),
		SupportDeleteSDR:             isBit3Set(b),
		SupportPartialAddSDR:         isBit2Set(b),
		SupportReserveSDRRepo:        isBit1Set(b),
		SupportGetSDRRepoAllocInfo:   isBit0Set(b),
	}
	return nil
}

func (res *GetSDRRepoInfoResponse) Format() string {

	s := ""
	if res.SDROperationSupport.SupportModalSDRRepoUpdate {
		if s != "" {
			s += "/ modal"
		} else {
			s += "modal"
		}
	}
	if res.SDROperationSupport.SupportNonModalSDRRepoUpdate {
		if s != "" {
			s += "/ non-modal"
		} else {
			s += "non-modal"
		}
	}

	return "" +
		fmt.Sprintf("SDR Version                         : %#02x\n", res.SDRVersion) +
		fmt.Sprintf("Record Count                        : %d\n", res.RecordCount) +
		fmt.Sprintf("Free Space                          : %d bytes\n", res.FreeSpaceBytes) +
		fmt.Sprintf("Most recent Addition                : %s\n", res.MostRecentAdditionTime.Format(timeFormat)) +
		fmt.Sprintf("Most recent Erase                   : %s\n", res.MostRecentEraseTime.Format(timeFormat)) +
		fmt.Sprintf("SDR overflow                        : %v\n", res.SDROperationSupport.Overflow) +
		fmt.Sprintf("SDR Repository Update Support       : %s\n", s) +
		fmt.Sprintf("Delete SDR supported                : %v\n", res.SDROperationSupport.SupportDeleteSDR) +
		fmt.Sprintf("Partial Add SDR supported           : %v\n", res.SDROperationSupport.SupportPartialAddSDR) +
		fmt.Sprintf("Reserve SDR repository supported    : %v\n", res.SDROperationSupport.SupportReserveSDRRepo) +
		fmt.Sprintf("SDR Repository Alloc info supported : %v\n", res.SDROperationSupport.SupportGetSDRRepoAllocInfo)
}

func (res *GetSDRRepoInfoResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (c *Client) GetSDRRepoInfo(ctx context.Context) (response *GetSDRRepoInfoResponse, err error) {
	request := &GetSDRRepoInfoRequest{}
	response = &GetSDRRepoInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

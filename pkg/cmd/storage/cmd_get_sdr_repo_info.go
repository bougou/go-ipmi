package storage

import (
	"fmt"
	"time"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *GetSDRRepoInfoRequest) Command() ipmi.Command {
	return ipmi.CommandGetSDRRepoInfo
}

func (res *GetSDRRepoInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 14)
	}

	res.SDRVersion, _, _ = ipmi.UnpackUint8(msg, 0)
	res.RecordCount, _, _ = ipmi.UnpackUint16L(msg, 1)
	res.FreeSpaceBytes, _, _ = ipmi.UnpackUint16L(msg, 3)

	addTS, _, _ := ipmi.UnpackUint32L(msg, 5)
	res.MostRecentAdditionTime = ipmi.ParseTimestamp(addTS)

	deleteTS, _, _ := ipmi.UnpackUint32L(msg, 9)
	res.MostRecentEraseTime = ipmi.ParseTimestamp(deleteTS)

	b, _, _ := ipmi.UnpackUint8(msg, 13)
	res.SDROperationSupport = SDROperationSupport{
		Overflow:                     ipmi.IsBit7Set(b),
		SupportModalSDRRepoUpdate:    ipmi.IsBit6Set(b),
		SupportNonModalSDRRepoUpdate: ipmi.IsBit5Set(b),
		SupportDeleteSDR:             ipmi.IsBit3Set(b),
		SupportPartialAddSDR:         ipmi.IsBit2Set(b),
		SupportReserveSDRRepo:        ipmi.IsBit1Set(b),
		SupportGetSDRRepoAllocInfo:   ipmi.IsBit0Set(b),
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
		fmt.Sprintf("Most recent Addition                : %s\n", res.MostRecentAdditionTime.Format(ipmi.TimeFormat)) +
		fmt.Sprintf("Most recent Erase                   : %s\n", res.MostRecentEraseTime.Format(ipmi.TimeFormat)) +
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

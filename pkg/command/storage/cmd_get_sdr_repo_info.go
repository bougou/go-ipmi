package storage

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
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

func (req *GetSDRRepoInfoRequest) Command() types.Command {
	return types.CommandGetSDRRepoInfo
}

func (res *GetSDRRepoInfoResponse) Pack() []byte {
	out := make([]byte, 14)
	types.PackUint8(res.SDRVersion, out, 0)
	types.PackUint16L(res.RecordCount, out, 1)
	types.PackUint16L(res.FreeSpaceBytes, out, 3)
	// LS-byte-first unsigned 32-bit Unix timestamp per §33.9; wraps at 2106-02-07.
	types.PackUint32L(uint32(res.MostRecentAdditionTime.Unix()), out, 5)
	types.PackUint32L(uint32(res.MostRecentEraseTime.Unix()), out, 9)
	var b uint8
	if res.SDROperationSupport.Overflow {
		b = types.SetBit7(b)
	}
	if res.SDROperationSupport.SupportModalSDRRepoUpdate {
		b = types.SetBit6(b)
	}
	if res.SDROperationSupport.SupportNonModalSDRRepoUpdate {
		b = types.SetBit5(b)
	}
	if res.SDROperationSupport.SupportDeleteSDR {
		b = types.SetBit3(b)
	}
	if res.SDROperationSupport.SupportPartialAddSDR {
		b = types.SetBit2(b)
	}
	if res.SDROperationSupport.SupportReserveSDRRepo {
		b = types.SetBit1(b)
	}
	if res.SDROperationSupport.SupportGetSDRRepoAllocInfo {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 13)
	return out
}

func (res *GetSDRRepoInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 14)
	}

	res.SDRVersion, _, _ = types.UnpackUint8(msg, 0)
	res.RecordCount, _, _ = types.UnpackUint16L(msg, 1)
	res.FreeSpaceBytes, _, _ = types.UnpackUint16L(msg, 3)

	addTS, _, _ := types.UnpackUint32L(msg, 5)
	res.MostRecentAdditionTime = types.ParseTimestamp(addTS)

	deleteTS, _, _ := types.UnpackUint32L(msg, 9)
	res.MostRecentEraseTime = types.ParseTimestamp(deleteTS)

	b, _, _ := types.UnpackUint8(msg, 13)
	res.SDROperationSupport = SDROperationSupport{
		Overflow:                     types.IsBit7Set(b),
		SupportModalSDRRepoUpdate:    types.IsBit6Set(b),
		SupportNonModalSDRRepoUpdate: types.IsBit5Set(b),
		SupportDeleteSDR:             types.IsBit3Set(b),
		SupportPartialAddSDR:         types.IsBit2Set(b),
		SupportReserveSDRRepo:        types.IsBit1Set(b),
		SupportGetSDRRepoAllocInfo:   types.IsBit0Set(b),
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
		fmt.Sprintf("Most recent Addition                : %s\n", res.MostRecentAdditionTime.Format(types.TimeFormat)) +
		fmt.Sprintf("Most recent Erase                   : %s\n", res.MostRecentEraseTime.Format(types.TimeFormat)) +
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

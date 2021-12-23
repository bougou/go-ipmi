package ipmi

import (
	"fmt"
	"time"
)

// 33.9 Get SDR Repository Info Command
type GetSDRRepoInfoRequest struct {
	// empty
}

type GetSDRRepoInfoResponse struct {
	SDRVersion               uint8  // version number of the SDR command set for the SDR Device. 51h for this specification.
	RecordCount              uint16 // LS Byte first
	FreeSpeceBytes           uint16 // LS Byte first
	MostRecentAddititionTime time.Time
	MostRecentEraseTime      time.Time

	SDROperationSupport SDROperationSupport
}

type SDROperationSupport struct {
	Overflow                     bool
	SupportModalSDRRepoUpdate    bool
	SupportNonModalSDRRepoUpdate bool
	SupportDeleteSDR             bool
	SupportParitialAddSDR        bool
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
		return ErrUnpackedDataTooShort
	}

	res.SDRVersion, _, _ = unpackUint8(msg, 0)
	res.RecordCount, _, _ = unpackUint16L(msg, 1)
	res.FreeSpeceBytes, _, _ = unpackUint16L(msg, 3)

	addTS, _, _ := unpackUint32L(msg, 5)
	res.MostRecentAddititionTime = parseTimestamp(addTS)

	deleteTS, _, _ := unpackUint32L(msg, 9)
	res.MostRecentEraseTime = parseTimestamp(deleteTS)

	b, _, _ := unpackUint8(msg, 13)
	res.SDROperationSupport = SDROperationSupport{
		Overflow:                     b&0x80 == 0x80, // bit 7 is set
		SupportModalSDRRepoUpdate:    b&0x40 == 0x40, // bit 6 is set
		SupportNonModalSDRRepoUpdate: b&0x20 == 0x20, // bit 5 is set
		SupportDeleteSDR:             b&0x08 == 0x08, // bit 3 is set
		SupportParitialAddSDR:        b&0x04 == 0x04, // bit 2 is set
		SupportReserveSDRRepo:        b&0x02 == 0x02, // bit 1 is set
		SupportGetSDRRepoAllocInfo:   b&0x01 == 0x01, // bit 0 is set
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

	return fmt.Sprintf(`SDR Version                         : %#02x
Record Count                        : %d
Free Space                          : %d bytes
Most recent Addition                : %s
Most recent Erase                   : %s
SDR overflow                        : %v
SDR Repository Update Support       : %s
Delete SDR supported                : %v
Partial Add SDR supported           : %v
Reserve SDR repository supported    : %v
SDR Repository Alloc info supported : %v`,
		res.SDRVersion,
		res.RecordCount,
		res.FreeSpeceBytes,
		res.MostRecentAddititionTime,
		res.MostRecentEraseTime,
		res.SDROperationSupport.Overflow,
		s,
		res.SDROperationSupport.SupportDeleteSDR,
		res.SDROperationSupport.SupportParitialAddSDR,
		res.SDROperationSupport.SupportReserveSDRRepo,
		res.SDROperationSupport.SupportGetSDRRepoAllocInfo,
	)
}

func (res *GetSDRRepoInfoResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (c *Client) GetSDRRepoInfo() (response *GetSDRRepoInfoResponse, err error) {
	request := &GetSDRRepoInfoRequest{}
	response = &GetSDRRepoInfoResponse{}
	err = c.Exchange(request, response)
	return
}

package storage

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
)

// GetSELInfoRequest (31.2) command returns the number of entries in the SEL.
type GetSELInfoRequest struct {
	// empty
}

type GetSELInfoResponse struct {
	SELVersion         uint8
	Entries            uint16
	FreeBytes          uint16
	RecentAdditionTime time.Time
	RecentEraseTime    time.Time
	OperationSupport   SELOperationSupport
}

type SELOperationSupport struct {
	Overflow        bool
	DeleteSEL       bool
	PartialAddSEL   bool
	ReserveSEL      bool
	GetSELAllocInfo bool
}

func (req *GetSELInfoRequest) Command() types.Command {
	return types.CommandGetSELInfo
}

func (req *GetSELInfoRequest) Pack() []byte {
	// empty request data
	return []byte{}
}

func (res *GetSELInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 14)
	}
	res.SELVersion, _, _ = types.UnpackUint8(msg, 0)
	res.Entries, _, _ = types.UnpackUint16L(msg, 1)
	res.FreeBytes, _, _ = types.UnpackUint16L(msg, 3)

	addTS, _, _ := types.UnpackUint32L(msg, 5)
	res.RecentAdditionTime = types.ParseTimestamp(addTS)

	eraseTS, _, _ := types.UnpackUint32L(msg, 9)
	res.RecentEraseTime = types.ParseTimestamp(eraseTS)

	b := msg[13]
	res.OperationSupport = SELOperationSupport{
		Overflow:        types.IsBit7Set(b),
		DeleteSEL:       types.IsBit3Set(b),
		PartialAddSEL:   types.IsBit2Set(b),
		ReserveSEL:      types.IsBit1Set(b),
		GetSELAllocInfo: types.IsBit0Set(b),
	}

	return nil
}

func (r *GetSELInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *GetSELInfoResponse) Format() string {
	var version string
	if res.SELVersion == 0x51 {
		version = "1.5"
	}

	var usedBytes int = int(res.Entries) * 16
	totalBytes := usedBytes + int(res.FreeBytes)
	var usedPct float64 = 100 * float64(usedBytes) / float64(totalBytes)

	return "" +
		"SEL Information\n" +
		fmt.Sprintf("Version                      : %s (v1.5, v2 compliant)\n", version) +
		fmt.Sprintf("Entries                      : %d\n", res.Entries) +
		fmt.Sprintf("Free Space                   : %d bytes\n", res.FreeBytes) +
		fmt.Sprintf("Percent Used                 : %.2f%%\n", usedPct) +
		fmt.Sprintf("Last Add Time                : %s\n", res.RecentAdditionTime.Format(types.TimeFormat)) +
		fmt.Sprintf("Last Del Time                : %s\n", res.RecentEraseTime.Format(types.TimeFormat)) +
		fmt.Sprintf("Overflow                     : %v\n", res.OperationSupport.Overflow) +
		fmt.Sprintf("Delete SEL supported         : %v\n", res.OperationSupport.DeleteSEL) +
		fmt.Sprintf("Partial Add SEL supported    : %v\n", res.OperationSupport.PartialAddSEL) +
		fmt.Sprintf("Reserve SEL supported        : %v\n", res.OperationSupport.ReserveSEL) +
		fmt.Sprintf("Get SEL Alloc Info supported : %v\n", res.OperationSupport.GetSELAllocInfo)
}

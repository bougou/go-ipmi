package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 31.9 Clear SEL Command
type ClearSELRequest struct {
	ReservationID        uint16 // LS Byte first
	GetErasureStatusFlag bool
}

type ClearSELResponse struct {
	ErasureProgressStatus uint8
}

func (req *ClearSELRequest) Pack() []byte {
	var out = make([]byte, 6)
	ipmi.PackUint16L(req.ReservationID, out, 0)
	ipmi.PackUint8('C', out, 2) // fixed 'C' char
	ipmi.PackUint8('L', out, 3) // fixed 'L' char
	ipmi.PackUint8('R', out, 4) // fixed 'R' char
	if req.GetErasureStatusFlag {
		ipmi.PackUint8(0x00, out, 5) //  get erasure status
	} else {
		ipmi.PackUint8(0xaa, out, 5) //  initiate erase
	}
	return out
}

func (req *ClearSELRequest) Command() ipmi.Command {
	return ipmi.CommandClearSEL
}

func (res *ClearSELResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ErasureProgressStatus, _, _ = ipmi.UnpackUint8(msg, 0)
	return nil
}

func (res *ClearSELResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *ClearSELResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

package storage

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 31.4 Reserve SEL Command
)

type ReserveSELRequest struct {
	// empty
}

type ReserveSELResponse struct {
	ReservationID uint16
}

func (req *ReserveSELRequest) Command() types.Command {
	return types.CommandReserveSEL
}

func (req *ReserveSELRequest) Pack() []byte {
	return nil
}

func (res *ReserveSELResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.ReservationID, _, _ = types.UnpackUint16L(msg, 0)
	return nil
}

func (*ReserveSELResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *ReserveSELResponse) Format() string {
	return ""
}

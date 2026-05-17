package storage

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 35.4 Reserve Device SDR Repository Command
)

type ReserveDeviceSDRRepoRequest struct {
	// empty
}

type ReserveDeviceSDRRepoResponse struct {
	ReservationID uint16
}

func (req *ReserveDeviceSDRRepoRequest) Command() ipmi.Command {
	return ipmi.CommandReserveDeviceSDRRepo
}

func (req *ReserveDeviceSDRRepoRequest) Pack() []byte {
	return []byte{}
}

func (res *ReserveDeviceSDRRepoResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ReservationID, _, _ = ipmi.UnpackUint16L(msg, 0)
	return nil
}

func (r *ReserveDeviceSDRRepoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ReserveDeviceSDRRepoResponse) Format() string {
	return ""
}

// This command is used to obtain a Reservation ID.

package storage

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 33.11 Reserve SDR Repository Command
)

type ReserveSDRRepoRequest struct {
	// empty
}

type ReserveSDRRepoResponse struct {
	ReservationID uint16
}

func (req *ReserveSDRRepoRequest) Command() types.Command {
	return types.CommandReserveSDRRepo
}

func (req *ReserveSDRRepoRequest) Pack() []byte {
	return []byte{}
}

func (res *ReserveSDRRepoResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ReservationID, _, _ = types.UnpackUint16L(msg, 0)
	return nil
}

func (r *ReserveSDRRepoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ReserveSDRRepoResponse) Format() string {
	return ""
}

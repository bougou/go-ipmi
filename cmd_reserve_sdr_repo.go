package ipmi

import "context"

// 33.11 Reserve SDR Repository Command
type ReserveSDRRepoRequest struct {
	// empty
}

type ReserveSDRRepoResponse struct {
	ReservationID uint16
}

func (req *ReserveSDRRepoRequest) Command() Command {
	return CommandReserveSDRRepo
}

func (req *ReserveSDRRepoRequest) Pack() []byte {
	return []byte{}
}

func (res *ReserveSDRRepoResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ReservationID, _, _ = unpackUint16L(msg, 0)
	return nil
}

func (r *ReserveSDRRepoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ReserveSDRRepoResponse) Format() string {
	return ""
}

func (c *Client) ReserveSDRRepo(ctx context.Context) (response *ReserveSDRRepoResponse, err error) {
	request := &ReserveSDRRepoRequest{}
	response = &ReserveSDRRepoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

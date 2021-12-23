package ipmi

import "fmt"

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
	packUint16L(req.ReservationID, out, 0)
	packUint8('C', out, 2) // fixed 'C' char
	packUint8('L', out, 3) // fixed 'L' char
	packUint8('R', out, 4) // fixed 'R' char
	if req.GetErasureStatusFlag {
		packUint8(0x00, out, 5) //  get erasure status
	} else {
		packUint8(0xaa, out, 5) //  initiate erase
	}
	return out
}

func (req *ClearSELRequest) Command() Command {
	return CommandClearSEL
}

func (res *ClearSELResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.ErasureProgressStatus, _, _ = unpackUint8(msg, 0)
	return nil
}

func (res *ClearSELResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *ClearSELResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) ClearSEL(reservationID uint16) (response *ClearSELResponse, err error) {
	request := &ClearSELRequest{
		ReservationID:        reservationID,
		GetErasureStatusFlag: false,
	}
	response = &ClearSELResponse{}
	err = c.Exchange(request, response)
	return
}

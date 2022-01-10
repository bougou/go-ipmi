package ipmi

// 31.4 Reserve SEL Command
type ReserveSELRequest struct {
	// empty
}

type ReserveSELResponse struct {
	ReservationID uint16
}

func (req *ReserveSELRequest) Command() Command {
	return CommandReserveSEL
}

func (req *ReserveSELRequest) Pack() []byte {
	return nil
}

func (res *ReserveSELResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	res.ReservationID, _, _ = unpackUint16L(msg, 0)
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

func (c *Client) ReserveSEL() (response *ReserveSELResponse, err error) {
	request := &ReserveSELRequest{}
	response = &ReserveSELResponse{}
	err = c.Exchange(request, response)
	return
}

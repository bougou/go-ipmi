package ipmi

import (
	"fmt"
	"time"
)

// 22.14 Get System GUID Command
type GetSystemGUIDRequest struct {
	// empty
}

type GetSystemGUIDResponse struct {
	// Note that the individual fields within the GUID are stored least-significant byte first
	GUID [16]byte
}

func (req *GetSystemGUIDRequest) Command() Command {
	return CommandGetSystemGUID
}

func (req *GetSystemGUIDRequest) Pack() []byte {
	return nil
}

func (res *GetSystemGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShort
	}
	b, _, _ := unpackBytes(msg, 0, 16)
	res.GUID = array16(b)
	return nil
}

func (*GetSystemGUIDResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSystemGUIDResponse) Format() string {
	u, err := ParseGUID(res.GUID[:])
	if err != nil {
		return ""
	}

	out := fmt.Sprintf(`System GUID  : %s`, u.String())
	out += fmt.Sprintf("\nGUID Version : %s", u.Version().String())

	if uint8(u.Version()) == 1 {
		sec, nsec := u.Time().UnixTime()
		out += fmt.Sprintf("\nTimestamp    : %s", time.Unix(sec, nsec).Format(time.RFC3339))
	}

	return out
}

func (c *Client) GetSystemGUID() (response *GetSystemGUIDResponse, err error) {
	request := &GetSystemGUIDRequest{}
	response = &GetSystemGUIDResponse{}
	err = c.Exchange(request, response)
	return
}

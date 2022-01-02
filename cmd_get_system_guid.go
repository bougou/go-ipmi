package ipmi

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// 22.14 Get System GUID Command
type GetSystemGUIDRequest struct {
	// empty
}

type GetSystemGUIDResponse struct {
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

type GUID struct {
	Node               [6]byte
	ClockSeqLow        uint8
	ClockSeqHigh       uint8
	TimeHighAndVersion uint16
	TimeMid            uint16
	TimeLow            uint32
}

type GUIDVersion uint8

const (
	GUIDVersionTimebased     GUIDVersion = 1 // 0001b
	GUIDVersionDCESec        GUIDVersion = 2 // 0010b
	GUIDVersionNamebaesdMD5  GUIDVersion = 3 // 0011b
	GUIDVersionRandom        GUIDVersion = 4 // 0100b
	GUIDVersionNamebasedSHA1 GUIDVersion = 5 // 0101b
)

func (res *GetSystemGUIDResponse) Format() string {
	uuidRFC4122MSB := make([]byte, 16)
	for i := 0; i < 16; i++ {
		uuidRFC4122MSB[i] = res.GUID[:][15-i]
	}
	u, err := uuid.FromBytes(uuidRFC4122MSB)
	if err != nil {
		return "Invalid UUID Bytes"
	}

	sec, nsec := u.Time().UnixTime()
	return fmt.Sprintf(`System GUID  : %s
Timestamp    : %s`,
		u.String(),
		time.Unix(sec, nsec).Format(time.RFC3339),
	)
}

func (c *Client) GetSystemGUID() (response *GetSystemGUIDResponse, err error) {
	request := &GetSystemGUIDRequest{}
	response = &GetSystemGUIDResponse{}
	err = c.Exchange(request, response)
	return
}

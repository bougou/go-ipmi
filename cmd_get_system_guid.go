package ipmi

import (
	"context"
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
		return ErrUnpackedDataTooShortWith(len(msg), 16)
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
	guidMode := GUIDModeSMBIOS
	u, err := ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}
	sec, nsec := u.Time().UnixTime()

	return "" +
		fmt.Sprintf("System GUID       : %s\n", u.String()) +
		fmt.Sprintf("UUID Encoding     : %s\n", guidMode) +
		fmt.Sprintf("UUID Version      : %s\n", UUIDVersionString(u)) +
		fmt.Sprintf("Timestamp         : %s\n", time.Unix(sec, nsec).Format(timeFormat)) +
		fmt.Sprintf("Timestamp(Legacy) : %s\n", IPMILegacyGUIDTime(u).Format(timeFormat))
}

func (c *Client) GetSystemGUID(ctx context.Context) (response *GetSystemGUIDResponse, err error) {
	request := &GetSystemGUIDRequest{}
	response = &GetSystemGUIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

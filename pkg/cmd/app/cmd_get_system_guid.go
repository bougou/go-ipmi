package app

import (
	"fmt"
	"time"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 22.14 Get System GUID Command
type GetSystemGUIDRequest struct {
	// empty
}

type GetSystemGUIDResponse struct {
	// Note that the individual fields within the GUID are stored least-significant byte first
	GUID [16]byte
}

func (req *GetSystemGUIDRequest) Command() ipmi.Command {
	return ipmi.CommandGetSystemGUID
}

func (req *GetSystemGUIDRequest) Pack() []byte {
	return nil
}

func (res *GetSystemGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 16)
	}
	b, _, _ := ipmi.UnpackBytes(msg, 0, 16)
	res.GUID = ipmi.Array16(b)
	return nil
}

func (*GetSystemGUIDResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSystemGUIDResponse) Format() string {
	guidMode := ipmi.GUIDModeSMBIOS
	u, err := ipmi.ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}
	sec, nsec := u.Time().UnixTime()

	return "" +
		fmt.Sprintf("System GUID       : %s\n", u.String()) +
		fmt.Sprintf("UUID Encoding     : %s\n", guidMode) +
		fmt.Sprintf("UUID Version      : %s\n", ipmi.UUIDVersionString(u)) +
		fmt.Sprintf("Timestamp         : %s\n", time.Unix(sec, nsec).Format(ipmi.TimeFormat)) +
		fmt.Sprintf("Timestamp(Legacy) : %s\n", ipmi.IPMILegacyGUIDTime(u).Format(ipmi.TimeFormat))
}

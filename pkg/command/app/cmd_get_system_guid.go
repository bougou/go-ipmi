package app

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.14 Get System GUID Command
type GetSystemGUIDRequest struct {
	// empty
}

type GetSystemGUIDResponse struct {
	// Note that the individual fields within the GUID are stored least-significant byte first
	GUID [16]byte
}

func (req *GetSystemGUIDRequest) Command() types.Command {
	return types.CommandGetSystemGUID
}

func (req *GetSystemGUIDRequest) Pack() []byte {
	return nil
}

func (res *GetSystemGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 16)
	}
	b, _, _ := types.UnpackBytes(msg, 0, 16)
	res.GUID = types.Array16(b)
	return nil
}

func (*GetSystemGUIDResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSystemGUIDResponse) Format() string {
	guidMode := types.GUIDModeSMBIOS
	u, err := types.ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}
	sec, nsec := u.Time().UnixTime()

	return "" +
		fmt.Sprintf("System GUID       : %s\n", u.String()) +
		fmt.Sprintf("UUID Encoding     : %s\n", guidMode) +
		fmt.Sprintf("UUID Version      : %s\n", types.UUIDVersionString(u)) +
		fmt.Sprintf("Timestamp         : %s\n", time.Unix(sec, nsec).Format(types.TimeFormat)) +
		fmt.Sprintf("Timestamp(Legacy) : %s\n", types.IPMILegacyGUIDTime(u).Format(types.TimeFormat))
}

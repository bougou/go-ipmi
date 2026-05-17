package app

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 20.8 Get Device GUID Command
type GetDeviceGUIDRequest struct {
	// empty
}

type GetDeviceGUIDResponse struct {
	GUID [16]byte
}

func (req *GetDeviceGUIDRequest) Command() ipmi.Command {
	return ipmi.CommandGetDeviceGUID
}

func (req *GetDeviceGUIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetDeviceGUIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDeviceGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	guid, _, _ := ipmi.UnpackBytes(msg, 0, 16)
	res.GUID = ipmi.Array16(guid)
	return nil
}

func (res *GetDeviceGUIDResponse) Format() string {
	guidMode := ipmi.GUIDModeSMBIOS
	u, err := ipmi.ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}

	return fmt.Sprintf("GUID: %s", u.String())
}

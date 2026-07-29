package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 20.8 Get Device GUID Command
type GetDeviceGUIDRequest struct {
	// empty
}

type GetDeviceGUIDResponse struct {
	GUID [16]byte
}

func (req *GetDeviceGUIDRequest) Command() types.Command {
	return types.CommandGetDeviceGUID
}

func (req *GetDeviceGUIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetDeviceGUIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDeviceGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	guid, _, _ := types.UnpackBytes(msg, 0, 16)
	res.GUID = types.Array16(guid)
	return nil
}

func (res *GetDeviceGUIDResponse) Format() string {
	guidMode := types.GUIDModeSMBIOS
	u, err := types.ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}

	return fmt.Sprintf("GUID: %s", u.String())
}

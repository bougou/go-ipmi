package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 34.1 Get FRU Inventory Area Info Command
type GetFRUInventoryAreaInfoRequest struct {
	FRUDeviceID uint8
}

type GetFRUInventoryAreaInfoResponse struct {
	AreaSizeBytes         uint16
	DeviceAccessedByWords bool // false means Device is accessed by Bytes
}

func (req *GetFRUInventoryAreaInfoRequest) Command() types.Command {
	return types.CommandGetFRUInventoryAreaInfo
}

func (req *GetFRUInventoryAreaInfoRequest) Pack() []byte {
	return []byte{req.FRUDeviceID}
}

func (res *GetFRUInventoryAreaInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.AreaSizeBytes, _, _ = types.UnpackUint16L(msg, 0)
	b, _, _ := types.UnpackUint8(msg, 2)
	res.DeviceAccessedByWords = types.IsBit0Set(b)
	return nil
}

func (r *GetFRUInventoryAreaInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetFRUInventoryAreaInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("FRU size = %d bytes (accessed by %s)\n", res.AreaSizeBytes, types.FormatBool(res.DeviceAccessedByWords, "words", "bytes"))
}

// This command returns overall the size of the FRU Inventory Area in this device, in bytes.

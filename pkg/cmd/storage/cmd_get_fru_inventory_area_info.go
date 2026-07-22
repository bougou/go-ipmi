package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// v2.0§34.1 Get FRU Inventory Area Info Command. Returns the FRU Inventory Area size in bytes.
type GetFRUInventoryAreaInfoRequest struct {
	FRUDeviceID uint8
}

type GetFRUInventoryAreaInfoResponse struct {
	AreaSizeBytes         uint16
	DeviceAccessedByWords bool // false: bytes; true: words (bit 0, v2.0§34.1 Table 34-2)
}

func (req *GetFRUInventoryAreaInfoRequest) Command() types.Command {
	return types.CommandGetFRUInventoryAreaInfo
}

func (req *GetFRUInventoryAreaInfoRequest) Pack() []byte {
	return []byte{req.FRUDeviceID}
}

func (req *GetFRUInventoryAreaInfoRequest) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	req.FRUDeviceID, _, _ = types.UnpackUint8(msg, 0)
	return nil
}

func (res *GetFRUInventoryAreaInfoResponse) Pack() []byte {
	out := make([]byte, 3)
	types.PackUint16L(res.AreaSizeBytes, out, 0)
	var b uint8
	if res.DeviceAccessedByWords {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 2)
	return out
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
	return fmt.Sprintf("FRU size = %d bytes (accessed by %s)\n",
		res.AreaSizeBytes, types.FormatBool(res.DeviceAccessedByWords, "words", "bytes"))
}

package ipmi

import (
	"context"
	"fmt"
)

// 34.1 Get FRU Inventory Area Info Command
type GetFRUInventoryAreaInfoRequest struct {
	FRUDeviceID uint8
}

type GetFRUInventoryAreaInfoResponse struct {
	AreaSizeBytes         uint16
	DeviceAccessedByWords bool // false means Device is accessed by Bytes
}

func (req *GetFRUInventoryAreaInfoRequest) Command() Command {
	return CommandGetFRUInventoryAreaInfo
}

func (req *GetFRUInventoryAreaInfoRequest) Pack() []byte {
	return []byte{req.FRUDeviceID}
}

func (res *GetFRUInventoryAreaInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.AreaSizeBytes, _, _ = unpackUint16L(msg, 0)
	b, _, _ := unpackUint8(msg, 2)
	res.DeviceAccessedByWords = isBit0Set(b)
	return nil
}

func (r *GetFRUInventoryAreaInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetFRUInventoryAreaInfoResponse) Format() string {
	return fmt.Sprintf(`fru.size = %d bytes (accessed by %s)`,
		res.AreaSizeBytes,
		formatBool(res.DeviceAccessedByWords, "words", "bytes"),
	)
}

// This command returns overall the size of the FRU Inventory Area in this device, in bytes.
func (c *Client) GetFRUInventoryAreaInfo(ctx context.Context, fruDeviceID uint8) (response *GetFRUInventoryAreaInfoResponse, err error) {
	request := &GetFRUInventoryAreaInfoRequest{
		FRUDeviceID: fruDeviceID,
	}
	response = &GetFRUInventoryAreaInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

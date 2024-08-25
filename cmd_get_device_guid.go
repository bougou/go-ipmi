package ipmi

import (
	"fmt"
)

// 20.8 Get Device GUID Command
type GetDeviceGUIDRequest struct {
	// empty
}

type GetDeviceGUIDResponse struct {
	GUID [16]byte
}

func (req *GetDeviceGUIDRequest) Command() Command {
	return CommandGetDeviceGUID
}

func (req *GetDeviceGUIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetDeviceGUIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDeviceGUIDResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	guid, _, _ := unpackBytes(msg, 0, 16)
	res.GUID = array16(guid)
	return nil
}

func (res *GetDeviceGUIDResponse) Format() string {
	guidMode := GUIDModeSMBIOS
	u, err := ParseGUID(res.GUID[:], guidMode)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}

	return fmt.Sprintf("GUID: %s", u.String())
}

func (c *Client) GetDeviceGUID() (response *GetDeviceGUIDResponse, err error) {
	request := &GetDeviceGUIDRequest{}
	response = &GetDeviceGUIDResponse{}
	err = c.Exchange(request, response)
	return
}

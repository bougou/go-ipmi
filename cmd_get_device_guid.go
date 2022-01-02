package ipmi

import (
	"fmt"

	"github.com/google/uuid"
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
		return ErrUnpackedDataTooShort
	}

	guid, _, _ := unpackBytes(msg, 0, 16)
	res.GUID = array16(guid)
	return nil
}

func (res *GetDeviceGUIDResponse) Format() string {

	uuidRFC4122MSB := make([]byte, 16)
	for i := 0; i < 16; i++ {
		uuidRFC4122MSB[i] = res.GUID[:][15-i]
	}
	u, err := uuid.FromBytes(uuidRFC4122MSB)
	if err != nil {
		return "Invalid UUID Bytes"
	}

	return fmt.Sprintf(`GUID: %s`, u.String())
}

func (c *Client) GetDeviceGUID() (response *GetDeviceGUIDResponse, err error) {
	request := &GetDeviceGUIDRequest{}
	response = &GetDeviceGUIDResponse{}
	err = c.Exchange(request, response)
	return
}

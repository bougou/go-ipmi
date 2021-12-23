package ipmi

import "fmt"

// 22.14 Get System GUID Command
type GetSystemGUIDRequest struct {
	// empty
}

type GetSystemGUIDResponse struct {
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
		return ErrUnpackedDataTooShort
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
	return fmt.Sprintf("%s", res)
}

func (c *Client) GetSystemGUID() (*GetSystemGUIDResponse, error) {
	req := &GetSystemGUIDRequest{}
	res := &GetSystemGUIDResponse{}

	if err := c.Exchange(req, res); err != nil {
		return nil, fmt.Errorf("client exchange failed, err: %s", err)
	}
	return res, nil

}

package ipmi

import "bytes"

// 22.29 Get User Name Command
type GetUsernameRequest struct {
	// [5:0] - User ID. 000000b = reserved. (User ID 1 is permanently associated with User 1, the null user name).
	UserID uint8
}

type GetUsernameResponse struct {
	Username string
}

func (req *GetUsernameRequest) Command() Command {
	return CommandGetUsername
}

func (req *GetUsernameRequest) Pack() []byte {
	return []byte{req.UserID}
}

func (res *GetUsernameResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUsernameResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShort
	}
	username, _, _ := unpackBytes(msg, 0, 16)
	res.Username = string(bytes.TrimRight(username, "\x00"))
	return nil
}

func (res *GetUsernameResponse) Format() string {
	return ""
}

func (c *Client) GetUsername(userID uint8) (response *GetUsernameResponse, err error) {
	request := &GetUsernameRequest{
		UserID: userID,
	}
	response = &GetUsernameResponse{}
	err = c.Exchange(request, response)
	return
}

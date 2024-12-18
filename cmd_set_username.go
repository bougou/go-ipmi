package ipmi

import "context"

// 22.28 Set User Name Command
type SetUsernameRequest struct {
	// [5:0] - User ID. 000000b = reserved. (User ID 1 is permanently associated with User 1, the null user name).
	UserID uint8

	// User Name String in ASCII, 16 bytes, max. Strings with fewer than 16
	// characters are terminated with a null (00h) character and 00h padded to 16
	// bytes. When the string is read back using the Get User Name command,
	// those bytes shall be returned as 0s.
	// Here if string length is longer than 16, it would be auto truncated.
	Username string
}

type SetUsernameResponse struct {
	GUID [16]byte
}

func (req *SetUsernameRequest) Command() Command {
	return CommandSetUsername
}

func (req *SetUsernameRequest) Pack() []byte {
	out := make([]byte, 17)
	packUint8(req.UserID, out, 0)

	username := padBytes(req.Username, 16, 0x00)
	packBytes(username, out, 1)
	return out
}

func (res *SetUsernameResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetUsernameResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetUsernameResponse) Format() string {
	return ""
}

func (c *Client) SetUsername(ctx context.Context, userID uint8, username string) (response *SetUsernameResponse, err error) {
	request := &SetUsernameRequest{
		UserID:   userID,
		Username: username,
	}
	response = &SetUsernameResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

package app

import (
	"bytes"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.29 Get User Name Command
type GetUsernameRequest struct {
	// [5:0] - User ID. 000000b = reserved. (User ID 1 is permanently associated with User 1, the null user name).
	UserID uint8
}

type GetUsernameResponse struct {
	Username string
}

func (req *GetUsernameRequest) Command() types.Command {
	return types.CommandGetUsername
}

func (req *GetUsernameRequest) Pack() []byte {
	return []byte{req.UserID}
}

func (res *GetUsernameResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUsernameResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 16)
	}
	username, _, _ := types.UnpackBytes(msg, 0, 16)
	res.Username = string(bytes.TrimRight(username, "\x00"))
	return nil
}

func (res *GetUsernameResponse) Format() string {
	return ""
}

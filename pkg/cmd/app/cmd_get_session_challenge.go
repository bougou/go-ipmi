package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.16
type GetSessionChallengeRequest struct {
	// Authentication Type for Challenge
	// indicating what type of authentication type the console wants to use.
	AuthType types.AuthType

	// Sixteen-bytes. All 0s for null user name (User 1)
	Username [16]byte
}

type GetSessionChallengeResponse struct {
	TemporarySessionID uint32 // LS byte first
	Challenge          [16]byte
}

func (req *GetSessionChallengeRequest) Command() types.Command {
	return types.CommandGetSessionChallenge
}

func (req *GetSessionChallengeRequest) Pack() []byte {
	out := make([]byte, 17)
	types.PackUint8(uint8(req.AuthType), out, 0)
	types.PackBytes(req.Username[:], out, 1)
	return out
}

func (res *GetSessionChallengeResponse) Unpack(msg []byte) error {
	if len(msg) < 20 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 20)
	}
	res.TemporarySessionID, _, _ = types.UnpackUint32L(msg, 0)
	b, _, _ := types.UnpackBytes(msg, 4, 16)
	res.Challenge = types.Array16(b)
	return nil
}

func (*GetSessionChallengeResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "invalid user name",
		0x82: "null user name (User 1) not enabled",
	}
}

func (res *GetSessionChallengeResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// The command selects which of the BMC-supported authentication types the Remote Console would like to use,
// and a username that selects which set of user information should be used for the session

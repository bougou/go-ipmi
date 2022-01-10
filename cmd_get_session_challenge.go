package ipmi

import "fmt"

// 22.16
type GetSessionChallengeRequest struct {
	// Authentication Type for Challenge
	// indicating what type of authentication type the console wants to use.
	AuthType AuthType

	// Sixteen-bytes. All 0s for null user name (User 1)
	Username [16]byte
}

type GetSessionChallengeResponse struct {
	TemporarySessionID uint32 // LS byte first
	Challenge          [16]byte
}

func (req *GetSessionChallengeRequest) Command() Command {
	return CommandGetSessionChallenge
}

func (req *GetSessionChallengeRequest) Pack() []byte {
	out := make([]byte, 17)
	packUint8(uint8(req.AuthType), out, 0)
	packBytes(req.Username[:], out, 1)
	return out
}

func (res *GetSessionChallengeResponse) Unpack(msg []byte) error {
	if len(msg) < 20 {
		return ErrUnpackedDataTooShort
	}
	res.TemporarySessionID, _, _ = unpackUint32L(msg, 0)
	b, _, _ := unpackBytes(msg, 4, 16)
	res.Challenge = array16(b)
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
func (c *Client) GetSessionChallenge() (response *GetSessionChallengeResponse, err error) {
	username := padBytes(c.Username, 16, 0x00)
	request := &GetSessionChallengeRequest{
		AuthType: c.session.authType,
		Username: array16(username),
	}

	response = &GetSessionChallengeResponse{}
	err = c.Exchange(request, response)
	if err != nil {
		return
	}

	c.session.v15.sessionID = response.TemporarySessionID
	c.session.v15.challenge = response.Challenge

	return
}

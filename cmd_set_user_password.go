package ipmi

// 22.30 Set User Password Command
type SetUserPasswordRequest struct {
	// [5:0] - User ID. 000000b = reserved. (User ID 1 is permanently associated with User 1, the null user name).
	UserID uint8

	// The BMC shall maintain an internal tag that indicates whether
	// the password was set as a 16-byte or as a 20-byte password.
	Stored20 bool

	Operation PasswordOperation

	Password string
}

type PasswordOperation uint8

const (
	PasswordOperationDisableUser  PasswordOperation = 0x00
	PasswordOperationEnableUser   PasswordOperation = 0x01
	PasswordOperationSetPassword  PasswordOperation = 0x02
	PasswordOperationTestPassword PasswordOperation = 0x03
)

type SetUserPasswordResponse struct {
	// empty
}

func (req *SetUserPasswordRequest) Command() Command {
	return CommandSetUserPassword
}

func (req *SetUserPasswordRequest) Pack() []byte {
	out := make([]byte, 2)
	b := req.UserID & 0x3f
	if req.Stored20 {
		b = setBit7(b)
	}
	packUint8(b, out, 0)
	packUint8(uint8(req.Operation)&0x03, out, 1)

	if req.Operation == PasswordOperationSetPassword || req.Operation == PasswordOperationTestPassword {
		var passwordStored []byte
		if req.Stored20 {
			passwordStored = padBytes(req.Password, 20, 0x00)
		} else {
			passwordStored = padBytes(req.Password, 16, 0x00)
		}
		out = append(out, passwordStored...)
	}

	return out
}

func (res *SetUserPasswordResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetUserPasswordResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetUserPasswordResponse) Format() string {
	return ""
}

func (c *Client) SetUserPassword(userID uint8, password string, stored20 bool) (response *SetUserPasswordResponse, err error) {
	request := &SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: PasswordOperationSetPassword,
		Password:  password,
	}
	response = &SetUserPasswordResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) TestUserPassword(userID uint8, password string, stored20 bool) (response *SetUserPasswordResponse, err error) {
	request := &SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: PasswordOperationTestPassword,
		Password:  password,
	}
	response = &SetUserPasswordResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) DisableUser(userID uint8) (err error) {
	request := &SetUserPasswordRequest{
		UserID:    userID,
		Operation: PasswordOperationDisableUser,
	}
	response := &SetUserPasswordResponse{}
	err = c.Exchange(request, response)
	return err
}

func (c *Client) EnableUser(userID uint8) (err error) {
	request := &SetUserPasswordRequest{
		UserID:    userID,
		Operation: PasswordOperationEnableUser,
	}
	response := &SetUserPasswordResponse{}
	err = c.Exchange(request, response)
	return err
}

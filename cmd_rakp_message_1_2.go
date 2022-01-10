package ipmi

import (
	"fmt"
)

const IPMI_MAX_USER_NAME_LENGTH = 16
const IPMI_RAKP1_MESSAGE_SIZE = 44

// 13.20 RAKP Message 1
type RAKPMessage1 struct {
	MessageTag uint8

	// The Managed System's Session ID for this session, returned by the Managed System on the
	// previous RMCP+ Open Session Response message.
	ManagedSystemSessionID uint32

	// 16 bytes
	RemoteConsoleRandomNumber [16]byte

	// bit 4
	// 0b = Username/Privilege lookup.
	// 1b = Name-only lookup.
	NameOnlyLookup                 bool
	RequestedMaximumPrivilegeLevel PrivilegeLevel

	UsernameLength uint8
	Username       []byte
}

type RAKPMessage2 struct {
	// authAlg describes the authentication algorithm was agreed upon in
	// the open session request/response phase.
	// We need to know that here so that we know how many bytes (if any) to read from the packet for KeyExchangeAuthenticationCode
	authAlg AuthAlg

	MessageTag uint8

	RmcpStatusCode uint8

	// The Remote Console Session ID specified by the RMCP+ Open Session Request message associated with this response.
	RemoteConsoleSessionID uint32

	// Random number generated/selected by the managed system.
	ManagedSystemRandomNumber [16]byte

	// The Globally Unique ID (GUID) of the Managed System.
	// This value is typically specified by the client system's SMBIOS implementation. See
	// 22.14, Get System GUID Command, for additional information
	ManagedSystemGUID [16]byte

	// An integrity check value over the relevant items specified by the RAKP algorithm for RAKP Message 2.
	// The size of this field depends on the specific Authentication Algorithm
	// This field may be 0-bytes (absent) for some algorithms (e.g. RAKP-none).
	//
	// see 13.31 for how the managed system generate this HMAC
	KeyExchangeAuthenticationCode []byte
}

func (req *RAKPMessage1) Command() Command {
	return CommandNone
}

func (r *RAKPMessage1) Pack() []byte {
	var msg = make([]byte, 28+len(r.Username))
	packUint8(r.MessageTag, msg, 0)
	packUint24L(0, msg, 1) // 3 bytes reserved
	packUint32L(r.ManagedSystemSessionID, msg, 4)
	packBytes((r.RemoteConsoleRandomNumber[:]), msg, 8)

	packUint8(r.Role(), msg, 24)
	packUint16L(0, msg, 25) // 2 bytes reserved

	packUint8(r.UsernameLength, msg, 27)
	packBytes(r.Username, msg, 28)
	return msg
}

// the combination of RequestedMaximumPrivilegeLevel and NameOnlyLookup field
// The whole byte should be stored to client session for computing auth code of rakp2
func (r *RAKPMessage1) Role() uint8 {
	privilegeLevel := uint8(r.RequestedMaximumPrivilegeLevel)
	if r.NameOnlyLookup {
		privilegeLevel = setBit4(privilegeLevel)
	}
	return privilegeLevel
}

func (res *RAKPMessage2) Unpack(msg []byte) error {
	if len(msg) < 40 {
		return ErrUnpackedDataTooShort
	}

	res.MessageTag = msg[0]
	res.RmcpStatusCode = msg[1]
	// 2 bytes reserved
	res.RemoteConsoleSessionID, _, _ = unpackUint32L(msg, 4)

	if res.RmcpStatusCode != uint8(RakpStatusNoErrors) {
		return fmt.Errorf("the return status of rakp2 has error: %v", res.RmcpStatusCode)
	}

	res.ManagedSystemRandomNumber = array16(msg[8:24])
	res.ManagedSystemGUID = array16(msg[24:40])

	var authCodeLen int = 0
	switch res.authAlg {
	case AuthAlgRAKP_None:
		break
	case AuthAlgRAKP_HMAC_MD5:
		authCodeLen = 16
	case AuthAlgRAKP_HMAC_SHA1:
		authCodeLen = 20
	case AuthAlgRAKP_HMAC_SHA256:
		authCodeLen = 32
	}
	if len(msg) < 40+authCodeLen {
		return fmt.Errorf("the unpacked data does not contain enough auth code")
	}
	res.KeyExchangeAuthenticationCode = make([]byte, authCodeLen)
	copy(res.KeyExchangeAuthenticationCode, msg[40:40+authCodeLen])

	return nil
}

func (*RAKPMessage2) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *RAKPMessage2) Format() string {
	return fmt.Sprintf("%v", res)
}

// ValidateRAKP2 validates RAKPMessage2 returned by BMC.
func (c *Client) ValidateRAKP2(rakp2 *RAKPMessage2) (bool, error) {
	if c.session.v20.consoleSessionID != rakp2.RemoteConsoleSessionID {
		return false, fmt.Errorf("session id not matched, cached console session id: %x, rapk2 returned session id: %x", c.session.v20.consoleSessionID, rakp2.RemoteConsoleSessionID)
	}

	// rakp2 authcode is valid
	authcode, err := c.generate_rakp2_authcode()
	if err != nil {
		return false, fmt.Errorf("generate rakp2 authcode failed, err: %s", err)
	}

	c.DebugBytes("rakp2 returned auth code", rakp2.KeyExchangeAuthenticationCode, 16)

	if !isByteSliceEqual(authcode, rakp2.KeyExchangeAuthenticationCode) {
		return false, fmt.Errorf("rakp2 authcode not equal, console: %x, bmc: %x", authcode, rakp2.KeyExchangeAuthenticationCode)
	}
	return true, nil
}

func (c *Client) RAKPMessage1() (response *RAKPMessage2, err error) {

	c.session.v20.consoleRand = array16(randomBytes(16))
	c.DebugBytes("console generate console random number", c.session.v20.consoleRand[:], 16)

	request := &RAKPMessage1{
		MessageTag:                     0,
		ManagedSystemSessionID:         c.session.v20.bmcSessionID, // set by previous RMCP+ Open Session Request
		RemoteConsoleRandomNumber:      c.session.v20.consoleRand,
		RequestedMaximumPrivilegeLevel: c.session.v20.maxPrivilegeLevel,
		NameOnlyLookup:                 true,
		UsernameLength:                 uint8(len(c.Username)),
		Username:                       []byte(c.Username),
	}

	c.session.v20.role = request.Role()

	response = &RAKPMessage2{
		authAlg: c.session.v20.authAlg,
	}
	c.session.v20.state = SessionStateRakp1Sent

	err = c.Exchange(request, response)
	if err != nil {
		return nil, err
	}

	// the following fields must be set before generate_sik/generate_k1/generate_k2
	c.session.v20.rakp2ReturnCode = response.RmcpStatusCode
	c.session.v20.bmcGUID = response.ManagedSystemGUID
	c.session.v20.bmcRand = response.ManagedSystemRandomNumber // will be used in rakp3 to generate authCode

	if _, err = c.ValidateRAKP2(response); err != nil {
		err = fmt.Errorf("validate rakp2 message failed, err: %s", err)
		return
	}

	c.session.v20.state = SessionStateRakp2Received

	return
}

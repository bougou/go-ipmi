package ipmi

import "fmt"

// 13.17 RMCP+ Open Session Request
type OpenSessionRequest struct {
	MessageTag                     uint8
	RequestedMaximumPrivilegeLevel PrivilegeLevel
	RemoteConsoleSessionID         uint32
	AuthenticationPayload
	IntegrityPayload
	ConfidentialityPayload
}

// 13.18 RMCP+ Open Session Response
type OpenSessionResponse struct {
	MessageTag             uint8
	RmcpStatusCode         uint8
	MaximumPrivilegeLevel  uint8
	RemoteConsoleSessionID uint32
	ManagedSystemSessionID uint32
	AuthenticationPayload
	IntegrityPayload
	ConfidentialityPayload
}

type AuthenticationPayload struct {
	// 00h = authentication algorithm
	PayloadType   uint8
	PayloadLength uint8 // Payload Length in bytes (1-based). The total length in bytes of the payload including the header (= 08h for this specification).
	AuthAlg       uint8
}

type IntegrityPayload struct {
	// 01h = integrity algorithm
	PayloadType   uint8
	PayloadLength uint8
	IntegrityAlg  uint8
}

type ConfidentialityPayload struct {
	// 02h = confidentiality algorithm
	PayloadType   uint8
	PayloadLength uint8
	CryptAlg      uint8
}

const (
	RmcpOpenSessionRequestSize     int = 32
	RmcpOpenSessionResponseSize    int = 36
	RmcpOpenSessionResponseMinSize int = 8
)

func (req *OpenSessionRequest) Command() Command {
	return CommandNone
}

func (req *OpenSessionRequest) Pack() []byte {
	var out = make([]byte, RmcpOpenSessionRequestSize)
	packUint8(req.MessageTag, out, 0)
	packUint8(uint8(req.RequestedMaximumPrivilegeLevel), out, 1)
	packUint16(0, out, 2) // 2 bytes reserved
	packUint32L(req.RemoteConsoleSessionID, out, 4)
	packBytes(req.AuthenticationPayload.Pack(), out, 8)
	packBytes(req.IntegrityPayload.Pack(), out, 16)
	packBytes(req.ConfidentialityPayload.Pack(), out, 24)
	return out
}

func (res *OpenSessionResponse) Unpack(data []byte) error {
	if len(data) < RmcpOpenSessionResponseMinSize {
		return ErrUnpackedDataTooShort
	}

	res.MessageTag, _, _ = unpackUint8(data, 0)
	res.RmcpStatusCode, _, _ = unpackUint8(data, 1)
	res.MaximumPrivilegeLevel, _, _ = unpackUint8(data, 2)
	// reserved
	res.RemoteConsoleSessionID, _, _ = unpackUint32L(data, 4)

	// If the previous message generated an error, then only the Status Code, Reserved, and Remote Console Session ID fields are returned.
	// See Table 13-, RMCP+ and RAKP Message Status Codes.
	// The session establishment in progress is discarded at the BMC, and the
	// remote console will need to start over with a new Open Session Request message.
	// (Since the BMC has not yet delivered a Managed System Session ID to the remote console,
	// it shouldn't be carrying any state information from the prior Open Session Request,
	// but if it has, that state should be discarded.)
	if res.RmcpStatusCode != uint8(RakpStatusNoErrors) {
		return nil
	}

	if len(data) < RmcpOpenSessionResponseSize {
		return ErrUnpackedDataTooShort
	}
	res.ManagedSystemSessionID, _, _ = unpackUint32L(data, 8)
	res.AuthenticationPayload.Unpack(data[12:20])
	res.IntegrityPayload.Unpack(data[20:28])
	res.ConfidentialityPayload.Unpack(data[28:36])
	return nil
}

func (*OpenSessionResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *OpenSessionResponse) Format() string {
	return fmt.Sprintf(`  Message tag                        : %#02x
  RMCP+ status                       : %#02x %s
  Maximum privilege level            : %#02x %s
  Console Session ID                 : %#0x
  BMC Session ID                     : %#0x
  Negotiated authenticatin algorithm : %#02x %s
  Negotiated integrity algorithm     : %#02x %s
  Negotiated encryption algorithm    : %#02x %s`,
		res.MessageTag,
		res.RmcpStatusCode, RakpStatus(res.RmcpStatusCode),
		res.MaximumPrivilegeLevel, PrivilegeLevel(res.MaximumPrivilegeLevel),
		res.RemoteConsoleSessionID,
		res.ManagedSystemSessionID,
		res.AuthAlg, AuthAlg(res.AuthAlg),
		res.IntegrityAlg, IntegrityAlg(res.IntegrityAlg),
		res.CryptAlg, CryptAlg(res.CryptAlg),
	)
}

func (c *Client) OpenSession() (response *OpenSessionResponse, err error) {
	bestSuiteID := findBestCipherSuite()
	authAlg, integrityAlg, cryptAlg, err := getCipherSuiteAlgorithms(bestSuiteID)
	if err != nil {
		return nil, fmt.Errorf("get cipher suite for id %0x failed, err: %s", bestSuiteID, err)
	}
	c.session.v20.requestedAuthAlg = authAlg
	c.session.v20.requestedIntegrityAlg = integrityAlg
	c.session.v20.requestedEncryptAlg = cryptAlg

	// Choose our session ID for easy recognition in the packet dump
	var remoteConsoleSessionID uint32 = 0xa0a1a2a3

	request := &OpenSessionRequest{
		MessageTag:                     0x00,
		RequestedMaximumPrivilegeLevel: 0, // Request the highest level matching proposed algorithms
		RemoteConsoleSessionID:         remoteConsoleSessionID,
		AuthenticationPayload: AuthenticationPayload{
			PayloadType:   0x00, // 0 means authentication algorithm
			PayloadLength: 8,
			AuthAlg:       uint8(c.session.v20.requestedAuthAlg),
		},
		IntegrityPayload: IntegrityPayload{
			PayloadType:   0x01, // 1 means integrity algorithm
			PayloadLength: 8,
			IntegrityAlg:  uint8(c.session.v20.requestedIntegrityAlg),
		},
		ConfidentialityPayload: ConfidentialityPayload{
			PayloadType:   0x02, // 2 means confidentiality algorithm
			PayloadLength: 8,
			CryptAlg:      uint8(c.session.v20.requestedEncryptAlg),
		},
	}

	response = &OpenSessionResponse{}

	c.session.v20.state = SessionStateOpenSessionSent

	err = c.Exchange(request, response)
	if err != nil {
		return nil, fmt.Errorf("client exchange failed, err: %s", err)
	}

	c.Debug("OPEN SESSION RESPONSE", response.Format())

	if response.RmcpStatusCode != uint8(RakpStatusNoErrors) {
		err = fmt.Errorf("rakp status code error: %x", response.RmcpStatusCode)
		return
	}

	c.session.v20.state = SessionStateOpenSessionReceived

	c.session.v20.authAlg = AuthAlg(response.AuthAlg)
	c.session.v20.integrityAlg = IntegrityAlg(response.IntegrityAlg)
	c.session.v20.cryptAlg = CryptAlg(response.CryptAlg)
	c.session.v20.maxPrivilegeLevel = PrivilegeLevel(response.MaximumPrivilegeLevel)
	c.session.v20.consoleSessionID = response.RemoteConsoleSessionID
	c.session.v20.bmcSessionID = response.ManagedSystemSessionID

	return
}

func (p *AuthenticationPayload) Pack() []byte {
	out := make([]byte, 8)
	packUint8(p.PayloadType, out, 0)
	packUint16(0, out, 1) // 2 bytes reserved
	packUint8(p.PayloadLength, out, 3)
	packUint8(p.AuthAlg, out, 4)
	packUint24(0, out, 5) // 3 bytes reserved
	return out
}

func (p *AuthenticationPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShort
	}
	p.PayloadType, _, _ = unpackUint8(msg, 0)
	// 2 bytes reserved
	p.PayloadLength, _, _ = unpackUint8(msg, 3)
	p.AuthAlg, _, _ = unpackUint8(msg, 4)
	// 3 bytes serverd
	return nil
}

func (p *IntegrityPayload) Pack() []byte {
	out := make([]byte, 8)
	packUint8(p.PayloadType, out, 0)
	packUint16(0, out, 1) // 2 bytes reserved
	packUint8(p.PayloadLength, out, 3)
	packUint8(p.IntegrityAlg, out, 4)
	packUint24(0, out, 5) // 3 bytes reserved
	return out
}

func (p *IntegrityPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShort
	}
	p.PayloadType, _, _ = unpackUint8(msg, 0)
	// 2 bytes reserved
	p.PayloadLength, _, _ = unpackUint8(msg, 3)
	p.IntegrityAlg, _, _ = unpackUint8(msg, 4)
	// 3 bytes serverd
	return nil
}

func (p *ConfidentialityPayload) Pack() []byte {
	out := make([]byte, 8)
	packUint8(p.PayloadType, out, 0)
	packUint16(0, out, 1) // 2 bytes reserved
	packUint8(p.PayloadLength, out, 3)
	packUint8(p.CryptAlg, out, 4)
	packUint24(0, out, 5) // 3 bytes reserved
	return out
}

func (p *ConfidentialityPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShort
	}
	p.PayloadType, _, _ = unpackUint8(msg, 0)
	// 2 bytes reserved
	p.PayloadLength, _, _ = unpackUint8(msg, 3)
	p.CryptAlg, _, _ = unpackUint8(msg, 4)
	// 3 bytes serverd
	return nil
}

// 13.24 RMCP+ and RAKP Message Status Codes
type RakpStatus uint8

const (
	RakpStatusNoErrors                        RakpStatus = 0x00
	RakpStatusInsufficientResourcesForSession RakpStatus = 0x01
	RakpStatusInvalidSessionID                RakpStatus = 0x02
	RakpStatusInvalidPayloadType              RakpStatus = 0x03
	RakpStatusInvalidAuthAlgorithm            RakpStatus = 0x04
	RakpStatusInvalidIntegrityAlgorithm       RakpStatus = 0x05
	RakpStatusNoMatchingAuthenticationPayload RakpStatus = 0x06
	RakpStatusNoMatchingIntegrityPayload      RakpStatus = 0x07
	RakpStatusInactiveSessionID               RakpStatus = 0x08
	RakpStatusInvalidRole                     RakpStatus = 0x09
	RakpStatusUnauthorizedRoleForRequested    RakpStatus = 0x0a
	RakpStatusInsufficientResourcesForRole    RakpStatus = 0x0b
	RakpStatusInvalidNameLength               RakpStatus = 0x0c
	RakpStatusUnauthorizedName                RakpStatus = 0x0d
	RakpStatusUnauthorizedGUID                RakpStatus = 0x0e
	RakpStatusInvalidIntegrityCheckValue      RakpStatus = 0x0f
	RakpStatusInvalidConfidentialityAlgorithm RakpStatus = 0x10
	RakpStatusNoCipherSuiteMatch              RakpStatus = 0x11
	RakpStatusIllegalParameter                RakpStatus = 0x12

	// 13h-FFh Reserved for future definition by this specification
)

func (r RakpStatus) String() string {
	m := map[RakpStatus]string{
		0x00: "No errors",
		0x01: "Insufficient resources to create a session",
		0x02: "Invalid Session ID",
		0x03: "Invalid payload type",
		0x04: "Invalid authentication algorithm",
		0x05: "Invalid integrity algorithm",
		0x06: "No matching authentication payload",
		0x07: "No matching integrity payload",
		0x08: "Inactive Session ID",
		0x09: "Invalid role",
		0x0a: "Unauthorized role or privilege level requested",
		0x0b: "Insufficient resources to create a session at the requested role",
		0x0c: "Invalid name length",
		0x0d: "Unauthorized name",
		0x0e: "Unauthorized GUID",
		0x0f: "Invalid integrity check value",
		0x10: "Invalid confidentiality algorithm",
		0x11: "No Cipher Suite match with proposed security algorithms",
		0x12: "Illegal or unrecognized parameter",
	}
	s, ok := m[r]
	if ok {
		return s
	}
	return ""
}

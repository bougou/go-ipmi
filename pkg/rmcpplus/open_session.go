package rmcpplus

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// Open Session / RAKP message sizes (v2.0§13.17–13.18).
const (
	OpenSessionRequestSize     = 32
	OpenSessionResponseSize    = 36
	OpenSessionResponseMinSize = 8
)

// OpenSessionRequest is an RMCP+ Open Session Request (v2.0§13.17).
type OpenSessionRequest struct {
	MessageTag                     uint8
	RequestedMaximumPrivilegeLevel types.PrivilegeLevel
	RemoteConsoleSessionID         uint32
	AuthenticationPayload
	IntegrityPayload
	ConfidentialityPayload
}

// OpenSessionResponse is an RMCP+ Open Session Response (v2.0§13.18).
type OpenSessionResponse struct {
	MessageTag             uint8
	RmcpStatusCode         types.RmcpStatusCode
	MaximumPrivilegeLevel  uint8
	RemoteConsoleSessionID uint32
	ManagedSystemSessionID uint32
	AuthenticationPayload
	IntegrityPayload
	ConfidentialityPayload
}

// AuthenticationPayload is the auth-algorithm record in Open Session (v2.0§13.17).
type AuthenticationPayload struct {
	PayloadType   uint8
	PayloadLength uint8
	AuthAlg       uint8
}

// IntegrityPayload is the integrity-algorithm record in Open Session.
type IntegrityPayload struct {
	PayloadType   uint8
	PayloadLength uint8
	IntegrityAlg  uint8
}

// ConfidentialityPayload is the confidentiality-algorithm record in Open Session.
type ConfidentialityPayload struct {
	PayloadType   uint8
	PayloadLength uint8
	CryptAlg      uint8
}

func (req *OpenSessionRequest) Command() types.Command {
	return types.CommandNone
}

func (req *OpenSessionRequest) Pack() []byte {
	out := make([]byte, OpenSessionRequestSize)
	types.PackUint8(req.MessageTag, out, 0)
	types.PackUint8(uint8(req.RequestedMaximumPrivilegeLevel), out, 1)
	types.PackUint16(0, out, 2)
	types.PackUint32L(req.RemoteConsoleSessionID, out, 4)
	types.PackBytes(req.AuthenticationPayload.Pack(), out, 8)
	types.PackBytes(req.IntegrityPayload.Pack(), out, 16)
	types.PackBytes(req.ConfidentialityPayload.Pack(), out, 24)
	return out
}

// Unpack parses an Open Session Request payload.
func (req *OpenSessionRequest) Unpack(data []byte) error {
	if len(data) < OpenSessionRequestSize {
		return types.ErrUnpackedDataTooShortWith(len(data), OpenSessionRequestSize)
	}
	req.MessageTag, _, _ = types.UnpackUint8(data, 0)
	priv, _, _ := types.UnpackUint8(data, 1)
	req.RequestedMaximumPrivilegeLevel = types.PrivilegeLevel(priv & 0x0F)
	req.RemoteConsoleSessionID, _, _ = types.UnpackUint32L(data, 4)
	if err := req.AuthenticationPayload.Unpack(data[8:16]); err != nil {
		return err
	}
	if err := req.IntegrityPayload.Unpack(data[16:24]); err != nil {
		return err
	}
	return req.ConfidentialityPayload.Unpack(data[24:32])
}

func (res *OpenSessionResponse) Unpack(data []byte) error {
	if len(data) < OpenSessionResponseMinSize {
		return types.ErrUnpackedDataTooShortWith(len(data), OpenSessionResponseMinSize)
	}

	res.MessageTag, _, _ = types.UnpackUint8(data, 0)
	b1, _, _ := types.UnpackUint8(data, 1)
	res.RmcpStatusCode = types.RmcpStatusCode(b1)
	res.MaximumPrivilegeLevel, _, _ = types.UnpackUint8(data, 2)
	res.RemoteConsoleSessionID, _, _ = types.UnpackUint32L(data, 4)

	// On error only Status Code, Reserved, and Remote Console Session ID are returned
	// (v2.0§13.18 / Table 13-15).
	if res.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		return nil
	}

	if len(data) < OpenSessionResponseSize {
		return types.ErrUnpackedDataTooShortWith(len(data), OpenSessionResponseSize)
	}
	res.ManagedSystemSessionID, _, _ = types.UnpackUint32L(data, 8)
	_ = res.AuthenticationPayload.Unpack(data[12:20])
	_ = res.IntegrityPayload.Unpack(data[20:28])
	_ = res.ConfidentialityPayload.Unpack(data[28:36])
	return nil
}

// Pack encodes a successful or error Open Session Response.
// On error (RmcpStatusCode != NoErrors) only the 8-byte short form is returned.
func (res *OpenSessionResponse) Pack() []byte {
	if res.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		out := make([]byte, OpenSessionResponseMinSize)
		types.PackUint8(res.MessageTag, out, 0)
		types.PackUint8(uint8(res.RmcpStatusCode), out, 1)
		types.PackUint8(res.MaximumPrivilegeLevel, out, 2)
		types.PackUint8(0, out, 3)
		types.PackUint32L(res.RemoteConsoleSessionID, out, 4)
		return out
	}
	out := make([]byte, OpenSessionResponseSize)
	types.PackUint8(res.MessageTag, out, 0)
	types.PackUint8(uint8(res.RmcpStatusCode), out, 1)
	types.PackUint8(res.MaximumPrivilegeLevel, out, 2)
	types.PackUint8(0, out, 3)
	types.PackUint32L(res.RemoteConsoleSessionID, out, 4)
	types.PackUint32L(res.ManagedSystemSessionID, out, 8)
	types.PackBytes(res.AuthenticationPayload.Pack(), out, 12)
	types.PackBytes(res.IntegrityPayload.Pack(), out, 20)
	types.PackBytes(res.ConfidentialityPayload.Pack(), out, 28)
	return out
}

func (*OpenSessionResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *OpenSessionResponse) Format() string {
	return "" +
		fmt.Sprintf("  Message tag                         : %#02x\n", res.MessageTag) +
		fmt.Sprintf("  RMCP+ status                        : %#02x %s\n", res.RmcpStatusCode, types.RmcpStatusCode(res.RmcpStatusCode)) +
		fmt.Sprintf("  Maximum privilege level             : %#02x %s\n", res.MaximumPrivilegeLevel, types.PrivilegeLevel(res.MaximumPrivilegeLevel)) +
		fmt.Sprintf("  Console Session ID                  : %#0x\n", res.RemoteConsoleSessionID) +
		fmt.Sprintf("  BMC Session ID                      : %#0x\n", res.ManagedSystemSessionID) +
		fmt.Sprintf("  Negotiated authentication algorithm : %#02x %s\n", res.AuthAlg, types.AuthAlg(res.AuthAlg)) +
		fmt.Sprintf("  Negotiated integrity algorithm      : %#02x %s\n", res.IntegrityAlg, types.IntegrityAlg(res.IntegrityAlg)) +
		fmt.Sprintf("  Negotiated encryption algorithm     : %#02x %s\n", res.CryptAlg, types.CryptAlg(res.CryptAlg))
}

func (p *AuthenticationPayload) Pack() []byte {
	out := make([]byte, 8)
	types.PackUint8(p.PayloadType, out, 0)
	types.PackUint16(0, out, 1)
	types.PackUint8(p.PayloadLength, out, 3)
	types.PackUint8(p.AuthAlg, out, 4)
	types.PackUint24(0, out, 5)
	return out
}

func (p *AuthenticationPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8)
	}
	p.PayloadType, _, _ = types.UnpackUint8(msg, 0)
	p.PayloadLength, _, _ = types.UnpackUint8(msg, 3)
	p.AuthAlg, _, _ = types.UnpackUint8(msg, 4)
	return nil
}

func (p *IntegrityPayload) Pack() []byte {
	out := make([]byte, 8)
	types.PackUint8(p.PayloadType, out, 0)
	types.PackUint16(0, out, 1)
	types.PackUint8(p.PayloadLength, out, 3)
	types.PackUint8(p.IntegrityAlg, out, 4)
	types.PackUint24(0, out, 5)
	return out
}

func (p *IntegrityPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8)
	}
	p.PayloadType, _, _ = types.UnpackUint8(msg, 0)
	p.PayloadLength, _, _ = types.UnpackUint8(msg, 3)
	p.IntegrityAlg, _, _ = types.UnpackUint8(msg, 4)
	return nil
}

func (p *ConfidentialityPayload) Pack() []byte {
	out := make([]byte, 8)
	types.PackUint8(p.PayloadType, out, 0)
	types.PackUint16(0, out, 1)
	types.PackUint8(p.PayloadLength, out, 3)
	types.PackUint8(p.CryptAlg, out, 4)
	types.PackUint24(0, out, 5)
	return out
}

func (p *ConfidentialityPayload) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8)
	}
	p.PayloadType, _, _ = types.UnpackUint8(msg, 0)
	p.PayloadLength, _, _ = types.UnpackUint8(msg, 3)
	p.CryptAlg, _, _ = types.UnpackUint8(msg, 4)
	return nil
}

// NewAlgorithmPayloads builds the three Open Session algorithm records for the
// given algorithms (payload type 00h/01h/02h, length 08h).
func NewAlgorithmPayloads(auth types.AuthAlg, integ types.IntegrityAlg, crypt types.CryptAlg) (AuthenticationPayload, IntegrityPayload, ConfidentialityPayload) {
	return AuthenticationPayload{PayloadType: 0x00, PayloadLength: 8, AuthAlg: uint8(auth)},
		IntegrityPayload{PayloadType: 0x01, PayloadLength: 8, IntegrityAlg: uint8(integ)},
		ConfidentialityPayload{PayloadType: 0x02, PayloadLength: 8, CryptAlg: uint8(crypt)}
}

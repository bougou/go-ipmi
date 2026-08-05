package rmcpplus

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	MaxUserNameLength = 16
	RAKP1MessageSize  = 44
)

// RAKPMessage1 is RAKP Message 1 (v2.0§13.20).
type RAKPMessage1 struct {
	MessageTag uint8

	ManagedSystemSessionID uint32

	RemoteConsoleRandomNumber [16]byte

	NameOnlyLookup                 bool
	RequestedMaximumPrivilegeLevel types.PrivilegeLevel

	UsernameLength uint8
	Username       []byte
}

// RAKPMessage2 is RAKP Message 2 (v2.0§13.21).
type RAKPMessage2 struct {
	// AuthAlg is the negotiated authentication algorithm; needed to know how
	// many KeyExchangeAuthenticationCode bytes to read.
	AuthAlg types.AuthAlg

	MessageTag uint8

	RmcpStatusCode types.RmcpStatusCode

	RemoteConsoleSessionID uint32

	ManagedSystemRandomNumber [16]byte

	ManagedSystemGUID [16]byte

	KeyExchangeAuthenticationCode []byte
}

func (req *RAKPMessage1) Command() types.Command {
	return types.CommandNone
}

func (r *RAKPMessage1) Pack() []byte {
	msg := make([]byte, 28+len(r.Username))
	types.PackUint8(r.MessageTag, msg, 0)
	types.PackUint24L(0, msg, 1)
	types.PackUint32L(r.ManagedSystemSessionID, msg, 4)
	types.PackBytes(r.RemoteConsoleRandomNumber[:], msg, 8)
	types.PackUint8(r.Role(), msg, 24)
	types.PackUint16L(0, msg, 25)
	types.PackUint8(r.UsernameLength, msg, 27)
	types.PackBytes(r.Username, msg, 28)
	return msg
}

// Unpack parses a RAKP Message 1 payload.
func (r *RAKPMessage1) Unpack(msg []byte) error {
	if len(msg) < 28 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 28)
	}
	r.MessageTag, _, _ = types.UnpackUint8(msg, 0)
	r.ManagedSystemSessionID, _, _ = types.UnpackUint32L(msg, 4)
	copy(r.RemoteConsoleRandomNumber[:], msg[8:24])
	role := msg[24]
	r.RequestedMaximumPrivilegeLevel = types.PrivilegeLevel(role & 0x0F)
	r.NameOnlyLookup = types.IsBit4Set(role)
	r.UsernameLength = msg[27]
	if int(28+r.UsernameLength) > len(msg) {
		return types.ErrUnpackedDataTooShortWith(len(msg), int(28+r.UsernameLength))
	}
	r.Username = make([]byte, r.UsernameLength)
	copy(r.Username, msg[28:28+r.UsernameLength])
	return nil
}

// Role returns the privilege byte combining RequestedMaximumPrivilegeLevel and
// NameOnlyLookup (stored for RAKP auth-code computation).
func (r *RAKPMessage1) Role() uint8 {
	privilegeLevel := uint8(r.RequestedMaximumPrivilegeLevel)
	if r.NameOnlyLookup {
		privilegeLevel = types.SetBit4(privilegeLevel)
	}
	return privilegeLevel
}

func (res *RAKPMessage2) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	res.MessageTag = msg[0]
	res.RmcpStatusCode = types.RmcpStatusCode(msg[1])
	res.RemoteConsoleSessionID, _, _ = types.UnpackUint32L(msg, 4)

	if res.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		return fmt.Errorf("the return status of rakp2 has error: %v", res.RmcpStatusCode)
	}

	if len(msg) < 40 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 40)
	}

	res.ManagedSystemRandomNumber = types.Array16(msg[8:24])
	res.ManagedSystemGUID = types.Array16(msg[24:40])

	authCodeLen := crypto.RAKP2AuthCodeLen(res.AuthAlg)
	if len(msg) < 40+authCodeLen {
		return fmt.Errorf("the unpacked data does not contain enough auth code")
	}
	res.KeyExchangeAuthenticationCode = make([]byte, authCodeLen)
	copy(res.KeyExchangeAuthenticationCode, msg[40:40+authCodeLen])
	return nil
}

// Pack encodes a RAKP Message 2 response (success or short error form).
func (res *RAKPMessage2) Pack() []byte {
	if res.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		out := make([]byte, 8)
		out[0] = res.MessageTag
		out[1] = uint8(res.RmcpStatusCode)
		types.PackUint32L(res.RemoteConsoleSessionID, out, 4)
		return out
	}
	out := make([]byte, 40+len(res.KeyExchangeAuthenticationCode))
	out[0] = res.MessageTag
	out[1] = uint8(res.RmcpStatusCode)
	types.PackUint32L(res.RemoteConsoleSessionID, out, 4)
	copy(out[8:24], res.ManagedSystemRandomNumber[:])
	copy(out[24:40], res.ManagedSystemGUID[:])
	copy(out[40:], res.KeyExchangeAuthenticationCode)
	return out
}

func (res *RAKPMessage2) Format() string {
	return fmt.Sprintf("%v", res)
}

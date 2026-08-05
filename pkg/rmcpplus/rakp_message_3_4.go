package rmcpplus

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

// RAKPMessage3 is RAKP Message 3 (v2.0§13.22).
type RAKPMessage3 struct {
	MessageTag uint8

	RmcpStatusCode types.RmcpStatusCode

	ManagedSystemSessionID uint32

	KeyExchangeAuthenticationCode []byte
}

// RAKPMessage4 is RAKP Message 4 (v2.0§13.23).
type RAKPMessage4 struct {
	AuthAlg types.AuthAlg

	MessageTag uint8

	RmcpStatusCode types.RmcpStatusCode

	MgmtConsoleSessionID uint32

	IntegrityCheckValue []byte
}

func (req *RAKPMessage3) Command() types.Command {
	return types.CommandNone
}

func (req *RAKPMessage3) Pack() []byte {
	msg := make([]byte, 8+len(req.KeyExchangeAuthenticationCode))
	types.PackUint8(req.MessageTag, msg, 0)
	types.PackUint8(uint8(req.RmcpStatusCode), msg, 1)
	types.PackUint16(0, msg, 2)
	types.PackUint32L(req.ManagedSystemSessionID, msg, 4)
	types.PackBytes(req.KeyExchangeAuthenticationCode, msg, 8)
	return msg
}

// Unpack parses a RAKP Message 3 payload. authAlg selects the expected auth-code length.
func (req *RAKPMessage3) Unpack(msg []byte, authAlg types.AuthAlg) error {
	authCodeLen := crypto.RAKP3AuthCodeLen(authAlg)
	if len(msg) < 8+authCodeLen {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8+authCodeLen)
	}
	req.MessageTag, _, _ = types.UnpackUint8(msg, 0)
	b1, _, _ := types.UnpackUint8(msg, 1)
	req.RmcpStatusCode = types.RmcpStatusCode(b1)
	req.ManagedSystemSessionID, _, _ = types.UnpackUint32L(msg, 4)
	req.KeyExchangeAuthenticationCode, _, _ = types.UnpackBytes(msg, 8, authCodeLen)
	return nil
}

func (res *RAKPMessage4) Unpack(msg []byte) error {
	authCodeLen := crypto.RAKP4ICVLen(res.AuthAlg)
	if len(msg) < 8+authCodeLen {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8+authCodeLen)
	}

	res.MessageTag, _, _ = types.UnpackUint8(msg, 0)
	b1, _, _ := types.UnpackUint8(msg, 1)
	res.RmcpStatusCode = types.RmcpStatusCode(b1)
	res.MgmtConsoleSessionID, _, _ = types.UnpackUint32L(msg, 4)
	res.IntegrityCheckValue, _, _ = types.UnpackBytes(msg, 8, authCodeLen)
	return nil
}

// Pack encodes a RAKP Message 4 response.
func (res *RAKPMessage4) Pack() []byte {
	out := make([]byte, 8+len(res.IntegrityCheckValue))
	types.PackUint8(res.MessageTag, out, 0)
	types.PackUint8(uint8(res.RmcpStatusCode), out, 1)
	types.PackUint16(0, out, 2)
	types.PackUint32L(res.MgmtConsoleSessionID, out, 4)
	types.PackBytes(res.IntegrityCheckValue, out, 8)
	return out
}

func (res *RAKPMessage4) Format() string {
	return fmt.Sprintf("%v", res)
}

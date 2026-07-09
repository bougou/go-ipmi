package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 26.1 SOL Activating Command
)

type SOLActivatingRequest struct {
	SessionState       uint8
	PayloadInstance    uint8
	FormatVersionMajor uint8
	FormatVersionMinor uint8
}

type SOLActivatingResponse struct {
}

func (req *SOLActivatingRequest) Command() types.Command {
	return types.CommandSOLActivating
}

func (req *SOLActivatingRequest) Pack() []byte {
	out := make([]byte, 4)
	types.PackUint8(req.SessionState, out, 0)
	types.PackUint8(req.PayloadInstance, out, 1)
	types.PackUint8(req.FormatVersionMajor, out, 2)
	types.PackUint8(req.FormatVersionMinor, out, 3)
	return out
}

func (res *SOLActivatingResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SOLActivatingResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SOLActivatingResponse) Format() string {
	return ""
}

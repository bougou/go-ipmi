package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 21.4 Get Command Sub-function Support Command
)

type GetCommandSubfunctionSupportRequest struct {
	ChannelNumber uint8

	NetFn types.NetFn
	LUN   uint8
	Cmd   uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type GetCommandSubfunctionSupportResponse struct {
	SpecificationType uint8
	ErrataVersion     uint8
	OEMGroupBody      uint8

	SpecificationVersion  uint8
	SpecificationRevision uint8

	// Todo
	SupportMask []byte
}

func (req *GetCommandSubfunctionSupportRequest) Command() types.Command {
	return types.CommandGetCommandSubfunctionSupport
}

func (req *GetCommandSubfunctionSupportRequest) Pack() []byte {
	out := make([]byte, 7)
	types.PackUint8(req.ChannelNumber, out, 0)

	types.PackUint8(uint8(req.NetFn)&0x3f, out, 1)
	types.PackUint8(req.LUN&0x03, out, 2)
	types.PackUint8(req.Cmd, out, 3)

	if uint8(req.NetFn) == 0x2c {
		types.PackUint8(req.CodeForNetFn2C, out, 4)
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		types.PackUint24L(req.OEMIANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetCommandSubfunctionSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}
	b, _, _ := types.UnpackUint8(msg, 0)
	res.SpecificationType = b >> 4
	res.ErrataVersion = b & 0x0f
	res.OEMGroupBody = b

	res.SpecificationVersion, _, _ = types.UnpackUint8(msg, 1)
	res.SpecificationRevision, _, _ = types.UnpackUint8(msg, 2)

	res.SupportMask, _, _ = types.UnpackBytes(msg, 3, 4)
	return nil
}

func (*GetCommandSubfunctionSupportResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetCommandSubfunctionSupportResponse) Format() string {
	// Todo
	return ""
}

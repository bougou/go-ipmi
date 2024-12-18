package ipmi

import "context"

// 21.4 Get Command Sub-function Support Command
type GetCommandSubfunctionSupportRequest struct {
	ChannelNumber uint8

	NetFn NetFn
	LUN   uint8
	Cmd   uint8

	CodeForNetFn2C uint8
	OEM_IANA       uint32 // 3 bytes only
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

func (req *GetCommandSubfunctionSupportRequest) Command() Command {
	return CommandGetCommandSubfunctionSupport
}

func (req *GetCommandSubfunctionSupportRequest) Pack() []byte {
	out := make([]byte, 7)
	packUint8(req.ChannelNumber, out, 0)

	packUint8(uint8(req.NetFn)&0x3f, out, 1)
	packUint8(req.LUN&0x03, out, 2)
	packUint8(req.Cmd, out, 3)

	if uint8(req.NetFn) == 0x2c {
		packUint8(req.CodeForNetFn2C, out, 4)
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		packUint24L(req.OEM_IANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetCommandSubfunctionSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}
	b, _, _ := unpackUint8(msg, 0)
	res.SpecificationType = b >> 4
	res.ErrataVersion = b & 0x0f
	res.OEMGroupBody = b

	res.SpecificationVersion, _, _ = unpackUint8(msg, 1)
	res.SpecificationRevision, _, _ = unpackUint8(msg, 2)

	res.SupportMask, _, _ = unpackBytes(msg, 3, 4)
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

func (c *Client) GetCommandSubfunctionSupport(ctx context.Context, channelNumber uint8, netFn NetFn, lun uint8, code uint8, oemIANA uint32) (response *GetCommandSubfunctionSupportResponse, err error) {
	request := &GetCommandSubfunctionSupportRequest{
		ChannelNumber:  channelNumber,
		NetFn:          netFn,
		LUN:            lun,
		CodeForNetFn2C: code,
		OEM_IANA:       oemIANA,
	}
	response = &GetCommandSubfunctionSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

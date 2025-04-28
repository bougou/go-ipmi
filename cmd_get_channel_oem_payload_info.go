package ipmi

import (
	"context"
	"fmt"
)

// 24.10 Get Channel OEM Payload Info Command
type GetChannelOEMPayloadInfoRequest struct {
	ChannelNumber uint8

	PayloadType PayloadType

	// OEM IANA. When Payload Type Number is 02h (OEM Explicit) this field
	// holds the OEM IANA for the OEM payload type to look up information for. Otherwise, this field is set to 00_00_00h.
	OEMIANA      uint32
	OEMPayloadID uint16
}

type GetChannelOEMPayloadInfoResponse struct {
	PayloadType  PayloadType
	OEMIANA      uint32
	OEMPayloadID uint16

	MajorVersion uint8
	MinorVersion uint8
}

func (req *GetChannelOEMPayloadInfoRequest) Pack() []byte {
	out := make([]byte, 7)

	out[0] = req.ChannelNumber
	out[1] = byte(req.PayloadType)

	packUint24L(req.OEMIANA, out, 2)
	packUint16L(req.OEMPayloadID, out, 5)

	return out
}

func (req *GetChannelOEMPayloadInfoRequest) Command() Command {
	return CommandGetChannelOEMPayloadInfo
}

func (res *GetChannelOEMPayloadInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "OEM Payload IANA and/or Payload ID not supported",
	}
}

func (res *GetChannelOEMPayloadInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShortWith(len(msg), 7)
	}

	res.PayloadType = PayloadType(msg[0])
	res.OEMIANA, _, _ = unpackUint24L(msg, 1)
	res.OEMPayloadID, _, _ = unpackUint16L(msg, 4)
	res.MajorVersion = msg[6] >> 4
	res.MinorVersion = msg[6] & 0x0f

	return nil
}

func (res *GetChannelOEMPayloadInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Payload Type   : %s (%d)\n", res.PayloadType.String(), res.PayloadType) +
		fmt.Sprintf("OEM IANA       : %d\n", res.OEMIANA) +
		fmt.Sprintf("OEM Payload ID : %d\n", res.OEMPayloadID) +
		fmt.Sprintf("Major Version  : %d\n", res.MajorVersion) +
		fmt.Sprintf("Minor Version  : %d\n", res.MinorVersion)
}

func (c *Client) GetChannelOEMPayloadInfo(ctx context.Context, request *GetChannelOEMPayloadInfoRequest) (response *GetChannelOEMPayloadInfoResponse, err error) {
	response = &GetChannelOEMPayloadInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

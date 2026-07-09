package transport

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 24.9 Get Channel Payload Version Command
type GetChannelPayloadVersionRequest struct {
	ChannelNumber uint8

	PayloadType types.PayloadType
}

type GetChannelPayloadVersionResponse struct {
	MajorVersion uint8
	MinorVersion uint8
}

func (req *GetChannelPayloadVersionRequest) Pack() []byte {
	return []byte{req.ChannelNumber, uint8(req.PayloadType)}
}

func (req *GetChannelPayloadVersionRequest) Command() types.Command {
	return types.CommandGetChannelPayloadVersion
}

func (res *GetChannelPayloadVersionResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Payload type not available on given channel",
	}
}

func (res *GetChannelPayloadVersionResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.MajorVersion = msg[0] >> 4
	res.MinorVersion = msg[0] & 0x0f

	return nil
}

func (res *GetChannelPayloadVersionResponse) Format() string {
	return "" +
		fmt.Sprintf("Major Version: %d\n", res.MajorVersion) +
		fmt.Sprintf("Minor Version: %d\n", res.MinorVersion)
}

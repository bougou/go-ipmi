package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.8 Read Event Message Buffer Command
)

type ReadEventMessageBufferRequest struct {
	// empty
}

type ReadEventMessageBufferResponse struct {
	// 16 bytes of data in SEL Record format
	MessageData [16]byte
}

func (req ReadEventMessageBufferRequest) Command() types.Command {
	return types.CommandReadEventMessageBuffer
}

func (req *ReadEventMessageBufferRequest) Pack() []byte {
	return []byte{}
}

func (res *ReadEventMessageBufferResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	b, _, _ := types.UnpackBytes(msg, 0, 16)
	res.MessageData = types.Array16(b)
	return nil
}

func (res *ReadEventMessageBufferResponse) Format() string {
	return ""
}

package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 22.8 Read Event Message Buffer Command
)

type ReadEventMessageBufferRequest struct {
	// empty
}

type ReadEventMessageBufferResponse struct {
	// 16 bytes of data in SEL Record format
	MessageData [16]byte
}

func (req ReadEventMessageBufferRequest) Command() ipmi.Command {
	return ipmi.CommandReadEventMessageBuffer
}

func (req *ReadEventMessageBufferRequest) Pack() []byte {
	return []byte{}
}

func (res *ReadEventMessageBufferResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	b, _, _ := ipmi.UnpackBytes(msg, 0, 16)
	res.MessageData = ipmi.Array16(b)
	return nil
}

func (*ReadEventMessageBufferResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: " data not available (queue / buffer empty)",
	}
}

func (res *ReadEventMessageBufferResponse) Format() string {
	return ""
}

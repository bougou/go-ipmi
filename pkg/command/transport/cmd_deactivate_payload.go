package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 24.2 Deactivate Payload Command
)

type DeactivatePayloadRequest struct {
	PayloadType     types.PayloadType
	PayloadInstance uint8
}

type DeactivatePayloadResponse struct {
}

func (req DeactivatePayloadRequest) Command() types.Command {
	return types.CommandDeactivatePayload
}

func (req *DeactivatePayloadRequest) Pack() []byte {
	out := make([]byte, 6)

	out[0] = byte(req.PayloadType)
	out[1] = req.PayloadInstance

	out[2] = 0
	out[3] = 0
	out[4] = 0
	out[5] = 0

	return out
}

func (res *DeactivatePayloadResponse) Unpack(msg []byte) error {
	return nil
}

func (res *DeactivatePayloadResponse) Format() string {
	return ""
}

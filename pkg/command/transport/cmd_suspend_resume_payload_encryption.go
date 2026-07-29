package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 24.3 Suspend/Resume Payload Encryption Command
type SuspendResumePayloadEncryptionRequest struct {
	PayloadType     types.PayloadType
	PayloadInstance uint8
	Operation       PayloadEncryptionOperation
}

type PayloadEncryptionOperation uint8

const (
	PayloadEncryptionOperationSuspend      PayloadEncryptionOperation = 0
	PayloadEncryptionOperationResume       PayloadEncryptionOperation = 1
	PayloadEncryptionOperationReinitialize PayloadEncryptionOperation = 2
)

type SuspendResumePayloadEncryptionResponse struct {
}

func (req *SuspendResumePayloadEncryptionRequest) Pack() []byte {
	out := make([]byte, 3)
	out[0] = byte(req.PayloadType)
	out[1] = req.PayloadInstance
	out[2] = byte(req.Operation)
	return out
}

func (req *SuspendResumePayloadEncryptionRequest) Command() types.Command {
	return types.CommandSuspendResumePayloadEncryption
}

func (res *SuspendResumePayloadEncryptionResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SuspendResumePayloadEncryptionResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SuspendResumePayloadEncryptionResponse) Format() string {
	return ""
}

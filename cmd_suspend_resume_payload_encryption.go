package ipmi

import (
	"context"
)

// 24.3 Suspend/Resume Payload Encryption Command
type SuspendResumePayloadEncryptionRequest struct {
	PayloadType     PayloadType
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

func (req *SuspendResumePayloadEncryptionRequest) Command() Command {
	return CommandSuspendResumePayloadEncryption
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

func (c *Client) SuspendResumePayloadEncryption(ctx context.Context, payloadType PayloadType, payloadInstance uint8, operation PayloadEncryptionOperation) (response *SuspendResumePayloadEncryptionResponse, err error) {
	request := &SuspendResumePayloadEncryptionRequest{
		PayloadType:     payloadType,
		PayloadInstance: payloadInstance,
		Operation:       operation,
	}
	response = &SuspendResumePayloadEncryptionResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

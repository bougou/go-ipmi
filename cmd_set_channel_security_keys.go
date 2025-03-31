package ipmi

import (
	"context"
	"fmt"
)

// 22.25 Set Channel Security Keys Command
type SetChannelSecurityKeysRequest struct {
	ChannelNumber uint8

	Operation ChannelSecurityKeysOperation

	KeyID uint8

	KeyValue []byte
}

type ChannelSecurityKeysOperation uint8

const (
	ChannelSecurityKeysOperationRead ChannelSecurityKeysOperation = 0
	ChannelSecurityKeysOperationSet  ChannelSecurityKeysOperation = 1
	ChannelSecurityKeysOperationLock ChannelSecurityKeysOperation = 2
)

func (operation ChannelSecurityKeysOperation) String() string {
	m := map[ChannelSecurityKeysOperation]string{
		ChannelSecurityKeysOperationRead: "read",
		ChannelSecurityKeysOperationSet:  "set",
		ChannelSecurityKeysOperationLock: "lock",
	}
	s, ok := m[operation]
	if ok {
		return s
	}
	return "Unknown"
}

type ChannelSecurityKeysLockStatus uint8

const (
	ChannelSecurityKeysLockStatus_NotLockable ChannelSecurityKeysLockStatus = 0
	ChannelSecurityKeysLockStatus_Locked      ChannelSecurityKeysLockStatus = 1
	ChannelSecurityKeysLockStatus_Unlocked    ChannelSecurityKeysLockStatus = 2
)

func (lockStatus ChannelSecurityKeysLockStatus) String() string {
	m := map[ChannelSecurityKeysLockStatus]string{
		ChannelSecurityKeysLockStatus_NotLockable: "not lockable",
		ChannelSecurityKeysLockStatus_Locked:      "locked",
		ChannelSecurityKeysLockStatus_Unlocked:    "unlocked",
	}
	s, ok := m[lockStatus]
	if ok {
		return s
	}
	return "Unknown"
}

type SetChannelSecurityKeysResponse struct {
	LockStatus ChannelSecurityKeysLockStatus
	KeyValue   []byte
}

func (req *SetChannelSecurityKeysRequest) Pack() []byte {
	out := make([]byte, 3+len(req.KeyValue))

	out[0] = req.ChannelNumber
	out[1] = byte(req.Operation)
	out[2] = req.KeyID

	if len(req.KeyValue) > 0 {
		packBytes(req.KeyValue, out, 3)
	}

	return out
}

func (req *SetChannelSecurityKeysRequest) Command() Command {
	return CommandSetChannelSecurityKeys
}

func (res *SetChannelSecurityKeysResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Cannot perform set / confirm. Key is locked",
		0x81: "insufficient key bytes",
		0x82: "too many key bytes",
		0x83: "key value does not meet criteria for specified type of key",
		0x84: "KR is not used.",
	}
}

func (res *SetChannelSecurityKeysResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.LockStatus = ChannelSecurityKeysLockStatus(msg[0])

	if len(msg) > 1 {
		res.KeyValue, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *SetChannelSecurityKeysResponse) Format() string {
	return fmt.Sprintf(`
		Lock Status : %s (%d)
		Key Value  : %v
`,
		res.LockStatus.String(), res.LockStatus,
		res.KeyValue,
	)
}

func (c *Client) SetChannelSecurityKeys(ctx context.Context, request *SetChannelSecurityKeysRequest) (response *SetChannelSecurityKeysResponse, err error) {
	response = &SetChannelSecurityKeysResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

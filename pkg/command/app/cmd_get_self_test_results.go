package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 20.4 Get Self Test Results Command
)

type GetSelfTestResultsRequest struct {
	// empty
}

type GetSelfTestResultsResponse struct {
	Byte1 uint8
	Byte2 uint8
}

func (req *GetSelfTestResultsRequest) Command() types.Command {
	return types.CommandGetSelfTestResults
}

func (req *GetSelfTestResultsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetSelfTestResultsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSelfTestResultsResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.Byte1, _, _ = types.UnpackUint8(msg, 0)
	res.Byte2, _, _ = types.UnpackUint8(msg, 1)
	return nil
}

func (res *GetSelfTestResultsResponse) Format() string {
	// Todo
	return ""
}

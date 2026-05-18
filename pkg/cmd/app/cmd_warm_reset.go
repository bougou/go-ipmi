package app

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 20.3 Warm Reset Command
)

type WarmResetRequest struct {
	// empty
}

type WarmResetResponse struct {
}

func (req *WarmResetRequest) Command() ipmi.Command {
	return ipmi.CommandWarmReset
}

func (req *WarmResetRequest) Pack() []byte {
	return []byte{}
}

func (res *WarmResetResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *WarmResetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *WarmResetResponse) Format() string {
	return ""
}

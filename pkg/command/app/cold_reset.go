package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 20.2 Cold Reset Command
)

type ColdResetRequest struct {
	// empty
}

type ColdResetResponse struct {
}

func (req *ColdResetRequest) Command() types.Command {
	return types.CommandColdReset
}

func (req *ColdResetRequest) Pack() []byte {
	return []byte{}
}

func (res *ColdResetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ColdResetResponse) Format() string {
	return ""
}

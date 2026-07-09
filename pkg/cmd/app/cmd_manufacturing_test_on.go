package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 20.4 20.5 Manufacturing Test On Command
)

type ManufacturingTestOnRequest struct {
	// empty
}

type ManufacturingTestOnResponse struct {
	// empty
}

func (req *ManufacturingTestOnRequest) Command() types.Command {
	return types.CommandManufacturingTestOn
}

func (req *ManufacturingTestOnRequest) Pack() []byte {
	return []byte{}
}

func (res *ManufacturingTestOnResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ManufacturingTestOnResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ManufacturingTestOnResponse) Format() string {
	// Todo
	return ""
}

// If the device supports a "manufacturing test mode", this command is reserved to turn that mode on.

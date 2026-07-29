package oem

import (
	"fmt"
	"strings"

	"github.com/bougou/go-ipmi/pkg/types"
)

type CommandGetSupermicroBiosVersionRequest struct {
}

type CommandGetSupermicroBiosVersionResponse struct {
	Version string
}

func (req *CommandGetSupermicroBiosVersionRequest) Command() types.Command {
	return types.CommandGetSupermicroBiosVersion
}

func (req *CommandGetSupermicroBiosVersionRequest) Pack() []byte {
	return []byte{0x00, 0x00}
}

func (res *CommandGetSupermicroBiosVersionResponse) Unpack(msg []byte) error {
	res.Version = string(msg)
	res.Version = strings.TrimSpace(res.Version)
	return nil
}

func (res *CommandGetSupermicroBiosVersionResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *CommandGetSupermicroBiosVersionResponse) Format() string {
	return "" +
		fmt.Sprintf("bios.version = %s\n", res.Version)
}

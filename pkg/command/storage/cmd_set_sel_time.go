package storage

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 31.11 Set SEL Time Command
type SetSELTimeRequest struct {
	Time time.Time
}

type SetSELTimeResponse struct {
}

func (req *SetSELTimeRequest) Pack() []byte {
	var out = make([]byte, 4)
	types.PackUint32L(uint32(req.Time.Unix()), out, 0)
	return out
}

func (req *SetSELTimeRequest) Command() types.Command {
	return types.CommandSetSELTime
}

func (res *SetSELTimeResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSELTimeResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *SetSELTimeResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

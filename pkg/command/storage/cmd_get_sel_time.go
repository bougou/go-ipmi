package storage

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 31.10 Get SEL Time Command
type GetSELTimeRequest struct {
	// empty
}

type GetSELTimeResponse struct {
	// Present Timestamp clock reading
	Time time.Time
}

func (req *GetSELTimeRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSELTimeRequest) Command() types.Command {
	return types.CommandGetSELTime
}

func (res *GetSELTimeResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	t, _, _ := types.UnpackUint32L(msg, 0)
	res.Time = types.ParseTimestamp(t)
	return nil
}

func (res *GetSELTimeResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSELTimeResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

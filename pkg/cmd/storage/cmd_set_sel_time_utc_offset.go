package storage

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 31.11a Set SEL Time UTC Offset
)

type SetSELTimeUTCOffsetRequest struct {
	// signed integer for the offset in minutes from UTC to SEL Time. (ranges from -1440 to 1440)
	MinutesOffset int16
}

type SetSELTimeUTCOffsetResponse struct {
	// empty
}

func (req *SetSELTimeUTCOffsetRequest) Pack() []byte {
	out := make([]byte, 2)

	a := types.TwoSComplementEncode(int32(req.MinutesOffset), 16)
	types.PackUint16L(uint16(a), out, 0)

	return out
}

func (req *SetSELTimeUTCOffsetRequest) Command() types.Command {
	return types.CommandSetSELTimeUTCOffset
}

func (res *SetSELTimeUTCOffsetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSELTimeUTCOffsetResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *SetSELTimeUTCOffsetResponse) Format() string {
	return ""
}

// SetSELTimeUTCOffset initializes and retrieve a UTC offset (timezone) that is associated with the SEL Time

package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 31.11a Get SEL Time UTC Offset
type GetSELTimeUTCOffsetRequest struct {
	// empty
}

type GetSELTimeUTCOffsetResponse struct {
	// signed integer for the offset in minutes from UTC to SEL Time.
	MinutesOffset int16
}

func (req *GetSELTimeUTCOffsetRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSELTimeUTCOffsetRequest) Command() types.Command {
	return types.CommandGetSELTimeUTCOffset
}

func (res *GetSELTimeUTCOffsetResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	b, _, _ := types.UnpackUint16L(msg, 0)
	c := types.TwoSComplement(uint32(b), 16)
	res.MinutesOffset = int16(c)
	return nil
}

func (res *GetSELTimeUTCOffsetResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSELTimeUTCOffsetResponse) Format() string {
	return fmt.Sprintf("Offset : %d", res.MinutesOffset)
}

// GetSELTimeUTCOffset is used to retrieve the SEL Time UTC Offset (timezone)

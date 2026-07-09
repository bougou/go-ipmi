package chassis

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.14 Get POH Counter Command
type GetPOHCounterRequest struct {
	// empty
}

type GetPOHCounterResponse struct {
	MinutesPerCount uint8
	CounterReading  uint32
}

func (res *GetPOHCounterResponse) Minutes() uint32 {
	return res.CounterReading * uint32(res.MinutesPerCount)
}

func (req *GetPOHCounterRequest) Command() types.Command {
	return types.CommandGetPOHCounter
}

func (req *GetPOHCounterRequest) Pack() []byte {
	return []byte{}
}

func (res *GetPOHCounterResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	res.MinutesPerCount, _, _ = types.UnpackUint8(msg, 0)
	res.CounterReading, _, _ = types.UnpackUint32L(msg, 1)
	return nil
}

func (r *GetPOHCounterResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPOHCounterResponse) Format() string {
	totalMinutes := res.Minutes()

	days := totalMinutes / 1440
	minutes := totalMinutes - days*1440
	hours := minutes / 60

	return "" +
		fmt.Sprintf("POH Counter       : %d days, %d hours\n", days, hours) +
		fmt.Sprintf("Minutes per count : %d\n", res.MinutesPerCount) +
		fmt.Sprintf("Counter reading   : %d\n", res.CounterReading)
}

// This command returns the present reading of the POH (Power-On Hours) counter, plus the number of counts per hour.

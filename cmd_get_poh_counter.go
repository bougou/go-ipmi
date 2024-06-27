package ipmi

import (
	"fmt"
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

func (req *GetPOHCounterRequest) Command() Command {
	return CommandGetPOHCounter
}

func (req *GetPOHCounterRequest) Pack() []byte {
	return []byte{}
}

func (res *GetPOHCounterResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	res.MinutesPerCount, _, _ = unpackUint8(msg, 0)
	res.CounterReading, _, _ = unpackUint32L(msg, 1)
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

	return fmt.Sprintf(`POH Counter       : %d days, %d hours
Minutes per count : %d
Counter reading   : %d`,
		days, hours,
		res.MinutesPerCount,
		res.CounterReading,
	)
}

// This command returns the present reading of the POH (Power-On Hours) counter, plus the number of counts per hour.
func (c *Client) GetPOHCounter() (response *GetPOHCounterResponse, err error) {
	request := &GetPOHCounterRequest{}
	response = &GetPOHCounterResponse{}
	err = c.Exchange(request, response)
	return
}

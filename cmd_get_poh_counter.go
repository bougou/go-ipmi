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

func (req *GetPOHCounterRequest) Command() Command {
	return CommandGetPOHCounter
}

func (req *GetPOHCounterRequest) Pack() []byte {
	return []byte{}
}

func (res *GetPOHCounterResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShort
	}

	res.MinutesPerCount, _, _ = unpackUint8(msg, 0)
	res.CounterReading, _, _ = unpackUint32L(msg, 1)
	return nil
}

func (r *GetPOHCounterResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPOHCounterResponse) Format() string {
	return fmt.Sprintf(`Minutes per count : %d
Counter reading   : %d`,
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

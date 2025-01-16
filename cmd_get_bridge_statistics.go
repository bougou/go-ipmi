package ipmi

import (
	"context"
	"fmt"
)

type GetBridgeStatisticsRequest struct {
	StatSelector uint8
}

type GetBridgeStatisticsResponse struct {
	StatSelector uint8
	Statistics   []byte
}

func (req *GetBridgeStatisticsRequest) Command() Command {
	return CommandGetBridgeStatistics
}

func (req *GetBridgeStatisticsRequest) Pack() []byte {
	return []byte{req.StatSelector}
}

func (res *GetBridgeStatisticsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.StatSelector = msg[0]

	if len(msg) > 1 {
		res.Statistics, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}
	return nil
}

func (res *GetBridgeStatisticsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetBridgeStatisticsResponse) Format() string {
	return fmt.Sprintf("#%02x %#02x", res.StatSelector, res.Statistics)
}

func (c *Client) GetBridgeStatistics(ctx context.Context, statSelector uint8) (response *GetBridgeStatisticsResponse, err error) {
	request := &GetBridgeStatisticsRequest{
		StatSelector: statSelector,
	}
	response = &GetBridgeStatisticsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

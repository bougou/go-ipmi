package ipmi

import (
	"context"
	"fmt"
)

// 23.4 Get IP/UDP/RMCP Statistics Command
type GetIPStatisticsRequest struct {
	ChannelNumber      uint8
	ClearAllStatistics bool
}

type GetIPStatisticsResponse struct {
	IPPacketsReceived           uint16
	IPHeaderErrorsReceived      uint16
	IPAddressErrorsReceived     uint16
	IPPacketsFragmentedReceived uint16
	IPPacketsTransmitted        uint16
	UDPPacketsReceived          uint16
	RMCPPacketsValidReceived    uint16
	UDPProxyPacketsReceived     uint16
	UDPProxyPacketsDropped      uint16
}

func (req *GetIPStatisticsRequest) Pack() []byte {
	out := make([]byte, 2)

	packUint8(req.ChannelNumber, out, 0)

	var b uint8
	if req.ClearAllStatistics {
		b = setBit0(b)
	}
	packUint8(b, out, 1)

	return out
}

func (req *GetIPStatisticsRequest) Command() Command {
	return CommandGetIPStatistics
}

func (res *GetIPStatisticsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetIPStatisticsResponse) Unpack(msg []byte) error {
	if len(msg) < 18 {
		return ErrUnpackedDataTooShortWith(len(msg), 18)
	}

	res.IPPacketsReceived, _, _ = unpackUint16L(msg, 0)
	res.IPHeaderErrorsReceived, _, _ = unpackUint16L(msg, 2)
	res.IPAddressErrorsReceived, _, _ = unpackUint16L(msg, 4)
	res.IPPacketsFragmentedReceived, _, _ = unpackUint16L(msg, 6)
	res.IPPacketsTransmitted, _, _ = unpackUint16L(msg, 8)
	res.UDPPacketsReceived, _, _ = unpackUint16L(msg, 10)
	res.RMCPPacketsValidReceived, _, _ = unpackUint16L(msg, 12)
	res.UDPProxyPacketsReceived, _, _ = unpackUint16L(msg, 14)
	res.UDPProxyPacketsDropped, _, _ = unpackUint16L(msg, 16)

	return nil
}

func (res *GetIPStatisticsResponse) Format() string {
	return fmt.Sprintf(`IP Rx Packet              : %d
IP Rx Header Errors       : %d
IP Rx Address Errors      : %d
IP Rx Fragmented          : %d
IP Tx Packet              : %d
UDP Rx Packet             : %d
RMCP Rx Valid             : %d
UDP Proxy Packet Received : %d
UDP Proxy Packet Dropped  : %d`,
		res.IPPacketsReceived,
		res.IPHeaderErrorsReceived,
		res.IPAddressErrorsReceived,
		res.IPPacketsFragmentedReceived,
		res.IPPacketsTransmitted,
		res.UDPPacketsReceived,
		res.RMCPPacketsValidReceived,
		res.UDPProxyPacketsReceived,
		res.UDPProxyPacketsDropped,
	)
}

func (c *Client) GetIPStatistics(ctx context.Context, channelNumber uint8, clearAllStatistics bool) (response *GetIPStatisticsResponse, err error) {
	request := &GetIPStatisticsRequest{
		ChannelNumber:      channelNumber,
		ClearAllStatistics: clearAllStatistics,
	}
	response = &GetIPStatisticsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

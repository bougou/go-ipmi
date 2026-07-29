package transport

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
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

	types.PackUint8(req.ChannelNumber, out, 0)

	var b uint8
	if req.ClearAllStatistics {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 1)

	return out
}

func (req *GetIPStatisticsRequest) Command() types.Command {
	return types.CommandGetIPStatistics
}

func (res *GetIPStatisticsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetIPStatisticsResponse) Unpack(msg []byte) error {
	if len(msg) < 18 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 18)
	}

	res.IPPacketsReceived, _, _ = types.UnpackUint16L(msg, 0)
	res.IPHeaderErrorsReceived, _, _ = types.UnpackUint16L(msg, 2)
	res.IPAddressErrorsReceived, _, _ = types.UnpackUint16L(msg, 4)
	res.IPPacketsFragmentedReceived, _, _ = types.UnpackUint16L(msg, 6)
	res.IPPacketsTransmitted, _, _ = types.UnpackUint16L(msg, 8)
	res.UDPPacketsReceived, _, _ = types.UnpackUint16L(msg, 10)
	res.RMCPPacketsValidReceived, _, _ = types.UnpackUint16L(msg, 12)
	res.UDPProxyPacketsReceived, _, _ = types.UnpackUint16L(msg, 14)
	res.UDPProxyPacketsDropped, _, _ = types.UnpackUint16L(msg, 16)

	return nil
}

func (res *GetIPStatisticsResponse) Format() string {
	return "" +
		fmt.Sprintf("IP Rx Packet              : %d\n", res.IPPacketsReceived) +
		fmt.Sprintf("IP Rx Header Errors       : %d\n", res.IPHeaderErrorsReceived) +
		fmt.Sprintf("IP Rx Address Errors      : %d\n", res.IPAddressErrorsReceived) +
		fmt.Sprintf("IP Rx Fragmented          : %d\n", res.IPPacketsFragmentedReceived) +
		fmt.Sprintf("IP Tx Packet              : %d\n", res.IPPacketsTransmitted) +
		fmt.Sprintf("UDP Rx Packet             : %d\n", res.UDPPacketsReceived) +
		fmt.Sprintf("RMCP Rx Valid             : %d\n", res.RMCPPacketsValidReceived) +
		fmt.Sprintf("UDP Proxy Packet Received : %d\n", res.UDPProxyPacketsReceived) +
		fmt.Sprintf("UDP Proxy Packet Dropped  : %d\n", res.UDPProxyPacketsDropped)
}

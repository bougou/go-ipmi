package transport

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 24.8 Get Channel Payload Support Command
type GetChannelPayloadSupportRequest struct {
	ChannelNumber uint8
}

type GetChannelPayloadSupportResponse struct {
	// Standard payload types
	PayloadTypeIPMI bool
	PayloadTypeSOL  bool
	PayloadTypeOEM  bool

	// Session setup payload types
	PayloadTypeRmcpOpenSessionRequest  bool
	PayloadTypeRmcpOpenSessionResponse bool
	PayloadTypeRAKPMessage1            bool
	PayloadTypeRAKPMessage2            bool
	PayloadTypeRAKPMessage3            bool
	PayloadTypeRAKPMessage4            bool

	// OEM payload types
	PayloadTypeOEM0 bool
	PayloadTypeOEM1 bool
	PayloadTypeOEM2 bool
	PayloadTypeOEM3 bool
	PayloadTypeOEM4 bool
	PayloadTypeOEM5 bool
	PayloadTypeOEM6 bool
	PayloadTypeOEM7 bool
}

func (req *GetChannelPayloadSupportRequest) Pack() []byte {
	return []byte{req.ChannelNumber}
}

func (req *GetChannelPayloadSupportRequest) Command() ipmi.Command {
	return ipmi.CommandGetChannelPayloadSupport
}

func (res *GetChannelPayloadSupportResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChannelPayloadSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	res.PayloadTypeIPMI = ipmi.IsBit0Set(msg[0])
	res.PayloadTypeSOL = ipmi.IsBit1Set(msg[0])
	res.PayloadTypeOEM = ipmi.IsBit2Set(msg[0])

	res.PayloadTypeRmcpOpenSessionRequest = ipmi.IsBit0Set(msg[2])
	res.PayloadTypeRmcpOpenSessionResponse = ipmi.IsBit1Set(msg[2])
	res.PayloadTypeRAKPMessage1 = ipmi.IsBit2Set(msg[2])
	res.PayloadTypeRAKPMessage2 = ipmi.IsBit3Set(msg[2])
	res.PayloadTypeRAKPMessage3 = ipmi.IsBit4Set(msg[2])
	res.PayloadTypeRAKPMessage4 = ipmi.IsBit5Set(msg[2])

	res.PayloadTypeOEM7 = ipmi.IsBit7Set(msg[4])
	res.PayloadTypeOEM6 = ipmi.IsBit6Set(msg[4])
	res.PayloadTypeOEM5 = ipmi.IsBit5Set(msg[4])
	res.PayloadTypeOEM4 = ipmi.IsBit4Set(msg[4])
	res.PayloadTypeOEM3 = ipmi.IsBit3Set(msg[4])
	res.PayloadTypeOEM2 = ipmi.IsBit2Set(msg[4])
	res.PayloadTypeOEM1 = ipmi.IsBit1Set(msg[4])
	res.PayloadTypeOEM0 = ipmi.IsBit0Set(msg[4])

	return nil
}

func (res *GetChannelPayloadSupportResponse) Format() string {
	return "" +
		fmt.Sprintf("PayloadTypeIPMI  : %v\n", ipmi.FormatBool(res.PayloadTypeIPMI, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeSOL   : %v\n", ipmi.FormatBool(res.PayloadTypeSOL, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM   : %v\n", ipmi.FormatBool(res.PayloadTypeOEM, "supported", "unsupported")) +

		fmt.Sprintf("PayloadTypeRmcpOpenSessionRequest  : %v\n", ipmi.FormatBool(res.PayloadTypeRmcpOpenSessionRequest, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRmcpOpenSessionResponse : %v\n", ipmi.FormatBool(res.PayloadTypeRmcpOpenSessionResponse, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage1            : %v\n", ipmi.FormatBool(res.PayloadTypeRAKPMessage1, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage2            : %v\n", ipmi.FormatBool(res.PayloadTypeRAKPMessage2, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage3            : %v\n", ipmi.FormatBool(res.PayloadTypeRAKPMessage3, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage4            : %v\n", ipmi.FormatBool(res.PayloadTypeRAKPMessage4, "supported", "unsupported")) +

		fmt.Sprintf("PayloadTypeOEM0 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM0, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM1 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM1, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM2 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM2, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM3 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM3, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM4 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM4, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM5 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM5, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM6 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM6, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM7 : %v\n", ipmi.FormatBool(res.PayloadTypeOEM7, "supported", "unsupported"))
}

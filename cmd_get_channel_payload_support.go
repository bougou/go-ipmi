package ipmi

import (
	"context"
	"fmt"
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

func (req *GetChannelPayloadSupportRequest) Command() Command {
	return CommandGetChannelPayloadSupport
}

func (res *GetChannelPayloadSupportResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChannelPayloadSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	res.PayloadTypeIPMI = isBit0Set(msg[0])
	res.PayloadTypeSOL = isBit1Set(msg[0])
	res.PayloadTypeOEM = isBit2Set(msg[0])

	res.PayloadTypeRmcpOpenSessionRequest = isBit0Set(msg[2])
	res.PayloadTypeRmcpOpenSessionResponse = isBit1Set(msg[2])
	res.PayloadTypeRAKPMessage1 = isBit2Set(msg[2])
	res.PayloadTypeRAKPMessage2 = isBit3Set(msg[2])
	res.PayloadTypeRAKPMessage3 = isBit4Set(msg[2])
	res.PayloadTypeRAKPMessage4 = isBit5Set(msg[2])

	res.PayloadTypeOEM7 = isBit7Set(msg[4])
	res.PayloadTypeOEM6 = isBit6Set(msg[4])
	res.PayloadTypeOEM5 = isBit5Set(msg[4])
	res.PayloadTypeOEM4 = isBit4Set(msg[4])
	res.PayloadTypeOEM3 = isBit3Set(msg[4])
	res.PayloadTypeOEM2 = isBit2Set(msg[4])
	res.PayloadTypeOEM1 = isBit1Set(msg[4])
	res.PayloadTypeOEM0 = isBit0Set(msg[4])

	return nil
}

func (res *GetChannelPayloadSupportResponse) Format() string {
	return "" +
		fmt.Sprintf("PayloadTypeIPMI  : %v\n", formatBool(res.PayloadTypeIPMI, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeSOL   : %v\n", formatBool(res.PayloadTypeSOL, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM   : %v\n", formatBool(res.PayloadTypeOEM, "supported", "unsupported")) +

		fmt.Sprintf("PayloadTypeRmcpOpenSessionRequest  : %v\n", formatBool(res.PayloadTypeRmcpOpenSessionRequest, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRmcpOpenSessionResponse : %v\n", formatBool(res.PayloadTypeRmcpOpenSessionResponse, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage1            : %v\n", formatBool(res.PayloadTypeRAKPMessage1, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage2            : %v\n", formatBool(res.PayloadTypeRAKPMessage2, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage3            : %v\n", formatBool(res.PayloadTypeRAKPMessage3, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeRAKPMessage4            : %v\n", formatBool(res.PayloadTypeRAKPMessage4, "supported", "unsupported")) +

		fmt.Sprintf("PayloadTypeOEM0 : %v\n", formatBool(res.PayloadTypeOEM0, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM1 : %v\n", formatBool(res.PayloadTypeOEM1, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM2 : %v\n", formatBool(res.PayloadTypeOEM2, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM3 : %v\n", formatBool(res.PayloadTypeOEM3, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM4 : %v\n", formatBool(res.PayloadTypeOEM4, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM5 : %v\n", formatBool(res.PayloadTypeOEM5, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM6 : %v\n", formatBool(res.PayloadTypeOEM6, "supported", "unsupported")) +
		fmt.Sprintf("PayloadTypeOEM7 : %v\n", formatBool(res.PayloadTypeOEM7, "supported", "unsupported"))
}

func (c *Client) GetChannelPayloadSupport(ctx context.Context, channelNumber uint8) (response *GetChannelPayloadSupportResponse, err error) {
	request := &GetChannelPayloadSupportRequest{
		ChannelNumber: channelNumber,
	}
	response = &GetChannelPayloadSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

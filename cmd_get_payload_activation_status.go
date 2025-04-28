package ipmi

import (
	"context"
	"fmt"
)

// 24.4 Get Payload Activation Status Command
type GetPayloadActivationStatusRequest struct {
	PayloadType PayloadType
}

type GetPayloadActivationStatusResponse struct {

	// [3:0] - Number of instances of given payload type that can be simultaneously activated on BMC. 1-based. 0h = reserved.
	InstanceCapacity uint8

	Instance01Activated bool
	Instance02Activated bool
	Instance03Activated bool
	Instance04Activated bool
	Instance05Activated bool
	Instance06Activated bool
	Instance07Activated bool
	Instance08Activated bool
	Instance09Activated bool
	Instance10Activated bool
	Instance11Activated bool
	Instance12Activated bool
	Instance13Activated bool
	Instance14Activated bool
	Instance15Activated bool
	Instance16Activated bool

	// Store the PayloadType specified in GetPayloadActivationStatusRequest
	PayloadType PayloadType
}

func (req *GetPayloadActivationStatusRequest) Pack() []byte {
	return []byte{byte(req.PayloadType)}
}

func (req *GetPayloadActivationStatusRequest) Command() Command {
	return CommandGetPayloadActivationStatus
}

func (res *GetPayloadActivationStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPayloadActivationStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.InstanceCapacity = msg[0] + 1 // plus 1 cause of 1-based.

	res.Instance01Activated = isBit0Set(msg[1])
	res.Instance02Activated = isBit1Set(msg[1])
	res.Instance03Activated = isBit2Set(msg[1])
	res.Instance04Activated = isBit3Set(msg[1])
	res.Instance05Activated = isBit4Set(msg[1])
	res.Instance06Activated = isBit5Set(msg[1])
	res.Instance07Activated = isBit6Set(msg[1])
	res.Instance08Activated = isBit7Set(msg[1])

	res.Instance09Activated = isBit0Set(msg[2])
	res.Instance10Activated = isBit1Set(msg[2])
	res.Instance11Activated = isBit2Set(msg[2])
	res.Instance12Activated = isBit3Set(msg[2])
	res.Instance13Activated = isBit4Set(msg[2])
	res.Instance14Activated = isBit5Set(msg[2])
	res.Instance15Activated = isBit6Set(msg[2])
	res.Instance16Activated = isBit7Set(msg[2])
	return nil
}

func (res *GetPayloadActivationStatusResponse) Format() string {
	return "" +
		fmt.Sprintf("Payload Type      : %s (%d)\n", res.PayloadType.String(), uint8(res.PayloadType)) +
		fmt.Sprintf("Instance Capacity : %d\n", res.InstanceCapacity) +
		fmt.Sprintf("Instance 01       : %s\n", formatBool(res.Instance01Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 02       : %s\n", formatBool(res.Instance02Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 03       : %s\n", formatBool(res.Instance03Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 04       : %s\n", formatBool(res.Instance04Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 05       : %s\n", formatBool(res.Instance05Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 06       : %s\n", formatBool(res.Instance06Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 07       : %s\n", formatBool(res.Instance07Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 08       : %s\n", formatBool(res.Instance08Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 09       : %s\n", formatBool(res.Instance09Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 10       : %s\n", formatBool(res.Instance10Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 11       : %s\n", formatBool(res.Instance11Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 12       : %s\n", formatBool(res.Instance12Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 13       : %s\n", formatBool(res.Instance13Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 14       : %s\n", formatBool(res.Instance14Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 15       : %s\n", formatBool(res.Instance15Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 16       : %s\n", formatBool(res.Instance16Activated, "activated", "deactivated"))
}

func (c *Client) GetPayloadActivationStatus(ctx context.Context, payloadType PayloadType) (response *GetPayloadActivationStatusResponse, err error) {
	request := &GetPayloadActivationStatusRequest{
		PayloadType: payloadType,
	}
	response = &GetPayloadActivationStatusResponse{}
	response.PayloadType = request.PayloadType
	err = c.Exchange(ctx, request, response)
	return
}

package transport

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 24.4 Get Payload Activation Status Command
type GetPayloadActivationStatusRequest struct {
	PayloadType ipmi.PayloadType
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
	PayloadType ipmi.PayloadType
}

func (req *GetPayloadActivationStatusRequest) Pack() []byte {
	return []byte{byte(req.PayloadType)}
}

func (req *GetPayloadActivationStatusRequest) Command() ipmi.Command {
	return ipmi.CommandGetPayloadActivationStatus
}

func (res *GetPayloadActivationStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPayloadActivationStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.InstanceCapacity = msg[0] + 1 // plus 1 cause of 1-based.

	res.Instance01Activated = ipmi.IsBit0Set(msg[1])
	res.Instance02Activated = ipmi.IsBit1Set(msg[1])
	res.Instance03Activated = ipmi.IsBit2Set(msg[1])
	res.Instance04Activated = ipmi.IsBit3Set(msg[1])
	res.Instance05Activated = ipmi.IsBit4Set(msg[1])
	res.Instance06Activated = ipmi.IsBit5Set(msg[1])
	res.Instance07Activated = ipmi.IsBit6Set(msg[1])
	res.Instance08Activated = ipmi.IsBit7Set(msg[1])

	res.Instance09Activated = ipmi.IsBit0Set(msg[2])
	res.Instance10Activated = ipmi.IsBit1Set(msg[2])
	res.Instance11Activated = ipmi.IsBit2Set(msg[2])
	res.Instance12Activated = ipmi.IsBit3Set(msg[2])
	res.Instance13Activated = ipmi.IsBit4Set(msg[2])
	res.Instance14Activated = ipmi.IsBit5Set(msg[2])
	res.Instance15Activated = ipmi.IsBit6Set(msg[2])
	res.Instance16Activated = ipmi.IsBit7Set(msg[2])
	return nil
}

func (res *GetPayloadActivationStatusResponse) Format() string {
	return "" +
		fmt.Sprintf("Payload Type      : %s (%d)\n", res.PayloadType.String(), uint8(res.PayloadType)) +
		fmt.Sprintf("Instance Capacity : %d\n", res.InstanceCapacity) +
		fmt.Sprintf("Instance 01       : %s\n", ipmi.FormatBool(res.Instance01Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 02       : %s\n", ipmi.FormatBool(res.Instance02Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 03       : %s\n", ipmi.FormatBool(res.Instance03Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 04       : %s\n", ipmi.FormatBool(res.Instance04Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 05       : %s\n", ipmi.FormatBool(res.Instance05Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 06       : %s\n", ipmi.FormatBool(res.Instance06Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 07       : %s\n", ipmi.FormatBool(res.Instance07Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 08       : %s\n", ipmi.FormatBool(res.Instance08Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 09       : %s\n", ipmi.FormatBool(res.Instance09Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 10       : %s\n", ipmi.FormatBool(res.Instance10Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 11       : %s\n", ipmi.FormatBool(res.Instance11Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 12       : %s\n", ipmi.FormatBool(res.Instance12Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 13       : %s\n", ipmi.FormatBool(res.Instance13Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 14       : %s\n", ipmi.FormatBool(res.Instance14Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 15       : %s\n", ipmi.FormatBool(res.Instance15Activated, "activated", "deactivated")) +
		fmt.Sprintf("Instance 16       : %s\n", ipmi.FormatBool(res.Instance16Activated, "activated", "deactivated"))
}

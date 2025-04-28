package ipmi

import (
	"context"
	"fmt"
	"strings"
)

// 35.13 Get Sensor Event Status Command
type GetSensorEventStatusRequest struct {
	SensorNumber uint8
}

// For event boolean value, true means the event has occurred.
type GetSensorEventStatusResponse struct {
	EventMessagesDisabled  bool
	SensorScanningDisabled bool
	ReadingUnavailable     bool

	SensorEventFlag
}

func (req *GetSensorEventStatusRequest) Command() Command {
	return CommandGetSensorEventStatus
}

func (req *GetSensorEventStatusRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorEventStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	b1, _, _ := unpackUint8(msg, 0)
	res.EventMessagesDisabled = !isBit7Set(b1)
	res.SensorScanningDisabled = !isBit6Set(b1)
	res.ReadingUnavailable = isBit5Set(b1)

	if len(msg) >= 2 {
		b2, _, _ := unpackUint8(msg, 1)
		res.SensorEvent_UNC_High_Assert = isBit7Set(b2)
		res.SensorEvent_UNC_Low_Assert = isBit6Set(b2)
		res.SensorEvent_LNR_High_Assert = isBit5Set(b2)
		res.SensorEvent_LNR_Low_Assert = isBit4Set(b2)
		res.SensorEvent_LCR_High_Assert = isBit3Set(b2)
		res.SensorEvent_LCR_Low_Assert = isBit2Set(b2)
		res.SensorEvent_LNC_High_Assert = isBit1Set(b2)
		res.SensorEvent_LNC_Low_Assert = isBit0Set(b2)
		res.SensorEvent_State_7_Assert = isBit7Set(b2)
		res.SensorEvent_State_6_Assert = isBit6Set(b2)
		res.SensorEvent_State_5_Assert = isBit5Set(b2)
		res.SensorEvent_State_4_Assert = isBit4Set(b2)
		res.SensorEvent_State_3_Assert = isBit3Set(b2)
		res.SensorEvent_State_2_Assert = isBit2Set(b2)
		res.SensorEvent_State_1_Assert = isBit1Set(b2)
		res.SensorEvent_State_0_Assert = isBit0Set(b2)
	}

	if len(msg) >= 3 {
		b3, _, _ := unpackUint8(msg, 2)
		res.SensorEvent_UNR_High_Assert = isBit3Set(b3)
		res.SensorEvent_UNR_Low_Assert = isBit2Set(b3)
		res.SensorEvent_UCR_High_Assert = isBit1Set(b3)
		res.SensorEvent_UCR_Low_Assert = isBit0Set(b3)
		res.SensorEvent_State_14_Assert = isBit6Set(b3)
		res.SensorEvent_State_13_Assert = isBit5Set(b3)
		res.SensorEvent_State_12_Assert = isBit4Set(b3)
		res.SensorEvent_State_11_Assert = isBit3Set(b3)
		res.SensorEvent_State_10_Assert = isBit2Set(b3)
		res.SensorEvent_State_9_Assert = isBit1Set(b3)
		res.SensorEvent_State_8_Assert = isBit0Set(b3)
	}

	if len(msg) >= 4 {
		b4, _, _ := unpackUint8(msg, 3)
		res.SensorEvent_UNC_High_Deassert = isBit7Set(b4)
		res.SensorEvent_UNC_Low_Deassert = isBit6Set(b4)
		res.SensorEvent_LNR_High_Deassert = isBit5Set(b4)
		res.SensorEvent_LNR_Low_Deassert = isBit4Set(b4)
		res.SensorEvent_LCR_High_Deassert = isBit3Set(b4)
		res.SensorEvent_LCR_Low_Deassert = isBit2Set(b4)
		res.SensorEvent_LNC_High_Deassert = isBit1Set(b4)
		res.SensorEvent_LNC_Low_Deassert = isBit0Set(b4)
		res.SensorEvent_State_7_Deassert = isBit7Set(b4)
		res.SensorEvent_State_6_Deassert = isBit6Set(b4)
		res.SensorEvent_State_5_Deassert = isBit5Set(b4)
		res.SensorEvent_State_4_Deassert = isBit4Set(b4)
		res.SensorEvent_State_3_Deassert = isBit3Set(b4)
		res.SensorEvent_State_2_Deassert = isBit2Set(b4)
		res.SensorEvent_State_1_Deassert = isBit1Set(b4)
		res.SensorEvent_State_0_Deassert = isBit0Set(b4)
	}

	if len(msg) >= 5 {
		b5, _, _ := unpackUint8(msg, 4)
		res.SensorEvent_UNR_High_Deassert = isBit3Set(b5)
		res.SensorEvent_UNR_Low_Deassert = isBit2Set(b5)
		res.SensorEvent_UCR_High_Deassert = isBit1Set(b5)
		res.SensorEvent_UCR_Low_Deassert = isBit0Set(b5)
		res.SensorEvent_State_14_Deassert = isBit6Set(b5)
		res.SensorEvent_State_13_Deassert = isBit5Set(b5)
		res.SensorEvent_State_12_Deassert = isBit4Set(b5)
		res.SensorEvent_State_11_Deassert = isBit3Set(b5)
		res.SensorEvent_State_10_Deassert = isBit2Set(b5)
		res.SensorEvent_State_9_Deassert = isBit1Set(b5)
		res.SensorEvent_State_8_Deassert = isBit0Set(b5)
	}

	return nil
}

func (res *GetSensorEventStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorEventStatusResponse) Format() string {

	all := res.SensorEventFlag.TrueEvents()
	asserted := SensorEvents(all).FilterAssert()
	deasserted := SensorEvents(all).FilterDeassert()

	var assertedStr = []string{}
	var deassertedStr = []string{}
	for _, v := range asserted {
		assertedStr = append(assertedStr, v.String())
	}
	for _, v := range deasserted {
		deassertedStr = append(deassertedStr, v.String())
	}

	return "" +
		fmt.Sprintf("Event Messages Disabled   : %v\n", res.EventMessagesDisabled) +
		fmt.Sprintf("Sensor Scanning Disabled  : %v\n", res.SensorScanningDisabled) +
		fmt.Sprintf("Reading State Unavailable : %v\n", res.ReadingUnavailable) +
		fmt.Sprintf("Occurred Assert Event     : %s\n", strings.Join(assertedStr, "\n - ")) +
		fmt.Sprintf("Occurred Deassert Event   : %s\n", strings.Join(deassertedStr, "\n -"))
}

func (c *Client) GetSensorEventStatus(ctx context.Context, sensorNumber uint8) (response *GetSensorEventStatusResponse, err error) {
	request := &GetSensorEventStatusRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorEventStatusResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

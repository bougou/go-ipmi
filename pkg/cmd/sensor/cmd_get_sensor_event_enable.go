package sensor

import (
	"fmt"
	"strings"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 35.11 Get Sensor Event Enable Command
type GetSensorEventEnableRequest struct {
	SensorNumber uint8
}

// For event , true means the event has enabled.
type GetSensorEventEnableResponse struct {
	EventMessagesDisabled  bool
	SensorScanningDisabled bool
	ipmi.SensorEventFlag
}

func (req *GetSensorEventEnableRequest) Command() ipmi.Command {
	return ipmi.CommandGetSensorEventEnable
}

func (req *GetSensorEventEnableRequest) Pack() []byte {
	out := make([]byte, 1)
	ipmi.PackUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorEventEnableResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	b1, _, _ := ipmi.UnpackUint8(msg, 0)
	res.EventMessagesDisabled = !ipmi.IsBit7Set(b1)
	res.SensorScanningDisabled = !ipmi.IsBit6Set(b1)

	if len(msg) >= 2 {
		b2, _, _ := ipmi.UnpackUint8(msg, 1)
		res.SensorEvent_UNC_High_Assert = ipmi.IsBit7Set(b2)
		res.SensorEvent_UNC_Low_Assert = ipmi.IsBit6Set(b2)
		res.SensorEvent_LNR_High_Assert = ipmi.IsBit5Set(b2)
		res.SensorEvent_LNR_Low_Assert = ipmi.IsBit4Set(b2)
		res.SensorEvent_LCR_High_Assert = ipmi.IsBit3Set(b2)
		res.SensorEvent_LCR_Low_Assert = ipmi.IsBit2Set(b2)
		res.SensorEvent_LNC_High_Assert = ipmi.IsBit1Set(b2)
		res.SensorEvent_LNC_Low_Assert = ipmi.IsBit0Set(b2)
		res.SensorEvent_State_7_Assert = ipmi.IsBit7Set(b2)
		res.SensorEvent_State_6_Assert = ipmi.IsBit6Set(b2)
		res.SensorEvent_State_5_Assert = ipmi.IsBit5Set(b2)
		res.SensorEvent_State_4_Assert = ipmi.IsBit4Set(b2)
		res.SensorEvent_State_3_Assert = ipmi.IsBit3Set(b2)
		res.SensorEvent_State_2_Assert = ipmi.IsBit2Set(b2)
		res.SensorEvent_State_1_Assert = ipmi.IsBit1Set(b2)
		res.SensorEvent_State_0_Assert = ipmi.IsBit0Set(b2)
	}

	if len(msg) >= 3 {
		b3, _, _ := ipmi.UnpackUint8(msg, 2)
		res.SensorEvent_UNR_High_Assert = ipmi.IsBit3Set(b3)
		res.SensorEvent_UNR_Low_Assert = ipmi.IsBit2Set(b3)
		res.SensorEvent_UCR_High_Assert = ipmi.IsBit1Set(b3)
		res.SensorEvent_UCR_Low_Assert = ipmi.IsBit0Set(b3)
		res.SensorEvent_State_14_Assert = ipmi.IsBit6Set(b3)
		res.SensorEvent_State_13_Assert = ipmi.IsBit5Set(b3)
		res.SensorEvent_State_12_Assert = ipmi.IsBit4Set(b3)
		res.SensorEvent_State_11_Assert = ipmi.IsBit3Set(b3)
		res.SensorEvent_State_10_Assert = ipmi.IsBit2Set(b3)
		res.SensorEvent_State_9_Assert = ipmi.IsBit1Set(b3)
		res.SensorEvent_State_8_Assert = ipmi.IsBit0Set(b3)
	}

	if len(msg) >= 4 {
		b4, _, _ := ipmi.UnpackUint8(msg, 3)
		res.SensorEvent_UNC_High_Deassert = ipmi.IsBit7Set(b4)
		res.SensorEvent_UNC_Low_Deassert = ipmi.IsBit6Set(b4)
		res.SensorEvent_LNR_High_Deassert = ipmi.IsBit5Set(b4)
		res.SensorEvent_LNR_Low_Deassert = ipmi.IsBit4Set(b4)
		res.SensorEvent_LCR_High_Deassert = ipmi.IsBit3Set(b4)
		res.SensorEvent_LCR_Low_Deassert = ipmi.IsBit2Set(b4)
		res.SensorEvent_LNC_High_Deassert = ipmi.IsBit1Set(b4)
		res.SensorEvent_LNC_Low_Deassert = ipmi.IsBit0Set(b4)
		res.SensorEvent_State_7_Deassert = ipmi.IsBit7Set(b4)
		res.SensorEvent_State_6_Deassert = ipmi.IsBit6Set(b4)
		res.SensorEvent_State_5_Deassert = ipmi.IsBit5Set(b4)
		res.SensorEvent_State_4_Deassert = ipmi.IsBit4Set(b4)
		res.SensorEvent_State_3_Deassert = ipmi.IsBit3Set(b4)
		res.SensorEvent_State_2_Deassert = ipmi.IsBit2Set(b4)
		res.SensorEvent_State_1_Deassert = ipmi.IsBit1Set(b4)
		res.SensorEvent_State_0_Deassert = ipmi.IsBit0Set(b4)
	}

	if len(msg) >= 5 {
		b5, _, _ := ipmi.UnpackUint8(msg, 4)
		res.SensorEvent_UNR_High_Deassert = ipmi.IsBit3Set(b5)
		res.SensorEvent_UNR_Low_Deassert = ipmi.IsBit2Set(b5)
		res.SensorEvent_UCR_High_Deassert = ipmi.IsBit1Set(b5)
		res.SensorEvent_UCR_Low_Deassert = ipmi.IsBit0Set(b5)
		res.SensorEvent_State_14_Deassert = ipmi.IsBit6Set(b5)
		res.SensorEvent_State_13_Deassert = ipmi.IsBit5Set(b5)
		res.SensorEvent_State_12_Deassert = ipmi.IsBit4Set(b5)
		res.SensorEvent_State_11_Deassert = ipmi.IsBit3Set(b5)
		res.SensorEvent_State_10_Deassert = ipmi.IsBit2Set(b5)
		res.SensorEvent_State_9_Deassert = ipmi.IsBit1Set(b5)
		res.SensorEvent_State_8_Deassert = ipmi.IsBit0Set(b5)
	}

	return nil
}

func (r *GetSensorEventEnableResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorEventEnableResponse) Format() string {
	all := res.SensorEventFlag.TrueEvents()
	asserted := ipmi.SensorEvents(all).FilterAssert()
	deasserted := ipmi.SensorEvents(all).FilterDeassert()

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
		fmt.Sprintf("Enabled Assert Event      : %s\n", strings.Join(assertedStr, "\n - ")) +
		fmt.Sprintf("Enabled Deassert Event    : %s\n", strings.Join(deassertedStr, "\n - "))
}

package ipmi

import (
	"context"
	"fmt"
)

// 35.2 Get Device SDR Info Command
type GetDeviceSDRInfoRequest struct {
	// true: Get SDR count. This returns the total number of SDRs in the device.
	// false: Get Sensor count. This returns the number of sensors implemented on LUN this command was addressed to.
	GetSDRCount bool
}

type GetDeviceSDRInfoResponse struct {
	getSDRCount bool

	Count uint8

	// 0b = static sensor population. The number of sensors handled by this
	// device is fixed, and a query shall return records for all sensors.
	//
	// 1b = dynamic sensor population. This device may have its sensor
	// population vary during "run time" (defined as any time other that
	// when an install operation is in progress).
	DynamicSensorPopulation bool

	LUN3HasSensors bool
	LUN2HasSensors bool
	LUN1HasSensors bool
	LUN0HasSensors bool

	// Four byte timestamp, or counter. Updated or incremented each time the
	// sensor population changes. This field is not provided if the flags indicate a
	// static sensor population.
	SensorPopulationChangeIndicator uint32
}

func (req *GetDeviceSDRInfoRequest) Command() Command {
	return CommandGetDeviceSDRInfo
}

func (req *GetDeviceSDRInfoRequest) Pack() []byte {
	var b uint8
	if req.GetSDRCount {
		b = setBit0(b)
	}
	return []byte{b}
}

func (res *GetDeviceSDRInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.Count, _, _ = unpackUint8(msg, 0)
	b, _, _ := unpackUint8(msg, 1)

	res.DynamicSensorPopulation = isBit7Set(b)
	res.LUN3HasSensors = isBit3Set(b)
	res.LUN2HasSensors = isBit2Set(b)
	res.LUN1HasSensors = isBit1Set(b)
	res.LUN0HasSensors = isBit0Set(b)

	if res.DynamicSensorPopulation {
		if len(msg) < 6 {
			return ErrUnpackedDataTooShortWith(len(msg), 6)
		}
		res.SensorPopulationChangeIndicator, _, _ = unpackUint32L(msg, 2)
	}

	return nil
}

func (r *GetDeviceSDRInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDeviceSDRInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Count              : %d (%s)\n", res.Count, formatBool(res.getSDRCount, "SDRs", "Sensors")) +
		fmt.Sprintf("Dynamic Population : %v\n", res.DynamicSensorPopulation) +
		fmt.Sprintf("LUN 0 has sensors  : %v\n", res.LUN0HasSensors) +
		fmt.Sprintf("LUN 1 has sensors  : %v\n", res.LUN1HasSensors) +
		fmt.Sprintf("LUN 2 has sensors  : %v\n", res.LUN2HasSensors) +
		fmt.Sprintf("LUN 3 has sensors  : %v\n", res.LUN3HasSensors)
}

// This command returns general information about the collection of sensors in a Dynamic Sensor Device.
func (c *Client) GetDeviceSDRInfo(ctx context.Context, getSDRCount bool) (response *GetDeviceSDRInfoResponse, err error) {
	request := &GetDeviceSDRInfoRequest{
		GetSDRCount: getSDRCount,
	}
	response = &GetDeviceSDRInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

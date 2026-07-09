package dcmi

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5]: 6.7.3 Get Temperature Readings Command
type GetDCMITemperatureReadingsRequest struct {
	SensorType          types.SensorType
	EntityID            types.EntityID
	EntityInstance      types.EntityInstance
	EntityInstanceStart uint8
}

type GetDCMITemperatureReadingsResponse struct {
	EntityID types.EntityID

	TotalEntityInstances     uint8
	TemperatureReadingsCount uint8
	TemperatureReadings      []DCMITemperatureReading
}

type DCMITemperatureReading struct {
	TemperatureReading int8
	EntityInstance     types.EntityInstance
	EntityID           types.EntityID
}

func (req *GetDCMITemperatureReadingsRequest) Pack() []byte {
	return []byte{types.GroupExtensionDCMI, byte(req.SensorType), byte(req.EntityID), byte(req.EntityInstance), req.EntityInstanceStart}
}

func (req *GetDCMITemperatureReadingsRequest) Command() types.Command {
	return types.CommandGetDCMITemperatureReadings
}

func (res *GetDCMITemperatureReadingsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "No Active Set Power Limit",
	}
}

func (res *GetDCMITemperatureReadingsResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalEntityInstances = msg[1]
	res.TemperatureReadingsCount = msg[2]

	if len(msg) < 3+int(res.TemperatureReadingsCount)*2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3+int(res.TemperatureReadingsCount)*2)
	}

	tempReadings := make([]DCMITemperatureReading, 0)
	for i := 0; i < int(res.TemperatureReadingsCount); i++ {
		r := DCMITemperatureReading{}

		v := msg[3+i*2]
		r.TemperatureReading = int8(v)

		r.EntityInstance = types.EntityInstance(msg[3+i*2+1])
		r.EntityID = res.EntityID

		tempReadings = append(tempReadings, r)
	}

	res.TemperatureReadings = tempReadings

	return nil
}

func (res *GetDCMITemperatureReadingsResponse) Format() string {
	return "" +
		fmt.Sprintf("Total entity instances         : %d\n", res.TotalEntityInstances) +
		fmt.Sprintf("Number of temperature readings : %d\n", res.TemperatureReadingsCount) +
		fmt.Sprintf("Temperature Readings           : %v\n", res.TemperatureReadings)
}

func FormatDCMITemperatureReadings(readings []DCMITemperatureReading) string {
	rows := make([]map[string]string, len(readings))

	for i, reading := range readings {
		rows[i] = map[string]string{
			"Entity ID":       fmt.Sprintf("%s(%#02x)", reading.EntityID.String(), uint8(reading.EntityID)),
			"Entity Instance": fmt.Sprintf("%d", reading.EntityInstance),
			"Temp. Readings":  fmt.Sprintf("%+d C", reading.TemperatureReading),
		}
	}

	headers := []string{
		"Entity ID",
		"Entity Instance",
		"Temp. Readings",
	}

	return types.RenderTable(headers, rows)
}

package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5]: 6.7.3 Get Temperature Readings Command
type GetDCMITemperatureReadingsRequest struct {
	SensorType          SensorType
	EntityID            EntityID
	EntityInstance      EntityInstance
	EntityInstanceStart uint8
}

type GetDCMITemperatureReadingsResponse struct {
	entityID EntityID

	TotalEntityInstances     uint8
	TemperatureReadingsCount uint8
	TemperatureReadings      []DCMITemperatureReading
}

type DCMITemperatureReading struct {
	TemperatureReading int8
	EntityInstance     EntityInstance
	EntityID           EntityID
}

func (req *GetDCMITemperatureReadingsRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, byte(req.SensorType), byte(req.EntityID), byte(req.EntityInstance), req.EntityInstanceStart}
}

func (req *GetDCMITemperatureReadingsRequest) Command() Command {
	return CommandGetDCMITemperatureReadings
}

func (res *GetDCMITemperatureReadingsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "No Active Set Power Limit",
	}
}

func (res *GetDCMITemperatureReadingsResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalEntityInstances = msg[1]
	res.TemperatureReadingsCount = msg[2]

	if len(msg) < 3+int(res.TemperatureReadingsCount)*2 {
		return ErrUnpackedDataTooShortWith(len(msg), 3+int(res.TemperatureReadingsCount)*2)
	}

	tempReadings := make([]DCMITemperatureReading, 0)
	for i := 0; i < int(res.TemperatureReadingsCount); i++ {
		r := DCMITemperatureReading{}

		v := msg[3+i*2]
		r.TemperatureReading = int8(v)

		r.EntityInstance = EntityInstance(msg[3+i*2+1])
		r.EntityID = res.entityID

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

func (c *Client) GetDCMITemperatureReadings(ctx context.Context, request *GetDCMITemperatureReadingsRequest) (response *GetDCMITemperatureReadingsResponse, err error) {
	response = &GetDCMITemperatureReadingsResponse{
		entityID: request.EntityID,
	}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMITemperatureReadingsForEntities(ctx context.Context, entityIDs ...EntityID) ([]DCMITemperatureReading, error) {
	out := make([]DCMITemperatureReading, 0)

	for _, entityID := range entityIDs {
		request := &GetDCMITemperatureReadingsRequest{
			SensorType:          SensorTypeTemperature,
			EntityID:            entityID,
			EntityInstance:      0x00,
			EntityInstanceStart: 0,
		}

		response, err := c.GetDCMITemperatureReadings(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("GetDCMITemperatureReadings failed for entityID (%#02x), err: %w", entityID, err)
		}

		out = append(out, response.TemperatureReadings...)
	}

	return out, nil
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

	return RenderTable(headers, rows)
}

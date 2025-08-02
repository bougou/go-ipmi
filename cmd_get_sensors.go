package ipmi

import (
	"context"
	"fmt"
	"strings"
)

type SensorFilterOption func(sensor *Sensor) bool

func SensorFilterOptionIsThreshold(sensor *Sensor) bool {
	return sensor.IsThreshold()
}

func SensorFilterOptionIsReadingValid(sensor *Sensor) bool {
	return sensor.IsReadingValid()
}

// Sensor is matched if the sensor type of the sensor is one of the given sensor types.
func SensorFilterOptionIsSensorType(sensorTypes ...SensorType) func(sensor *Sensor) bool {
	return func(sensor *Sensor) bool {
		for _, sensorType := range sensorTypes {
			if sensor.SensorType == sensorType {
				return true
			}
		}
		return false
	}
}

// GetSensors retrieves all sensors with their current readings and status.
//
// Filter behavior:
//   - If no filter options are provided, returns all sensors
//   - If filter options are provided, returns only sensors that pass ALL filters (logical AND)
//   - For logical OR filtering, use GetSensorsAny instead
//
// Example usage:
//
//	// Get all fan sensors
//	sensors, err := client.GetSensors(ctx, ipmi.SensorFilterOptionIsSensorType(ipmi.SensorTypeFan))
//
//	// Get all temperature sensors with valid readings
//	sensors, err := client.GetSensors(ctx,
//	    ipmi.SensorFilterOptionIsSensorType(ipmi.SensorTypeTemperature),
//	    ipmi.SensorFilterOptionIsReadingValid)
func (c *Client) GetSensors(ctx context.Context, filterOptions ...SensorFilterOption) ([]*Sensor, error) {
	var out = make([]*Sensor, 0)

	sdrs, err := c.GetSDRs(ctx, SDRRecordTypeFullSensor, SDRRecordTypeCompactSensor)
	if err != nil {
		return nil, fmt.Errorf("GetSDRs failed, err: %w", err)
	}

	for _, sdr := range sdrs {
		sensor, err := c.sdrToSensor(ctx, sdr)
		if err != nil {
			return nil, fmt.Errorf("sdrToSensor failed, err: %w", err)
		}

		var choose bool = true
		for _, filterOption := range filterOptions {
			if !filterOption(sensor) {
				choose = false
				break
			}
		}

		if choose {
			out = append(out, sensor)
		}
	}

	return out, nil
}

// GetSensorsAny retrieves all sensors with their current readings and status.
//
// Filter behavior:
//   - If no filter options are provided, returns all sensors
//   - If filter options are provided, returns sensors that pass ANY of the filters (logical OR)
//   - For logical AND filtering, use GetSensors instead
//
// Example usage:
//
//	// Get sensors that are either temperature or voltage sensors
//	sensors, err := client.GetSensorsAny(ctx,
//	    ipmi.SensorFilterOptionIsSensorType(ipmi.SensorTypeTemperature),
//	    ipmi.SensorFilterOptionIsSensorType(ipmi.SensorTypeVoltage))
func (c *Client) GetSensorsAny(ctx context.Context, filterOptions ...SensorFilterOption) ([]*Sensor, error) {
	var out = make([]*Sensor, 0)

	sdrs, err := c.GetSDRs(ctx, SDRRecordTypeFullSensor, SDRRecordTypeCompactSensor)
	if err != nil {
		return nil, fmt.Errorf("GetSDRs failed, err: %w", err)
	}

	for _, sdr := range sdrs {
		sensor, err := c.sdrToSensor(ctx, sdr)
		if err != nil {
			return nil, fmt.Errorf("sdrToSensor failed, err: %w", err)
		}

		var choose bool = false
		for _, filterOption := range filterOptions {
			if filterOption(sensor) {
				choose = true
				break
			}
		}

		if choose {
			out = append(out, sensor)
		}
	}

	return out, nil
}

// GetSensorByID returns the sensor with current reading and status by specified sensor number.
func (c *Client) GetSensorByID(ctx context.Context, sensorNumber uint8) (*Sensor, error) {
	sdr, err := c.GetSDRBySensorID(ctx, sensorNumber)
	if err != nil {
		return nil, fmt.Errorf("GetSDRBySensorID failed, err: %w", err)
	}

	c.Debug("SDR", sdr)

	sensor, err := c.sdrToSensor(ctx, sdr)
	if err != nil {
		return nil, fmt.Errorf("GetSensorFromSDR failed, err: %w", err)
	}

	return sensor, nil
}

// GetSensorByName returns the sensor with current reading and status by specified sensor name.
func (c *Client) GetSensorByName(ctx context.Context, sensorName string) (*Sensor, error) {
	sdr, err := c.GetSDRBySensorName(ctx, sensorName)
	if err != nil {
		return nil, fmt.Errorf("GetSDRBySensorName failed, err: %w", err)
	}

	c.Debug("SDR", sdr)

	sensor, err := c.sdrToSensor(ctx, sdr)
	if err != nil {
		return nil, fmt.Errorf("GetSensorFromSDR failed, err: %w", err)
	}

	return sensor, nil
}

// sdrToSensor converts a Sensor Data Record (SDR) to a Sensor struct.
//
// Requirements:
//   - Only Full and Compact SDR records are supported
//   - Returns error for other record types
//
// This function performs additional IPMI commands to fetch sensor-related values
// that are not stored in the SDR record, including:
//   - Current sensor readings
//   - Threshold values
//   - Event status
//   - Hysteresis values
//
// The function handles both Full and Compact SDR record types, populating
// the appropriate fields based on the record type.
func (c *Client) sdrToSensor(ctx context.Context, sdr *SDR) (*Sensor, error) {
	if sdr == nil {
		return nil, fmt.Errorf("nil sdr parameter")
	}

	sensor := &Sensor{
		SDRRecordType:    sdr.RecordHeader.RecordType,
		HasAnalogReading: sdr.HasAnalogReading(),
	}

	switch sdr.RecordHeader.RecordType {
	case SDRRecordTypeFullSensor:
		sensor.GeneratorID = sdr.Full.GeneratorID
		sensor.Number = uint8(sdr.Full.SensorNumber)
		sensor.Name = strings.TrimSpace(string(sdr.Full.IDStringBytes))
		sensor.SensorUnit = sdr.Full.SensorUnit
		sensor.SensorType = sdr.Full.SensorType
		sensor.EventReadingType = sdr.Full.SensorEventReadingType
		sensor.SensorInitialization = sdr.Full.SensorInitialization
		sensor.SensorCapabilities = sdr.Full.SensorCapabilities
		sensor.EntityID = sdr.Full.SensorEntityID
		sensor.EntityInstance = sdr.Full.SensorEntityInstance

		sensor.Threshold.LinearizationFunc = sdr.Full.LinearizationFunc
		sensor.Threshold.ReadingFactors = sdr.Full.ReadingFactors

	case SDRRecordTypeCompactSensor:
		sensor.GeneratorID = sdr.Compact.GeneratorID
		sensor.Number = uint8(sdr.Compact.SensorNumber)
		sensor.Name = strings.TrimSpace(string(sdr.Compact.IDStringBytes))
		sensor.SensorUnit = sdr.Compact.SensorUnit
		sensor.SensorType = sdr.Compact.SensorType
		sensor.EventReadingType = sdr.Compact.SensorEventReadingType
		sensor.SensorInitialization = sdr.Compact.SensorInitialization
		sensor.SensorCapabilities = sdr.Compact.SensorCapabilities
		sensor.EntityID = sdr.Compact.SensorEntityID
		sensor.EntityInstance = sdr.Compact.SensorEntityInstance

	default:
		return nil, fmt.Errorf("only support Full or Compact SDR record type, input is %s", sdr.RecordHeader.RecordType)
	}

	sensorOwner := uint8(sensor.GeneratorID.OwnerID())
	sensorLUN := uint8(sensor.GeneratorID.LUN())
	commandContext := &CommandContext{}
	commandContext.
		WithResponderAddr(sensorOwner).
		WithResponderLUN(sensorLUN)

	ctx = WithCommandContext(ctx, commandContext)
	c.Debug("Set CommandContext:", commandContext)

	if err := c.fillSensorReading(ctx, sensor); err != nil {
		return nil, fmt.Errorf("fillSensorReading failed, err: %w", err)
	}

	// notPresent is filled/set by fillSensorReading
	if sensor.notPresent {
		c.Debug(fmt.Sprintf(":( Sensor [%s](%#02x) not present\n", sensor.Name, sensor.Number), "")
		return sensor, nil
	}

	// scanningDisabled is filled/set by fillSensorReading
	if sensor.scanningDisabled {
		// Sensor scanning disabled, no need to continue
		c.Debug(fmt.Sprintf(":( Sensor [%s](%#02x) scanning disabled\n", sensor.Name, sensor.Number), "")
		return sensor, nil
	}

	if !sensor.EventReadingType.IsThreshold() || !sensor.SensorUnit.IsAnalog() {
		if err := c.fillSensorDiscrete(ctx, sensor); err != nil {
			return nil, fmt.Errorf("fillSensorDiscrete failed, err: %w", err)
		}
	} else {
		if err := c.fillSensorThreshold(ctx, sensor); err != nil {
			return nil, fmt.Errorf("fillSensorThreshold failed, err: %w", err)
		}
	}

	return sensor, nil
}

func (c *Client) fillSensorReading(ctx context.Context, sensor *Sensor) error {
	c.Debug("try to fill sensor reading for sensor", sensor.Number)

	readingRes, err := c.GetSensorReading(ctx, sensor.Number)
	c.Debug("GetSensorReading response", readingRes.Format())

	if isErrOfCompletionCodes(err, uint8(CompletionCodeRequestedDataNotPresent)) {
		c.Debugf("GetSensorReading for sensor %#02x failed, err: %s", sensor.Number, err)
		sensor.notPresent = true
		return nil
	}

	if _canIgnoreSensorErr(err) != nil {
		return fmt.Errorf("GetSensorReading for sensor %#02x failed, err: %w", sensor.Number, err)
	}

	sensor.Raw = readingRes.Reading
	sensor.Value = sensor.ConvertReading(readingRes.Reading)

	sensor.scanningDisabled = readingRes.SensorScanningDisabled
	sensor.readingAvailable = !readingRes.ReadingUnavailable
	sensor.Threshold.ThresholdStatus = readingRes.ThresholdStatus()

	sensor.Discrete.ActiveStates = readingRes.ActiveStates
	sensor.Discrete.optionalData1 = readingRes.optionalData1
	sensor.Discrete.optionalData2 = readingRes.optionalData2

	return nil
}

// fillSensorDiscrete retrieves and fills extra sensor attributes for given discrete sensor.
func (c *Client) fillSensorDiscrete(ctx context.Context, sensor *Sensor) error {
	statusRes, err := c.GetSensorEventStatus(ctx, sensor.Number)
	if _canIgnoreSensorErr(err) != nil {
		return fmt.Errorf("GetSensorEventStatus for sensor %#02x failed, err: %w", sensor.Number, err)
	}
	sensor.OccurredEvents = statusRes.SensorEventFlag.TrueEvents()
	return nil
}

// fillSensorThreshold retrieves and fills sensor attributes for given threshold sensor.
func (c *Client) fillSensorThreshold(ctx context.Context, sensor *Sensor) error {
	if sensor.SDRRecordType != SDRRecordTypeFullSensor {
		return nil
	}

	// If Non Linear, should update the ReadingFactors
	// see 36.2 Non-Linear Sensors
	if sensor.Threshold.LinearizationFunc.IsNonLinear() {
		factorsRes, err := c.GetSensorReadingFactors(ctx, sensor.Number, sensor.Raw)
		if _canIgnoreSensorErr(err) != nil {
			return fmt.Errorf("GetSensorReadingFactors for sensor %#02x failed, err: %w", sensor.Number, err)
		}
		sensor.Threshold.ReadingFactors = factorsRes.ReadingFactors
	}

	thresholdRes, err := c.GetSensorThresholds(ctx, sensor.Number)
	if _canIgnoreSensorErr(err) != nil {
		return fmt.Errorf("GetSensorThresholds for sensor %#02x failed, err: %w", sensor.Number, err)
	}
	sensor.Threshold.Mask.UNR.Readable = thresholdRes.UNR_Readable
	sensor.Threshold.Mask.UCR.Readable = thresholdRes.UCR_Readable
	sensor.Threshold.Mask.UNC.Readable = thresholdRes.UNC_Readable
	sensor.Threshold.Mask.LNR.Readable = thresholdRes.LNR_Readable
	sensor.Threshold.Mask.LCR.Readable = thresholdRes.LCR_Readable
	sensor.Threshold.Mask.LNC.Readable = thresholdRes.LNC_Readable
	sensor.Threshold.LNC_Raw = thresholdRes.LNC_Raw
	sensor.Threshold.LCR_Raw = thresholdRes.LCR_Raw
	sensor.Threshold.LNR_Raw = thresholdRes.LNR_Raw
	sensor.Threshold.UNC_Raw = thresholdRes.UNC_Raw
	sensor.Threshold.UCR_Raw = thresholdRes.UCR_Raw
	sensor.Threshold.UNR_Raw = thresholdRes.UNR_Raw
	sensor.Threshold.LNC = sensor.ConvertReading(thresholdRes.LNC_Raw)
	sensor.Threshold.LCR = sensor.ConvertReading(thresholdRes.LCR_Raw)
	sensor.Threshold.LNR = sensor.ConvertReading(thresholdRes.LNR_Raw)
	sensor.Threshold.UNC = sensor.ConvertReading(thresholdRes.UNC_Raw)
	sensor.Threshold.UCR = sensor.ConvertReading(thresholdRes.UCR_Raw)
	sensor.Threshold.UNR = sensor.ConvertReading(thresholdRes.UNR_Raw)

	hysteresisRes, err := c.GetSensorHysteresis(ctx, sensor.Number)
	if _canIgnoreSensorErr(err) != nil {
		return fmt.Errorf("GetSensorHysteresis for sensor %#02x failed, err: %w", sensor.Number, err)
	}
	sensor.Threshold.PositiveHysteresisRaw = hysteresisRes.PositiveRaw
	sensor.Threshold.NegativeHysteresisRaw = hysteresisRes.NegativeRaw
	sensor.Threshold.PositiveHysteresis = sensor.ConvertSensorHysteresis(hysteresisRes.PositiveRaw)
	sensor.Threshold.NegativeHysteresis = sensor.ConvertSensorHysteresis(hysteresisRes.NegativeRaw)

	return nil
}

func _canIgnoreSensorErr(err error) error {
	canIgnore := buildCanIgnoreFn(
		// the following completion codes CAN be ignored,
		// it normally means the sensor device does not exist or the sensor device does not recognize the IPMI command
		uint8(CompletionCodeRequestedDataNotPresent),
		uint8(CompletionCodeIllegalCommand),
		uint8(CompletionCodeInvalidCommand),
	)

	return canIgnore(err)
}

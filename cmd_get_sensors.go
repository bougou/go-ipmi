package ipmi

import (
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

func SensorFilterOptionIsSensorType(sensorType SensorType) func(sensor *Sensor) bool {
	return func(sensor *Sensor) bool {
		return sensor.SensorType == sensorType
	}
}

// GetSensors returns all sensors with their current readings and status.
// If there's no filter options, it returns all sensors.
// If there exists filter options, it only returns the sensors those
// passed ALL filter options.
func (c *Client) GetSensors(filterOptions ...SensorFilterOption) ([]*Sensor, error) {
	var out = make([]*Sensor, 0)

	sdrs, err := c.GetSDRs(SDRRecordTypeFullSensor, SDRRecordTypeCompactSensor)
	if err != nil {
		return nil, fmt.Errorf("GetSDRs failed, err: %s", err)
	}

	for _, sdr := range sdrs {
		sensor, err := c.sdrToSensor(sdr)
		if err != nil {
			return nil, fmt.Errorf("sdrToSensor failed, err: %s", err)
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

// GetSensor returns the sensor with current reading and status by specified sensor number.
func (c *Client) GetSensorByID(sensorNumber uint8) (*Sensor, error) {
	sdr, err := c.GetSDRBySensorID(sensorNumber)
	if err != nil {
		return nil, fmt.Errorf("GetSDRBySensorID failed, err: %s", err)
	}

	sensor, err := c.sdrToSensor(sdr)
	if err != nil {
		return nil, fmt.Errorf("GetSensorFromSDR failed, err: %s", err)
	}

	return sensor, nil
}

// GetSensor returns the sensor with current reading and status by specified sensor name.
func (c *Client) GetSensorByName(sensorName string) (*Sensor, error) {
	sdr, err := c.GetSDRBySensorName(sensorName)
	if err != nil {
		return nil, fmt.Errorf("GetSDRBySensorName failed, err: %s", err)
	}

	sensor, err := c.sdrToSensor(sdr)
	if err != nil {
		return nil, fmt.Errorf("GetSensorFromSDR failed, err: %s", err)
	}

	return sensor, nil
}

// sdrToSensor convert SDR record to Sensor struct.
//
// Only Full and Compact SDR records are meaningful here. Pass SDRs with other record types will return error.
//
// This function will fetch other sensor-related values which are not stored in SDR by other IPMI commands.
func (c *Client) sdrToSensor(sdr *SDR) (*Sensor, error) {
	if sdr == nil {
		return nil, fmt.Errorf("nil sdr parameter")
	}

	sensor := &Sensor{
		SDRRecordType:    sdr.RecordHeader.RecordType,
		HasAnalogReading: sdr.HasAnalogReading(),
	}

	switch sdr.RecordHeader.RecordType {
	case SDRRecordTypeFullSensor:
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

	c.Debug("Sensor:", sensor)
	c.Debug("Get Sensor", fmt.Sprintf("Sensor Name: %s, Sensor Number: %#02x\n", sensor.Name, sensor.Number))

	if err := c.fillSensorReading(sensor); err != nil {
		return nil, fmt.Errorf("fillSensorReading failed, err: %s", err)
	}

	// scanningDisabled is filled/set by fillSensorReading
	if sensor.scanningDisabled {
		// Sensor scanning disabled, no need to continue
		c.Debug(fmt.Sprintf(":( Sensor [%s](%#02x) scanning disabled\n", sensor.Name, sensor.Number), "")
		return sensor, nil
	}

	if !sensor.EventReadingType.IsThreshold() || !sensor.SensorUnit.IsAnalog() {
		if err := c.fillSensorDiscrete(sensor); err != nil {
			return nil, fmt.Errorf("fillSensorDiscrete failed, err: %s", err)
		}
	} else {
		if err := c.fillSensorThreshold(sensor); err != nil {
			return nil, fmt.Errorf("fillSensorThreshold failed, err: %s", err)
		}
	}

	return sensor, nil
}

func (c *Client) fillSensorReading(sensor *Sensor) error {

	readingRes, err := c.GetSensorReading(sensor.Number)
	if err != nil {
		if _canSafelyIgnoredResponseError(err) {
			c.Debug(fmt.Sprintf("GetSensorReading for sensor %#02x failed but skipped", sensor.Number), err)
			return nil
		}
		return fmt.Errorf("GetSensorReading for sensor %#02x failed, err: %s", sensor.Number, err)
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
func (c *Client) fillSensorDiscrete(sensor *Sensor) error {
	statusRes, err := c.GetSensorEventStatus(sensor.Number)
	if err != nil {
		if _canSafelyIgnoredResponseError(err) {
			c.Debug(fmt.Sprintf("GetSensorEventStatus for sensor %#02x failed but skipped", sensor.Number), err)
			return nil
		}
		return fmt.Errorf("GetSensorEventStatus for sensor %#02x failed, err: %s", sensor.Number, err)
	}
	sensor.OccurredEvents = statusRes.SensorEventFlag.TrueEvents()
	return nil
}

// fillSensorThreshold retrieves and fills sensor attributes for given threshold sensor.
func (c *Client) fillSensorThreshold(sensor *Sensor) error {
	if sensor.SDRRecordType != SDRRecordTypeFullSensor {
		return nil
	}

	// If Non Linear, should update the ReadingFactors
	// see 36.2 Non-Linear Sensors
	if sensor.Threshold.LinearizationFunc.IsNonLinear() {
		factorsRes, err := c.GetSensorReadingFactors(sensor.Number, sensor.Raw)
		if err != nil {
			if _canSafelyIgnoredResponseError(err) {
				c.Debug(fmt.Sprintf("GetSensorReadingFactors for sensor %#02x failed but skipped", sensor.Number), err)
				return nil
			}
			return fmt.Errorf("GetSensorReadingFactors for sensor %#02x failed, err: %s", sensor.Number, err)
		}
		sensor.Threshold.ReadingFactors = factorsRes.ReadingFactors
	}

	thresholdRes, err := c.GetSensorThresholds(sensor.Number)
	if err != nil {
		if _canSafelyIgnoredResponseError(err) {
			c.Debug(fmt.Sprintf("GetSensorThresholds for sensor %#02x failed but skipped", sensor.Number), err)
			return nil
		}
		return fmt.Errorf("GetSensorThresholds for sensor %#02x failed, err: %s", sensor.Number, err)
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

	hysteresisRes, err := c.GetSensorHysteresis(sensor.Number)
	if err != nil {
		if _canSafelyIgnoredResponseError(err) {
			c.Debug(fmt.Sprintf("GetSensorHysteresis for sensor %#02x failed but skipped", sensor.Number), err)
			return nil
		}
		return fmt.Errorf("GetSensorHysteresis for sensor %#02x failed, err: %s", sensor.Number, err)
	}
	sensor.Threshold.PositiveHysteresisRaw = hysteresisRes.PositiveRaw
	sensor.Threshold.NegativeHysteresisRaw = hysteresisRes.NegativeRaw
	sensor.Threshold.PositiveHysteresis = sensor.ConvertSensorHysteresis(hysteresisRes.PositiveRaw)
	sensor.Threshold.NegativeHysteresis = sensor.ConvertSensorHysteresis(hysteresisRes.NegativeRaw)

	return nil
}

// If the err is a ResponseError and the completion code wrapped
// in ResponseError can be safely ignored
func _canSafelyIgnoredResponseError(err error) bool {
	if respErr, ok := err.(*ResponseError); ok {
		cc := respErr.CompletionCode()
		if cc == CompletionCodeRequestedDataNotPresent || cc == CompletionCodeIllegalCommand {
			// above completion codes CAN be ignored
			// it normally means the sensor device does not exist or the sensor device does not recognize the IPMI command
			return true
		}
	}
	return false
}

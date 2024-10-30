package ipmi

import (
	"fmt"
	"time"
)

// GetDCMIPowerReadingRequest represents a "Get Power Reading" request according
// to section 6.6.1 of the [DCMI specification v1.5].
//
// Currently, only the basic "System Power Statistics" mode is supported, not
// the extended mode.
//
// [DCMI specification v1.5]: https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf
type GetDCMIPowerReadingRequest struct {
	// TODO add support for extended mode...
}

// GetDCMIPowerReadingResponse represents a response to a [GetDCMIPowerReadingRequest].
type GetDCMIPowerReadingResponse struct {
	// Current Power in watts
	CurrentPower uint16
	// Minimum Power over sampling duration in watts
	MinimumPower uint16
	// Maximum Power over sampling duration in watts
	MaximumPower uint16
	// Average Power over sampling duration in watts
	AveragePower uint16
	// IPMI Specification based Time Stamp
	//
	// For Mode 02h (not yet supported), the time stamp specifies the end of the
	// averaging window.
	Timestamp uint32
	// Statistics reporting time period
	//
	// For Mode 01h, timeframe in milliseconds, over which the controller
	// collects statistics. For Mode 02h (not yet supported), timeframe reflects
	// the Averaging Time period in units.
	ReportingPeriod uint32
	// True if power measurements are available, false otherwise.
	PowerMeasurementActive bool
}

func (req *GetDCMIPowerReadingRequest) Pack() []byte {
	// second byte 0x01 = "basic" System Power Statistics
	return []byte{GroupExtensionDCMI, 0x01, 0x00, 0x00}
}

func (req *GetDCMIPowerReadingRequest) Command() Command {
	return CommandGetDCMIPowerReading
}

func (res *GetDCMIPowerReadingResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIPowerReadingResponse) Unpack(msg []byte) error {
	if len(msg) < 18 {
		return ErrUnpackedDataTooShortWith(len(msg), 19)
	}

	var off int

	if grpExt, _, _ := unpackUint8(msg, 0); grpExt != GroupExtensionDCMI {
		return fmt.Errorf("unexpected group extension ID in response: expected %d, found %d", GroupExtensionDCMI, grpExt)
	}

	res.CurrentPower, off, _ = unpackUint16L(msg, 1)
	res.MinimumPower, off, _ = unpackUint16L(msg, off)
	res.MaximumPower, off, _ = unpackUint16L(msg, off)
	res.AveragePower, off, _ = unpackUint16L(msg, off)
	res.Timestamp, off, _ = unpackUint32L(msg, off)
	res.ReportingPeriod, off, _ = unpackUint32L(msg, off)

	state, _, _ := unpackUint8(msg, off)
	res.PowerMeasurementActive = isBit6Set(state)

	return nil
}

func (res *GetDCMIPowerReadingResponse) Format() string {
	ts := time.Unix(int64(res.Timestamp), 0)
	return "Instantaneous power reading:                 " + fmt.Sprintf("%5d", res.CurrentPower) + " Watts\n" +
		"Minimum during sampling period:              " + fmt.Sprintf("%5d", res.MinimumPower) + " Watts\n" +
		"Maximum during sampling period:              " + fmt.Sprintf("%5d", res.MaximumPower) + " Watts\n" +
		"Average power reading over sample period:    " + fmt.Sprintf("%5d", res.CurrentPower) + " Watts\n" +
		"IPMI timestamp:                           " + ts.Format("01/02/06 15:04:05 UTC") + "\n" +
		"Sampling period:                          " + fmt.Sprintf("%08d", res.ReportingPeriod/1000) + " Seconds\n" +
		"Power reading state is:                   " + formatBool(res.PowerMeasurementActive, "activated", "deactivated")
}

// GetDCMIPowerReading sends a DCMI "Get Power Reading" command.
// See [GetDCMIPowerReadingRequest] for details.
func (c *Client) GetDCMIPowerReading() (response *GetDCMIPowerReadingResponse, err error) {
	request := &GetDCMIPowerReadingRequest{}
	response = &GetDCMIPowerReadingResponse{}
	err = c.Exchange(request, response)
	return
}

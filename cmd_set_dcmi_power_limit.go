package ipmi

import "context"

// The Set Power Limit command sets the power limit parameters on the system.
// The power limit defines a threshold which, if exceeded for a configurable amount of time,
// will trigger a system power off and/or event logging action.
//
// If the limit is already active, the Set Power Limit command may immediately change the limit that is in effect.
// However, software should always explicitly activate the limit using the Activate/Deactivate power limit
// command to ensure the setting takes effect.
//
// [DCMI specification v1.5]: 6.6.3 Set Power Limit
type SetDCMIPowerLimitRequest struct {
	ExceptionAction DCMIExceptionAction
	// Power Limit Requested in Watts
	PowerLimitRequested uint16
	// Maximum time taken to limit the power after the platform power has reached
	// the power limit before the Exception Action will be taken.
	CorrectionTimeLimitMilliSec uint32
	// Management application Statistics Sampling period in seconds
	StatisticsSamplingPeriodSec uint16
}

type SetDCMIPowerLimitResponse struct {
}

func (req *SetDCMIPowerLimitRequest) Pack() []byte {
	out := make([]byte, 15)

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ExceptionAction), out, 4)
	packUint16L(req.PowerLimitRequested, out, 5)
	packUint32L(req.CorrectionTimeLimitMilliSec, out, 7)
	packUint16L(req.StatisticsSamplingPeriodSec, out, 13)

	return out
}

func (req *SetDCMIPowerLimitRequest) Command() Command {
	return CommandSetDCMIPowerLimit
}

func (res *SetDCMIPowerLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x84: "Power Limit out of range",
		0x85: "Correction Time out of range",
		0x89: "Statistics Reporting Period out of range",
	}
}

func (res *SetDCMIPowerLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIPowerLimitResponse) Format() string {
	return ""
}

// SetDCMIPowerLimit sends a DCMI "Get Power Reading" command.
// See [SetDCMIPowerLimitRequest] for details.
func (c *Client) SetDCMIPowerLimit(ctx context.Context, request *SetDCMIPowerLimitRequest) (response *SetDCMIPowerLimitResponse, err error) {
	response = &SetDCMIPowerLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

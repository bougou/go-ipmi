package dcmi

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// The Set Power Limit command sets the power limit parameters on the system.
	// The power limit defines a threshold which, if exceeded for a configurable amount of time,
	// will trigger a system power off and/or event logging action.
	//
	// If the limit is already active, the Set Power Limit command may immediately change the limit that is in effect.
	// However, software should always explicitly activate the limit using the Activate/Deactivate power limit
	// command to ensure the setting takes effect.
	//
	// [DCMI specification v1.5]: 6.6.3 Set Power Limit
)

type SetDCMIPowerLimitRequest struct {
	ExceptionAction types.DCMIExceptionAction
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

	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(uint8(req.ExceptionAction), out, 4)
	types.PackUint16L(req.PowerLimitRequested, out, 5)
	types.PackUint32L(req.CorrectionTimeLimitMilliSec, out, 7)
	types.PackUint16L(req.StatisticsSamplingPeriodSec, out, 13)

	return out
}

func (req *SetDCMIPowerLimitRequest) Command() types.Command {
	return types.CommandSetDCMIPowerLimit
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
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIPowerLimitResponse) Format() string {
	return ""
}

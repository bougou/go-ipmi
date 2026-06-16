package dcmi

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5]: 6.6.2 Get Power Limit
type GetDCMIPowerLimitRequest struct {
}

type GetDCMIPowerLimitResponse struct {
	ExceptionAction ipmi.DCMIExceptionAction
	// Power Limit Requested in Watts
	PowerLimitRequested uint16
	// Maximum time taken to limit the power after the platform power has reached
	// the power limit before the Exception Action will be taken.
	CorrectionTimeLimitMilliSec uint32
	// Management application Statistics Sampling period in seconds
	StatisticsSamplingPeriodSec uint16
}

func (req *GetDCMIPowerLimitRequest) Pack() []byte {
	return []byte{ipmi.GroupExtensionDCMI, 0x00, 0x00}
}

func (req *GetDCMIPowerLimitRequest) Command() ipmi.Command {
	return ipmi.CommandGetDCMIPowerLimit
}

func (res *GetDCMIPowerLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "No Active Set Power Limit",
	}
}

func (res *GetDCMIPowerLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 14)
	}

	if err := ipmi.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	exceptionAction, _, _ := ipmi.UnpackUint8(msg, 3)
	res.ExceptionAction = ipmi.DCMIExceptionAction(exceptionAction)
	res.PowerLimitRequested, _, _ = ipmi.UnpackUint16L(msg, 4)
	res.CorrectionTimeLimitMilliSec, _, _ = ipmi.UnpackUint32L(msg, 6)
	res.StatisticsSamplingPeriodSec, _, _ = ipmi.UnpackUint16L(msg, 12)

	return nil
}

func (res *GetDCMIPowerLimitResponse) Format() string {
	return "" +
		fmt.Sprintf("Power limit exception action : %s\n", res.ExceptionAction.String()) +
		fmt.Sprintf("Power limit requested        : %d Watts\n", res.PowerLimitRequested) +
		fmt.Sprintf("Correction Time Limit        : %d Milliseconds\n", res.CorrectionTimeLimitMilliSec) +
		fmt.Sprintf("Statistics Sampling period   : %d Seconds\n", res.StatisticsSamplingPeriodSec)

}

package dcmi

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// [DCMI specification v1.5]: 6.6.4 Activate/Deactivate Power Limit
)

type ActivateDCMIPowerLimitRequest struct {
	Activate bool
}

type ActivateDCMIPowerLimitResponse struct {
}

func (req *ActivateDCMIPowerLimitRequest) Pack() []byte {
	activate := uint8(0x00)
	if req.Activate {
		activate = 0x01
	}

	return []byte{ipmi.GroupExtensionDCMI, activate, 0x00, 0x00}
}

func (req *ActivateDCMIPowerLimitRequest) Command() ipmi.Command {
	return ipmi.CommandActivateDCMIPowerLimit
}

func (res *ActivateDCMIPowerLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ActivateDCMIPowerLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	if err := ipmi.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *ActivateDCMIPowerLimitResponse) Format() string {
	return ""

}

// ActivateDCMIPowerLimit activate or deactivate the power limit set.
// Setting the param 'activate' to true means to activate the power limit, false means to deactivate the power limit

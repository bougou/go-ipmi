package dcmi

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// [DCMI specification v1.5]: 6.7.2 Set Thermal Limit Command
)

type SetDCMIThermalLimitRequest struct {
	EntityID       types.EntityID // Entity ID = 37h or 40h (Inlet Temperature)
	EntityInstance types.EntityInstance

	ExceptionAction_PowerOffAndLogSEL bool
	ExceptionAction_LogSELOnly        bool // ignored if ExceptionAction_PowerOffAndLogSEL is true

	TemperatureLimit uint8
	ExceptionTimeSec uint16
}

type SetDCMIThermalLimitResponse struct {
}

func (req *SetDCMIThermalLimitRequest) Pack() []byte {
	out := make([]byte, 7)
	out[0] = types.GroupExtensionDCMI
	out[1] = byte(req.EntityID)
	out[2] = byte(req.EntityInstance)

	exceptionAction := uint8(0)
	if req.ExceptionAction_PowerOffAndLogSEL {
		exceptionAction = types.SetBit6(exceptionAction)
	}
	if req.ExceptionAction_LogSELOnly {
		exceptionAction = types.SetBit5(exceptionAction)
	}
	out[3] = exceptionAction

	out[4] = req.TemperatureLimit
	types.PackUint16L(req.ExceptionTimeSec, out, 5)
	return out
}

func (req *SetDCMIThermalLimitRequest) Command() types.Command {
	return types.CommandSetDCMIThermalLimit
}

func (res *SetDCMIThermalLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIThermalLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIThermalLimitResponse) Format() string {
	return ""
}

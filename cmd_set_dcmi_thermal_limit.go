package ipmi

import "context"

// [DCMI specification v1.5]: 6.7.2 Set Thermal Limit Command
type SetDCMIThermalLimitRequest struct {
	EntityID       EntityID // Entity ID = 37h or 40h (Inlet Temperature)
	EntityInstance EntityInstance

	ExceptionAction_PowerOffAndLogSEL bool
	ExceptionAction_LogSELOnly        bool // ignored if ExceptionAction_PowerOffAndLogSEL is true

	TemperatureLimit uint8
	ExceptionTimeSec uint16
}

type SetDCMIThermalLimitResponse struct {
}

func (req *SetDCMIThermalLimitRequest) Pack() []byte {
	out := make([]byte, 7)
	out[0] = GroupExtensionDCMI
	out[1] = byte(req.EntityID)
	out[2] = byte(req.EntityInstance)

	exceptionAction := uint8(0)
	if req.ExceptionAction_PowerOffAndLogSEL {
		exceptionAction = setBit6(exceptionAction)
	}
	if req.ExceptionAction_LogSELOnly {
		exceptionAction = setBit5(exceptionAction)
	}
	out[3] = exceptionAction

	out[4] = req.TemperatureLimit
	packUint16L(req.ExceptionTimeSec, out, 5)
	return out
}

func (req *SetDCMIThermalLimitRequest) Command() Command {
	return CommandSetDCMIThermalLimit
}

func (res *SetDCMIThermalLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIThermalLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIThermalLimitResponse) Format() string {
	return ""
}

func (c *Client) SetDCMIThermalLimit(ctx context.Context, request *SetDCMIThermalLimitRequest) (response *SetDCMIThermalLimitResponse, err error) {
	response = &SetDCMIThermalLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

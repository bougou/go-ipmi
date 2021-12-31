package ipmi

type ChassisControl uint8

const (
	ChassisControlPowerDown           ChassisControl = 0
	ChassisControlPowerUp             ChassisControl = 1
	ChassisControlPowerCycle          ChassisControl = 2
	ChassisControlHardwareRest        ChassisControl = 3
	ChassisControlDiagnosticInterrupt ChassisControl = 4
	ChassisControlSoftShutdown        ChassisControl = 5
)

// 28.3 Chassis Control Command
type ChassisControlRequest struct {
	ChassisControl ChassisControl
}

type ChassisControlResponse struct {
}

func (req *ChassisControlRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(uint8(req.ChassisControl), out, 0)
	return out
}

func (req *ChassisControlRequest) Command() Command {
	return CommandChassisControl
}

func (res *ChassisControlResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ChassisControlResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ChassisControlResponse) Format() string {
	return ""
}

func (c *Client) ChassisControl(control ChassisControl) (response *ChassisControlResponse, err error) {
	request := &ChassisControlRequest{
		ChassisControl: control,
	}
	response = &ChassisControlResponse{}
	err = c.Exchange(request, response)
	return
}

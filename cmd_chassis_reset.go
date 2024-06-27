package ipmi

// 28.4 Chassis Reset Command
type ChassisResetRequest struct {
	// empty
}

type ChassisResetResponse struct {
	// empty
}

func (req *ChassisResetRequest) Pack() []byte {
	return []byte{}
}

func (req *ChassisResetRequest) Command() Command {
	return CommandChassisReset
}

func (res *ChassisResetResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ChassisResetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ChassisResetResponse) Format() string {
	return ""
}

// This command was used with early versions of the ICMB.
// It has been superseded by the Chassis Control command
// For host systems, this corresponds to a system hard reset.
func (c *Client) ChassisReset() (response *ChassisResetResponse, err error) {
	request := &ChassisResetRequest{}
	response = &ChassisResetResponse{}
	err = c.Exchange(request, response)
	return
}

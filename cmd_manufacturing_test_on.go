package ipmi

// 20.4 20.5 Manufacturing Test On Command
type ManufacturingTestOnRequest struct {
	// empty
}

type ManufacturingTestOnResponse struct {
	// empty
}

func (req *ManufacturingTestOnRequest) Command() Command {
	return CommandManufacturingTestOn
}

func (req *ManufacturingTestOnRequest) Pack() []byte {
	return []byte{}
}

func (res *ManufacturingTestOnResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ManufacturingTestOnResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ManufacturingTestOnResponse) Format() string {
	// Todo
	return ""
}

// If the device supports a "manufacturing test mode", this command is reserved to turn that mode on.
func (c *Client) ManufacturingTestOn() (response *ManufacturingTestOnResponse, err error) {
	request := &ManufacturingTestOnRequest{}
	response = &ManufacturingTestOnResponse{}
	err = c.Exchange(request, response)
	return
}

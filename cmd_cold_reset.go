package ipmi

// 20.2 Cold Reset Command
type ColdResetRequest struct {
	// empty
}

type ColdResetResponse struct {
}

func (req *ColdResetRequest) Command() Command {
	return CommandColdReset
}

func (req *ColdResetRequest) Pack() []byte {
	return []byte{}
}

func (res *ColdResetResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ColdResetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *ColdResetResponse) Format() string {
	return ""
}

func (c *Client) ColdReset() (err error) {
	request := &ColdResetRequest{}
	response := &ColdResetResponse{}
	err = c.Exchange(request, response)
	return
}

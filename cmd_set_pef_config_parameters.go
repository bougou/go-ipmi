package ipmi

// 30.3 Set PEF Configuration Parameters Command
type SetPEFConfigParametersRequest struct {
	ParamSelector uint8
	ConfigData    []byte
}

type SetPEFConfigParametersResponse struct {
	// empty
}

func (req *SetPEFConfigParametersRequest) Command() Command {
	return CommandSetPEFConfigParameters
}

func (req *SetPEFConfigParametersRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 1+len(req.ConfigData))

	// out[0] = req.ParamSelector
	packUint8(req.ParamSelector, out, 0)
	if len(req.ConfigData) > 0 {
		packBytes(req.ConfigData, out, 1)
	}
	return out
}

func (res *SetPEFConfigParametersResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetPEFConfigParametersResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",

		// (This completion code provides a way to recognize that another party has already 'claimed' the parameters)"
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",

		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetPEFConfigParametersResponse) Format() string {
	return ""
}

func (c *Client) SetPEFConfigParameters() (response *SetPEFConfigParametersResponse, err error) {
	request := &SetPEFConfigParametersRequest{}
	response = &SetPEFConfigParametersResponse{}
	err = c.Exchange(request, response)
	return
}

package ipmi

import "context"

// 30.3 Set PEF Configuration Parameters Command
type SetPEFConfigParamsRequest struct {
	ParamSelector uint8
	ParamData     []byte
}

type SetPEFConfigParamsResponse struct {
	// empty
}

func (req *SetPEFConfigParamsRequest) Command() Command {
	return CommandSetPEFConfigParams
}

func (req *SetPEFConfigParamsRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 1+len(req.ParamData))

	// out[0] = req.ParamSelector
	packUint8(uint8(req.ParamSelector), out, 0)
	if len(req.ParamData) > 0 {
		packBytes(req.ParamData, out, 1)
	}
	return out
}

func (res *SetPEFConfigParamsResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetPEFConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		// (This completion code provides a way to recognize that another party has already 'claimed' the parameters)"
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetPEFConfigParamsResponse) Format() string {
	return ""
}

// Todo
func (c *Client) SetPEFConfigParams(ctx context.Context) (response *SetPEFConfigParamsResponse, err error) {
	request := &SetPEFConfigParamsRequest{}
	response = &SetPEFConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

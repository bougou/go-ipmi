package ipmi

// 23.1 Set LAN Configuration Parameters Command
type SetLanConfigParamsRequest struct {
	ChannelNumber int8
	ParamSelector LanParamSelector
	ConfigData    []byte
}

type SetLanConfigParamsResponse struct {
	// emtpy
}

func (req *SetLanConfigParamsRequest) Pack() []byte {
	return []byte{}
}

func (req *SetLanConfigParamsRequest) Command() Command {
	return CommandSetLanConfigParams
}

func (res *SetLanConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetLanConfigParamsResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetLanConfigParamsResponse) Format() string {
	return ""
}

// Todo
func (c *Client) SetLanConfigParams() (response *SetLanConfigParamsResponse, err error) {
	request := &SetLanConfigParamsRequest{}
	response = &SetLanConfigParamsResponse{}
	err = c.Exchange(request, response)
	return
}

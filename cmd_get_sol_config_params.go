package ipmi

// 26.3 Get SOL Configuration Parameters Command
type GetSOLConfigParamsRequest struct {
	GetParameterRevisionOnly bool
	ChannelNumber            uint8
	ParameterSelector        SOLConfigParamSelector
	SetSelector              uint8
	BlockSelector            uint8
}

type GetSOLConfigParamsResponse struct {
	ParameterRevision uint8
	ParameterData     []byte
}

func (req *GetSOLConfigParamsRequest) Command() Command {
	return CommandGetSOLConfigParams
}

func (req *GetSOLConfigParamsRequest) Pack() []byte {
	out := make([]byte, 4)
	b := req.ChannelNumber
	if req.GetParameterRevisionOnly {
		b = setBit7(b)
	}

	packUint8(b, out, 0)
	packUint8(uint8(req.ParameterSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packUint8(req.BlockSelector, out, 3)
	return out
}

func (res *GetSOLConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSOLConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.ParameterRevision = msg[0]
	res.ParameterData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetSOLConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) GetSOLConfigParams(channelNumber uint8, paramSelector SOLConfigParamSelector) (response *GetSOLConfigParamsResponse, err error) {
	request := &GetSOLConfigParamsRequest{
		ChannelNumber:     channelNumber,
		ParameterSelector: paramSelector,
		SetSelector:       0x00,
		BlockSelector:     0x00,
	}
	response = &GetSOLConfigParamsResponse{}
	err = c.Exchange(request, response)
	return
}

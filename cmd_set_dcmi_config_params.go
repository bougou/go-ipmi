package ipmi

import "context"

// [DCMI specification v1.5] 6.1.2 Set DCMI Configuration Parameters
type SetDCMIConfigParamsRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
	ParamData     []byte
}

type SetDCMIConfigParamsResponse struct {
}

func (req *SetDCMIConfigParamsRequest) Pack() []byte {
	out := make([]byte, 3+len(req.ParamData))

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packBytes(req.ParamData, out, 3)

	return out

}

func (req *SetDCMIConfigParamsRequest) Command() Command {
	return CommandSetDCMIConfigParams
}

func (res *SetDCMIConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) SetDCMIConfigParams(ctx context.Context, param DCMIConfigParameter) (response *SetDCMIConfigParamsResponse, err error) {
	paramSelector, setSelector := param.DCMIConfigParameter()
	paramData := param.Pack()

	request := &SetDCMIConfigParamsRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		ParamData:     paramData,
	}
	response = &SetDCMIConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

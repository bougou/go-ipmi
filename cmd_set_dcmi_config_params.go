package ipmi

import "context"

// [DCMI specification v1.5] 6.1.2 Set DCMI Configuration Parameters
type SetDCMIConfigParamRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
	ParamData     []byte
}

type SetDCMIConfigParamResponse struct {
}

func (req *SetDCMIConfigParamRequest) Pack() []byte {
	out := make([]byte, 3+len(req.ParamData))

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packBytes(req.ParamData, out, 3)

	return out

}

func (req *SetDCMIConfigParamRequest) Command() Command {
	return CommandSetDCMIConfigParam
}

func (res *SetDCMIConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	return nil
}

func (res *SetDCMIConfigParamResponse) Format() string {
	return ""
}

func (c *Client) SetDCMIConfigParam(ctx context.Context, param DCMIConfigParameter) (response *SetDCMIConfigParamResponse, err error) {
	paramSelector, setSelector := param.DCMIConfigParameter()
	paramData := param.Pack()

	request := &SetDCMIConfigParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		ParamData:     paramData,
	}
	response = &SetDCMIConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

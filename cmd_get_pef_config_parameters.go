package ipmi

import "fmt"

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParametersRequest struct {
	// [7] - 1b = get parameter revision only. 0b = get parameter
	// [6:0] - Parameter selector
	GetRevisionOnly bool
	ParamSelector   PEFConfigParamSelector

	SetSelector   uint8 // 00h if parameter does not require a Set Selector
	BlockSelector uint8 // 00h if parameter does not require a block number

}

type PEFConfigParamSelector uint8

const (
	PEFConfigParamSelector_SetInProgress                    PEFConfigParamSelector = 0x00
	PEFConfigParamSelector_Control                          PEFConfigParamSelector = 0x01
	PEFConfigParamSelector_ActionGlobalControl              PEFConfigParamSelector = 0x02
	PEFConfigParamSelector_StartupDelay                     PEFConfigParamSelector = 0x03
	PEFConfigParamSelector_AlertStartDelay                  PEFConfigParamSelector = 0x04
	PEFConfigParamSelector_NumberOfEventFilter              PEFConfigParamSelector = 0x05
	PEFConfigParamSelector_EventFilterTable                 PEFConfigParamSelector = 0x06
	PEFConfigParamSelector_EventFilterTableData1            PEFConfigParamSelector = 0x07
	PEFConfigParamSelector_NumberOfAlertPolicyEntries       PEFConfigParamSelector = 0x08
	PEFConfigParamSelector_AlertPolicyTable                 PEFConfigParamSelector = 0x09
	PEFConfigParamSelector_SystemGUID                       PEFConfigParamSelector = 0x0a
	PEFConfigParamSelector_NumberOfAlertStrings             PEFConfigParamSelector = 0x0b
	PEFConfigParamSelector_AlertStringKeys                  PEFConfigParamSelector = 0x0c
	PEFConfigParamSelector_AlertStrings                     PEFConfigParamSelector = 0x0d
	PEFConfigParamSelector_NumberOfGroupControlTableEntries PEFConfigParamSelector = 0x0e
	PEFConfigParamSelector_GroupControlTable                PEFConfigParamSelector = 0x0f

	// 96:127
	// OEM Parameters (optional. Non-volatile or volatile as specified by OEM)
	// This range is available for special OEM configuration parameters.
	// The OEM is identified according to the Manufacturer ID field returned by the Get Device ID command.
)

type GetPEFConfigParametersResponse struct {
	// Parameter revision.
	//
	// Format:
	//  - MSN = present revision.
	//  - LSN = oldest revision parameter is backward compatible with.
	//  - 11h for parameters in this specification.
	Revision uint8

	// ConfigData data bytes are not returned when the 'get parameter revision only' bit is 1b.
	ConfigData []byte
}

func (req *GetPEFConfigParametersRequest) Command() Command {
	return CommandGetPEFConfigParameters
}

func (req *GetPEFConfigParametersRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	b0 := uint8(req.ParamSelector) & 0x3f
	if req.GetRevisionOnly {
		b0 = setBit7(b0)
	}
	packUint8(b0, out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)

	return out
}

func (res *GetPEFConfigParametersResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.Revision = msg[0]
	if len(msg) > 1 {
		res.ConfigData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParametersResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParametersResponse) Format() string {
	return fmt.Sprintf(`
Parameter Revision           : %#02x (%d)
Configuration Parameter Data : %# 02x`,
		res.Revision, res.Revision,
		res.ConfigData,
	)
}

func (c *Client) GetPEFConfigParameters(getRevisionOnly bool, paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetPEFConfigParametersResponse, err error) {
	request := &GetPEFConfigParametersRequest{
		GetRevisionOnly: getRevisionOnly,
		ParamSelector:   paramSelector,
		SetSelector:     setSelector,
		BlockSelector:   blockSelector,
	}
	response = &GetPEFConfigParametersResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) GetPEFConfigSystemUUID() (string, error) {
	res, err := c.GetPEFConfigParameters(false, PEFConfigParamSelector_SystemGUID, 0, 0)
	if err != nil {
		return "", fmt.Errorf("GetPEFConfigParameters failed, err: %s", err)
	}

	u, err := ParseGUID(res.ConfigData)
	if err != nil {
		return "", fmt.Errorf("ParseGUID failed, err: %s", err)
	}

	return u.String(), nil
}

package ipmi

import (
	"context"
	"fmt"
)

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParamsRequest struct {
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

type GetPEFConfigParamsResponse struct {
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

func (req *GetPEFConfigParamsRequest) Command() Command {
	return CommandGetPEFConfigParams
}

func (req *GetPEFConfigParamsRequest) Pack() []byte {
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

func (res *GetPEFConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.Revision = msg[0]
	if len(msg) > 1 {
		res.ConfigData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParamsResponse) Format() string {
	return fmt.Sprintf(`
Parameter Revision           : %#02x (%d)
Configuration Parameter Data : %# 02x`,
		res.Revision, res.Revision,
		res.ConfigData,
	)
}

func (c *Client) GetPEFConfigParams(ctx context.Context, getRevisionOnly bool, paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetPEFConfigParamsResponse, err error) {
	request := &GetPEFConfigParamsRequest{
		GetRevisionOnly: getRevisionOnly,
		ParamSelector:   paramSelector,
		SetSelector:     setSelector,
		BlockSelector:   blockSelector,
	}
	response = &GetPEFConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFConfigParams_SystemUUID(ctx context.Context) (param *PEFConfigParam_SystemUUID, err error) {
	res, err := c.GetPEFConfigParams(ctx, false, PEFConfigParamSelector_SystemGUID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("GetPEFConfigParams failed, err: %s", err)
	}

	param = &PEFConfigParam_SystemUUID{}
	if err := param.Unpack(res.ConfigData); err != nil {
		return nil, fmt.Errorf("unpack")
	}

	return param, nil
}

// Used to fill in the GUID field in a PET Trap.
type PEFConfigParam_SystemUUID struct {
	// [7:1] - reserved
	// [0]
	//	1b = BMC uses following value in PET Trap.
	//	0b = BMC ignores following value and uses value returned from Get System GUID command instead.
	UseGUID bool
	GUID    [16]byte
}

func (param *PEFConfigParam_SystemUUID) Unpack(configData []byte) error {
	if len(configData) < 17 {
		return ErrUnpackedDataTooShortWith(len(configData), 17)
	}

	param.UseGUID = isBit0Set(configData[0])
	param.GUID = array16(configData[1:17])
	return nil
}

func (param *PEFConfigParam_SystemUUID) Format() string {
	u, err := ParseGUID(param.GUID[:], GUIDModeSMBIOS)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}

	out := ""
	out += fmt.Sprintf("UseGUID:   %v\n", param.UseGUID)
	out += fmt.Sprintf("GUID:      %s\n", u.String())
	return out
}

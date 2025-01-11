package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5] 6.1.3 Get DCMI Configuration Parameters Command
type GetDCMIConfigParamRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
}

type GetDCMIConfigParamResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMIConfigParamRequest) Pack() []byte {
	out := make([]byte, 3)

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)

	return out
}

func (req *GetDCMIConfigParamRequest) Command() Command {
	return CommandGetDCMIConfigParam
}

func (res *GetDCMIConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.MajorVersion, _, _ = unpackUint8(msg, 1)
	res.MinorVersion, _, _ = unpackUint8(msg, 2)
	res.ParamRevision, _, _ = unpackUint8(msg, 3)
	res.ParamData, _, _ = unpackBytes(msg, 4, len(msg)-4)

	return nil
}

func (res *GetDCMIConfigParamResponse) Format() string {
	return ""
}

func (c *Client) GetDCMIConfigParam(ctx context.Context, paramSelector DCMIConfigParamSelector, setSelector uint8) (response *GetDCMIConfigParamResponse, err error) {
	request := &GetDCMIConfigParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
	}
	response = &GetDCMIConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIConfigParamFor(ctx context.Context, param DCMIConfigParameter) error {
	if isNilDCMIConfigParameter(param) {
		return nil
	}

	paramSelector, setSelector := param.DCMIConfigParameter()

	request := &GetDCMIConfigParamRequest{ParamSelector: paramSelector, SetSelector: setSelector}
	response := &GetDCMIConfigParamResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetDCMIConfigParams(ctx context.Context) (*DCMIConfigParams, error) {
	dcmiConfigParams := &DCMIConfigParams{
		ActivateDHCP:           &DCMIConfigParam_ActivateDHCP{},
		DiscoveryConfiguration: &DCMIConfigParam_DiscoveryConfiguration{},
		DHCPTiming1:            &DCMIConfigParam_DHCPTiming1{},
		DHCPTiming2:            &DCMIConfigParam_DHCPTiming2{},
		DHCPTiming3:            &DCMIConfigParam_DHCPTiming3{},
	}

	if err := c.GetDCMIConfigParamsFor(ctx, dcmiConfigParams); err != nil {
		return nil, err
	}

	return dcmiConfigParams, nil
}

func (c *Client) GetDCMIConfigParamsFor(ctx context.Context, dcmiConfigParams *DCMIConfigParams) error {
	if dcmiConfigParams == nil {
		return nil
	}

	if dcmiConfigParams.ActivateDHCP != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.ActivateDHCP); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DiscoveryConfiguration != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DiscoveryConfiguration); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming1 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming1); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming2 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming2); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming3 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming3); err != nil {
			return err
		}
	}

	return nil
}

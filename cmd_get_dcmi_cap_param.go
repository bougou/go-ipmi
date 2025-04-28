package ipmi

import (
	"context"
	"fmt"
)

// GetDCMICapParamRequest provides version information for DCMI and information about
// the mandatory and optional DCMI capabilities that are available on the particular platform.
//
// The command is session-less and can be called similar to the Get Authentication Capability command.
// This command is a bare-metal provisioning command, and the availability of features does not imply
// the features are configured.
//
// [DCMI specification v1.5] 6.1.1 Get DCMI Capabilities Info Command
type GetDCMICapParamRequest struct {
	ParamSelector DCMICapParamSelector
}

type GetDCMICapParamResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMICapParamRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, byte(req.ParamSelector)}
}

func (req *GetDCMICapParamRequest) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	req.ParamSelector = DCMICapParamSelector(msg[1])

	return nil
}

func (req *GetDCMICapParamRequest) Command() Command {
	return CommandGetDCMICapParam
}

func (res *GetDCMICapParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMICapParamResponse) Pack() []byte {
	out := make([]byte, 4+len(res.ParamData))

	out[0] = GroupExtensionDCMI
	out[1] = res.MajorVersion
	out[2] = res.MinorVersion
	out[3] = res.ParamRevision
	copy(out[4:], res.ParamData)

	return out
}

func (res *GetDCMICapParamResponse) Unpack(msg []byte) error {
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

func (res *GetDCMICapParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Major version  : %d\n", res.MajorVersion) +
		fmt.Sprintf("Minor version  : %d\n", res.MinorVersion) +
		fmt.Sprintf("Param revision : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param data     : %v\n", res.ParamData)
}

func (c *Client) GetDCMICapParam(ctx context.Context, paramSelector DCMICapParamSelector) (response *GetDCMICapParamResponse, err error) {
	request := &GetDCMICapParamRequest{ParamSelector: paramSelector}
	response = &GetDCMICapParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMICapParamFor(ctx context.Context, param DCMICapParameter) error {
	if isNilDCMICapParameter(param) {
		return nil
	}

	paramSelector := param.DCMICapParameter()

	request := &GetDCMICapParamRequest{ParamSelector: paramSelector}
	response := &GetDCMICapParamResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param data failed, err: %w", err)
	}

	return nil
}

func (c *Client) GetDCMICapParams(ctx context.Context) (*DCMICapParams, error) {
	dcmiCapParams := &DCMICapParams{
		SupportedDCMICapabilities:               &DCMICapParam_SupportedDCMICapabilities{},
		MandatoryPlatformAttributes:             &DCMICapParam_MandatoryPlatformAttributes{},
		OptionalPlatformAttributes:              &DCMICapParam_OptionalPlatformAttributes{},
		ManageabilityAccessAttributes:           &DCMICapParam_ManageabilityAccessAttributes{},
		EnhancedSystemPowerStatisticsAttributes: &DCMICapParam_EnhancedSystemPowerStatisticsAttributes{},
	}

	if err := c.GetDCMICapParamsFor(ctx, dcmiCapParams); err != nil {
		return nil, err
	}

	return dcmiCapParams, nil
}

func (c *Client) GetDCMICapParamsFor(ctx context.Context, dcmiCapParams *DCMICapParams) error {
	if dcmiCapParams == nil {
		return nil
	}

	if dcmiCapParams.SupportedDCMICapabilities != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.SupportedDCMICapabilities); err != nil {
			return err
		}
	}

	if dcmiCapParams.MandatoryPlatformAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.MandatoryPlatformAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.OptionalPlatformAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.OptionalPlatformAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.ManageabilityAccessAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.ManageabilityAccessAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.EnhancedSystemPowerStatisticsAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.EnhancedSystemPowerStatisticsAttributes); err != nil {
			return err
		}
	}

	return nil
}

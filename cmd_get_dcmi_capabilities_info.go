package ipmi

import (
	"context"
	"fmt"
)

// GetDCMICapabilitiesInfoRequest provides version information for DCMI and information about
// the mandatory and optional DCMI capabilities that are available on the particular platform.
//
// The command is session-less and can be called similar to the Get Authentication Capability command.
// This command is a bare-metal provisioning command, and the availability of features does not imply
// the features are configured.
//
// [DCMI specification v1.5] 6.1.1 Get DCMI Capabilities Info Command
type GetDCMICapabilitiesInfoRequest struct {
	ParamSelector DCMICapParamSelector
}

type GetDCMICapabilitiesInfoResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMICapabilitiesInfoRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, byte(req.ParamSelector)}
}

func (req *GetDCMICapabilitiesInfoRequest) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	req.ParamSelector = DCMICapParamSelector(msg[1])

	return nil
}

func (req *GetDCMICapabilitiesInfoRequest) Command() Command {
	return CommandGetDCMICapabilitiesInfo
}

func (res *GetDCMICapabilitiesInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMICapabilitiesInfoResponse) Pack() []byte {
	out := make([]byte, 4+len(res.ParamData))

	out[0] = GroupExtensionDCMI
	out[1] = res.MajorVersion
	out[2] = res.MinorVersion
	out[3] = res.ParamRevision
	copy(out[4:], res.ParamData)

	return out
}

func (res *GetDCMICapabilitiesInfoResponse) Unpack(msg []byte) error {
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

func (res *GetDCMICapabilitiesInfoResponse) Format() string {
	return fmt.Sprintf(`
  Major version  : %d
  Minor version  : %d
  Param revision : %d
	Param data     : %v`,
		res.MajorVersion,
		res.MinorVersion,
		res.ParamRevision,
		res.ParamData,
	)
}

func (c *Client) GetDCMICapabilitiesInfo(ctx context.Context, paramSelector DCMICapParamSelector) (response *GetDCMICapabilitiesInfoResponse, err error) {
	request := &GetDCMICapabilitiesInfoRequest{ParamSelector: paramSelector}
	response = &GetDCMICapabilitiesInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMICapabilitiesInfoFor(ctx context.Context, param DCMICapParameter) error {
	paramSelector := param.DCMICapParamSelector()

	request := &GetDCMICapabilitiesInfoRequest{ParamSelector: paramSelector}
	response := &GetDCMICapabilitiesInfoResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param data failed, err: %w", err)
	}

	return nil
}

func (c *Client) DiscoveryDCMICapabilities(ctx context.Context) (*DCMICapabilities, error) {
	dcmiCap := &DCMICapabilities{}

	{
		param := DCMICapParam_SupportedDCMICapabilities{}
		if err := c.GetDCMICapabilitiesInfoFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiCap.SupportedDCMICapabilities = param
	}

	{
		param := DCMICapParam_MandatoryPlatformAttributes{}
		if err := c.GetDCMICapabilitiesInfoFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiCap.MandatoryPlatformAttributes = param
	}

	{
		param := DCMICapParam_OptionalPlatformAttributes{}
		if err := c.GetDCMICapabilitiesInfoFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiCap.OptionalPlatformAttributes = param
	}

	{
		param := DCMICapParam_ManageabilityAccessAttributes{}
		if err := c.GetDCMICapabilitiesInfoFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiCap.ManageabilityAccessAttributes = param
	}

	{
		param := DCMICapParam_EnhancedSystemPowerStatisticsAttributes{}
		if err := c.GetDCMICapabilitiesInfoFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiCap.EnhancedSystemPowerStatisticsAttributes = param
	}

	return dcmiCap, nil
}

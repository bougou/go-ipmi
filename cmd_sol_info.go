package ipmi

import (
	"context"
	"fmt"
)

func (c *Client) SOLInfo(ctx context.Context, channelNumber uint8) (*SOLConfigParam, error) {
	solConfigParam := &SOLConfigParam{}

	params := []SOLConfigParamSelector{
		SOLConfigParamSelector_SetInProgress,
		SOLConfigParamSelector_SOLEnable,
		SOLConfigParamSelector_SOLAuthentication,
		SOLConfigParamSelector_Character,
		SOLConfigParamSelector_SOLRetry,
		SOLConfigParamSelector_NonVolatileBitRate,
		SOLConfigParamSelector_VolatileBitRate,
		SOLConfigParamSelector_PayloadChannel,
		SOLConfigParamSelector_PayloadPort,
	}

	for _, param := range params {
		res, err := c.GetSOLConfigParams(ctx, channelNumber, param)
		if err != nil {
			return nil, fmt.Errorf("GetSOLConfigParams for %d failed, err: %s", uint8(param), err)
		}

		if err = ParseSOLParamData(param, res.ParamData, solConfigParam); err != nil {
			return nil, fmt.Errorf("ParseSOLParamData failed, err: %s", err)
		}
	}

	return solConfigParam, nil
}

package ipmi

import (
	"context"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

type GetSystemInfoParamRequest struct {
	GetParamRevisionOnly bool
	ParamSelector        SystemInfoParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSystemInfoParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSystemInfoParamRequest) Pack() []byte {
	out := make([]byte, 4)

	var b uint8
	b = setOrClearBit7(b, req.GetParamRevisionOnly)
	out[0] = b

	out[1] = uint8(req.ParamSelector)
	out[2] = req.SetSelector
	out[3] = req.BlockSelector

	return out
}

func (req *GetSystemInfoParamRequest) Command() Command {
	return CommandGetSystemInfoParam
}

func (res *GetSystemInfoParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemInfoParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision, _, _ = unpackUint8(msg, 0)
	if len(msg) > 1 {
		res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSystemInfoParamResponse) Format() string {
	return fmt.Sprintf(`
        Param Revision: %d
        Param Data: %v
`, res.ParamRevision, res.ParamData)
}

func (c *Client) GetSystemInfoParam(ctx context.Context, paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) (response *GetSystemInfoParamResponse, err error) {
	request := &GetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &GetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemInfoParamFor(ctx context.Context, param SystemInfoParameter) error {
	if isNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, setSelector, blockSelector := param.SystemInfoParameter()
	response, err := c.GetSystemInfoParam(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSystemInfoParams(ctx context.Context) (*SystemInfoParams, error) {
	systemInfo := &SystemInfoParams{
		SetInProgress:          &SystemInfoParam_SetInProgress{},
		SystemFirmwareVersions: make([]*SystemInfoParam_SystemFirmwareVersion, 0),
		SystemNames:            make([]*SystemInfoParam_SystemName, 0),
		PrimaryOSNames:         make([]*SystemInfoParam_PrimaryOSName, 0),
		OSNames:                make([]*SystemInfoParam_OSName, 0),
		OSVersions:             make([]*SystemInfoParam_OSVersion, 0),
		BMCURLs:                make([]*SystemInfoParam_BMCURL, 0),
		ManagementURLs:         make([]*SystemInfoParam_ManagementURL, 0),
	}

	if err := c.GetSystemInfoParamsFor(ctx, systemInfo); err != nil {
		return nil, err
	}

	return systemInfo, nil
}

func (c *Client) GetSystemInfoParamsFor(ctx context.Context, params *SystemInfoParams) error {
	if params == nil {
		return nil
	}

	canIgnore := buildCanIgnoreFn(
		0x80, // parameter not supported
	)

	if err := c.GetSystemInfoParamFor(ctx, params.SetInProgress); canIgnore(err) != nil {
		return err
	}

	if params.SystemFirmwareVersions != nil {
		if len(params.SystemFirmwareVersions) == 0 {
			var setsCount uint8

			p := &SystemInfoParam_SystemFirmwareVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			} else {
				stringLength := uint8(p.BlockData[1])
				setsCount = stringLength/16 + 1
			}

			params.SystemFirmwareVersions = make([]*SystemInfoParam_SystemFirmwareVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_SystemFirmwareVersion{
					SetSelector: i,
				}
				params.SystemFirmwareVersions[i] = p
			}
		}

		for _, param := range params.SystemFirmwareVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.SystemNames != nil {
		if len(params.SystemNames) == 0 {
			var setsCount uint8

			p := &SystemInfoParam_SystemName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			} else {
				stringLength := uint8(p.BlockData[1])
				setsCount = stringLength/16 + 1
			}

			params.SystemNames = make([]*SystemInfoParam_SystemName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_SystemName{
					SetSelector: i,
				}
				params.SystemNames[i] = p
			}
		}

		for _, param := range params.SystemNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.PrimaryOSNames != nil {
		if len(params.PrimaryOSNames) == 0 {
			var setsCount uint8

			p := &SystemInfoParam_PrimaryOSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			} else {
				stringLength := uint8(p.BlockData[1])
				setsCount = stringLength/16 + 1
			}

			params.PrimaryOSNames = make([]*SystemInfoParam_PrimaryOSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_PrimaryOSName{
					SetSelector: i,
				}
				params.PrimaryOSNames[i] = p
			}
		}

		for _, param := range params.PrimaryOSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSNames != nil {
		if len(params.OSNames) == 0 {
			var setsCount uint8

			p := &SystemInfoParam_OSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			} else {
				stringLength := uint8(p.BlockData[1])
				setsCount = stringLength/16 + 1
			}

			params.OSNames = make([]*SystemInfoParam_OSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_OSName{
					SetSelector: i,
				}
				params.OSNames[i] = p
			}
		}

		for _, param := range params.OSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSVersions != nil {
		if len(params.OSVersions) == 0 {
			var setsCount uint8

			p := &SystemInfoParam_OSVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			} else {
				stringLength := uint8(p.BlockData[1])
				setsCount = stringLength/16 + 1
			}

			params.OSVersions = make([]*SystemInfoParam_OSVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_OSVersion{
					SetSelector: i,
				}
				params.OSVersions[i] = p
			}
		}

		for _, param := range params.OSVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.BMCURLs != nil {
		if len(params.BMCURLs) == 0 {
			p := &SystemInfoParam_BMCURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}

			//stringDataType := uint8(p.BlockData[0])
			stringLength := uint8(p.BlockData[1]) // string length 1-based
			setsCount := stringLength/16 + 1

			params.BMCURLs = make([]*SystemInfoParam_BMCURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_BMCURL{
					SetSelector: i,
				}
				params.BMCURLs[i] = p
			}
		}

		for _, param := range params.BMCURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.ManagementURLs != nil {
		if len(params.ManagementURLs) == 0 {
			p := &SystemInfoParam_ManagementURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}

			//stringDataType := uint8(p.BlockData[0])
			stringLength := uint8(p.BlockData[1]) // string length 1-based
			setsCount := stringLength/16 + 1

			params.ManagementURLs = make([]*SystemInfoParam_ManagementURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &SystemInfoParam_ManagementURL{
					SetSelector: i,
				}
				params.ManagementURLs[i] = p
			}
		}

		for _, param := range params.ManagementURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	systemInfoParams, err := c.GetSystemInfoParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetSystemInfo failed, err: %w", err)
	}
	return systemInfoParams.ToSystemInfo(), nil
}

func getSystemInfoStringMeta(params []interface{}) (s string, stringDataRaw []byte, stringDataType uint8, stringDataLength uint8) {
	if len(params) == 0 {
		return
	}

	array := make([]SystemInfoParameter, 0)
	for _, param := range params {
		v, ok := param.(SystemInfoParameter)
		if ok {
			array = append(array, v)
		}
	}

	allBlockData := make([]byte, 0)

	for _, p := range array {
		_, setSelector, _ := p.SystemInfoParameter()
		paramData := p.Pack()
		blockData := paramData[1:]
		if setSelector == 0 {
			stringDataType = blockData[0]
			stringDataLength = blockData[1]
		}
		allBlockData = append(allBlockData, blockData[:]...)
	}

	stringDataRaw = allBlockData[2 : stringDataLength+2]

	switch stringDataType {
	// 0h = ASCII+Latin1
	// 1h = UTF-8
	// 2h = UNICODE
	// all other = reserved.
	case 0x00:
		s = string(stringDataRaw)
	case 0x01:
		s = string(stringDataRaw)
	case 0x02:
		// here, suppose UTF-16
		u16 := make([]uint16, len(stringDataRaw)/2)
		for i := 0; i < len(u16); i++ {
			u16[i] = binary.BigEndian.Uint16(stringDataRaw[i*2 : i*2+2])
		}
		// Decode UTF-16 to UTF-8
		runes := utf16.Decode(u16)
		s = string(runes)
	default:
		s = string(stringDataRaw)
	}

	return
}

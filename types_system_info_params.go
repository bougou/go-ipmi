package ipmi

import (
	"fmt"
)

type SystemInfoParamSelector uint8

const (
	SystemInfoParamSelector_SetInProgress         SystemInfoParamSelector = 0
	SystemInfoParamSelector_SystemFirmwareVersion SystemInfoParamSelector = 1
	SystemInfoParamSelector_SystemName            SystemInfoParamSelector = 2
	SystemInfoParamSelector_PrimaryOSName         SystemInfoParamSelector = 3
	SystemInfoParamSelector_OSName                SystemInfoParamSelector = 4
	SystemInfoParamSelector_OSVersion             SystemInfoParamSelector = 5
	SystemInfoParamSelector_BMCURL                SystemInfoParamSelector = 6
	SystemInfoParamSelector_ManagementURL         SystemInfoParamSelector = 7
)

func (paramSelector SystemInfoParamSelector) String() string {
	m := map[SystemInfoParamSelector]string{
		SystemInfoParamSelector_SetInProgress:         "Set In Progress",
		SystemInfoParamSelector_SystemFirmwareVersion: "System Firmware Version",
		SystemInfoParamSelector_SystemName:            "System Name",
		SystemInfoParamSelector_PrimaryOSName:         "Primary OS Name",
		SystemInfoParamSelector_OSName:                "OS Name",
		SystemInfoParamSelector_OSVersion:             "OS Version",
		SystemInfoParamSelector_BMCURL:                "BMC URL",
		SystemInfoParamSelector_ManagementURL:         "Management URL",
	}
	s, ok := m[paramSelector]
	if ok {
		return s
	}
	return "Unknown"
}

type SystemInfoParameter interface {
	SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8)
	Parameter
}

var (
	_ SystemInfoParameter = (*SystemInfoParam_SetInProgress)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_SystemFirmwareVersion)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_SystemName)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_PrimaryOSName)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_OSName)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_OSVersion)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_BMCURL)(nil)
	_ SystemInfoParameter = (*SystemInfoParam_ManagementURL)(nil)
)

func isNilSystemInfoParamete(param SystemInfoParameter) bool {
	switch v := param.(type) {
	// MUSTnot put multiple types on the same case.
	case *SystemInfoParam_SetInProgress:
		return v == nil
	case *SystemInfoParam_SystemFirmwareVersion:
		return v == nil
	case *SystemInfoParam_SystemName:
		return v == nil
	case *SystemInfoParam_PrimaryOSName:
		return v == nil
	case *SystemInfoParam_OSName:
		return v == nil
	case *SystemInfoParam_OSVersion:
		return v == nil
	case *SystemInfoParam_BMCURL:
		return v == nil
	case *SystemInfoParam_ManagementURL:
		return v == nil
	default:
		return false
	}
}

type SystemInfoParams struct {
	SetInProgress          *SystemInfoParam_SetInProgress
	SystemFirmwareVersions []*SystemInfoParam_SystemFirmwareVersion
	SystemNames            []*SystemInfoParam_SystemName
	PrimaryOSNames         []*SystemInfoParam_PrimaryOSName
	OSNames                []*SystemInfoParam_OSName
	OSVersions             []*SystemInfoParam_OSVersion
	BMCURLs                []*SystemInfoParam_BMCURL
	ManagementURLs         []*SystemInfoParam_ManagementURL
}

type SystemInfo struct {
	SetInProgress         SetInProgress
	SystemFirmwareVersion string
	SystemName            string
	PrimaryOSName         string
	OSName                string
	OSVersion             string
	BMCURL                string
	ManagementURL         string
}

func (systemInfoParams *SystemInfoParams) ToSystemInfo() *SystemInfo {
	systemInfo := &SystemInfo{
		SetInProgress: systemInfoParams.SetInProgress.Value,
	}

	systemInfo.SystemFirmwareVersion, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.SystemFirmwareVersions))
	systemInfo.SystemName, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.SystemNames))
	systemInfo.PrimaryOSName, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.PrimaryOSNames))
	systemInfo.OSName, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.OSNames))
	systemInfo.OSVersion, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.OSVersions))
	systemInfo.BMCURL, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.BMCURLs))
	systemInfo.ManagementURL, _, _, _ = getSystemInfoStringMeta(convertToInterfaceSlice(systemInfoParams.ManagementURLs))

	return systemInfo
}

func (systemInfoParams *SystemInfoParams) Format() string {
	format := func(param SystemInfoParameter) string {
		if isNilSystemInfoParamete(param) {
			return ""
		}
		paramSelector, _, _ := param.SystemInfoParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%02d] %-24s : %s", paramSelector, paramSelector.String(), content)
	}

	formatArray := func(params []interface{}) string {
		if len(params) == 0 {
			return ""
		}
		out := ""
		for _, param := range params {
			v, ok := param.(SystemInfoParameter)
			if ok {
				out += format(v)
			}
		}
		s, stringDataRaw, stringDataType, stringDataLength := getSystemInfoStringMeta(params)
		out += fmt.Sprintf(`
        String Data Type   : %d
        String Data Length : %d
        String Data Raw    : %v
        String Data        : %s
`, stringDataType, stringDataLength, stringDataRaw, s,
		)

		return out
	}

	out := ""
	out += format(systemInfoParams.SetInProgress)
	out += formatArray(convertToInterfaceSlice(systemInfoParams.SystemFirmwareVersions))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.SystemNames))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.PrimaryOSNames))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.OSNames))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.OSVersions))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.BMCURLs))
	out += formatArray(convertToInterfaceSlice(systemInfoParams.ManagementURLs))

	return out
}

func (systemInfo *SystemInfo) Format() string {
	return fmt.Sprintf(`
Set In Progress         : %s
System Firmware Version : %s
System Name             : %s
Primary OS Name         : %s
OS Name                 : %s
OS Version              : %s
BVM URL                 : %s
Management URL          : %s
`,
		systemInfo.SetInProgress,
		systemInfo.SystemFirmwareVersion,
		systemInfo.SystemName,
		systemInfo.PrimaryOSName,
		systemInfo.OSName,
		systemInfo.OSVersion,
		systemInfo.BMCURL,
		systemInfo.ManagementURL,
	)
}

type SystemInfoParam_SetInProgress struct {
	Value SetInProgress
}

func (p *SystemInfoParam_SetInProgress) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_SetInProgress, 0, 0
}

func (p *SystemInfoParam_SetInProgress) Pack() []byte {
	return []byte{uint8(p.Value)}
}

func (p *SystemInfoParam_SetInProgress) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	p.Value = SetInProgress(data[0])
	return nil
}

func (p *SystemInfoParam_SetInProgress) Format() string {
	return p.Value.String()
}

type SystemInfoParam_SystemFirmwareVersion struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_SystemFirmwareVersion) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_SystemFirmwareVersion, 0, 0
}

func (p *SystemInfoParam_SystemFirmwareVersion) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_SystemFirmwareVersion) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_SystemFirmwareVersion) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_SystemName struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_SystemName) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_SystemName, 0, 0
}

func (p *SystemInfoParam_SystemName) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_SystemName) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_SystemName) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_PrimaryOSName struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_PrimaryOSName) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_PrimaryOSName, 0, 0
}

func (p *SystemInfoParam_PrimaryOSName) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_PrimaryOSName) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_PrimaryOSName) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_OSName struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_OSName) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_OSName, 0, 0
}

func (p *SystemInfoParam_OSName) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_OSName) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_OSName) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_OSVersion struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_OSVersion) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_OSVersion, 0, 0
}

func (p *SystemInfoParam_OSVersion) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_OSVersion) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_OSVersion) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_BMCURL struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_BMCURL) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_BMCURL, 0, 0
}

func (p *SystemInfoParam_BMCURL) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_BMCURL) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_BMCURL) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

type SystemInfoParam_ManagementURL struct {
	SetSelector uint8
	BlockData   [16]byte
}

func (p *SystemInfoParam_ManagementURL) SystemInfoParameter() (paramSelector SystemInfoParamSelector, setSelector uint8, blockSelector uint8) {
	return SystemInfoParamSelector_ManagementURL, 0, 0
}

func (p *SystemInfoParam_ManagementURL) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData[:], out, 1)
	return out
}

func (p *SystemInfoParam_ManagementURL) Unpack(data []byte) error {
	if len(data) < 1+len(p.BlockData) {
		return ErrUnpackedDataTooShortWith(len(data), 1+len(p.BlockData))
	}
	p.SetSelector = data[0]
	copy(p.BlockData[:], data[1:])
	return nil
}

func (p *SystemInfoParam_ManagementURL) Format() string {
	return fmt.Sprintf(`
        Set Selector : %d
        Block Data   : %02x
`, p.SetSelector, p.BlockData)
}

package ipmi

import (
	"fmt"
)

type DCMICapParameter interface {
	DCMICapParameter() DCMICapParamSelector
	Parameter
}

var (
	_ DCMICapParameter = (*DCMICapParam_SupportedDCMICapabilities)(nil)
	_ DCMICapParameter = (*DCMICapParam_MandatoryPlatformAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_OptionalPlatformAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_ManageabilityAccessAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_EnhancedSystemPowerStatisticsAttributes)(nil)
)

type DCMICapParamSelector uint8

const (
	DCMICapParamSelector_SupportedDCMICapabilities               = DCMICapParamSelector(0x01)
	DCMICapParamSelector_MandatoryPlatformAttributes             = DCMICapParamSelector(0x02)
	DCMICapParamSelector_OptionalPlatformAttributes              = DCMICapParamSelector(0x03)
	DCMICapParamSelector_ManageabilityAccessAttributes           = DCMICapParamSelector(0x04)
	DCMICapParamSelector_EnhancedSystemPowerStatisticsAttributes = DCMICapParamSelector(0x05)
)

func (dcmiCapParamSelector DCMICapParamSelector) String() string {
	m := map[DCMICapParamSelector]string{
		DCMICapParamSelector_SupportedDCMICapabilities:               "Supported DCMI capabilities",
		DCMICapParamSelector_MandatoryPlatformAttributes:             "Mandatory platform attributes",
		DCMICapParamSelector_OptionalPlatformAttributes:              "Optional platform attributes",
		DCMICapParamSelector_ManageabilityAccessAttributes:           "Manageability access attributes",
		DCMICapParamSelector_EnhancedSystemPowerStatisticsAttributes: "Enhanced system power statistics attributes",
	}
	s, ok := m[dcmiCapParamSelector]
	if ok {
		return s
	}

	return "Unknown"
}

type DCMICapabilities struct {
	SupportedDCMICapabilities               *DCMICapParam_SupportedDCMICapabilities
	MandatoryPlatformAttributes             *DCMICapParam_MandatoryPlatformAttributes
	OptionalPlatformAttributes              *DCMICapParam_OptionalPlatformAttributes
	ManageabilityAccessAttributes           *DCMICapParam_ManageabilityAccessAttributes
	EnhancedSystemPowerStatisticsAttributes *DCMICapParam_EnhancedSystemPowerStatisticsAttributes
}

func (dcmiCap *DCMICapabilities) Format() string {
	format := func(param DCMICapParameter) string {
		paramSelector := param.DCMICapParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		content += "\n"
		return fmt.Sprintf("[%02d] %-44s: %s", paramSelector, paramSelector.String(), content)
	}

	out := ""

	if dcmiCap.SupportedDCMICapabilities != nil {
		out += format(dcmiCap.SupportedDCMICapabilities)
	}

	if dcmiCap.MandatoryPlatformAttributes != nil {
		out += format(dcmiCap.MandatoryPlatformAttributes)
	}

	if dcmiCap.OptionalPlatformAttributes != nil {
		out += format(dcmiCap.OptionalPlatformAttributes)
	}

	if dcmiCap.ManageabilityAccessAttributes != nil {
		out += format(dcmiCap.ManageabilityAccessAttributes)
	}

	if dcmiCap.EnhancedSystemPowerStatisticsAttributes != nil {
		out += format(dcmiCap.EnhancedSystemPowerStatisticsAttributes)
	}

	return out
}

type DCMICapParam_SupportedDCMICapabilities struct {
	SupportPowerManagement bool
	SupportInBandKCS       bool
	SupportOutOfBandSerial bool
	SupportOutOfBandLAN    bool
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) DCMICapParameter() DCMICapParamSelector {
	return DCMICapParamSelector_SupportedDCMICapabilities
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Pack() []byte {
	return []byte{}
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Unpack(paramData []byte) error {
	if len(paramData) < 3 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	dcmiCap.SupportPowerManagement = isBit0Set(paramData[1])
	dcmiCap.SupportInBandKCS = isBit0Set(paramData[2])
	dcmiCap.SupportOutOfBandSerial = isBit1Set(paramData[2])
	dcmiCap.SupportOutOfBandLAN = isBit2Set(paramData[2])

	return nil
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Format() string {
	return fmt.Sprintf(`
        Optional platform capabilities
            Power management                  (%s)

        Manageability access capabilities
            In-band KCS channel               (%s)
            Out-of-band serial TMODE          (%s)
            Out-of-band secondary LAN channel (%s)
`,
		formatBool(dcmiCap.SupportPowerManagement, "available", "unavailable"),
		formatBool(dcmiCap.SupportInBandKCS, "available", "unavailable"),
		formatBool(dcmiCap.SupportOutOfBandSerial, "available", "unavailable"),
		formatBool(dcmiCap.SupportOutOfBandLAN, "available", "unavailable"),
	)
}

type DCMICapParam_MandatoryPlatformAttributes struct {
	SELAutoRolloverEnabled           bool
	EntireSELFlushUponRollOver       bool
	RecordLevelSELFlushUponRollOver  bool
	SELEntriesCount                  uint16 //only 12 bits, [11-0] Number of SEL entries (Maximum 4096)
	TemperatrureSamplingFrequencySec uint8
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) DCMICapParameter() DCMICapParamSelector {
	return DCMICapParamSelector_MandatoryPlatformAttributes
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Pack() []byte {
	return []byte{}
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 5 {
		return ErrUnpackedDataTooShortWith(len(paramData), 5)
	}

	b1 := paramData[1]
	dcmiCap.SELAutoRolloverEnabled = isBit7Set(b1)
	dcmiCap.EntireSELFlushUponRollOver = isBit6Set(b1)
	dcmiCap.RecordLevelSELFlushUponRollOver = isBit5Set(b1)

	b_0_1, _, _ := unpackUint16L(paramData, 0)
	dcmiCap.SELEntriesCount = b_0_1 & 0x0FFF

	dcmiCap.TemperatrureSamplingFrequencySec = paramData[4]

	return nil
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Format() string {
	return fmt.Sprintf(`
        SEL Attributes:
            SEL automatic rollover is  (%s)
            %d SEL entries

        Identification Attributes:

        Temperature Monitoring Attributes:
            Temperature sampling frequency is %d seconds
`,
		formatBool(dcmiCap.SELAutoRolloverEnabled, "enabled", "disabled"),
		dcmiCap.SELEntriesCount,
		dcmiCap.TemperatrureSamplingFrequencySec,
	)
}

type DCMICapParam_OptionalPlatformAttributes struct {
	PowerMgmtDeviceSlaveAddr         uint8
	PewerMgmtControllerChannelNumber uint8
	DeviceRevision                   uint8
}

func (dcmiCap *DCMICapParam_OptionalPlatformAttributes) DCMICapParameter() DCMICapParamSelector {
	return DCMICapParamSelector_OptionalPlatformAttributes
}

func (param *DCMICapParam_OptionalPlatformAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_OptionalPlatformAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 2 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	param.PowerMgmtDeviceSlaveAddr = paramData[0]
	param.PewerMgmtControllerChannelNumber = paramData[1] & 0xF0
	param.DeviceRevision = paramData[1] & 0x0F

	return nil
}

func (param *DCMICapParam_OptionalPlatformAttributes) Format() string {
	return fmt.Sprintf(`
        Power Management:
            Slave address of device : %#02x
            Channel number is       : %#02x %s
            Device revision is      : %d
`,
		param.PowerMgmtDeviceSlaveAddr,
		param.PewerMgmtControllerChannelNumber,
		formatBool(param.PewerMgmtControllerChannelNumber == 0, "(Primary BMC)", ""),
		param.DeviceRevision,
	)
}

type DCMICapParam_ManageabilityAccessAttributes struct {
	PrimaryLANChannelNumber   uint8
	SecondaryLANChannelNumber uint8
	SerialChannelNumber       uint8
}

func (dcmiCap *DCMICapParam_ManageabilityAccessAttributes) DCMICapParameter() DCMICapParamSelector {
	return DCMICapParamSelector_ManageabilityAccessAttributes
}

func (param *DCMICapParam_ManageabilityAccessAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_ManageabilityAccessAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 3 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	param.PrimaryLANChannelNumber = paramData[0]
	param.SecondaryLANChannelNumber = paramData[1]
	param.SerialChannelNumber = paramData[2]

	return nil
}
func (param *DCMICapParam_ManageabilityAccessAttributes) Format() string {
	return fmt.Sprintf(`
        Primary LAN channel number   : %d is %s
        Secondary LAN channel number : %d is %s
        Serial channel number        : %d is %s
`,
		param.PrimaryLANChannelNumber,
		formatBool(param.PrimaryLANChannelNumber != 0xFF, "available", "unavailable"),
		param.SecondaryLANChannelNumber,
		formatBool(param.SecondaryLANChannelNumber != 0xFF, "available", "unavailable"),
		param.SerialChannelNumber,
		formatBool(param.SerialChannelNumber != 0xFF, "available", "unavailable"),
	)
}

type DCMICapParam_EnhancedSystemPowerStatisticsAttributes struct {
	RollingCount                 uint8
	RollingAverageTimePeriodsSec []int
}

func (dcmiCap *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) DCMICapParameter() DCMICapParamSelector {
	return DCMICapParamSelector_EnhancedSystemPowerStatisticsAttributes
}

func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.RollingCount = paramData[0]

	rollingCount := int(param.RollingCount)
	if len(paramData) < 1+rollingCount {
		return ErrNotEnoughDataWith("rolling average time periods", len(paramData), 1+rollingCount)
	}

	periodsData, _, _ := unpackBytes(paramData, 1, rollingCount)
	for _, periodData := range periodsData {
		durationUnit := periodData >> 6
		durationNumber := periodData & 0x3F

		durationSec := 0
		switch durationUnit {
		case 0b00: // seconds
			durationSec = int(durationNumber)
		case 0b01: // minutes
			durationSec = int(durationNumber) * 60
		case 0b10: // hours
			durationSec = int(durationNumber) * 60 * 60
		case 0b11: // days
			durationSec = int(durationNumber) * 60 * 60 * 24
		}

		param.RollingAverageTimePeriodsSec = append(param.RollingAverageTimePeriodsSec, durationSec)
	}

	return nil
}
func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Format() string {
	return fmt.Sprintf(`
        Rolling count                : %d
        Rolling average time periods : %v
`,
		param.RollingCount,
		param.RollingAverageTimePeriodsSec,
	)
}

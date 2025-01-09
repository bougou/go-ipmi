package ipmi

import "fmt"

type DCMIConfigParameter interface {
	DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8)
	Parameter
}

var (
	_ DCMIConfigParameter = (*DCMIConfigParam_ActivateDHCP)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DiscoveryConfiguration)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming1)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming2)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming3)(nil)
)

type DCMIConfigParamSelector uint8

const (
	DCMIConfigParamSelector_ActivateDHCP           DCMIConfigParamSelector = 0x01
	DCMIConfigParamSelector_DiscoveryConfiguration DCMIConfigParamSelector = 0x02
	DCMIConfigParamSelector_DHCPTiming1            DCMIConfigParamSelector = 0x03
	DCMIConfigParamSelector_DHCPTiming2            DCMIConfigParamSelector = 0x04
	DCMIConfigParamSelector_DHCPTiming3            DCMIConfigParamSelector = 0x05
)

func (paramSelector DCMIConfigParamSelector) String() string {
	m := map[DCMIConfigParamSelector]string{
		DCMIConfigParamSelector_ActivateDHCP:           "Activate DHCP",
		DCMIConfigParamSelector_DiscoveryConfiguration: "Discovery Configuration",
		DCMIConfigParamSelector_DHCPTiming1:            "DHCP Timing1",
		DCMIConfigParamSelector_DHCPTiming2:            "DHCP Timing2",
		DCMIConfigParamSelector_DHCPTiming3:            "DHCP Timing3",
	}

	if v, ok := m[paramSelector]; ok {
		return v
	}

	return "Unknown"
}

type DCMIConfig struct {
	ActivateDHCP           *DCMIConfigParam_ActivateDHCP
	DiscoveryConfiguration *DCMIConfigParam_DiscoveryConfiguration
	DHCPTiming1            *DCMIConfigParam_DHCPTiming1
	DHCPTiming2            *DCMIConfigParam_DHCPTiming2
	DHCPTiming3            *DCMIConfigParam_DHCPTiming3
}

func (dcmiConfig *DCMIConfig) Format() string {
	out := ""

	format := func(param DCMIConfigParameter) string {
		paramSelector, _ := param.DCMIConfigParameter()

		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%02d] %-24s: %s", paramSelector, paramSelector.String(), content)
	}

	if dcmiConfig.ActivateDHCP != nil {
		out = format(dcmiConfig.ActivateDHCP)
	}

	if dcmiConfig.DiscoveryConfiguration != nil {
		out += format(dcmiConfig.DiscoveryConfiguration)
	}

	if dcmiConfig.DHCPTiming1 != nil {
		out += format(dcmiConfig.DHCPTiming1)
	}

	if dcmiConfig.DHCPTiming2 != nil {
		out += format(dcmiConfig.DHCPTiming2)
	}

	if dcmiConfig.DHCPTiming3 != nil {
		out += format(dcmiConfig.DHCPTiming3)
	}

	return out
}

type DCMIConfigParam_ActivateDHCP struct {
	// Writing 01h to this parameter will trigger DHCP protocol restart using the latest parameter
	// settings, if DHCP is enabled. This can be used to ensure that the other DHCP configuration
	// parameters take effect immediately. Otherwise, the parameters may not take effect until the
	// next time the protocol restarts or a protocol timeout or lease expiration occurs. This is not a
	// non-volatile setting. It is only used to trigger a restart of the DHCP protocol.
	//
	// This parameter shall always return 0x00 when read.
	Activate bool
}

func (param *DCMIConfigParam_ActivateDHCP) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_ActivateDHCP, 0
}

func (param *DCMIConfigParam_ActivateDHCP) Pack() []byte {
	b := uint8(0)
	if param.Activate {
		b = 1
	}

	return []byte{b}
}

func (param *DCMIConfigParam_ActivateDHCP) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.Activate = paramData[0] == 1
	return nil
}

func (param *DCMIConfigParam_ActivateDHCP) Format() string {
	return fmt.Sprintf(`%v`, param.Activate)
}

type DCMIConfigParam_DiscoveryConfiguration struct {
	RandomBackoffEnabled     bool
	IncludeDHCPOption60And43 bool
	IncludeDHCPOption12      bool
}

func (param *DCMIConfigParam_DiscoveryConfiguration) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DiscoveryConfiguration, 0
}

func (param *DCMIConfigParam_DiscoveryConfiguration) Pack() []byte {
	b := uint8(0)
	if param.RandomBackoffEnabled {
		b = setBit7(b)
	}

	if param.IncludeDHCPOption60And43 {
		b = setBit1(b)
	}

	if param.IncludeDHCPOption12 {
		b = setBit0(b)
	}

	return []byte{b}
}

func (param *DCMIConfigParam_DiscoveryConfiguration) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.RandomBackoffEnabled = isBit7Set(paramData[0])
	param.IncludeDHCPOption60And43 = isBit1Set(paramData[0])
	param.IncludeDHCPOption12 = isBit0Set(paramData[0])
	return nil
}

func (param *DCMIConfigParam_DiscoveryConfiguration) Format() string {
	return fmt.Sprintf(`
        Random Backoff Enabled          : %v
        Include DHCPOption60AndOption43 : %v (Vendor class identifier using DCMI IANA, plus Vendor class-specific Information)
        Include DHCPOption12            : %v (Management Controller ID String)
`,
		param.RandomBackoffEnabled,
		formatBool(param.IncludeDHCPOption60And43, "enabled", "disabled"),
		formatBool(param.IncludeDHCPOption12, "enabled", "disabled"),
	)
}

type DCMIConfigParam_DHCPTiming1 struct {
	// This parameter sets the amount of time between the first attempt to reach a server and the
	// second attempt to reach a server.
	//
	// Each time a message is sent the timeout interval between messages is incremented by
	// twice the current interval multiplied by a pseudo random number between zero and one
	// if random back-off is enabled, or multiplied by one if random back-off is disabled.
	//
	// The recommended default is four seconds
	InitialTimeoutIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming1) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DHCPTiming1, 0
}

func (param *DCMIConfigParam_DHCPTiming1) Pack() []byte {
	return []byte{param.InitialTimeoutIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming1) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.InitialTimeoutIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming1) Format() string {
	return fmt.Sprintf(`
        Initial timeout interval : %d seconds
`,
		param.InitialTimeoutIntervalSec,
	)
}

type DCMIConfigParam_DHCPTiming2 struct {
	// This parameter determines the amount of time that must pass between the time that the
	// client initially tries to determine its address and the time that it decides that it cannot contact
	// a server. If the last lease is expired, the client will restart the protocol after the defined retry
	// interval. The recommended default timeout is two minutes. After server contact timeout, the
	// client must wait for Server Contact Retry Interval before attempting to contact the server
	// again.
	ServerContactTimeoutIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming2) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DHCPTiming2, 0
}

func (param *DCMIConfigParam_DHCPTiming2) Pack() []byte {
	return []byte{param.ServerContactTimeoutIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming2) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.ServerContactTimeoutIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming2) Format() string {
	return fmt.Sprintf(`
        Server contact timeout interval: %d seconds
`,
		param.ServerContactTimeoutIntervalSec)
}

type DCMIConfigParam_DHCPTiming3 struct {
	// This is the period between DHCP retries after Server contact timeout interval expires. This
	// parameter determines the time that must pass after the client has determined that there is no
	// DHCP server present before it tries again to contact a DHCP server.
	//
	// The recommended default timeout is sixty-four seconds
	ServerContactRetryIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming3) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DHCPTiming3, 0
}

func (param *DCMIConfigParam_DHCPTiming3) Pack() []byte {
	return []byte{param.ServerContactRetryIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming3) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.ServerContactRetryIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming3) Format() string {
	return fmt.Sprintf(`
        Server contact retry interval: %d seconds
`,
		param.ServerContactRetryIntervalSec,
	)
}

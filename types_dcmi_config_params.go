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
		DCMIConfigParamSelector_ActivateDHCP:           "ActivateDHCP",
		DCMIConfigParamSelector_DiscoveryConfiguration: "DiscoveryConfiguration",
		DCMIConfigParamSelector_DHCPTiming1:            "DHCPTiming1",
		DCMIConfigParamSelector_DHCPTiming2:            "DHCPTiming2",
		DCMIConfigParamSelector_DHCPTiming3:            "DHCPTiming3",
	}

	if v, ok := m[paramSelector]; ok {
		return v
	}

	return "Unknown"
}

type DCMIConfig struct {
	ActivateDHCP           DCMIConfigParam_ActivateDHCP
	DiscoveryConfiguration DCMIConfigParam_DiscoveryConfiguration
	DHCPTiming1            DCMIConfigParam_DHCPTiming1
	DHCPTiming2            DCMIConfigParam_DHCPTiming2
	DHCPTiming3            DCMIConfigParam_DHCPTiming3
}

func (dcmiConfig *DCMIConfig) Format() string {
	return fmt.Sprintf(`
%s
%s
%s
%s
%s`,
		dcmiConfig.ActivateDHCP.Format(),
		dcmiConfig.DiscoveryConfiguration.Format(),
		dcmiConfig.DHCPTiming1.Format(),
		dcmiConfig.DHCPTiming2.Format(),
		dcmiConfig.DHCPTiming3.Format(),
	)
}

type DCMIConfigParam_ActivateDHCP struct {
	SetSelector uint8

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
	return DCMIConfigParamSelector_ActivateDHCP, param.SetSelector
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
	return fmt.Sprintf(`
Activate DHCP: %v
`,
		param.Activate,
	)
}

type DCMIConfigParam_DiscoveryConfiguration struct {
	SetSelector uint8

	RandomBackoffEnabled     bool
	IncludeDHCPOption60And43 bool
	IncludeDHCPOption12      bool
}

func (param *DCMIConfigParam_DiscoveryConfiguration) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DiscoveryConfiguration, param.SetSelector
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
DHCP Discovery method:
    Random Backoff Enabled:             %v
    Include DHCPOption60AndOption43:    %v (Vendor class identifier using DCMI IANA, plus Vendor class
-specific Information)
    Include DHCPOption12:               %v (Management Controller ID String)
`,
		param.RandomBackoffEnabled,
		formatBool(param.IncludeDHCPOption60And43, "enabled", "disabled"),
		formatBool(param.IncludeDHCPOption12, "enabled", "disabled"),
	)
}

type DCMIConfigParam_DHCPTiming1 struct {
	SetSelector uint8

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
	return DCMIConfigParamSelector_DHCPTiming1, param.SetSelector
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
	return fmt.Sprintf(`Initial timeout interval: %d seconds`,
		param.InitialTimeoutIntervalSec,
	)
}

type DCMIConfigParam_DHCPTiming2 struct {
	SetSelector uint8

	// This parameter determines the amount of time that must pass between the time that the
	// client initially tries to determine its address and the time that it decides that it cannot contact
	// a server. If the last lease is expired, the client will restart the protocol after the defined retry
	// interval. The recommended default timeout is two minutes. After server contact timeout, the
	// client must wait for Server Contact Retry Interval before attempting to contact the server
	// again.
	ServerContactTimeoutIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming2) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DHCPTiming2, param.SetSelector
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
	return fmt.Sprintf(`Server contact timeout interval: %d seconds`,
		param.ServerContactTimeoutIntervalSec)
}

type DCMIConfigParam_DHCPTiming3 struct {
	SetSelector uint8

	// This is the period between DHCP retries after Server contact timeout interval expires. This
	// parameter determines the time that must pass after the client has determined that there is no
	// DHCP server present before it tries again to contact a DHCP server.
	//
	// The recommended default timeout is sixty-four seconds
	ServerContactRetryIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming3) DCMIConfigParameter() (paramSelector DCMIConfigParamSelector, setSelector uint8) {
	return DCMIConfigParamSelector_DHCPTiming3, param.SetSelector
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
	return fmt.Sprintf(`Server contact retry interval: %d seconds`,
		param.ServerContactRetryIntervalSec,
	)
}

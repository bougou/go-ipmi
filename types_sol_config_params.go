package ipmi

import "fmt"

type SOLConfigParamSelector uint8

const (
	SOLConfigParamSelector_SetInProgress      SOLConfigParamSelector = 0x00
	SOLConfigParamSelector_SOLEnable          SOLConfigParamSelector = 0x01
	SOLConfigParamSelector_SOLAuthentication  SOLConfigParamSelector = 0x02
	SOLConfigParamSelector_Character          SOLConfigParamSelector = 0x03
	SOLConfigParamSelector_SOLRetry           SOLConfigParamSelector = 0x04
	SOLConfigParamSelector_NonVolatileBitRate SOLConfigParamSelector = 0x05
	SOLConfigParamSelector_VolatileBitRate    SOLConfigParamSelector = 0x06
	SOLConfigParamSelector_PayloadChannel     SOLConfigParamSelector = 0x07
	SOLConfigParamSelector_PayloadPort        SOLConfigParamSelector = 0x08
)

func (p SOLConfigParamSelector) String() string {
	m := map[SOLConfigParamSelector]string{
		SOLConfigParamSelector_SetInProgress:      "Set In Progress",
		SOLConfigParamSelector_SOLEnable:          "SOL Enable",
		SOLConfigParamSelector_SOLAuthentication:  "SOL Authentication",
		SOLConfigParamSelector_Character:          "Character",
		SOLConfigParamSelector_SOLRetry:           "SOL Retry",
		SOLConfigParamSelector_NonVolatileBitRate: "Non-Volatile Bit Rate",
		SOLConfigParamSelector_VolatileBitRate:    "Volatile Bit Rate",
		SOLConfigParamSelector_PayloadChannel:     "Payload Channel",
		SOLConfigParamSelector_PayloadPort:        "Payload Port",
	}

	s, ok := m[p]
	if ok {
		return s
	}

	return "Unknown"
}

type SOLConfigParameter interface {
	SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8)
	Parameter
}

var (
	_ SOLConfigParameter = (*SOLConfigParam_SetInProgress)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_SOLEnable)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_SOLAuthentication)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_Character)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_SOLRetry)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_NonVolatileBitRate)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_VolatileBitRate)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_PayloadChannel)(nil)
	_ SOLConfigParameter = (*SOLConfigParam_PayloadPort)(nil)
)

func isNilSOLConfigParameter(param SOLConfigParameter) bool {
	switch v := param.(type) {
	case *SOLConfigParam_SetInProgress:
		return v == nil
	case *SOLConfigParam_SOLEnable:
		return v == nil
	case *SOLConfigParam_SOLAuthentication:
		return v == nil
	case *SOLConfigParam_Character:
		return v == nil
	case *SOLConfigParam_SOLRetry:
		return v == nil
	case *SOLConfigParam_NonVolatileBitRate:
		return v == nil
	case *SOLConfigParam_VolatileBitRate:
		return v == nil
	case *SOLConfigParam_PayloadChannel:
		return v == nil
	case *SOLConfigParam_PayloadPort:
		return v == nil
	default:
		return false
	}
}

type SOLConfigParams struct {
	SetInProgress      *SOLConfigParam_SetInProgress
	SOLEnable          *SOLConfigParam_SOLEnable
	SOLAuthentication  *SOLConfigParam_SOLAuthentication
	Character          *SOLConfigParam_Character
	SOLRetry           *SOLConfigParam_SOLRetry
	NonVolatileBitRate *SOLConfigParam_NonVolatileBitRate
	VolatileBitRate    *SOLConfigParam_VolatileBitRate
	PayloadChannel     *SOLConfigParam_PayloadChannel
	PayloadPort        *SOLConfigParam_PayloadPort
}

func (p *SOLConfigParams) Format() string {
	format := func(param SOLConfigParameter) string {
		if isNilSOLConfigParameter(param) {
			return ""
		}
		paramSelector, _, _ := param.SOLConfigParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%2d] %-22s : %s", paramSelector, paramSelector.String(), content)
	}

	out := ""
	out += format(p.SetInProgress)
	out += format(p.SOLEnable)
	out += format(p.SOLAuthentication)
	out += format(p.Character)
	out += format(p.SOLRetry)
	out += format(p.NonVolatileBitRate)
	out += format(p.VolatileBitRate)
	out += format(p.PayloadChannel)
	out += format(p.PayloadPort)

	return out
}

type SOLConfigParam_SetInProgress struct {
	Value SetInProgressState
}

func (p *SOLConfigParam_SetInProgress) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_SetInProgress, 0x00, 0x00
}

func (p *SOLConfigParam_SetInProgress) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	p.Value = SetInProgressState(paramData[0])
	return nil
}

func (p *SOLConfigParam_SetInProgress) Pack() []byte {
	return []byte{byte(p.Value)}
}

func (p *SOLConfigParam_SetInProgress) Format() string {
	return p.Value.String()
}

type SOLConfigParam_SOLEnable struct {
	EnableSOLPayload bool
}

func (p *SOLConfigParam_SOLEnable) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_SOLEnable, 0x00, 0x00
}

func (p *SOLConfigParam_SOLEnable) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.EnableSOLPayload = isBit0Set(paramData[0])
	return nil
}

func (p *SOLConfigParam_SOLEnable) Pack() []byte {
	var b uint8 = 0x00
	b = setOrClearBit0(b, p.EnableSOLPayload)
	return []byte{b}
}

func (p *SOLConfigParam_SOLEnable) Format() string {
	return fmt.Sprintf("%v", p.EnableSOLPayload)
}

type SOLConfigParam_SOLAuthentication struct {
	ForceEncryption     bool
	ForceAuthentication bool
	PrivilegeLevel      uint8
}

func (p *SOLConfigParam_SOLAuthentication) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_SOLAuthentication, 0x00, 0x00
}

func (p *SOLConfigParam_SOLAuthentication) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	b := paramData[0]
	p.ForceEncryption = isBit7Set(b)
	p.ForceAuthentication = isBit6Set(b)
	p.PrivilegeLevel = b & 0x0f

	return nil
}

func (p *SOLConfigParam_SOLAuthentication) Pack() []byte {
	var b uint8 = 0x00
	if p.ForceEncryption {
		b = setBit7(b)
	}
	if p.ForceAuthentication {
		b = setBit6(b)
	}
	b |= p.PrivilegeLevel
	return []byte{b}
}

func (p *SOLConfigParam_SOLAuthentication) Format() string {
	return "" +
		fmt.Sprintf("Force Encryption     : %v\n", p.ForceEncryption) +
		fmt.Sprintf("Force Authentication : %v\n", p.ForceAuthentication) +
		fmt.Sprintf("Privilege Level      : %#02x\n", p.PrivilegeLevel)
}

type SOLConfigParam_Character struct {
	AccumulateInterval5Millis uint8
	SendThreshold             uint8
}

func (p *SOLConfigParam_Character) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_Character, 0x00, 0x00
}

func (p *SOLConfigParam_Character) Unpack(paramData []byte) error {
	if len(paramData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 byte")
	}
	p.AccumulateInterval5Millis = paramData[0]
	p.SendThreshold = paramData[1]

	return nil
}

func (p *SOLConfigParam_Character) Pack() []byte {
	return []byte{p.AccumulateInterval5Millis, p.SendThreshold}
}

func (p *SOLConfigParam_Character) Format() string {
	return "" +
		fmt.Sprintf("Accumulate Interval (ms) : %d\n", p.AccumulateInterval5Millis*5) +
		fmt.Sprintf("Send Threshold           : %d\n", p.SendThreshold)
}

type SOLConfigParam_SOLRetry struct {
	// 1-based. 0 = no retries after packet is transmitted. Packet will be
	// dropped if no ACK/NACK received by time retries expire.
	RetryCount uint8

	// 1-based. Retry Interval in 10 ms increments. Sets the time that
	// the BMC will wait before the first retry and the time between retries when sending
	// SOL packets to the remote console.
	// 00h: Retries sent back-to-back

	RetryInterval10Millis uint8
}

func (p *SOLConfigParam_SOLRetry) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_SOLRetry, 0x00, 0x00
}

func (p *SOLConfigParam_SOLRetry) Unpack(paramData []byte) error {
	if len(paramData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 byte")
	}
	p.RetryCount = paramData[0]
	p.RetryInterval10Millis = paramData[1]

	return nil
}

func (p *SOLConfigParam_SOLRetry) Pack() []byte {
	return []byte{p.RetryCount, p.RetryInterval10Millis}
}

func (p *SOLConfigParam_SOLRetry) Format() string {
	return "" +
		fmt.Sprintf("Retry Count         : %d\n", p.RetryCount) +
		fmt.Sprintf("Retry Interval (ms) : %d\n", p.RetryInterval10Millis*10)
}

type SOLConfigParam_NonVolatileBitRate struct {
	BitRate uint8
}

func (p *SOLConfigParam_NonVolatileBitRate) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_NonVolatileBitRate, 0x00, 0x00
}

func (p *SOLConfigParam_NonVolatileBitRate) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.BitRate = paramData[0]

	return nil
}

func (p *SOLConfigParam_NonVolatileBitRate) Pack() []byte {
	return []byte{p.BitRate}
}

func (p *SOLConfigParam_NonVolatileBitRate) Format() string {
	return fmt.Sprintf("%.1f kbps (%d)", bitRateKBPS(p.BitRate), p.BitRate)
}

func bitRateKBPS(bitRate uint8) float64 {
	m := map[uint8]float64{
		0x06: 9.6,
		0x07: 19.2,
		0x08: 38.4,
		0x09: 57.6,
		0x0a: 115.2,
	}
	s, ok := m[bitRate]
	if ok {
		return s
	}
	return 0
}

type SOLConfigParam_VolatileBitRate struct {
	BitRate uint8
}

func (p *SOLConfigParam_VolatileBitRate) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_VolatileBitRate, 0x00, 0x00
}

func (p *SOLConfigParam_VolatileBitRate) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.BitRate = paramData[0]

	return nil
}

func (p *SOLConfigParam_VolatileBitRate) Pack() []byte {
	return []byte{p.BitRate}
}

func (p *SOLConfigParam_VolatileBitRate) Format() string {
	return fmt.Sprintf("%.1f kbps (%d)", bitRateKBPS(p.BitRate), p.BitRate)
}

type SOLConfigParam_PayloadChannel struct {
	ChannelNumber uint8
}

func (p *SOLConfigParam_PayloadChannel) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_PayloadChannel, 0x00, 0x00
}

func (p *SOLConfigParam_PayloadChannel) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.ChannelNumber = paramData[0]

	return nil
}

func (p *SOLConfigParam_PayloadChannel) Pack() []byte {
	return []byte{p.ChannelNumber}
}

func (p *SOLConfigParam_PayloadChannel) Format() string {
	return fmt.Sprintf("%d", p.ChannelNumber)
}

type SOLConfigParam_PayloadPort struct {
	Port uint16
}

func (p *SOLConfigParam_PayloadPort) SOLConfigParameter() (paramSelector SOLConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return SOLConfigParamSelector_PayloadPort, 0x00, 0x00
}

func (p *SOLConfigParam_PayloadPort) Unpack(paramData []byte) error {
	if len(paramData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 byte")
	}

	p.Port, _, _ = unpackUint16L(paramData, 0)

	return nil
}

func (p *SOLConfigParam_PayloadPort) Pack() []byte {
	out := make([]byte, 2)
	packUint16L(p.Port, out, 0)
	return out
}

func (p *SOLConfigParam_PayloadPort) Format() string {
	return fmt.Sprintf("%d", p.Port)
}

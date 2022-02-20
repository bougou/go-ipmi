package ipmi

import "fmt"

type SOLConfigParam struct {
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

func (p *SOLConfigParam) Format() string {
	return fmt.Sprintf(`Set in progress                 : %s
Enabled                         : %s
Force Encryption                : %v
Force Authentication            : %v
Privilege Level                 : %d
Character Accumulate Level (ms) : %d
Character Send Threshold        : %d
Retry Count                     : %d
Retry Interval (ms)             : %d
Volatile Bit Rate (kbps)        : %.1f
Non-Volatile Bit Rate (kbps)    : %.1f
Payload Channel                 : %d (%#02x)
Payload Port                    : %d`,
		p.SetInProgress.Format(),
		p.SOLEnable.Format(),
		p.SOLAuthentication.ForceEncryption,
		p.SOLAuthentication.ForceAuthentication,
		p.SOLAuthentication.PrivilegeLevel,
		p.Character.AccumulateInterval5Millis*5,
		p.Character.SendThreshold,
		p.SOLRetry.RetryCount,
		p.SOLRetry.RetryInterval10Millis*10,
		p.VolatileBitRate.KBPS(),
		p.NonVolatileBitRate.KBPS(),
		*p.PayloadChannel, *p.PayloadChannel,
		*p.PayloadPort,
	)
}

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

func ParseSOLParamData(paramSelector SOLConfigParamSelector, paramData []byte, solConfigParam *SOLConfigParam) error {
	var err error

	switch paramSelector {
	case SOLConfigParamSelector_SetInProgress:
		var tmp uint8
		p := (*SOLConfigParam_SetInProgress)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.SetInProgress = p

	case SOLConfigParamSelector_SOLEnable:
		var tmp bool
		p := (*SOLConfigParam_SOLEnable)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.SOLEnable = p

	case SOLConfigParamSelector_SOLAuthentication:
		p := &SOLConfigParam_SOLAuthentication{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.SOLAuthentication = p

	case SOLConfigParamSelector_Character:
		p := &SOLConfigParam_Character{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.Character = p

	case SOLConfigParamSelector_SOLRetry:
		p := &SOLConfigParam_SOLRetry{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.SOLRetry = p

	case SOLConfigParamSelector_NonVolatileBitRate:
		var tmp uint8
		p := (*SOLConfigParam_NonVolatileBitRate)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.NonVolatileBitRate = p

	case SOLConfigParamSelector_VolatileBitRate:
		var tmp uint8
		p := (*SOLConfigParam_VolatileBitRate)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.VolatileBitRate = p

	case SOLConfigParamSelector_PayloadChannel:
		var tmp uint8
		p := (*SOLConfigParam_PayloadChannel)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.PayloadChannel = p

	case SOLConfigParamSelector_PayloadPort:
		var tmp uint16
		p := (*SOLConfigParam_PayloadPort)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		solConfigParam.PayloadPort = p
	}

	if err != nil {
		return fmt.Errorf("unpack paramData for paramSelector (%d) failed, err: %s", paramSelector, err)
	}
	return nil
}

type SOLConfigParam_SetInProgress uint8

func (p *SOLConfigParam_SetInProgress) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = SOLConfigParam_SetInProgress(paramData[0])
	return nil
}

func (p *SOLConfigParam_SetInProgress) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *SOLConfigParam_SetInProgress) Format() string {
	switch *p {
	case 0:
		return "set complete"
	case 1:
		return "set in progess"
	case 2:
		return "commit write"
	}
	return ""
}

type SOLConfigParam_SOLEnable bool

func (p *SOLConfigParam_SOLEnable) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = SOLConfigParam_SOLEnable(isBit0Set(paramData[0]))
	return nil
}

func (p *SOLConfigParam_SOLEnable) Pack() []byte {
	var b uint8 = 0x00
	if *p {
		b = setBit0(b)
	}
	return []byte{b}
}

func (p *SOLConfigParam_SOLEnable) Format() string {
	return fmt.Sprintf("%v", *p)
}

type SOLConfigParam_SOLAuthentication struct {
	ForceEncryption     bool
	ForceAuthentication bool
	PrivilegeLevel      uint8
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
	return fmt.Sprintf(`
Force Entryption     : %v
Force Authentication : %v
Privilege Level      : %#02x`, p.ForceEncryption, p.ForceAuthentication, p.PrivilegeLevel)
}

type SOLConfigParam_Character struct {
	AccumulateInterval5Millis uint8
	SendThreshold             uint8
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
	return fmt.Sprintf(`
Accumulate Interval : %d
Send Threshold      : %d`, p.AccumulateInterval5Millis, p.SendThreshold)
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
	return fmt.Sprintf(`
Retry Count    : %d
Retry Interval : %d`, p.RetryCount, p.RetryInterval10Millis)
}

type SOLConfigParam_NonVolatileBitRate uint8

func (p *SOLConfigParam_NonVolatileBitRate) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = SOLConfigParam_NonVolatileBitRate(paramData[0])

	return nil
}

func (p *SOLConfigParam_NonVolatileBitRate) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *SOLConfigParam_NonVolatileBitRate) Format() string {
	return fmt.Sprintf("%d", *p)
}

func (p *SOLConfigParam_NonVolatileBitRate) KBPS() float64 {
	m := map[SOLConfigParam_NonVolatileBitRate]float64{
		0x06: 9.6,
		0x07: 19.2,
		0x08: 38.4,
		0x09: 57.6,
		0x0a: 115.2,
	}
	s, ok := m[*p]
	if ok {
		return s
	}
	return 0
}

type SOLConfigParam_VolatileBitRate uint8

func (p *SOLConfigParam_VolatileBitRate) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = SOLConfigParam_VolatileBitRate(paramData[0])

	return nil
}

func (p *SOLConfigParam_VolatileBitRate) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *SOLConfigParam_VolatileBitRate) Format() string {
	return fmt.Sprintf("%d", *p)
}

func (p *SOLConfigParam_VolatileBitRate) KBPS() float64 {
	m := map[SOLConfigParam_VolatileBitRate]float64{
		0x06: 9.6,
		0x07: 19.2,
		0x08: 38.4,
		0x09: 57.6,
		0x0a: 115.2,
	}
	s, ok := m[*p]
	if ok {
		return s
	}
	return 0
}

type SOLConfigParam_PayloadChannel uint8

func (p *SOLConfigParam_PayloadChannel) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = SOLConfigParam_PayloadChannel(paramData[0])

	return nil
}

func (p *SOLConfigParam_PayloadChannel) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *SOLConfigParam_PayloadChannel) Format() string {
	return fmt.Sprintf("%d", *p)
}

type SOLConfigParam_PayloadPort uint16

func (p *SOLConfigParam_PayloadPort) Unpack(paramData []byte) error {
	if len(paramData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 byte")
	}

	b, _, _ := unpackUint16L(paramData, 0)
	*p = SOLConfigParam_PayloadPort(b)

	return nil
}

func (p *SOLConfigParam_PayloadPort) Pack() []byte {
	out := make([]byte, 2)
	packUint16L(uint16(*p), out, 0)
	return out
}

func (p *SOLConfigParam_PayloadPort) Format() string {
	return fmt.Sprintf("%d", *p)
}

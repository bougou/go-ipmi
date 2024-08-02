package ipmi

import (
	"fmt"
	"net"
)

// 23.2 Get LAN Configuration Parameters Command
type GetLanConfigParamsRequest struct {
	ChannelNumber uint8
	ParamSelector LanParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

type GetLanConfigParamsResponse struct {
	ParameterVersion uint8
	ConfigData       []byte
}

func (req *GetLanConfigParamsRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packUint8(req.BlockSelector, out, 3)
	return out
}

func (req *GetLanConfigParamsRequest) Command() Command {
	return CommandGetLanConfigParams
}

func (res *GetLanConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *GetLanConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ParameterVersion, _, _ = unpackUint8(msg, 0)
	res.ConfigData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetLanConfigParamsResponse) Format() string {
	out := `
ParameterVersion:      %d
ConfigData:            %v
Length of Config Data: %d
`

	return fmt.Sprintf(out, res.ParameterVersion, res.ConfigData, len(res.ConfigData))
}

func (c *Client) GetLanConfigParams(channelNumber uint8, paramSelector LanParamSelector) (response *GetLanConfigParamsResponse, err error) {
	request := &GetLanConfigParamsRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
	}
	response = &GetLanConfigParamsResponse{}
	err = c.Exchange(request, response)
	return
}

// GetLanConfig will fetch all Lan information.
func (c *Client) GetLanConfig(channelNumber uint8) (*LanConfig, error) {
	lanConfig := &LanConfig{}

	for _, lanParam := range LanParams {
		paramSelector := LanParamSelector(lanParam.Selector)

		res, err := c.GetLanConfigParams(channelNumber, paramSelector)
		if err != nil {
			resErr, ok := err.(*ResponseError)
			if !ok {
				return nil, fmt.Errorf("not ResponseError")
			}

			cc := resErr.CompletionCode()
			ccString := StrCC(res, uint8(cc))

			switch uint8(cc) {
			case 0x80,
				uint8(CompletionCodeParameterOutOfRange),
				uint8(CompletionCodeRequestDataFieldInvalid):
				c.Debugf("paramSelector (%#02x) %s, cc: (%#02x) %s\n", uint8(paramSelector), paramSelector, uint8(cc), ccString)
				continue
			default:
				// other completion codes are treated as error.
				// including 0x00 which means cc is successful, but other part failed
				return nil, fmt.Errorf("get lan config param (%s) failed, err: %s", paramSelector, err)
			}
		}

		if err := FillLanConfig(lanConfig, paramSelector, res.ConfigData); err != nil {
			return nil, fmt.Errorf("get lan config param (%s) failed, err: %s", paramSelector, err)
		}
	}

	return lanConfig, nil
}

// FillLanConfig will set the corresponding field of lanConfig according to paramSelector and paramData.
func FillLanConfig(lanConfig *LanConfig, paramSelector LanParamSelector, paramData []byte) error {
	var lanParam LanParam
	for _, v := range LanParams {
		if v.Selector == paramSelector {
			lanParam = v
			break
		}
	}

	// pre-check the length of paramData to avoid array index out-of-bound access panic.
	if uint8(len(paramData)) < lanParam.DataSize {
		return fmt.Errorf("the data for param (%s) is too short, input (%d), required (%d)", paramSelector, len(paramData), lanParam.DataSize)
	}

	switch lanParam.Selector {
	case LanParam_SetInProgress:
		lanConfig.SetInProgress = SetInProgress(paramData[0])

	case LanParam_AuthTypeSupported:
		b := paramData[0]
		lanConfig.AuthTypeSupport = AuthTypeSupport{
			OEM:      isBit5Set(b),
			Password: isBit4Set(b),
			MD5:      isBit2Set(b),
			MD2:      isBit1Set(b),
		}

	case LanParam_AuthTypeEnables:
		lanConfig.AuthTypeEnables = AuthTypeEnables{
			Callback: AuthTypeEnabled{
				OEM:      isBit5Set(paramData[0]),
				Password: isBit4Set(paramData[0]),
				MD5:      isBit2Set(paramData[0]),
				MD2:      isBit1Set(paramData[0]),
			},
			User: AuthTypeEnabled{
				OEM:      isBit5Set(paramData[1]),
				Password: isBit4Set(paramData[1]),
				MD5:      isBit2Set(paramData[1]),
				MD2:      isBit1Set(paramData[1]),
			},
			Operator: AuthTypeEnabled{
				OEM:      isBit5Set(paramData[2]),
				Password: isBit4Set(paramData[2]),
				MD5:      isBit2Set(paramData[2]),
				MD2:      isBit1Set(paramData[2]),
			},
			Admin: AuthTypeEnabled{
				OEM:      isBit5Set(paramData[3]),
				Password: isBit4Set(paramData[3]),
				MD5:      isBit2Set(paramData[3]),
				MD2:      isBit1Set(paramData[3]),
			},
			OEM: AuthTypeEnabled{
				OEM:      isBit5Set(paramData[4]),
				Password: isBit4Set(paramData[4]),
				MD5:      isBit2Set(paramData[4]),
				MD2:      isBit1Set(paramData[4]),
			},
		}

	case LanParam_IP:
		lanConfig.IP = net.IPv4(paramData[0], paramData[1], paramData[2], paramData[3])

	case LanParam_IPSource:
		lanConfig.IPSource = IPAddressSource(paramData[0])

	case LanParam_MAC:
		lanConfig.MAC = net.HardwareAddr(paramData[0:6])

	case LanParam_SubnetMask:
		lanConfig.SubnetMask = net.IPv4(paramData[0], paramData[1], paramData[2], paramData[3])

	case LanParam_IPv4HeaderParams:
		lanConfig.IPHeaderParams = IPHeaderParams{
			TTL:        paramData[0],
			Flags:      (paramData[1] & 0xc0) >> 5,
			Precedence: (paramData[2] & 0xc0) >> 5,
			TOS:        (paramData[2] & 0x1f) >> 1,
		}

	case LanParam_PrimaryRMCPPort:
		lanConfig.PrimaryRMCPPort, _, _ = unpackUint16L(paramData[0:2], 0)

	case LanParam_SecondaryRMCPPort:
		lanConfig.SecondaryRMCPPort, _, _ = unpackUint16L(paramData[0:2], 0)

	case LanParam_ARPControl:
		lanConfig.ARPControl = ARPControl{
			ARPResponseEnabled:   isBit1Set(paramData[0]),
			GratuitousARPEnabled: isBit0Set(paramData[0]),
		}

	case LanParam_GratuitousARPInterval:
		// Gratuitous ARP interval in 500 millisecond increments. 0-based.
		lanConfig.GratuitousARPIntervalMilliSec = int32(paramData[0]) * 500

	case LanParam_DefaultGatewayIP:
		lanConfig.DefaultGatewayIP = net.IP(paramData[0:4])

	case LanParam_DefaultGatewayMAC:
		lanConfig.DefaultGatewayMAC = net.HardwareAddr(paramData[0:6])

	case LanParam_BackupGatewayIP:
		lanConfig.BackupGatewayIP = net.IP(paramData[0:4])

	case LanParam_BackupGatewayMAC:
		lanConfig.BackupGatewayMAC = net.HardwareAddr(paramData[0:6])

	case LanParam_CommunityString:
		lanConfig.CommunityString = NewCommunityString(string(paramData))

	case LanParam_AlertDestinationsNumber:
		lanConfig.AlertDestinationsNumber = paramData[0]

	case LanParam_AlertDestinationType:
		lanConfig.AlertDestinationType = AlertDestinationType{
			SetSelector:             paramData[0],
			AlertSupportAcknowledge: isBit7Set(paramData[1]),
			DestinationType:         paramData[1] & 0x07,
			AlertAcknowledgeTimeout: paramData[2],
			Retries:                 paramData[3] & 0x7f,
		}

	case LanParam_AlertDestinationAddress:
		alertDestinationAddress := AlertDestinationAddress{
			SetSelector:   paramData[0],
			AddressFormat: (paramData[1] & 0xf0) >> 4,
		}

		if alertDestinationAddress.AddressFormat == 0 {

			// IPv4 and MAC
			alertDestinationAddress.IP4UseBackupGateway = isBit0Set(paramData[2])
			alertDestinationAddress.IP4IP = net.IP(paramData[3:7])
			alertDestinationAddress.IP4MAC = net.HardwareAddr(paramData[7:12])

		} else if alertDestinationAddress.AddressFormat == 1 {

			// IPv6
			if len(paramData) < 18 {
				return fmt.Errorf("the data for param (%s) is too short, input (%d), required (%d), AddressFormat is IPv6", paramSelector, len(paramData), 18)
			}
			alertDestinationAddress.IP6IP = net.IP(paramData[2:17])
		}

		lanConfig.AlertDestinationAddress = alertDestinationAddress

	case LanParam_VLANID:
		lanConfig.VLANEnabled = isBit7Set(paramData[1])
		lanConfig.VLANID = (uint16(paramData[1]&0x0f) << 8) | uint16(paramData[0])

	case LanParam_VLANPriority:
		lanConfig.VLANPriority = paramData[0] & 0x07

	case LanParam_CipherSuiteEntrySupport:
		lanConfig.RMCPCipherSuitesCount = paramData[0] & 0x1f

	case LanParam_CipherSuiteEntries:
		ids := []CipherSuiteID{}
		var count uint8 = 0
		for i, v := range paramData {
			if i == 0 {
				// first byte is Reserved
				continue
			}
			if count+1 > lanConfig.RMCPCipherSuitesCount {
				break
			}
			ids = append(ids, CipherSuiteID(v))
			count += 1
		}
		lanConfig.RMCPCipherSuiteEntries = ids

	case LanParam_CipherSuitePrivilegeLevels:
		levels := []PrivilegeLevel{}
		for i, v := range paramData {
			if i == 0 {
				// first byte is reserved
				continue
			}
			levels = append(levels, PrivilegeLevel(v&0x0f))
			levels = append(levels, PrivilegeLevel(v&0xf0>>4))
		}
		lanConfig.RMCPCipherSuitesMaxPrivLevel = levels

	case LanParam_AlertDestinationVLAN:
		lanConfig.AlertDestinationVLAN = AlertDestinationVLAN{
			SetSelector:   paramData[0],
			AddressFormat: (paramData[1] & 0xf0) >> 4,
			VLANID:        (uint16(paramData[3]&0x0f) << 8) | uint16(paramData[2]),
			CFI:           isBit4Set(paramData[3]),
			Priority:      (paramData[3] & 0xe0) >> 5,
		}

	case LanParam_BadPasswordThreshold:
		resetInterval, _, _ := unpackUint16L(paramData, 2)
		lockInterval, _, _ := unpackUint16L(paramData, 4)

		lanConfig.BadPasswordThreshold = BadPasswordThreshold{
			GenerateSessionAuditEvent:    isBit0Set(paramData[0]),
			Threshold:                    paramData[1],
			AttemptCountResetIntervalSec: uint32(resetInterval) * 10,
			UserLockoutIntervalSec:       uint32(lockInterval) * 10,
		}

	case LanParam_IP6Support:
		lanConfig.IP6Support = IP6Support{
			SupportIP6AlertDestination: isBit2Set(paramData[0]),
			CanUseBothIP4AndIP6:        isBit1Set(paramData[0]),
			CanUseIP6Only:              isBit0Set(paramData[0]),
		}

		// Todo IP6 params parse

	}

	return nil
}

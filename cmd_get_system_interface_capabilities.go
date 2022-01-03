package ipmi

// 22.9 Get System Interface Capabilities Command
type GetSystemInterfaceCapabilitiesRequest struct {
	SystemInterfaceType SystemInterfaceType
}

type GetSystemInterfaceCapabilitiesResponse struct {
	// For System Interface Type = SSIF
	TranscationSupportMask uint8
	PECSupported           bool
	SSIFVersion            uint8
	InputMessageSizeBytes  uint8
	OutputMessageSizeBytes uint8

	// For System Interface Type = KCS or SMIC
	SystemInterfaceVersion       uint8
	InputMaximumMessageSizeBytes uint8
}

type SystemInterfaceType uint8

const (
	SystemInterfaceTypeSSIF SystemInterfaceType = 0x00
	SystemInterfaceTypeKCS  SystemInterfaceType = 0x01
	SystemInterfaceTypeSMIC SystemInterfaceType = 0x02
)

func (req *GetSystemInterfaceCapabilitiesRequest) Command() Command {
	return CommandGetSystemInterfaceCapabilities
}

func (req *GetSystemInterfaceCapabilitiesRequest) Pack() []byte {
	return []byte{uint8(req.SystemInterfaceType)}
}

func (res *GetSystemInterfaceCapabilitiesResponse) Unpack(msg []byte) error {
	// at least 3 bytes
	if len(msg) < 3 {
		return ErrUnpackedDataTooShort
	}

	// For System Interface Type = SSIF:
	b, _, _ := unpackUint8(msg, 1)
	res.TranscationSupportMask = b >> 6
	res.PECSupported = isBit3Set(b)
	res.SSIFVersion = b & 0x07
	res.InputMessageSizeBytes, _, _ = unpackUint8(msg, 2)

	// For System Interface Type = KCS or SMIC
	res.SystemInterfaceVersion = b & 0x07
	res.InputMaximumMessageSizeBytes, _, _ = unpackUint8(msg, 2)

	if len(msg) >= 4 {
		res.OutputMessageSizeBytes, _, _ = unpackUint8(msg, 3)
	}
	return nil
}

func (*GetSystemInterfaceCapabilitiesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSystemInterfaceCapabilitiesResponse) Format() string {
	return ""
}

func (c *Client) GetSystemInterfaceCapabilities(interfaceType SystemInterfaceType) (response *GetSystemInterfaceCapabilitiesResponse, err error) {
	request := &GetSystemInterfaceCapabilitiesRequest{
		SystemInterfaceType: interfaceType,
	}
	response = &GetSystemInterfaceCapabilitiesResponse{}
	err = c.Exchange(request, response)
	return
}

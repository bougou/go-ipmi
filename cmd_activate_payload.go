package ipmi

import "context"

// 24.1 Activate Payload Command
type ActivatePayloadRequest struct {
	PayloadType     PayloadType
	PayloadInstance uint8

	EnableEncryption     bool
	EnableAuthentication bool

	EnableTestMode bool

	// [3:2] - Shared Serial Alert Behavior
	//         The following settings are determine what happens to serial alerts
	//         if IPMI over Serial and SOL are sharing the same baseboard serial controller.
	//   - 11b: Reserved
	//   - 10b: Serial/modem alerts succeed while SOL active.
	//   - 01b: Serial/modem alerts deferred while SOL active.
	//   - 00b: Serial/modem alerts fail while SOL active.
	SharedSerialAlertBehavior SharedSerialAlertBehavior

	// [1] - SOL startup handshake
	//  - 0b: BMC asserts CTS and DCD/DSR to baseboard upon activation.
	//  - 1b: CTS and DCD/DSR remain deasserted after activation.
	//        Remote console must send an SOL Payload packet with control field settings to assert CTS and DCD/DSR.
	//        (This enables the remote console to first alter volatile configuration settings before hardware handshake is released).
	SOLStartupHandshake bool
}

type SharedSerialAlertBehavior uint8

const (
	SharedSerialAlertBehavior_Fail    SharedSerialAlertBehavior = 0
	SharedSerialAlertBehavior_Defer   SharedSerialAlertBehavior = 1
	SharedSerialAlertBehavior_Success SharedSerialAlertBehavior = 2
)

type ActivatePayloadResponse struct {
	TestModeEnabled bool

	InboundPayloadSize  uint16
	OutboundPayloadSize uint16

	PayloadUDPPort uint16
	PayloadVLANID  uint16
}

func (req ActivatePayloadRequest) Command() Command {
	return CommandActivatePayload
}

func (req *ActivatePayloadRequest) Pack() []byte {
	out := make([]byte, 6)

	out[0] = byte(req.PayloadType)
	out[1] = req.PayloadInstance

	var b2 uint8
	b2 = (uint8(req.SharedSerialAlertBehavior) & 0x03) << 2
	b2 = setOrClearBit7(b2, req.EnableEncryption)
	b2 = setOrClearBit6(b2, req.EnableAuthentication)
	b2 = setOrClearBit5(b2, req.EnableTestMode)
	b2 = setOrClearBit1(b2, req.SOLStartupHandshake)
	out[2] = b2

	out[3] = 0
	out[4] = 0
	out[5] = 0

	return out
}

func (res *ActivatePayloadResponse) Unpack(msg []byte) error {
	if len(msg) < 12 {
		return ErrUnpackedDataTooShortWith(len(msg), 12)
	}

	res.TestModeEnabled = isBit0Set(msg[0])

	res.InboundPayloadSize, _, _ = unpackUint16L(msg, 4)
	res.OutboundPayloadSize, _, _ = unpackUint16L(msg, 6)
	res.PayloadUDPPort, _, _ = unpackUint16L(msg, 8)
	res.PayloadVLANID, _, _ = unpackUint16L(msg, 10)

	return nil
}

func (*ActivatePayloadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Payload already active on another session",
		0x81: "Payload type is disabled",
		0x82: "Payload activation limit reached",
		0x83: "Cannot activate payload with encryption",
		0x84: "Cannot activate payload without encryption",
	}
}

func (res *ActivatePayloadResponse) Format() string {
	return ""
}

func (c *Client) ActivatePayload(ctx context.Context, request *ActivatePayloadRequest) (response *ActivatePayloadResponse, err error) {
	response = &ActivatePayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

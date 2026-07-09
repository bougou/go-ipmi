package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.11 Master Write-Read Command
)

type MasterWriteReadRequest struct {
	// [7:4] channel number (Ignored when bus type = 1b)
	ChannelNumber uint8
	// [3:1] bus ID, 0-based (always 000b for public bus [bus type = 0b])
	BusID uint8
	// [0] bus type:
	// - 0b = public (e.g. IPMB or PCI Management Bus.
	//   The channel number value is used to select the target bus.)
	// - 1b = private bus (The bus ID value is used to select the target bus.)
	BusTypeIsPrivate bool

	SlaveAddress uint8

	ReadCount uint8

	// Data to write. This command should support at least 35 bytes of write data
	Data []byte
}

type MasterWriteReadResponse struct {
	// Bytes read from specified slave address.
	// This field will be absent if the read count is 0.
	// The controller terminates the I2C transaction with a STOP condition after reading the requested number of bytes.
	Data []byte
}

func (req *MasterWriteReadRequest) Command() types.Command {
	return types.CommandMasterWriteRead
}

func (req *MasterWriteReadRequest) Pack() []byte {
	out := make([]byte, 3+len(req.Data))

	var b uint8 = req.ChannelNumber << 4
	b |= (req.BusID << 1) & 0x0e
	if req.BusTypeIsPrivate {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 0)
	types.PackUint8(req.SlaveAddress, out, 1)
	types.PackUint8(req.ReadCount, out, 2)
	types.PackBytes(req.Data, out, 3)

	return out
}

func (res *MasterWriteReadResponse) Unpack(msg []byte) error {
	res.Data, _, _ = types.UnpackBytes(msg, 0, len(msg))
	return nil
}

func (*MasterWriteReadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
		0x84: "Truncated Read",
	}
}

func (res *MasterWriteReadResponse) Format() string {
	return ""
}

package transport

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 23.3 Suspend BMC ARPs Command
)

type SuspendARPsRequest struct {
	ChannelNumber        uint8
	SuspendARP           bool
	SuspendGratuitousARP bool
}

type SuspendARPsResponse struct {
	// Present state of ARP suspension

	IsARPOccurring           bool
	IsGratuitousARPOccurring bool
}

func (req *SuspendARPsRequest) Pack() []byte {
	out := make([]byte, 2)

	types.PackUint8(req.ChannelNumber, out, 0)

	var b uint8
	if req.SuspendARP {
		b = types.SetBit1(b)
	}
	if req.SuspendGratuitousARP {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 1)

	return out
}

func (req *SuspendARPsRequest) Command() types.Command {
	return types.CommandSuspendARPs
}

func (res *SuspendARPsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *SuspendARPsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := types.UnpackUint8(msg, 0)
	res.IsARPOccurring = types.IsBit1Set(b)
	res.IsGratuitousARPOccurring = types.IsBit0Set(b)
	return nil
}

func (res *SuspendARPsResponse) Format() string {
	return ""
}

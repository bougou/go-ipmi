package ipmi

// 23.3 Suspend BMC ARPs Command
type SuspendBMCARPsRequest struct {
	ChannelNumber        uint8
	SuspendARP           bool
	SuspendGratuitousARP bool
}

type SuspendBMCARPsResponse struct {
	// Present state of ARP suspension

	IsARPOccurring           bool
	IsGratuitousARPOccurring bool
}

func (req *SuspendBMCARPsRequest) Pack() []byte {
	out := make([]byte, 2)

	packUint8(req.ChannelNumber, out, 0)

	var b uint8
	if req.SuspendARP {
		b = setBit1(b)
	}
	if req.SuspendGratuitousARP {
		b = setBit0(b)
	}
	packUint8(b, out, 1)

	return out
}

func (req *SuspendBMCARPsRequest) Command() Command {
	return CommandSuspendBMCArps
}

func (res *SuspendBMCARPsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *SuspendBMCARPsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	b, _, _ := unpackUint8(msg, 0)
	res.IsARPOccurring = isBit1Set(b)
	res.IsGratuitousARPOccurring = isBit0Set(b)
	return nil
}

func (res *SuspendBMCARPsResponse) Format() string {
	return ""
}

func (c *Client) SuspendBMCArps(channelNumber uint8, suspendARP bool, suspendGratuitousARP bool) (response *SuspendBMCARPsResponse, err error) {
	request := &SuspendBMCARPsRequest{
		ChannelNumber:        channelNumber,
		SuspendARP:           suspendARP,
		SuspendGratuitousARP: suspendGratuitousARP,
	}
	response = &SuspendBMCARPsResponse{}
	err = c.Exchange(request, response)
	return
}

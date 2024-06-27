package ipmi

// 21.2 Get NetFn Support Command
type GetNetFnSupportRequest struct {
	ChannelNumber uint8
}

type GetNetFnSupportResponse struct {
	LUN3Support LUNSupport
	LUN2Support LUNSupport
	LUN1Support LUNSupport
	LUN0Support LUNSupport

	// Todo
	NetFnPairsSupport []byte
}

type LUNSupport uint8

func (l LUNSupport) String() string {
	m := map[LUNSupport]string{
		0x00: "no commands supported",
		0x01: "commands exist on LUN - no restriction",
		0x02: "commands exist on LUN - restricted",
		0x03: "reserved",
	}
	s, ok := m[l]
	if ok {
		return s
	}
	return ""
}

func (req *GetNetFnSupportRequest) Command() Command {
	return CommandGetNetFnSupport
}

func (req *GetNetFnSupportRequest) Pack() []byte {
	return []byte{req.ChannelNumber}
}

func (res *GetNetFnSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 17 {
		return ErrUnpackedDataTooShortWith(len(msg), 17)
	}
	b, _, _ := unpackUint8(msg, 0)
	res.LUN3Support = LUNSupport(b >> 6)
	res.LUN2Support = LUNSupport((b & 0x3f) >> 4)
	res.LUN1Support = LUNSupport((b & 0x0f) >> 2)
	res.LUN0Support = LUNSupport(b & 0x03)

	res.NetFnPairsSupport, _, _ = unpackBytes(msg, 1, 16)
	return nil
}

func (*GetNetFnSupportResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetNetFnSupportResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetNetFnSupport(channelNumber uint8) (response *GetNetFnSupportResponse, err error) {
	request := &GetNetFnSupportRequest{
		ChannelNumber: channelNumber,
	}
	response = &GetNetFnSupportResponse{}
	err = c.Exchange(request, response)
	return
}

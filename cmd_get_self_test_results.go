package ipmi

// 20.4 Get Self Test Results Command
type GetSelfTestResultsRequest struct {
	// empty
}

type GetSelfTestResultsResponse struct {
	Byte1 uint8
	Byte2 uint8
}

func (req *GetSelfTestResultsRequest) Command() Command {
	return CommandGetSelfTestResults
}

func (req *GetSelfTestResultsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetSelfTestResultsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSelfTestResultsResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	res.Byte1, _, _ = unpackUint8(msg, 0)
	res.Byte2, _, _ = unpackUint8(msg, 1)
	return nil
}

func (res *GetSelfTestResultsResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetSelfTestResults() (response *GetSelfTestResultsResponse, err error) {
	request := &GetSelfTestResultsRequest{}
	response = &GetSelfTestResultsResponse{}
	err = c.Exchange(request, response)
	return
}

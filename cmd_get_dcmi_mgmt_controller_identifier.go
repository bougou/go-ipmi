package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5]: 6.4.6.1 Get Management Controller Identifier String Command
type GetDCMIMgmtControllerIdentifierRequest struct {
	Offset uint8
}

type GetDCMIMgmtControllerIdentifierResponse struct {
	// ID String Length Count of non-null characters starting from offset 0 up to the first null.
	// Note: The Maximum length of the Identifier String is specified as 64 bytes including the null character,
	// therefore the range for this return is 0-63.
	IDStrLength uint8

	IDStr []byte
}

func (req *GetDCMIMgmtControllerIdentifierRequest) Pack() []byte {
	// Number of bytes to read (16 bytes maximum)
	// using the fixed (maximum) value is OK here.
	var readBytes = uint8(0x10)
	return []byte{GroupExtensionDCMI, req.Offset, readBytes}
}

func (req *GetDCMIMgmtControllerIdentifierRequest) Command() Command {
	return CommandGetDCMIMgmtControllerIdentifier
}

func (res *GetDCMIMgmtControllerIdentifierResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIMgmtControllerIdentifierResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.IDStrLength = msg[1]
	res.IDStr, _, _ = unpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetDCMIMgmtControllerIdentifierResponse) Format() string {
	return fmt.Sprintf("[%s] (returned length: %d,total length: %d)", string(res.IDStr), len(res.IDStr), res.IDStrLength)
}

// GetDCMIMgmtControllerIdentifier sends a DCMI "Get Asset Tag" command.
// See [GetDCMIMgmtControllerIdentifierRequest] for details.
func (c *Client) GetDCMIMgmtControllerIdentifier(ctx context.Context, offset uint8) (response *GetDCMIMgmtControllerIdentifierResponse, err error) {
	request := &GetDCMIMgmtControllerIdentifierRequest{Offset: offset}
	response = &GetDCMIMgmtControllerIdentifierResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIMgmtControllerIdentifierFull(ctx context.Context) ([]byte, error) {
	id := make([]byte, 0)
	offset := uint8(0)
	for {
		resp, err := c.GetDCMIMgmtControllerIdentifier(ctx, offset)
		if err != nil {
			return nil, fmt.Errorf("GetDCMIMgmtControllerIdentifier failed, err: %s", err)
		}
		id = append(id, resp.IDStr...)
		if resp.IDStrLength <= offset+uint8(len(resp.IDStr)) {
			break
		}
		offset += uint8(len(resp.IDStr))
	}

	return id, nil
}

package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5] 6.4.6.2 Set Management Controller Identifier String Command
type SetDCMIMgmtControllerIdentifierRequest struct {
	Offset     uint8
	WriteBytes uint8
	IDStr      []byte
}

type SetDCMIMgmtControllerIdentifierResponse struct {
	// Total Asset Tag Length.
	// This is the length in bytes of the stored Asset Tag after the Set operation has completed.
	// The Asset Tag length shall be set to the sum of the offset to write plus bytes to write.
	// For example, if offset to write is 32 and bytes to write is 4, the Total Asset Tag Length returned will be 36.
	TotalLength uint8
}

func (req *SetDCMIMgmtControllerIdentifierRequest) Pack() []byte {
	out := make([]byte, 3+len(req.IDStr))
	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(req.Offset, out, 1)
	packUint8(req.WriteBytes, out, 2)
	packBytes(req.IDStr, out, 3)
	return out
}

func (req *SetDCMIMgmtControllerIdentifierRequest) Command() Command {
	return CommandSetDCMIMgmtControllerIdentifier
}

func (res *SetDCMIMgmtControllerIdentifierResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalLength = msg[1]

	return nil
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Format() string {
	return fmt.Sprintf("Total Length: %d", res.TotalLength)
}

func (c *Client) SetDCMIMgmtControllerIdentifier(ctx context.Context, offset uint8, writeBytes uint8, idStr []byte) (response *SetDCMIMgmtControllerIdentifierResponse, err error) {
	request := &SetDCMIMgmtControllerIdentifierRequest{
		Offset:     offset,
		WriteBytes: writeBytes,
		IDStr:      idStr,
	}
	response = &SetDCMIMgmtControllerIdentifierResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetDCMIMgmtControllerIdentifierFull(ctx context.Context, idStr []byte) (err error) {
	if len(idStr) > 63 {
		return fmt.Errorf("the id str must be at most 63 bytes")
	}

	// make sure idStr null terminated
	if idStr[len(idStr)-1] != 0x00 {
		idStr = append(idStr, 0x00)
	}

	var offset uint8 = 0
	var writeBytes uint8 = 16
	if len(idStr) < 16 {
		writeBytes = uint8(len(idStr))
	}

	for {
		offsetEnd := offset + writeBytes
		_, err := c.SetDCMIMgmtControllerIdentifier(ctx, offset, writeBytes, idStr[offset:offsetEnd])
		if err != nil {
			return fmt.Errorf("SetDCMIMgmtControllerIdentifier failed, err: %w", err)
		}

		offset = offset + writeBytes
		if offset >= uint8(len(idStr)) {
			break
		}
		if offset+writeBytes > uint8(len(idStr)) {
			writeBytes = uint8(len(idStr)) - offset
		}
	}

	return nil
}

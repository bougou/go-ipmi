package ipmi

import (
	"context"
	"fmt"
)

// GetDCMIAssetTagRequest represents a "Get Asset Tag" request according
// to section 6.4.2 of the [DCMI specification v1.5].
//
// While the asset tag is allowed to be up to 64 bytes, each request will always
// return at most 16 bytes. The response also indicates the total length of the
// asset tag. If it is greater than 16 bytes, additional requests have to be
// performed, setting the offset accordingly.
//
// [DCMI specification v1.5]: https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf
type GetDCMIAssetTagRequest struct {
	Offset uint8
}

type GetDCMIAssetTagResponse struct {
	// At most 16 bytes of the asset tag, starting from the request's offset
	AssetTag []byte
	// The total length of the asset tag
	TotalLength uint8
}

func (req *GetDCMIAssetTagRequest) Pack() []byte {
	// Number of bytes to read (16 bytes maximum)
	// using the fixed (maximum) value is OK here.
	var readBytes = uint8(0x10)
	return []byte{GroupExtensionDCMI, req.Offset, readBytes}
}

func (req *GetDCMIAssetTagRequest) Command() Command {
	return CommandGetDCMIAssetTag
}

func (res *GetDCMIAssetTagResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIAssetTagResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalLength, _, _ = unpackUint8(msg, 1)
	if len(msg) > 2 {
		res.AssetTag, _, _ = unpackBytesMost(msg, 2, 16)
	}

	return nil
}

func (res *GetDCMIAssetTagResponse) Format() string {
	return fmt.Sprintf("%s (length: %d, total length: %d)", string(res.AssetTag), len(res.AssetTag), res.TotalLength)
}

// GetDCMIAssetTag sends a DCMI "Get Asset Tag" command.
// See [GetDCMIAssetTagRequest] for details.
func (c *Client) GetDCMIAssetTag(ctx context.Context, offset uint8) (response *GetDCMIAssetTagResponse, err error) {
	request := &GetDCMIAssetTagRequest{Offset: offset}
	response = &GetDCMIAssetTagResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

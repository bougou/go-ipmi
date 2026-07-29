package dcmi

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5] 6.4.3 Set Asset Tag Command
type SetDCMIAssetTagRequest struct {
	// The offset is relative to the first character of the Asset Tag data.
	// Offset to write (0 to 62)
	// C9h shall be returned if offset >62, offset+bytes to write >63, or bytes to write >16.
	Offset uint8
	// Number of bytes to write (16 bytes maximum).
	// The command shall set the overall length of the Asset Tag (in bytes) to
	// the value (offset to write + bytes to write). Any pre-existing Asset Tag
	// bytes at offsets past that length are automatically deleted.
	WriteBytes uint8

	// The Asset Tag shall be encoded using either UTF-8 with Byte Order Mark or ASCII+Latin1 encoding.
	// The maximum size of the Asset Tag shall be 63 bytes, including Byte Order Mark, if provided.
	AssetTag []byte
}

type SetDCMIAssetTagResponse struct {
	// Total Asset Tag Length.
	// This is the length in bytes of the stored Asset Tag after the Set operation has completed.
	// The Asset Tag length shall be set to the sum of the offset to write plus bytes to write.
	// For example, if offset to write is 32 and bytes to write is 4, the Total Asset Tag Length returned will be 36.
	TotalLength uint8
}

func (req *SetDCMIAssetTagRequest) Pack() []byte {
	out := make([]byte, 3+len(req.AssetTag))
	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(req.Offset, out, 1)
	types.PackUint8(req.WriteBytes, out, 2)
	types.PackBytes(req.AssetTag, out, 3)
	return out
}

func (req *SetDCMIAssetTagRequest) Command() types.Command {
	return types.CommandSetDCMIAssetTag
}

func (res *SetDCMIAssetTagResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIAssetTagResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalLength = msg[1]

	return nil
}

func (res *SetDCMIAssetTagResponse) Format() string {
	return fmt.Sprintf("Total Length: %d", res.TotalLength)
}

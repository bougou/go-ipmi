package dcmi

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
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
	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(req.Offset, out, 1)
	types.PackUint8(req.WriteBytes, out, 2)
	types.PackBytes(req.IDStr, out, 3)
	return out
}

func (req *SetDCMIMgmtControllerIdentifierRequest) Command() types.Command {
	return types.CommandSetDCMIMgmtControllerIdentifier
}

func (res *SetDCMIMgmtControllerIdentifierResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalLength = msg[1]

	return nil
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Format() string {
	return fmt.Sprintf("Total Length: %d", res.TotalLength)
}

// make sure idStr null terminated

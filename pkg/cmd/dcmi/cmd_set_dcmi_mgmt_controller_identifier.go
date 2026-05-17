package dcmi

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
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
	ipmi.PackUint8(ipmi.GroupExtensionDCMI, out, 0)
	ipmi.PackUint8(req.Offset, out, 1)
	ipmi.PackUint8(req.WriteBytes, out, 2)
	ipmi.PackBytes(req.IDStr, out, 3)
	return out
}

func (req *SetDCMIMgmtControllerIdentifierRequest) Command() ipmi.Command {
	return ipmi.CommandSetDCMIMgmtControllerIdentifier
}

func (res *SetDCMIMgmtControllerIdentifierResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if err := ipmi.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalLength = msg[1]

	return nil
}

func (res *SetDCMIMgmtControllerIdentifierResponse) Format() string {
	return fmt.Sprintf("Total Length: %d", res.TotalLength)
}

// make sure idStr null terminated

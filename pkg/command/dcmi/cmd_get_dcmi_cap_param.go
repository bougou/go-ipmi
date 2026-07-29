package dcmi

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// GetDCMICapParamRequest provides version information for DCMI and information about
// the mandatory and optional DCMI capabilities that are available on the particular platform.
//
// The command is session-less and can be called similar to the Get Authentication Capability command.
// This command is a bare-metal provisioning command, and the availability of features does not imply
// the features are configured.
//
// [DCMI specification v1.5] 6.1.1 Get DCMI Capabilities Info Command
type GetDCMICapParamRequest struct {
	ParamSelector types.DCMICapParamSelector
}

type GetDCMICapParamResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMICapParamRequest) Pack() []byte {
	return []byte{types.GroupExtensionDCMI, byte(req.ParamSelector)}
}

func (req *GetDCMICapParamRequest) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	req.ParamSelector = types.DCMICapParamSelector(msg[1])

	return nil
}

func (req *GetDCMICapParamRequest) Command() types.Command {
	return types.CommandGetDCMICapParam
}

func (res *GetDCMICapParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMICapParamResponse) Pack() []byte {
	out := make([]byte, 4+len(res.ParamData))

	out[0] = types.GroupExtensionDCMI
	out[1] = res.MajorVersion
	out[2] = res.MinorVersion
	out[3] = res.ParamRevision
	copy(out[4:], res.ParamData)

	return out
}

func (res *GetDCMICapParamResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.MajorVersion, _, _ = types.UnpackUint8(msg, 1)
	res.MinorVersion, _, _ = types.UnpackUint8(msg, 2)
	res.ParamRevision, _, _ = types.UnpackUint8(msg, 3)
	res.ParamData, _, _ = types.UnpackBytes(msg, 4, len(msg)-4)

	return nil
}

func (res *GetDCMICapParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Major version  : %d\n", res.MajorVersion) +
		fmt.Sprintf("Minor version  : %d\n", res.MinorVersion) +
		fmt.Sprintf("Param revision : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param data     : %v\n", res.ParamData)
}

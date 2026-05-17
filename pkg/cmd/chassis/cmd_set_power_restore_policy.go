package chassis

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 28.8 Set Power Restore Policy Command
type SetPowerRestorePolicyRequest struct {
	PowerRestorePolicy
}

type SetPowerRestorePolicyResponse struct {
	SupportPolicyAlwaysOn  bool // chassis supports always powering up after AC/mains returns
	SupportPolicyPrevious  bool // chassis supports restoring power to state that was in effect when AC/mains was lost
	SupportPolicyAlwaysOff bool // chassis supports staying powered off after AC/mains returns
}

func (req *SetPowerRestorePolicyRequest) Pack() []byte {
	out := make([]byte, 1)
	ipmi.PackUint8(uint8(req.PowerRestorePolicy), out, 0)
	return out
}

func (req *SetPowerRestorePolicyRequest) Command() ipmi.Command {
	return ipmi.CommandSetPowerRestorePolicy
}

func (res *SetPowerRestorePolicyResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetPowerRestorePolicyResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	b, _, _ := ipmi.UnpackUint8(msg, 0)
	res.SupportPolicyAlwaysOff = ipmi.IsBit0Set(b)
	res.SupportPolicyPrevious = ipmi.IsBit1Set(b)
	res.SupportPolicyAlwaysOn = ipmi.IsBit2Set((b))
	return nil
}

func (res *SetPowerRestorePolicyResponse) Format() string {
	return "" +
		fmt.Sprintf("Policy always-off : %s\n", ipmi.FormatBool(res.SupportPolicyAlwaysOff, "supported", "unsupported")) +
		fmt.Sprintf("Policy always-on  : %s\n", ipmi.FormatBool(res.SupportPolicyAlwaysOn, "supported", "unsupported")) +
		fmt.Sprintf("Policy previous   : %s\n", ipmi.FormatBool(res.SupportPolicyPrevious, "supported", "unsupported"))
}

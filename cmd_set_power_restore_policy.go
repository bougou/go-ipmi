package ipmi

import (
	"context"
	"fmt"
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
	packUint8(uint8(req.PowerRestorePolicy), out, 0)
	return out
}

func (req *SetPowerRestorePolicyRequest) Command() Command {
	return CommandSetPowerRestorePolicy
}

func (res *SetPowerRestorePolicyResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetPowerRestorePolicyResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	b, _, _ := unpackUint8(msg, 0)
	res.SupportPolicyAlwaysOff = isBit0Set(b)
	res.SupportPolicyPrevious = isBit1Set(b)
	res.SupportPolicyAlwaysOn = isBit2Set((b))
	return nil
}

func (res *SetPowerRestorePolicyResponse) Format() string {
	return fmt.Sprintf(`Policy always-off : %s"
Policy always-on  : %s
Policy previous   :"%s`,
		formatBool(res.SupportPolicyAlwaysOff, "supported", "unsupported"),
		formatBool(res.SupportPolicyAlwaysOff, "supported", "unsupported"),
		formatBool(res.SupportPolicyAlwaysOff, "supported", "unsupported"),
	)

}

func (c *Client) SetPowerRestorePolicy(ctx context.Context, policy PowerRestorePolicy) (response *SetPowerRestorePolicyResponse, err error) {
	request := &SetPowerRestorePolicyRequest{
		PowerRestorePolicy: policy,
	}
	response = &SetPowerRestorePolicyResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

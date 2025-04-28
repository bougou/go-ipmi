package ipmi

import (
	"context"
	"fmt"
)

// 30.2 Arm PEF Postpone Timer Command
type ArmPEFPostponeTimerRequest struct {
	// PEF Postpone Timeout, in seconds. 01h -> 1 second.
	//
	//  00h = disable Postpone Timer (PEF will immediately handle events, if enabled).
	//        The BMC automatically disables the timer whenever the system
	//        enters a sleep state, is powered down, or reset.
	//  01h - FDh = arm timer.
	//        Timer will automatically start counting down from given value
	//        when the last-processed event Record ID is not equal to the last
	//        received event's Record ID.
	//  FEh = Temporary PEF disable.
	//        The PEF Postpone timer does not countdown from the value.
	//        The BMC automatically re-enables PEF (if enabled in the PEF configuration parameters)
	//        and sets the PEF Postpone timeout to 00h whenever the system
	//        enters a sleep state, is powered down, or reset. Software can
	//        cancel this disable by setting this parameter to 00h or 01h-FDh.
	//  FFh = get present countdown value
	Timeout uint8
}

type ArmPEFPostponeTimerResponse struct {
	// Present timer countdown value
	PresentValue uint8
}

func (req *ArmPEFPostponeTimerRequest) Command() Command {
	return CommandArmPEFPostponeTimer
}

func (req *ArmPEFPostponeTimerRequest) Pack() []byte {
	// empty request data
	return []byte{req.Timeout}
}

func (res *ArmPEFPostponeTimerResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.PresentValue = msg[0]
	return nil
}

func (r *ArmPEFPostponeTimerResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ArmPEFPostponeTimerResponse) Format() string {
	return "" +
		fmt.Sprintf("Present timer countdown value : %d (%#02x)\n", res.PresentValue, res.PresentValue)
}

func (c *Client) ArmPEFPostponeTimer(ctx context.Context, timeout uint8) (response *ArmPEFPostponeTimerResponse, err error) {
	request := &ArmPEFPostponeTimerRequest{
		Timeout: timeout,
	}
	response = &ArmPEFPostponeTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

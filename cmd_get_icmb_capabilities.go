package ipmi

import (
	"context"
	"fmt"
)

type GetICMBCapabilitiesRequest struct {
}

type GetICMBCapabilitiesResponse struct {
	ICMBVersion             uint8
	ICMBRevision            uint8
	SupportConnectorID      bool
	SupportManagementBridge bool
	SupportPeripheralBridge bool
	SupportChassisControl   bool
}

func (req *GetICMBCapabilitiesRequest) Command() Command {
	return CommandGetICMBCapabilities
}

func (req *GetICMBCapabilitiesRequest) Pack() []byte {
	return []byte{}
}

func (res *GetICMBCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ICMBVersion = msg[0]
	res.ICMBRevision = msg[1]

	b := msg[2]
	res.SupportConnectorID = isBit0Set(b)
	res.SupportManagementBridge = isBit1Set(b)
	res.SupportPeripheralBridge = isBit2Set(b)
	res.SupportChassisControl = isBit3Set(b)

	return nil
}

func (res *GetICMBCapabilitiesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetICMBCapabilitiesResponse) Format() string {
	return fmt.Sprintf(`
        ICMB Version              : %d
				ICMB Revision             : %d
				Support Connector ID      : %v
				Support Management Bridge : %v
				Support Peripheral Bridge : %v
				Support Chassis Control: %v
`,
		res.ICMBVersion,
		res.ICMBRevision,
		res.SupportConnectorID,
		res.SupportManagementBridge,
		res.SupportPeripheralBridge,
		res.SupportChassisControl)
}

func (c *Client) GetICMBCapabilities(ctx context.Context) (response *GetICMBCapabilitiesResponse, err error) {
	request := &GetICMBCapabilitiesRequest{}
	response = &GetICMBCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

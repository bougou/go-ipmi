package client

import (
	"context"
	"fmt"
	"os"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// openipmi holds the state for the "open" (local) interface.
//
// The local interface talks to the BMC directly through the host system
// interface. The concrete transport is platform specific:
//   - linux:   the OpenIPMI kernel driver (/dev/ipmiN) via ioctl.
//   - windows: the Microsoft_IPMI WMI provider (ipmidrv.sys).
//   - others:  not supported.
type openipmi struct {
	myAddr         uint8
	msgID          int64
	targetAddr     uint8
	targetChannel  uint8
	targetIPMBAddr uint8
	transitAddr    uint8
	transitLUN     uint8

	file *os.File // /dev/ipmi0 (linux only)
}

// exchangeOpen sends a request through the local system interface and unpacks
// the response. It is platform agnostic and relies on the platform specific
// openSendRequest implementation to perform the actual transport.
func (c *Client) exchangeOpen(ctx context.Context, request ipmi.Request, response ipmi.Response) error {
	if c.openipmi.targetAddr != 0 && c.openipmi.targetAddr != c.openipmi.myAddr {

	} else {
		// otherwise use system interface
		c.Debugf("\nSending request [%s] (%#02x) to System Interface\n", request.Command().Name, request.Command().ID)
	}

	recv, err := c.openSendRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("openSendRequest failed, err: %w", err)
	}

	c.DebugBytes("recv data", recv, 16)
	c.Debugf("\n\n")

	// recv[0] is cc
	if len(recv) < 1 {
		return fmt.Errorf("recv data at least contains one completion code byte")
	}

	ccode := recv[0]
	if ccode != 0x00 {
		return ipmi.NewResponseError(
			ipmi.CompletionCode(ccode),
			fmt.Sprintf("ipmiRes CompletionCode (%#02x) is not normal: %s", ccode, ipmi.StrCC(response, ccode)),
		)
	}

	var unpackData = []byte{}
	if len(recv) > 1 {
		unpackData = recv[1:]
	}

	if err := response.Unpack(unpackData); err != nil {
		return ipmi.NewResponseError(
			ipmi.CompletionCode(recv[0]),
			fmt.Sprintf("unpack response failed, err: %s", err),
		)
	}

	c.Debug("<< Command Response", response)
	return nil
}

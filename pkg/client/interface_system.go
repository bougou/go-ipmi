package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/open"
	"github.com/bougou/go-ipmi/pkg/types"
)

// openipmi holds the state for the "open" (local) interface.
//
// The local interface talks to the BMC directly through the host system
// interface. The concrete transport is platform specific and is abstracted
// by open.Backend:
//   - linux:   open.DeviceBackend, the OpenIPMI kernel driver (/dev/ipmiN)
//     via ioctl.
//   - windows: open.COMBackend or open.PowerShellBackend, the
//     Microsoft_IPMI WMI provider (ipmidrv.sys), selected at runtime via
//     Client.openBackendPref.
//   - others:  not supported.
//
// Addressing (ipmitool open.c / Linux openipmi):
//   - myAddr is this interface's IPMB address (default BMC_SA 0x20).
//   - targetAddr / targetChannel are the session-level destination
//     (ipmitool -t / -b). Zero or equal to myAddr means system interface.
//   - Per-request CommandContext.responderAddr / responderLUN override the
//     session target for a single exchange (e.g. satellite sensor owners).
type openipmi struct {
	myAddr         uint8
	msgID          int64
	targetAddr     uint8
	targetChannel  uint8
	targetIPMBAddr uint8
	transitAddr    uint8
	transitLUN     uint8

	// backend is the connected open.Backend transport. nil until
	// ConnectOpen succeeds.
	backend open.Backend
}

// exchangeOpen sends a request through the local system interface and unpacks
// the response. It is platform agnostic; the actual transport is performed
// by the connected open.Backend.
func (c *Client) exchangeOpen(ctx context.Context, request types.Request, response types.Response) error {
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
		return types.NewResponseError(
			types.CompletionCode(ccode),
			fmt.Sprintf("ipmiRes CompletionCode (%#02x) is not normal: %s", ccode, types.StrCC(request.Command(), ccode)),
		)
	}

	var unpackData = []byte{}
	if len(recv) > 1 {
		unpackData = recv[1:]
	}

	if err := response.Unpack(unpackData); err != nil {
		return types.NewResponseError(
			types.CompletionCode(recv[0]),
			fmt.Sprintf("unpack response failed, err: %s", err),
		)
	}

	c.Debug("<< Command Response", response)
	return nil
}

// openDestination resolves the effective Open Interface destination for one
// request. CommandContext responder fields override the session-level
// openipmi.targetAddr / targetChannel; missing values fall back to BMC_SA
// and LUN 0.
func (c *Client) openDestination(ctx context.Context) (targetAddr, channel, lun uint8) {
	targetAddr = c.openipmi.targetAddr
	if targetAddr == 0 {
		targetAddr = types.BMC_SA
	}
	channel = c.openipmi.targetChannel
	lun = 0

	commandContext := GetCommandContext(ctx)
	if commandContext != nil {
		c.Debug("Got CommandContext:", commandContext)
		if commandContext.responderAddr != nil {
			targetAddr = *commandContext.responderAddr
		}
		if commandContext.responderLUN != nil {
			lun = *commandContext.responderLUN
		}
	}
	return targetAddr, channel, lun
}

// buildOpenRequest converts a high-level command Request into the
// transport-neutral open.Request consumed by all Open Interface backends.
func (c *Client) buildOpenRequest(ctx context.Context, request types.Request) *open.Request {
	cmdData := request.Pack()
	c.DebugBytes("cmd data", cmdData, 16)

	targetAddr, channel, lun := c.openDestination(ctx)
	req := &open.Request{
		NetFn:         uint8(request.Command().NetFn),
		Cmd:           uint8(request.Command().ID),
		LUN:           lun,
		Data:          cmdData,
		TargetAddr:    targetAddr,
		TargetChannel: channel,
		MyAddr:        c.openipmi.myAddr,
	}

	if req.UsesIPMB() {
		c.Debugf("\nSending request [%s] (%#02x) to IPMB target @ %#02x ch=%d lun=%d (from %#02x)\n",
			request.Command().Name, request.Command().ID, targetAddr, channel, lun, c.openipmi.myAddr)
	} else {
		c.Debugf("\nSending request [%s] (%#02x) to System Interface (lun=%d)\n",
			request.Command().Name, request.Command().ID, lun)
	}

	c.Debug("open.Request", req)
	return req
}

// closeOpen releases the connected Open Interface backend. It is a no-op
// when the interface was never connected.
func (c *Client) closeOpen(ctx context.Context) error {
	if c.openipmi == nil || c.openipmi.backend == nil {
		return nil
	}
	return c.openipmi.backend.Close(ctx)
}

// openSendRequest converts the high-level request to an open.Request and
// sends it through the connected backend, returning the raw "cc + payload"
// response bytes.
func (c *Client) openSendRequest(ctx context.Context, request types.Request) ([]byte, error) {
	if c.openipmi == nil || c.openipmi.backend == nil {
		return nil, fmt.Errorf("Open Interface is not connected; call Connect first")
	}
	return c.openipmi.backend.Send(ctx, c.buildOpenRequest(ctx, request), c.timeout)
}

//go:build linux
// +build linux

package client

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"unsafe"

	"github.com/bougou/go-ipmi/open"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// ConnectOpen try to initialize the client by open the device of linux ipmi driver.
func (c *Client) ConnectOpen(ctx context.Context, devnum int32) error {
	c.Debugf("Using ipmi device %d\n", devnum)

	// try the following devices
	var (
		ipmiDev1 string = fmt.Sprintf("/dev/ipmi%d", devnum)
		ipmiDev2 string = fmt.Sprintf("/dev/ipmi/%d", devnum)
		ipmiDev3 string = fmt.Sprintf("/dev/ipmidev/%d", devnum)
	)

	var file *os.File
	var tryOpenFile = func(ipmiDev string) {
		if file != nil {
			return
		}
		if f, err := os.OpenFile(ipmiDev, os.O_RDWR, 0); err != nil {
			c.Debugf("can not open ipmi dev (%s), err: %w\n", ipmiDev, err)
		} else {
			file = f
			c.Debugf("opened ipmi dev file: %v\n", ipmiDev)
		}
	}
	tryOpenFile(ipmiDev1)
	tryOpenFile(ipmiDev2)
	tryOpenFile(ipmiDev3)

	if file == nil {
		return fmt.Errorf("ipmi dev file not opened")
	}

	c.Debugf("opened ipmi dev file: %v, descriptor is: %d\n", file, file.Fd())
	// set opened ipmi dev file
	c.openipmi.file = file

	var receiveEvents uint32 = 1
	if err := open.IOCTL(c.openipmi.file.Fd(), open.IPMICTL_SET_GETS_EVENTS_CMD, uintptr(unsafe.Pointer(&receiveEvents))); err != nil {
		return fmt.Errorf("ioctl failed, cloud not enable event receiver, err: %w", err)
	}

	return nil
}

// closeOpen closes the ipmi dev file.
func (c *Client) closeOpen(ctx context.Context) error {
	if err := c.openipmi.file.Close(); err != nil {
		return fmt.Errorf("close open file failed, err: %w", err)
	}
	return nil
}

func (c *Client) openSendRequest(ctx context.Context, request ipmi.Request) ([]byte, error) {

	var dataPtr *byte

	cmdData := request.Pack()
	c.DebugBytes("cmdData", cmdData, 16)
	if len(cmdData) > 0 {
		dataPtr = &cmdData[0]
	}

	c.DebugBytes("cmd data", cmdData, 16)

	msg := &open.IPMI_MSG{
		NetFn:   uint8(request.Command().NetFn),
		Cmd:     uint8(request.Command().ID),
		Data:    dataPtr,
		DataLen: uint16(len(cmdData)),
	}

	addr := &open.IPMI_SYSTEM_INTERFACE_ADDR{
		AddrType: open.IPMI_SYSTEM_INTERFACE_ADDR_TYPE,
		Channel:  open.IPMI_BMC_CHANNEL,
		LUN:      0,
	}

	commandContext := GetCommandContext(ctx)
	if commandContext != nil {
		c.Debug("Got CommandContext:", commandContext)

		if commandContext.responderAddr != nil {
		}
		if commandContext.responderLUN != nil {
			addr.LUN = *commandContext.responderLUN
		}
		if commandContext.requesterAddr != nil {
		}
		if commandContext.requesterLUN != nil {
		}
	}

	req := &open.IPMI_REQ{
		Addr:    addr,
		AddrLen: int(unsafe.Sizeof(addr)),
		MsgID:   rand.Int63(),
		Msg:     *msg,
	}

	c.Debug("IPMI_REQ", req)
	return open.SendCommand(c.openipmi.file, req, c.timeout)
}

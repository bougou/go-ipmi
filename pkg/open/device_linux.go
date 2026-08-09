//go:build linux
// +build linux

package open

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Linux transport for the Open Interface: the OpenIPMI kernel driver
// (ipmi_devintf, /dev/ipmiN). Kernel UAPI types and ioctl numbers live in
// uapi_linux.go; this file is the DeviceBackend behavior.

// DeviceBackend implements Backend on top of the Linux OpenIPMI kernel
// driver (/dev/ipmiN).
type DeviceBackend struct {
	file *os.File
}

var _ Backend = (*DeviceBackend)(nil)

// OpenDevice opens the OpenIPMI character device for the given device
// number. It probes /dev/ipmiN, /dev/ipmi/N and /dev/ipmidev/N in order and
// returns the first path that opens successfully. When no path can be
// opened, the returned error lists every tried path with its individual
// failure reason.
func OpenDevice(devnum int32) (*os.File, error) {
	paths := []string{
		fmt.Sprintf("/dev/ipmi%d", devnum),
		fmt.Sprintf("/dev/ipmi/%d", devnum),
		fmt.Sprintf("/dev/ipmidev/%d", devnum),
	}

	var failures []string
	for _, path := range paths {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s (err: %s)", path, err))
			continue
		}
		return f, nil
	}

	return nil, fmt.Errorf("open ipmi device %d failed, tried: %s", devnum, strings.Join(failures, "; "))
}

// Connect opens the device and enables the event receiver. The devnum
// selects which /dev/ipmiN device family member to open.
func (b *DeviceBackend) Connect(ctx context.Context, devnum int32) error {
	file, err := OpenDevice(devnum)
	if err != nil {
		return err
	}

	var receiveEvents uint32 = 1
	if err := IOCTL(file.Fd(), IPMICTL_SET_GETS_EVENTS_CMD, uintptr(unsafe.Pointer(&receiveEvents))); err != nil {
		// Do not leak the opened fd when enabling events fails.
		file.Close()
		return fmt.Errorf("ioctl failed, could not enable event receiver, err: %w", err)
	}

	b.file = file
	return nil
}

// Close closes the device file. Closing a backend that was never connected
// is a no-op.
func (b *DeviceBackend) Close(ctx context.Context) error {
	if b.file == nil {
		return nil
	}
	if err := b.file.Close(); err != nil {
		return fmt.Errorf("close open file failed, err: %w", err)
	}
	b.file = nil
	return nil
}

// Send maps the transport-neutral Request onto a Linux openipmi ioctl
// round trip and returns the response in the canonical "cc + payload" form.
func (b *DeviceBackend) Send(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	if b.file == nil {
		return nil, fmt.Errorf("device backend not connected")
	}
	if req == nil {
		return nil, fmt.Errorf("nil open request")
	}
	return sendCommand(b.file, req, timeout)
}

func setReq(fd uintptr, op uintptr, req *IPMI_REQ) error {
	err := IOCTL(fd, op, uintptr(unsafe.Pointer(req)))
	runtime.KeepAlive(req)
	return err
}

func getRecv(fd uintptr, op uintptr, recv *IPMI_RECV) error {
	err := IOCTL(fd, op, uintptr(unsafe.Pointer(recv)))
	runtime.KeepAlive(recv)
	return err
}

// sendCommand builds the openipmi ioctl structs from Request and performs
// one send/receive round trip.
func sendCommand(file *os.File, req *Request, timeout time.Duration) ([]byte, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	var dataPtr *byte
	if len(req.Data) > 0 {
		dataPtr = &req.Data[0]
	}

	// Pass TargetAddr as-is: 0 means system interface (ipmitool open.c).
	addrPtr, addrLen, addrKeep := buildRequestAddr(
		req.EffectiveMyAddr(),
		req.TargetAddr,
		req.TargetChannel,
		req.LUN,
	)

	kernelReq := &IPMI_REQ{
		Addr:    addrPtr,
		AddrLen: addrLen,
		MsgID:   rand.Int63(),
		Msg: IPMI_MSG{
			NetFn:   req.NetFn,
			Cmd:     req.Cmd,
			Data:    dataPtr,
			DataLen: uint16(len(req.Data)),
		},
	}

	fd := file.Fd()
	for {
		switch err := setReq(fd, IPMICTL_SEND_COMMAND, kernelReq); {
		case err == syscall.EINTR:
			continue
		case err != nil:
			return nil, fmt.Errorf("setReq failed, err: %w", err)
		}
		break
	}

	// Use the generic IPMI_ADDR buffer for the receive side so the kernel
	// can return either a system-interface or IPMB source address
	// (ipmitool open.c does the same with struct ipmi_addr).
	recvAddr := &IPMI_ADDR{}
	recvBuf := make([]byte, IPMI_BUF_SIZE)
	recv := &IPMI_RECV{
		Addr:    unsafe.Pointer(recvAddr),
		AddrLen: int(unsafe.Sizeof(*recvAddr)),
		Msg: IPMI_MSG{
			Data:    &recvBuf[0],
			DataLen: IPMI_BUF_SIZE,
		},
	}

	var result []byte
	var rerr error

	readMsgFunc := func(fd uintptr) bool {
		if err := getRecv(fd, IPMICTL_RECEIVE_MSG_TRUNC, recv); err != nil {
			rerr = fmt.Errorf("getRecv failed, err: %w", err)
			return false
		}

		if recv.MsgID != kernelReq.MsgID {
			rerr = fmt.Errorf("received msg id not match")
			return false
		}

		if recv.Msg.DataLen >= IPMI_BUF_SIZE {
			rerr = fmt.Errorf("received data length longer than buf size: %d > %d", recv.Msg.DataLen, IPMI_BUF_SIZE)
		} else {
			// recvBuf[0] is completion code.
			result = recvBuf[:recv.Msg.DataLen:recv.Msg.DataLen]
			rerr = nil
		}
		return true
	}

	conn, err := file.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("failed to get syscall conn from file: %s", err)
	}
	if err := file.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline on file: %s", err)
	}
	if err := conn.Read(readMsgFunc); err != nil {
		return nil, fmt.Errorf("failed to read from syscall conn: %s", err)
	}

	// Keep Go-backed pointers alive across the ioctl round trip.
	runtime.KeepAlive(req.Data)
	runtime.KeepAlive(addrKeep)
	runtime.KeepAlive(kernelReq)
	runtime.KeepAlive(recvAddr)
	return result, rerr
}

// buildRequestAddr selects the Linux openipmi destination address,
// mirroring ipmitool open.c (ipmi_openipmi_send_cmd):
//
//	if targetAddr != 0 && targetAddr != myAddr → IPMI_IPMB_ADDR
//	else                                       → IPMI_SYSTEM_INTERFACE_ADDR
//
// keep must stay alive across the ioctl that consumes ptr.
func buildRequestAddr(myAddr, targetAddr, channel, lun uint8) (ptr unsafe.Pointer, addrLen int, keep any) {
	if targetAddr != 0 && targetAddr != myAddr {
		ipmb := &IPMI_IPMB_ADDR{
			AddrType:  IPMI_IPMB_ADDR_TYPE,
			Channel:   uint16(channel & 0x0f),
			SlaveAddr: targetAddr,
			LUN:       lun,
		}
		return unsafe.Pointer(ipmb), int(unsafe.Sizeof(*ipmb)), ipmb
	}

	sys := &IPMI_SYSTEM_INTERFACE_ADDR{
		AddrType: IPMI_SYSTEM_INTERFACE_ADDR_TYPE,
		Channel:  IPMI_BMC_CHANNEL,
		LUN:      lun,
	}
	return unsafe.Pointer(sys), int(unsafe.Sizeof(*sys)), sys
}

// addrTypeOf returns the addr_type of a buildRequestAddr result, or -1 if nil.
func addrTypeOf(ptr unsafe.Pointer) int32 {
	if ptr == nil {
		return -1
	}
	return *(*int32)(ptr)
}

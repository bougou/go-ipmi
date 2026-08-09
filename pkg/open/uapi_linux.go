//go:build linux
// +build linux

package open

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Linux OpenIPMI UAPI: ioctl encoding helpers, kernel argument structs, and
// ipmi_devintf command numbers (include/uapi/linux/ipmi.h,
// arch/*/include/uapi/asm/ioctl.h). Used only by DeviceBackend.

// --- ioctl number encoding -------------------------------------------------

// see: https://github.com/torvalds/linux/blob/master/arch/alpha/include/uapi/asm/ioctl.h

// cSpell:disable
const (
	IOC_NRBITS   = 8
	IOC_TYPEBITS = 8
	IOC_SIZEBITS = 14
	IOC_DIRBITS  = 2

	IOC_NONE  = 0x0
	IOC_READ  = 0x1
	IOC_WRITE = 0x2

	IOC_NRMASK   = ((1 << IOC_NRBITS) - 1)
	IOC_TYPEMASK = ((1 << IOC_TYPEBITS) - 1)
	IOC_SIZEMASK = ((1 << IOC_SIZEBITS) - 1)
	IOC_DIRMASK  = ((1 << IOC_DIRBITS) - 1)

	IOC_NRSHIFT   = 0
	IOC_TYPESHIFT = (IOC_NRSHIFT + IOC_NRBITS)
	IOC_SIZESHIFT = (IOC_TYPESHIFT + IOC_TYPEBITS)
	IOC_DIRSHIFT  = (IOC_SIZESHIFT + IOC_SIZEBITS)

	IOC_IN        = (IOC_WRITE << IOC_DIRSHIFT)
	IOC_OUT       = (IOC_READ << IOC_DIRSHIFT)
	IOC_INOUT     = ((IOC_WRITE | IOC_READ) << IOC_DIRSHIFT)
	IOCSIZE_MASK  = (IOC_SIZEMASK << IOC_SIZESHIFT)
	IOCSIZE_SHIFT = (IOC_SIZESHIFT)
)

func IOC(dir, typ, nr, size uintptr) uintptr {
	return (dir << IOC_DIRSHIFT) | (typ << IOC_TYPESHIFT) | (nr << IOC_NRSHIFT) | (size << IOC_SIZESHIFT)
}

func IO(typ, nr uintptr) uintptr         { return IOC(IOC_NONE, typ, nr, 0) }
func IOR(typ, nr, size uintptr) uintptr  { return IOC(IOC_READ, typ, nr, size) }
func IOW(typ, nr, size uintptr) uintptr  { return IOC(IOC_WRITE, typ, nr, size) }
func IOWR(typ, nr, size uintptr) uintptr { return IOC(IOC_READ|IOC_WRITE, typ, nr, size) }
func IOC_DIR(nr uintptr) uintptr         { return (nr >> IOC_DIRSHIFT) & IOC_DIRMASK }
func IOC_TYPE(nr uintptr) uintptr        { return (nr >> IOC_TYPESHIFT) & IOC_TYPEMASK }
func IOC_NR(nr uintptr) uintptr          { return (nr >> IOC_NRSHIFT) & IOC_NRMASK }
func IOC_SIZE(nr uintptr) uintptr        { return (nr >> IOC_SIZESHIFT) & IOC_SIZEMASK }

// IOCTL issues SYS_IOCTL against an openipmi file descriptor.
func IOCTL(fd, name, data uintptr) error {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, name, data)
	if ep != 0 {
		return fmt.Errorf("syscall err: (%#02x) %s", uint8(ep), syscall.Errno(ep))
	}
	return nil
}

// --- openipmi structs (linux/ipmi.h) ----------------------------------------

const (
	IPMI_BUF_SIZE      = 1024
	IPMI_MAX_ADDR_SIZE = 32

	// Channel for talking directly with the BMC (system interface only).
	IPMI_BMC_CHANNEL = 0xf

	IPMI_NUM_CHANNELS = 0x10

	IPMI_RESPONSE_RECV_TYPE     = 1
	IPMI_ASYNC_EVENT_RECV_TYPE  = 2
	IPMI_CMD_RECV_TYPE          = 3
	IPMI_RESPONSE_RESPONSE_TYPE = 4
	IPMI_OEM_RECV_TYPE          = 5

	IPMI_MAINTENANCE_MODE_AUTO = 0
	IPMI_MAINTENANCE_MODE_OFF  = 1
	IPMI_MAINTENANCE_MODE_ON   = 2

	IPMI_SYSTEM_INTERFACE_ADDR_TYPE = 0x0c
	IPMI_IPMB_ADDR_TYPE             = 0x01
	IPMI_IPMB_BROADCAST_ADDR_TYPE   = 0x41 // broadcast get device id (IPMI 1.5 §17.9)
	IPMI_IPMB_DIRECT_ADDR_TYPE      = 0x81
	IPMI_LAN_ADDR_TYPE              = 0x04
)

// IPMI_ADDR is the generic openipmi address buffer.
type IPMI_ADDR struct {
	AddrType int32
	Channel  uint16
	Data     [IPMI_MAX_ADDR_SIZE]byte
}

// IPMI_SYSTEM_INTERFACE_ADDR is used for direct BMC system-interface messages.
type IPMI_SYSTEM_INTERFACE_ADDR struct {
	AddrType int32
	Channel  uint16
	LUN      uint8
}

// IPMI_IPMB_ADDR is used for IPMB / broadcast-IPMB destinations.
type IPMI_IPMB_ADDR struct {
	AddrType  int32
	Channel   uint16
	SlaveAddr uint8
	LUN       uint8
}

// IPMI_IPMB_DIRECT_ADDR is for messages received directly from an IPMB.
type IPMI_IPMB_DIRECT_ADDR struct {
	AddrType  int32
	Channel   uint16
	SlaveAddr uint8
	RsLUN     uint8
	RqLUN     uint8
}

// IPMI_LAN_ADDR is an address to/from a LAN interface bridged by the BMC.
type IPMI_LAN_ADDR struct {
	AddrType      int32
	Channel       uint16
	Privilege     uint8
	SessionHandle uint8
	RemoteSWID    uint8
	LocalSWID     uint8
	LUN           uint8
}

// IPMI_MSG is the openipmi message body (no addressing). Response data's
// first byte is the completion code. Size is 16 on amd64.
type IPMI_MSG struct {
	NetFn   uint8
	Cmd     uint8
	DataLen uint16
	Data    *byte
}

func (msg *IPMI_MSG) MsgData() ([]byte, error) {
	if msg.DataLen >= IPMI_BUF_SIZE {
		return nil, fmt.Errorf("received data length longer than buf size: %d > %d", msg.DataLen, IPMI_BUF_SIZE)
	}
	recvBuf := unsafe.Slice(msg.Data, msg.DataLen)
	return recvBuf[:msg.DataLen:msg.DataLen], nil
}

// IPMI_REQ mirrors struct ipmi_req. Size is 40 on amd64; ioctl numbers encode it.
// DeviceBackend only — not the Backend.Send public protocol (see Request).
type IPMI_REQ struct {
	Addr    unsafe.Pointer
	AddrLen int
	MsgID   int64
	Msg     IPMI_MSG
}

// IPMI_RECV mirrors struct ipmi_recv. Size is 48 on amd64.
type IPMI_RECV struct {
	RecvType int
	Addr     unsafe.Pointer
	AddrLen  int
	MsgID    int64
	Msg      IPMI_MSG
}

// IPMI_REQ_SETTIME mirrors struct ipmi_req_settime.
type IPMI_REQ_SETTIME struct {
	Req             IPMI_REQ
	Retries         int32
	RetryTimeMillis uint32
}

// IPMI_CMDSPEC registers for commands from other entities on this interface.
type IPMI_CMDSPEC struct {
	NetFn uint8
	Cmd   uint8
}

type IPMI_CMDSPEC_CHANS struct {
	NetFn int
	Cmd   int
	Chans int
}

type IPMI_CHANNEL_LUN_ADDRESS_SET struct {
	Channel uint16
	Value   uint8
}

type IPMI_TIMING_PARAMS struct {
	Retries         int
	RetryTimeMillis uint
}

// --- ipmi_devintf ioctl command numbers ------------------------------------

const IPMI_IOC_MAGIC uintptr = 'i'

var (
	IPMICTL_SEND_COMMAND         = IOW(IPMI_IOC_MAGIC, 13, unsafe.Sizeof(IPMI_REQ{}))
	IPMICTL_SEND_COMMAND_SETTIME = IOW(IPMI_IOC_MAGIC, 21, unsafe.Sizeof(IPMI_REQ_SETTIME{}))

	IPMICTL_RECEIVE_MSG       = IOWR(IPMI_IOC_MAGIC, 12, unsafe.Sizeof(IPMI_RECV{}))
	IPMICTL_RECEIVE_MSG_TRUNC = IOWR(IPMI_IOC_MAGIC, 11, unsafe.Sizeof(IPMI_RECV{}))

	IPMICTL_REGISTER_FOR_CMD   = IOR(IPMI_IOC_MAGIC, 14, unsafe.Sizeof(IPMI_CMDSPEC{}))
	IPMICTL_UNREGISTER_FOR_CMD = IOR(IPMI_IOC_MAGIC, 15, unsafe.Sizeof(IPMI_CMDSPEC{}))

	IPMICTL_REGISTER_FOR_CMD_CHANS   = IOR(IPMI_IOC_MAGIC, 28, unsafe.Sizeof(IPMI_CMDSPEC_CHANS{}))
	IPMICTL_UNREGISTER_FOR_CMD_CHANS = IOR(IPMI_IOC_MAGIC, 29, unsafe.Sizeof(IPMI_CMDSPEC_CHANS{}))

	IPMICTL_SET_GETS_EVENTS_CMD = IOW(IPMI_IOC_MAGIC, 16, unsafe.Sizeof(uint32(0)))

	IPMICTL_SET_MY_CHANNEL_ADDRESS_CMD = IOR(IPMI_IOC_MAGIC, 24, unsafe.Sizeof(IPMI_CHANNEL_LUN_ADDRESS_SET{}))
	IPMICTL_GET_MY_CHANNEL_ADDRESS_CMD = IOR(IPMI_IOC_MAGIC, 25, unsafe.Sizeof(IPMI_CHANNEL_LUN_ADDRESS_SET{}))
	IPMICTL_SET_MY_CHANNEL_LUN_CMD     = IOR(IPMI_IOC_MAGIC, 26, unsafe.Sizeof(IPMI_CHANNEL_LUN_ADDRESS_SET{}))
	IPMICTL_GET_MY_CHANNEL_LUN_CMD     = IOR(IPMI_IOC_MAGIC, 27, unsafe.Sizeof(IPMI_CHANNEL_LUN_ADDRESS_SET{}))

	/* Legacy interfaces, these only set IPMB 0. */
	IPMICTL_SET_MY_ADDRESS_CMD = IOR(IPMI_IOC_MAGIC, 17, unsafe.Sizeof(uint32(0)))
	IPMICTL_GET_MY_ADDRESS_CMD = IOR(IPMI_IOC_MAGIC, 18, unsafe.Sizeof(uint32(0)))
	IPMICTL_SET_MY_LUN_CMD     = IOR(IPMI_IOC_MAGIC, 19, unsafe.Sizeof(uint32(0)))
	IPMICTL_GET_MY_LUN_CMD     = IOR(IPMI_IOC_MAGIC, 20, unsafe.Sizeof(uint32(0)))

	IPMICTL_SET_TIMING_PARAMS_CMD = IOR(IPMI_IOC_MAGIC, 22, unsafe.Sizeof(IPMI_TIMING_PARAMS{}))
	IPMICTL_GET_TIMING_PARAMS_CMD = IOR(IPMI_IOC_MAGIC, 23, unsafe.Sizeof(IPMI_TIMING_PARAMS{}))

	IPMICTL_GET_MAINTENANCE_MODE_CMD = IOR(IPMI_IOC_MAGIC, 30, unsafe.Sizeof(uint32(0)))
	IPMICTL_SET_MAINTENANCE_MODE_CMD = IOW(IPMI_IOC_MAGIC, 31, unsafe.Sizeof(uint32(0)))
)

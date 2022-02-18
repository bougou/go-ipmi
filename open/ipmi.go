package open

import (
	"fmt"
	"time"
	"unsafe"
)

// see: https://github.com/u-root/u-root/blob/v0.8.0/pkg/ipmi/ipmi.go

// Note, this file basically is a Go conversion of https://github.com/torvalds/linux/blob/master/include/uapi/linux/ipmi.h

const (
	IPMI_IOC_MAGIC uintptr = 'i'

	IPMI_BUF_SIZE                        = 1024
	IPMI_FILE_READ_TIMEOUT time.Duration = time.Second * 10
	IPMI_MAX_ADDR_SIZE                   = 32

	// Channel for talking directly with the BMC.  When using this
	// channel, This is for the system interface address type only.
	IPMI_BMC_CHANNEL = 0xf

	IPMI_NUM_CHANNELS = 0x10

	// Receive types for messages coming from the receive interface.
	// This is used for the receive in-kernel interface and in the receive IOCTL.
	//
	// The "IPMI_RESPONSE_RESPONSE_TYPE" is a little strange sounding, but
	// it allows you to get the message results when you send a response message.

	IPMI_RESPONSE_RECV_TYPE     = 1
	IPMI_ASYNC_EVENT_RECV_TYPE  = 2
	IPMI_CMD_RECV_TYPE          = 3
	IPMI_RESPONSE_RESPONSE_TYPE = 4
	IPMI_OEM_RECV_TYPE          = 5

	IPMI_MAINTENANCE_MODE_AUTO = 0
	IPMI_MAINTENANCE_MODE_OFF  = 1
	IPMI_MAINTENANCE_MODE_ON   = 2
)

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

	IPMICTL_GET_MAINTENCANCE_MODE_CMD = IOR(IPMI_IOC_MAGIC, 30, unsafe.Sizeof(uint32(0)))
	IPMICTL_SET_MAINTENCANCE_MODE_CMD = IOW(IPMI_IOC_MAGIC, 31, unsafe.Sizeof(uint32(0)))
)

// IPMI_ADDR wraps different IPMI ADDR TYPE data to one struct.
// IPMI ADDR TYPE (Channel Meidum Type), see: 6.5 Channel Medium Type
type IPMI_ADDR struct {
	AddrType int32
	Channel  uint16
	Data     [IPMI_MAX_ADDR_SIZE]byte // Addr Data
}

const IPMI_SYSTEM_INTERFACE_ADDR_TYPE = 0x0c

// IPMI_SYSTEM_INTERFACE_ADDR holds addr data of addr type IPMI_SYSTEM_INTERFACE_ADDR_TYPE.
type IPMI_SYSTEM_INTERFACE_ADDR struct {
	AddrType int32
	Channel  uint16
	LUN      uint8
}

const IPMI_IPMB_ADDR_TYPE = 0x01
const IPMI_IPMB_BROADCAST_ADDR_TYPE = 0x41 // Used for broadcast get device id as described in section 17.9 of the IPMI 1.5 manual.

// IPMI_IPMB_ADDR holds addr data of addr type IPMI_IPMB_ADDR_TYPE or IPMI_IPMB_BROADCAST_ADDR_TYPE.
//
// It represents an IPMB address.
type IPMI_IPMB_ADDR struct {
	AddrType  int32
	Channel   uint16
	SlaveAddr uint8
	LUN       uint8
}

const IPMI_IPMB_DIRECT_ADDR_TYPE = 0x81

// IPMI_IPMB_DIRECT_ADDR holds addr data of addr type IPMI_IPMB_DIRECT_ADDR_TYPE.
//
// Used for messages received directly from an IPMB that have not gone
// through a MC. This is for systems that sit right on an IPMB so
// they can receive commands and respond to them.
type IPMI_IPMB_DIRECT_ADDR struct {
	AddrType  int32
	Channel   uint16
	SlaveAddr uint8
	RsLUN     uint8
	RqLUN     uint8
}

const IPMI_LAN_ADDR_TYPE = 0x04

// IPMI_LAN_ADDR holds addr data of addr type IPMI_LAN_ADDR_TYPE.
//
// A LAN Address. This is an address to/from a LAN interface bridged
// by the BMC, not an address actually out on the LAN.
type IPMI_LAN_ADDR struct {
	AddrType      int32
	Channel       uint16
	Privilege     uint8
	SessionHandle uint8
	RemoteSWID    uint8
	LocalSWID     uint8
	LUN           uint8
}

// IPMI_MSG holds a raw IPMI message without any addressing. This covers both
// commands and responses. The completion code is always the first
// byte of data in the response (as the spec shows the messages laid out).
//
// unsafe.Sizeof of IPMI_MSG is 1+1+2+(4)+8=16.
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

// unsafe.Sizeof of IPMI_REQ is 8+8(4+4)+8+16 = 40.
type IPMI_REQ struct {
	Addr    *IPMI_SYSTEM_INTERFACE_ADDR
	AddrLen int

	// The sequence number for the message.  This
	// exact value will be reported back in the
	// response to this request if it is a command.
	// If it is a response, this will be used as
	// the sequence value for the response.
	MsgID int64
	Msg   IPMI_MSG
}

// unsafe.Sizeof of IPMI_RECV is 8(4+4)+8+8(4+4)+8+16 = 48.
type IPMI_RECV struct {
	RecvType int
	Addr     *IPMI_SYSTEM_INTERFACE_ADDR
	AddrLen  int
	MsgID    int64
	Msg      IPMI_MSG
}

type IPMI_REQ_SETTIME struct {
	Req             IPMI_REQ
	Retries         int32
	RetryTimeMillis uint32
}

// Register to get commands from other entities on this interface
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

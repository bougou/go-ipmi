package open

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

func SetReq(fd uintptr, op uintptr, req *IPMI_REQ) error {
	err := IOCTL(fd, op, uintptr(unsafe.Pointer(req)))
	runtime.KeepAlive(req)
	return err
}

func GetRecv(fd uintptr, op uintptr, recv *IPMI_RECV) error {
	err := IOCTL(fd, op, uintptr(unsafe.Pointer(recv)))
	runtime.KeepAlive(recv)
	return err
}

func SendCommand(file *os.File, req *IPMI_REQ) ([]byte, error) {
	fd := file.Fd()

	for {
		switch err := SetReq(fd, IPMICTL_SEND_COMMAND, req); {
		case err == syscall.EINTR:
			continue
		case err != nil:
			return nil, fmt.Errorf("SetReq failed, err: %s", err)
		}
		break
	}

	recvBuf := make([]byte, IPMI_BUF_SIZE)
	recv := &IPMI_RECV{
		Addr:    req.Addr,
		AddrLen: req.AddrLen,
		Msg: IPMI_MSG{
			Data:    &recvBuf[0],
			DataLen: IPMI_BUF_SIZE,
		},
	}

	var result []byte
	var rerr error

	readMsgFunc := func(fd uintptr) bool {
		if err := GetRecv(fd, IPMICTL_RECEIVE_MSG_TRUNC, recv); err != nil {
			rerr = fmt.Errorf("GetRecv failed, err: %s", err)
			return false
		}

		if recv.MsgID != req.MsgID {
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
	if err := file.SetReadDeadline(time.Now().Add(IPMI_FILE_READ_TIMEOUT)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline on file: %s", err)
	}
	if err := conn.Read(readMsgFunc); err != nil {
		return nil, fmt.Errorf("failed to read from syscall conn: %s", err)
	}

	return result, rerr
}

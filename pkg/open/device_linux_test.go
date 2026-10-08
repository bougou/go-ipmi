//go:build linux
// +build linux

package open

import (
	"bytes"
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func TestBuildRequestAddrSystemInterface(t *testing.T) {
	cases := []struct {
		name                string
		myAddr, target, lun uint8
		channel             uint8
	}{
		{name: "target equals myAddr", myAddr: 0x20, target: 0x20, lun: 1},
		{name: "target zero", myAddr: 0x20, target: 0, lun: 0},
		{name: "both satellite same", myAddr: 0x88, target: 0x88, lun: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ptr, n, keep := buildRequestAddr(tc.myAddr, tc.target, tc.channel, tc.lun)
			if ptr == nil || keep == nil {
				t.Fatal("expected non-nil addr")
			}
			if addrTypeOf(ptr) != IPMI_SYSTEM_INTERFACE_ADDR_TYPE {
				t.Fatalf("addr type: got %#x, want SYSTEM_INTERFACE %#x", addrTypeOf(ptr), IPMI_SYSTEM_INTERFACE_ADDR_TYPE)
			}
			sys, ok := keep.(*IPMI_SYSTEM_INTERFACE_ADDR)
			if !ok {
				t.Fatalf("keep type: got %T, want *IPMI_SYSTEM_INTERFACE_ADDR", keep)
			}
			if sys.Channel != IPMI_BMC_CHANNEL {
				t.Fatalf("channel: got %#x, want BMC_CHANNEL %#x", sys.Channel, IPMI_BMC_CHANNEL)
			}
			if sys.LUN != tc.lun {
				t.Fatalf("lun: got %d, want %d", sys.LUN, tc.lun)
			}
			if n != uint32(unsafe.Sizeof(*sys)) {
				t.Fatalf("addrLen: got %d, want %d", n, unsafe.Sizeof(*sys))
			}
		})
	}
}

func TestBuildRequestAddrIPMB(t *testing.T) {
	ptr, n, keep := buildRequestAddr(0x20, 0x2c, 0x03, 0x01)
	if addrTypeOf(ptr) != IPMI_IPMB_ADDR_TYPE {
		t.Fatalf("addr type: got %#x, want IPMB %#x", addrTypeOf(ptr), IPMI_IPMB_ADDR_TYPE)
	}
	ipmb, ok := keep.(*IPMI_IPMB_ADDR)
	if !ok {
		t.Fatalf("keep type: got %T, want *IPMI_IPMB_ADDR", keep)
	}
	if ipmb.SlaveAddr != 0x2c {
		t.Fatalf("slave: got %#02x", ipmb.SlaveAddr)
	}
	if ipmb.Channel != 0x03 {
		t.Fatalf("channel: got %#x, want 0x03", ipmb.Channel)
	}
	if ipmb.LUN != 0x01 {
		t.Fatalf("lun: got %d, want 1", ipmb.LUN)
	}
	if n != uint32(unsafe.Sizeof(*ipmb)) {
		t.Fatalf("addrLen: got %d, want %d", n, unsafe.Sizeof(*ipmb))
	}
}

func TestBuildRequestAddrChannelMasked(t *testing.T) {
	_, _, keep := buildRequestAddr(0x20, 0x2c, 0xf5, 0)
	ipmb := keep.(*IPMI_IPMB_ADDR)
	if ipmb.Channel != 0x05 {
		t.Fatalf("channel low-nibble: got %#x, want 0x05", ipmb.Channel)
	}
}

func TestIPMIReqSizeMatchesKernel(t *testing.T) {
	want := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = 40
	}
	if unsafe.Sizeof(IPMI_REQ{}) != want {
		t.Fatalf("IPMI_REQ size: got %d, want %d", unsafe.Sizeof(IPMI_REQ{}), want)
	}
}

// The pipe supplies real poller waits and deadlines; only the device ioctls
// are replaced, so these tests need no IPMI hardware.
func TestSendCommandPreCanceled(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sent := false
	_, err = sendCommand(ctx, read, &Request{}, time.Second,
		func(uintptr, uintptr, *IPMI_REQ) error { sent = true; return syscall.EIO },
		func(uintptr, uintptr, *IPMI_RECV) error { t.Fatal("unexpected receive"); return nil })
	if sent || !errors.Is(err, context.Canceled) {
		t.Fatalf("sent=%v error=%v, want no send and context.Canceled", sent, err)
	}
	backend := &DeviceBackend{file: read}
	if _, err := backend.Send(ctx, &Request{}, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("backend error=%v, want context.Canceled", err)
	}
}

func TestSendCommandCancellationAndReuse(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancel"
		if deadline {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer read.Close()
			defer write.Close()
			ctx, cancel := context.WithCancel(context.Background())
			want := context.Canceled
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Millisecond)
				want = context.DeadlineExceeded
			}
			defer cancel()
			waiting := make(chan struct{}, 1)
			done := make(chan error, 1)
			go func() {
				_, err := sendCommand(ctx, read, &Request{}, time.Second,
					func(uintptr, uintptr, *IPMI_REQ) error { return nil },
					func(uintptr, uintptr, *IPMI_RECV) error {
						select {
						case waiting <- struct{}{}:
						default:
						}
						return syscall.EAGAIN
					})
				done <- err
			}()
			select {
			case <-waiting:
			case err := <-done:
				t.Fatalf("command returned before reaching receive wait: %v", err)
			}
			if !deadline {
				cancel()
			}
			select {
			case err = <-done:
				if !errors.Is(err, want) {
					t.Errorf("error=%v, want %v", err, want)
				}
			case <-time.After(500 * time.Millisecond):
				t.Error("receive did not stop promptly after context completion")
				// Join the operation even when testing the unfixed implementation.
				_ = read.SetReadDeadline(time.Now())
				<-done
			}
			var msgID int
			result, err := sendCommand(context.Background(), read, &Request{}, time.Second,
				func(_ uintptr, _ uintptr, req *IPMI_REQ) error { msgID = req.MsgID; return nil },
				func(_ uintptr, _ uintptr, recv *IPMI_RECV) error {
					recv.MsgID = msgID
					*recv.Msg.Data = 0
					recv.Msg.DataLen = 1
					return nil
				})
			if err != nil || len(result) != 1 || result[0] != 0 {
				t.Fatalf("next command: response=%v error=%v", result, err)
			}
		})
	}
}

func TestSendCommandTransportTimeout(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	_, err = sendCommand(context.Background(), read, &Request{}, 20*time.Millisecond,
		func(uintptr, uintptr, *IPMI_REQ) error { return nil },
		func(uintptr, uintptr, *IPMI_RECV) error { return syscall.EAGAIN })
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("error=%v, want transport timeout", err)
	}
}

func TestSendCommandPayload(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "empty"},
		{name: "nonempty", data: []byte{0x01, 0x80, 0xff}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer read.Close()
			defer write.Close()
			req := &Request{NetFn: 6, Cmd: 1, Data: tc.data}
			want := []byte{0, 0x80, 0xff}
			var msgID int
			result, err := sendCommand(context.Background(), read, req, time.Second,
				func(_ uintptr, op uintptr, kernelReq *IPMI_REQ) error {
					if op != IPMICTL_SEND_COMMAND || kernelReq.Msg.NetFn != req.NetFn || kernelReq.Msg.Cmd != req.Cmd {
						t.Fatalf("unexpected request: op=%#x msg=%+v", op, kernelReq.Msg)
					}
					data := unsafe.Slice(kernelReq.Msg.Data, kernelReq.Msg.DataLen)
					if !bytes.Equal(data, tc.data) || (len(tc.data) == 0 && kernelReq.Msg.Data != nil) {
						t.Fatalf("request data=%x pointer=%p, want %x", data, kernelReq.Msg.Data, tc.data)
					}
					msgID = kernelReq.MsgID
					return nil
				},
				func(_ uintptr, op uintptr, recv *IPMI_RECV) error {
					if op != IPMICTL_RECEIVE_MSG_TRUNC || int(recv.Msg.DataLen) < len(want) {
						t.Fatalf("unexpected receive: op=%#x capacity=%d", op, recv.Msg.DataLen)
					}
					recv.MsgID = msgID
					copy(unsafe.Slice(recv.Msg.Data, recv.Msg.DataLen), want)
					recv.Msg.DataLen = uint16(len(want))
					return nil
				})
			if err != nil || !bytes.Equal(result, want) {
				t.Fatalf("response=%x error=%v, want %x", result, err, want)
			}
		})
	}
}

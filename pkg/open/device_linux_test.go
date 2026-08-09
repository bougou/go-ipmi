//go:build linux
// +build linux

package open

import (
	"testing"
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
			if n != int(unsafe.Sizeof(*sys)) {
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
	if n != int(unsafe.Sizeof(*ipmb)) {
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
	if unsafe.Sizeof(IPMI_REQ{}) != 40 {
		t.Fatalf("IPMI_REQ size: got %d, want 40", unsafe.Sizeof(IPMI_REQ{}))
	}
}

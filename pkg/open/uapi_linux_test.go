//go:build linux

package open

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestLinuxUAPI compares layouts and commands with the target's real C headers.
// CC selects a cross-compiler; IPMI_UAPI_EXEC optionally runs its output under
// an emulator. Go's -exec option must use the same target when cross-testing.
func TestLinuxUAPI(t *testing.T) {
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	compiler, err := exec.LookPath(cc)
	if err != nil {
		if os.Getenv("CC") != "" {
			t.Fatal(err)
		}
		t.Skip("C ABI oracle requires a C compiler and Linux headers")
	}
	probe := filepath.Join(t.TempDir(), "uapi")
	compile := exec.Command(compiler, "-std=c11", "-Wall", "-Wextra", "-Werror", "-static", "-o", probe, "testdata/uapi.c")
	t.Logf("compile C ABI oracle: %q", compile.Args)
	if out, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("compile oracle: %v\n%s", err, out)
	}
	args := append(strings.Fields(os.Getenv("IPMI_UAPI_EXEC")), probe)
	run := exec.Command(args[0], args[1:]...)
	t.Logf("run C ABI oracle: %q", run.Args)
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run oracle: %v\n%s", err, output)
	}
	want := map[string]uint64{}
	reader := bytes.NewReader(output)
	for {
		var name string
		var value uint64
		if _, err := fmt.Fscan(reader, &name, &value); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("parse oracle: %v", err)
		}
		want[name] = value
	}
	got := map[string]uint64{}
	for _, layout := range []struct {
		value        any
		name, fields string
	}{
		{IPMI_ADDR{}, "ipmi_addr", "AddrType:addr_type Channel:channel Data:data"},
		{IPMI_SYSTEM_INTERFACE_ADDR{}, "ipmi_system_interface_addr", "AddrType:addr_type Channel:channel LUN:lun"},
		{IPMI_IPMB_ADDR{}, "ipmi_ipmb_addr", "AddrType:addr_type Channel:channel SlaveAddr:slave_addr LUN:lun"},
		{IPMI_IPMB_DIRECT_ADDR{}, "ipmi_ipmb_direct_addr", "AddrType:addr_type Channel:channel SlaveAddr:slave_addr RsLUN:rs_lun RqLUN:rq_lun"},
		{IPMI_LAN_ADDR{}, "ipmi_lan_addr", "AddrType:addr_type Channel:channel Privilege:privilege SessionHandle:session_handle RemoteSWID:remote_SWID LocalSWID:local_SWID LUN:lun"},
		{IPMI_MSG{}, "ipmi_msg", "NetFn:netfn Cmd:cmd DataLen:data_len Data:data"},
		{IPMI_REQ{}, "ipmi_req", "Addr:addr AddrLen:addr_len MsgID:msgid Msg:msg"},
		{IPMI_RECV{}, "ipmi_recv", "RecvType:recv_type Addr:addr AddrLen:addr_len MsgID:msgid Msg:msg"},
		{IPMI_REQ_SETTIME{}, "ipmi_req_settime", "Req:req Retries:retries RetryTimeMillis:retry_time_ms"},
		{IPMI_CMDSPEC{}, "ipmi_cmdspec", "NetFn:netfn Cmd:cmd"},
		{IPMI_CMDSPEC_CHANS{}, "ipmi_cmdspec_chans", "NetFn:netfn Cmd:cmd Chans:chans"},
		{IPMI_CHANNEL_LUN_ADDRESS_SET{}, "ipmi_channel_lun_address_set", "Channel:channel Value:value"},
		{IPMI_TIMING_PARAMS{}, "ipmi_timing_parms", "Retries:retries RetryTimeMillis:retry_time_ms"},
	} {
		typ := reflect.TypeOf(layout.value)
		got[layout.name+".size"] = uint64(typ.Size())
		got[layout.name+".align"] = uint64(typ.Align())
		for _, mapping := range strings.Fields(layout.fields) {
			names := strings.Split(mapping, ":")
			field, ok := typ.FieldByName(names[0])
			if !ok {
				t.Fatalf("missing field: %s.%s", typ, names[0])
			}
			prefix := layout.name + "." + names[1]
			got[prefix+".offset"] = uint64(field.Offset)
			got[prefix+".size"] = uint64(field.Type.Size())
			var signed uint64
			if k := field.Type.Kind(); k >= reflect.Int && k <= reflect.Int64 {
				signed = 1
			}
			got[prefix+".signed"] = signed
		}
	}
	for name, value := range map[string]uintptr{
		"IPMICTL_SEND_COMMAND":               IPMICTL_SEND_COMMAND,
		"IPMICTL_SEND_COMMAND_SETTIME":       IPMICTL_SEND_COMMAND_SETTIME,
		"IPMICTL_RECEIVE_MSG":                IPMICTL_RECEIVE_MSG,
		"IPMICTL_RECEIVE_MSG_TRUNC":          IPMICTL_RECEIVE_MSG_TRUNC,
		"IPMICTL_REGISTER_FOR_CMD":           IPMICTL_REGISTER_FOR_CMD,
		"IPMICTL_UNREGISTER_FOR_CMD":         IPMICTL_UNREGISTER_FOR_CMD,
		"IPMICTL_REGISTER_FOR_CMD_CHANS":     IPMICTL_REGISTER_FOR_CMD_CHANS,
		"IPMICTL_UNREGISTER_FOR_CMD_CHANS":   IPMICTL_UNREGISTER_FOR_CMD_CHANS,
		"IPMICTL_SET_GETS_EVENTS_CMD":        IPMICTL_SET_GETS_EVENTS_CMD,
		"IPMICTL_SET_MY_CHANNEL_ADDRESS_CMD": IPMICTL_SET_MY_CHANNEL_ADDRESS_CMD,
		"IPMICTL_GET_MY_CHANNEL_ADDRESS_CMD": IPMICTL_GET_MY_CHANNEL_ADDRESS_CMD,
		"IPMICTL_SET_MY_CHANNEL_LUN_CMD":     IPMICTL_SET_MY_CHANNEL_LUN_CMD,
		"IPMICTL_GET_MY_CHANNEL_LUN_CMD":     IPMICTL_GET_MY_CHANNEL_LUN_CMD,
		"IPMICTL_SET_MY_ADDRESS_CMD":         IPMICTL_SET_MY_ADDRESS_CMD,
		"IPMICTL_GET_MY_ADDRESS_CMD":         IPMICTL_GET_MY_ADDRESS_CMD,
		"IPMICTL_SET_MY_LUN_CMD":             IPMICTL_SET_MY_LUN_CMD,
		"IPMICTL_GET_MY_LUN_CMD":             IPMICTL_GET_MY_LUN_CMD,
		"IPMICTL_SET_TIMING_PARMS_CMD":       IPMICTL_SET_TIMING_PARAMS_CMD,
		"IPMICTL_GET_TIMING_PARMS_CMD":       IPMICTL_GET_TIMING_PARAMS_CMD,
		"IPMICTL_GET_MAINTENANCE_MODE_CMD":   IPMICTL_GET_MAINTENANCE_MODE_CMD,
		"IPMICTL_SET_MAINTENANCE_MODE_CMD":   IPMICTL_SET_MAINTENANCE_MODE_CMD,
		"_IOC_NRBITS":                        IOC_NRBITS,
		"_IOC_TYPEBITS":                      IOC_TYPEBITS,
		"_IOC_SIZEBITS":                      IOC_SIZEBITS,
		"_IOC_DIRBITS":                       IOC_DIRBITS,
		"_IOC_NONE":                          IOC_NONE,
		"_IOC_READ":                          IOC_READ,
		"_IOC_WRITE":                         IOC_WRITE,
		"_IOC_NRMASK":                        IOC_NRMASK,
		"_IOC_TYPEMASK":                      IOC_TYPEMASK,
		"_IOC_SIZEMASK":                      IOC_SIZEMASK,
		"_IOC_DIRMASK":                       IOC_DIRMASK,
		"_IOC_NRSHIFT":                       IOC_NRSHIFT,
		"_IOC_TYPESHIFT":                     IOC_TYPESHIFT,
		"_IOC_SIZESHIFT":                     IOC_SIZESHIFT,
		"_IOC_DIRSHIFT":                      IOC_DIRSHIFT,
		"IO":                                 IO(0x69, 0x55),
		"IOR":                                IOR(0x69, 0x55, 4),
		"IOW":                                IOW(0x69, 0x55, 4),
		"IOWR":                               IOWR(0x69, 0x55, 4),
	} {
		got[name] = uint64(value)
	}
	if len(got) != len(want) {
		t.Fatalf("oracle entries: Go=%d C=%d", len(got), len(want))
	}
	for name, want := range want {
		if got, ok := got[name]; !ok || got != want {
			t.Errorf("%s: Go=%#x, C=%#x", name, got, want)
		}
	}
	// Decode C-produced values as well as comparing the encoded numbers.
	command := uintptr(want["IOWR"])
	if IOC_DIR(command) != uintptr(want["_IOC_READ"]|want["_IOC_WRITE"]) || IOC_TYPE(command) != 0x69 || IOC_NR(command) != 0x55 || IOC_SIZE(command) != 4 {
		t.Errorf("incorrect decoded ioctl %#x", command)
	}
}

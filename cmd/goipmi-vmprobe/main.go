// goipmi-vmprobe is a minimal OpenIPMI VM protocol client for development and
// E2E testing: it connects to a stream socket served by goipmi-server's VM
// protocol mode (GOIPMI_SERVER_VM_SOCKET), sends one IPMI command, and reports
// the completion code and response. It stands in for QEMU's ipmi-bmc-extern or
// the in-guest OpenIPMI driver, which are the real consumers of that socket.
//
// Usage:
//
//	goipmi-vmprobe -socket /tmp/vm.sock                      # Get Device ID (default)
//	goipmi-vmprobe -socket /tmp/vm.sock -netfn 6 -cmd 1
//	goipmi-vmprobe -socket /tmp/vm.sock -netfn 0 -cmd 2 -data 1  # Chassis Control power up
//
// Exit status is 0 only when a response arrives with completion code 0x00.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bougou/go-ipmi/pkg/vmproto"
)

// parseDataBytes parses a comma-separated list of request data bytes, each in
// decimal or 0x-hex, rejecting anything that is not a byte in 0..255. An empty
// string is no data.
func parseDataBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]byte, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// Exactly "decimal or 0x-hex": base 16 only for a 0x/0X prefix, base 10
		// otherwise. Not base 0, which would read a leading zero as octal
		// (010 -> 8) and accept 0b/0o/underscores, none of which are meant here.
		base, num := 10, p
		if len(p) > 2 && (p[0:2] == "0x" || p[0:2] == "0X") {
			base, num = 16, p[2:]
		}
		v, err := strconv.ParseUint(num, base, 8) // 8 bits bounds 0..255
		if err != nil {
			return nil, fmt.Errorf("invalid byte %q: want a value 0..255 (decimal or 0x-hex)", p)
		}
		out = append(out, byte(v))
	}
	return out, nil
}

func main() {
	socket := flag.String("socket", "", "unix socket path served by goipmi-server VM protocol mode")
	netFn := flag.Uint("netfn", 0x06, "request NetFn (default App)")
	cmd := flag.Uint("cmd", 0x01, "command byte (default Get Device ID)")
	dataStr := flag.String("data", "", "comma-separated request data bytes, decimal or 0x-hex (e.g. 1 or 0x01,0xff)")
	network := flag.String("net", "unix", "socket network: unix or tcp")
	timeout := flag.Duration("timeout", 5*time.Second, "response timeout")
	flag.Parse()

	if *socket == "" {
		fmt.Fprintln(os.Stderr, "goipmi-vmprobe: -socket is required")
		os.Exit(2)
	}

	data, err := parseDataBytes(*dataStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goipmi-vmprobe: -data: %v\n", err)
		os.Exit(2)
	}

	conn, err := net.DialTimeout(*network, *socket, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goipmi-vmprobe: dial %s %s: %v\n", *network, *socket, err)
		os.Exit(1)
	}
	defer conn.Close() //nolint:errcheck

	cc, resp, err := vmproto.NewClient(conn, *timeout).Command(byte(*netFn), byte(*cmd), data...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goipmi-vmprobe: command: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("completion code: 0x%02x\n", cc)
	fmt.Printf("response (%d bytes): % x\n", len(resp), resp)
	if cc != 0x00 {
		os.Exit(1)
	}
}

package vmproto

import (
	"fmt"
	"net"
	"time"
)

// Client speaks the console side of the OpenIPMI VM protocol (QEMU's
// ipmi-bmc-extern wire format) over a byte stream, the counterpart to
// [VMServer]. QEMU and the in-guest OpenIPMI driver are the real consumers of
// the server; this client stands in for them in simulation and end-to-end
// testing.
type Client struct {
	conn    net.Conn
	timeout time.Duration
	msgID   byte
}

// NewClient wraps conn, a stream connection to a [VMServer] listener, as a
// VM-protocol client. timeout bounds each response read; pass 0 for no
// deadline.
func NewClient(conn net.Conn, timeout time.Duration) *Client {
	return &Client{conn: conn, timeout: timeout}
}

// Command sends one IPMI request over the VM protocol and returns the response
// completion code and data. The system interface carries one transaction at a
// time, so it writes the request and reads the matching response.
func (c *Client) Command(netFn, cmd byte, data ...byte) (cc byte, resp []byte, err error) {
	c.msgID++

	msg := append([]byte{c.msgID, netFn << 2, cmd}, data...)
	msg = append(msg, -vmChecksum(msg))
	if _, err := c.conn.Write(vmEncode(msg)); err != nil {
		return 0, nil, fmt.Errorf("write request: %w", err)
	}

	reply, err := c.readMessage()
	if err != nil {
		return 0, nil, err
	}
	if len(reply) < 5 {
		return 0, nil, fmt.Errorf("short response: % x", reply)
	}
	if reply[0] != c.msgID {
		return 0, nil, fmt.Errorf("response message ID %d, want %d", reply[0], c.msgID)
	}
	// The response NetFn is the request NetFn with the low bit set (§5.1).
	if got := reply[1] >> 2; got != netFn|1 {
		return 0, nil, fmt.Errorf("response NetFn %#x, want %#x", got, netFn|1)
	}
	if reply[2] != cmd {
		return 0, nil, fmt.Errorf("response command %#x, want %#x", reply[2], cmd)
	}
	return reply[3], reply[4 : len(reply)-1], nil
}

// readMessage reads one unescaped, checksum-verified message from the stream.
func (c *Client) readMessage() ([]byte, error) {
	if c.timeout > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return nil, err
		}
	}

	var (
		acc      []byte
		inEscape bool
		buf      = make([]byte, 1)
	)
	for {
		if _, err := c.conn.Read(buf); err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		switch b := buf[0]; b {
		case vmMsgChar:
			if len(acc) < 4 {
				return nil, fmt.Errorf("response too short: % x", acc)
			}
			if vmChecksum(acc) != 0 {
				return nil, fmt.Errorf("response checksum mismatch: % x", acc)
			}
			return acc, nil
		case vmEscapeChar:
			inEscape = true
		default:
			if inEscape {
				b &^= 0x10
				inEscape = false
			}
			acc = append(acc, b)
		}
	}
}

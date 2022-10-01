package ipmi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

// ConnectTool try to initialize the client.
func (c *Client) ConnectTool(devnum int32) error {
	return nil
}

// closeTool closes the ipmi dev file.
func (c *Client) closeTool() error {
	return nil
}

func (c *Client) exchangeTool(request Request, response Response) error {
	data := request.Pack()
	msg := make([]byte, 2+len(data))
	msg[0] = uint8(request.Command().NetFn)
	msg[1] = uint8(request.Command().ID)
	copy(msg[2:], data)

	args := append([]string{"raw"}, rawEncode(msg)...)

	path := c.Host
	if path == "" {
		path = "ipmitool"
	}

	cmd := exec.Command(path, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ipmitool run failed, err: %s", err)
	}

	output := stdout.String()
	resp := rawDecode(strings.TrimSpace(output))
	if err := response.Unpack(resp); err != nil {
		return fmt.Errorf("unpack response failed, err: %s", err)
	}

	return nil
}

func rawDecode(data string) []byte {
	var buf bytes.Buffer

	for _, s := range strings.Split(data, " ") {
		b, err := hex.DecodeString(s)
		if err != nil {
			panic(err)
		}

		_, err = buf.Write(b)
		if err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func rawEncode(data []byte) []string {
	n := len(data)
	buf := make([]string, 0, n)

	// ipmitool needs every byte to be a separate argument
	for i := 0; i < n; i++ {
		buf = append(buf, "0x"+hex.EncodeToString(data[i:i+1]))
	}

	return buf
}

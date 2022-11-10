package ipmi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	toolError = regexp.MustCompile(`^Unable to send RAW command \(channel=0x(?P<channel>[0-9a-fA-F]+) netfn=0x(?P<netfn>[0-9a-fA-F]+) lun=0x(?P<lun>[0-9a-fA-F]+) cmd=0x(?P<cmd>[0-9a-fA-F]+) rsp=0x(?P<rsp>[0-9a-fA-F]+)\): (?P<message>.*)`)
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

	c.Debugf(">>> Run cmd: \n>>> %s\n", cmd.String())
	err := cmd.Run()
	if err != nil {
		if bytes.HasPrefix(stderr.Bytes(), []byte("Unable to send RAW command")) {
			submatches := toolError.FindSubmatch(stderr.Bytes())
			if len(submatches) == 7 && len(submatches[5]) == 2 {
				code, err := strconv.ParseUint(string(submatches[5]), 16, 0)
				if err != nil {
					return fmt.Errorf("CompletionCode parse failed, err: %s", err)
				}
				return &ResponseError{
					completionCode: CompletionCode(uint8(code)),
					description:    fmt.Sprintf("Raw command failed, err: %s", string(submatches[6])),
				}
			}
		}
		return fmt.Errorf("ipmitool run failed, err: %s", err)
	}

	output := stdout.String()
	resp, err := rawDecode(strings.TrimSpace(output))
	if err != nil {
		return fmt.Errorf("decode response failed, err: %s", err)
	}
	if err := response.Unpack(resp); err != nil {
		return fmt.Errorf("unpack response failed, err: %s", err)
	}

	return nil
}

func rawDecode(data string) ([]byte, error) {
	var buf bytes.Buffer

	data = strings.ReplaceAll(data, "\n", "")
	for _, s := range strings.Split(data, " ") {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(b)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
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

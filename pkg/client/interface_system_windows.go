//go:build windows
// +build windows

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// On Windows there is no OpenIPMI character device. Local (in-band) access to
// the BMC is provided by the Microsoft IPMI driver (ipmidrv.sys), which is
// surfaced through the Microsoft_IPMI WMI class in the root\wmi namespace.
//
// Rather than pulling in a COM/WMI dependency, the implementation drives the
// provider through PowerShell (consistent with how the "tool" interface shells
// out to ipmitool). Each request invokes the Microsoft_IPMI RequestResponse
// method and the result is marshalled back as JSON.

// winIPMIResponse mirrors the JSON emitted by the PowerShell helper script.
type winIPMIResponse struct {
	// CompletionCode is the driver/method level status of the WMI call
	// (0 means the request was delivered). It is distinct from the IPMI
	// completion code, which is the first byte of ResponseData.
	CompletionCode uint32 `json:"CompletionCode"`
	// ResponseData is the comma separated list of response bytes. The first
	// byte is the IPMI completion code.
	ResponseData     string `json:"ResponseData"`
	ResponseDataSize uint32 `json:"ResponseDataSize"`
}

// ConnectOpen verifies that the Microsoft_IPMI WMI provider is available.
func (c *Client) ConnectOpen(ctx context.Context, devnum int32) error {
	c.Debugf("Using Microsoft_IPMI WMI provider (root\\wmi)\n")

	script := `$ErrorActionPreference = 'Stop'
$ipmi = Get-CimInstance -Namespace 'root/wmi' -ClassName 'Microsoft_IPMI' -ErrorAction Stop | Select-Object -First 1
if ($null -eq $ipmi) { throw 'Microsoft_IPMI WMI instance not found' }
'ok'`

	if _, err := c.runPowerShell(ctx, script); err != nil {
		return fmt.Errorf("Microsoft_IPMI WMI provider not available (is the IPMI driver installed and are you running as administrator?), err: %w", err)
	}

	return nil
}

// closeOpen is a no-op for the WMI backed local interface.
func (c *Client) closeOpen(ctx context.Context) error {
	return nil
}

func (c *Client) openSendRequest(ctx context.Context, request ipmi.Request) ([]byte, error) {
	cmdData := request.Pack()
	c.DebugBytes("cmd data", cmdData, 16)

	parts := make([]string, len(cmdData))
	for i, b := range cmdData {
		parts[i] = strconv.Itoa(int(b))
	}
	dataCSV := strings.Join(parts, ",")

	netFn := uint8(request.Command().NetFn)
	cmd := uint8(request.Command().ID)
	responderAddr := ipmi.BMC_SA
	var lun uint8 = 0

	commandContext := GetCommandContext(ctx)
	if commandContext != nil {
		c.Debug("Got CommandContext:", commandContext)
		if commandContext.responderAddr != nil {
			responderAddr = *commandContext.responderAddr
		}
		if commandContext.responderLUN != nil {
			lun = *commandContext.responderLUN
		}
	}

	script := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$data = [byte[]]@(%s)
$arguments = @{
  NetworkFunction = [byte]%d
  Command = [byte]%d
  Lun = [byte]%d
  ResponderAddress = [byte]%d
  RequestData = $data
  RequestDataSize = [uint32]$data.Length
}
$ipmi = Get-CimInstance -Namespace 'root/wmi' -ClassName 'Microsoft_IPMI' -ErrorAction Stop | Select-Object -First 1
if ($null -eq $ipmi) { throw 'Microsoft_IPMI WMI instance not found' }
$res = Invoke-CimMethod -InputObject $ipmi -MethodName 'RequestResponse' -Arguments $arguments -ErrorAction Stop
$rd = $res.ResponseData
if ($null -eq $rd) { $rdStr = '' } else { $rdStr = (($rd | ForEach-Object { [int]$_ }) -join ',') }
[pscustomobject]@{
  CompletionCode = [uint32]$res.CompletionCode
  ResponseData = $rdStr
  ResponseDataSize = [uint32]$res.ResponseDataSize
} | ConvertTo-Json -Compress`, dataCSV, netFn, cmd, lun, responderAddr)

	out, err := c.runPowerShell(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("Microsoft_IPMI RequestResponse failed, err: %w", err)
	}

	var resp winIPMIResponse
	if err := json.Unmarshal(bytes.TrimSpace(out), &resp); err != nil {
		return nil, fmt.Errorf("parse Microsoft_IPMI response failed, err: %w, output: %s", err, string(out))
	}

	if resp.CompletionCode != 0 {
		return nil, fmt.Errorf("Microsoft_IPMI RequestResponse returned method completion code %#x", resp.CompletionCode)
	}

	recv, err := parseByteCSV(resp.ResponseData)
	if err != nil {
		return nil, fmt.Errorf("parse Microsoft_IPMI response data failed, err: %w", err)
	}

	// recv[0] is the IPMI completion code, recv[1:] is the response data.
	return recv, nil
}

// runPowerShell executes the given PowerShell script and returns its stdout.
func (c *Client) runPowerShell(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-NonInteractive", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// parseByteCSV parses a comma separated list of unsigned byte values.
func parseByteCSV(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []byte{}, nil
	}

	fields := strings.Split(s, ",")
	out := make([]byte, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseUint(strings.TrimSpace(f), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %q: %w", f, err)
		}
		out[i] = byte(v)
	}
	return out, nil
}

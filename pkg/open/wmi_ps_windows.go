//go:build windows
// +build windows

package open

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PowerShell transport for the Open Interface on Windows.
//
// Same contract as WindowsConn: tunnel a raw IPMI request through the
// Microsoft_IPMI::RequestResponse WMI method and return the raw response,
// whose first byte is the IPMI completion code. Where WindowsConn drives
// the provider in-process via COM, this transport shells out to
// powershell.exe once per request, which costs a few hundred milliseconds
// of process-startup + CIM-enumeration overhead per call (measured
// 0.4–0.6 s per request end-to-end on a live BMC). It exists purely as a
// fallback for hosts where the COM path is unavailable; nothing should
// prefer it for bulk work.
//
// Stateless by design: each request is self-contained, so the "connection"
// carries no resources. OpenPowerShellConn still performs a provider probe
// so callers get the same Open/Send/Close lifecycle as WindowsConn and
// learn about a missing driver or broken WMI at connect time, not at the
// first request.
//
// For the overview of the two Windows transports and the auto-fallback
// policy, see wmi_com_windows.go.

// winIPMIResponse mirrors the JSON emitted by the PowerShell helper script.
type winIPMIResponse struct {
	// CompletionCode is the IPMI completion code returned by the BMC and
	// propagated through the Microsoft_IPMI WMI provider. 0x00 means the
	// command completed normally; any non-zero value is a real IPMI
	// completion code (generic per IPMI v2.0 §5.2, or command-specific per
	// the command's spec section) and is surfaced to callers as a
	// ResponseError by the client layer so that higher-level logic (e.g.
	// GetUsers skipping empty slots on 0xCC) can react to it.
	CompletionCode uint32 `json:"CompletionCode"`
	// ResponseData is the comma separated list of response bytes. On success
	// (CompletionCode == 0) the first byte is the IPMI completion code
	// (0x00) followed by the command's response payload, matching the
	// on-wire layout the client layer expects.
	ResponseData     string `json:"ResponseData"`
	ResponseDataSize uint32 `json:"ResponseDataSize"`
}

// PowerShellConn is the PowerShell-based transport to the Microsoft_IPMI
// WMI provider. It holds no state; it exists so the PowerShell path has
// the same Open/Send/Close lifecycle shape as WindowsConn.
type PowerShellConn struct{}

// OpenPowerShellConn verifies that the Microsoft_IPMI provider is reachable
// via PowerShell's CIM cmdlets. This is the probe counterpart to
// OpenWindowsConn's WMI connect: it fails fast when the IPMI driver is
// missing or the caller lacks rights, instead of failing at first Send.
func OpenPowerShellConn(ctx context.Context) (*PowerShellConn, error) {
	script := `$ErrorActionPreference = 'Stop'
$ipmi = Get-CimInstance -Namespace 'root/wmi' -ClassName 'Microsoft_IPMI' -ErrorAction Stop | Select-Object -First 1
if ($null -eq $ipmi) { throw 'Microsoft_IPMI WMI instance not found' }
'ok'`

	if _, err := runPowerShell(ctx, script); err != nil {
		return nil, fmt.Errorf("Microsoft_IPMI WMI provider not available (is the IPMI driver installed and are you running as administrator?), err: %w", err)
	}
	return &PowerShellConn{}, nil
}

// Close releases nothing — PowerShellConn holds no resources. The method
// exists so the lifecycle mirrors WindowsConn.Close.
func (c *PowerShellConn) Close() error {
	return nil
}

// SendCommand issues a single IPMI request via PowerShell's Invoke-CimMethod
// and returns the response in the canonical "cc + payload" form. Non-zero
// completion codes are NOT wrapped here — the client layer reads recv[0]
// and wraps them, matching the Linux and WindowsConn contracts.
//
// ResponderAddress / Lun come from req.EffectiveTarget() and req.LUN.
// The Microsoft_IPMI WMI provider documents ResponderAddress as ignored for
// the system interface on many hosts, but CommandContext overrides are still
// forwarded so behaviour matches the COM transport and the Linux IPMB path.
//
// timeout bounds the PowerShell process: the context handed to
// exec.CommandContext is wrapped with context.WithTimeout, mirroring how
// WindowsConn.SendCommand enforces its timeout. timeout == 0 falls back to
// DefaultTimeout like the other transports.
func (c *PowerShellConn) SendCommand(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("PowerShellConn is not initialised")
	}
	if req == nil {
		return nil, fmt.Errorf("nil open request")
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmdData := req.Data
	parts := make([]string, len(cmdData))
	for i, b := range cmdData {
		parts[i] = strconv.Itoa(int(b))
	}
	dataCSV := strings.Join(parts, ",")

	netfn := req.NetFn
	cmd := req.Cmd
	responderAddr := req.EffectiveTarget()
	lun := req.LUN

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
} | ConvertTo-Json -Compress`, dataCSV, netfn, cmd, lun, responderAddr)

	out, err := runPowerShell(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("Microsoft_IPMI RequestResponse failed, err: %w", err)
	}

	return parseWinIPMIResponse(out)
}

// parseWinIPMIResponse decodes the JSON emitted by the PowerShell helper and
// returns the response bytes in the canonical "cc + payload" form.
//
// The Microsoft_IPMI provider returns the full IPMI response including the
// leading completion code byte in ResponseData. The CompletionCode property
// mirrors ResponseData[0]. Prepending it separately shifts every downstream
// field by one. Return ResponseData as-is (truncated to ResponseDataSize when
// present, matching the COM transport) and let the client layer read recv[0]
// as the cc.
func parseWinIPMIResponse(out []byte) ([]byte, error) {
	var resp winIPMIResponse
	if err := json.Unmarshal(bytes.TrimSpace(out), &resp); err != nil {
		return nil, fmt.Errorf("parse Microsoft_IPMI response failed, err: %w, output: %s", err, string(out))
	}

	recv, err := parseByteCSV(resp.ResponseData)
	if err != nil {
		return nil, fmt.Errorf("parse Microsoft_IPMI response data failed, err: %w", err)
	}

	if resp.ResponseDataSize > 0 && int(resp.ResponseDataSize) <= len(recv) {
		recv = recv[:resp.ResponseDataSize]
	}

	if len(recv) == 0 {
		return nil, fmt.Errorf("ResponseData empty (no completion code)")
	}
	return recv, nil
}

// runPowerShell executes the given PowerShell script and returns its stdout.
func runPowerShell(ctx context.Context, script string) ([]byte, error) {
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

// PowerShellBackend implements Backend by talking to the Microsoft_IPMI WMI
// provider through PowerShellConn, which shells out to powershell.exe once
// per request. It holds only the connection handle; all transport internals
// (script construction, process spawning, JSON parsing) live on
// PowerShellConn.
type PowerShellBackend struct {
	conn *PowerShellConn
}

var _ Backend = (*PowerShellBackend)(nil)

// Connect probes the WMI provider via PowerShell. The devnum is ignored:
// Windows exposes a single Microsoft_IPMI instance.
func (b *PowerShellBackend) Connect(ctx context.Context, devnum int32) error {
	conn, err := OpenPowerShellConn(ctx)
	if err != nil {
		return fmt.Errorf("open WMI Microsoft_IPMI provider via PowerShell failed, err: %w", err)
	}
	b.conn = conn
	return nil
}

func (b *PowerShellBackend) Close(ctx context.Context) error {
	if b.conn == nil {
		return nil
	}
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("close PowerShell WMI connection failed, err: %w", err)
	}
	b.conn = nil
	return nil
}

func (b *PowerShellBackend) Send(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	if b.conn == nil {
		return nil, fmt.Errorf("wmi-ps backend not connected")
	}
	return b.conn.SendCommand(ctx, req, timeout)
}

// ConnectPowerShellBackend constructs a PowerShellBackend and connects it,
// returning the ready-to-use Backend. Intended as a factory callback for
// ResolveBackend.
func ConnectPowerShellBackend(ctx context.Context) (Backend, error) {
	b := &PowerShellBackend{}
	if err := b.Connect(ctx, 0); err != nil {
		return nil, err
	}
	return b, nil
}

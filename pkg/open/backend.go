// Package open provides in-band (System Interface) access to a local BMC.
//
//   - Linux:   OpenIPMI character device (/dev/ipmiN) via ioctl
//   - Windows: Microsoft_IPMI WMI provider (ipmidrv.sys) via COM or PowerShell
//
// Callers build a transport-neutral Request and send it through a Backend.
// Platform backends map Request onto their native plumbing and return the
// raw response as "completion code + payload".
package open

import (
	"context"
	"fmt"
	"time"
)

// DefaultTimeout is used when Backend.Send is called with timeout == 0.
const DefaultTimeout time.Duration = time.Second * 10

// BMCAddr is the conventional BMC slave address (0x20).
const BMCAddr uint8 = 0x20

// Backend abstracts an Open Interface transport: a local (in-band) path to
// the host BMC that tunnels a single system-interface-style IPMI request
// and returns the raw response.
//
// Implementations:
//   - linux:   DeviceBackend (OpenIPMI /dev/ipmiN)
//   - windows: COMBackend / PowerShellBackend (Microsoft_IPMI WMI)
//
// All implementations return a byte slice whose first byte is the IPMI
// completion code and remaining bytes are the response payload.
type Backend interface {
	Connect(ctx context.Context, devnum int32) error
	Close(ctx context.Context) error
	Send(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error)
}

// Request is the transport-neutral Open Interface request handed to
// Backend.Send. It carries the system-interface command body
// (NetFn/LUN/Cmd/Data — IPMI v2.0 §9.2 / §10.14 / §11.1) plus a logical
// destination. Individual backends map this onto their native plumbing:
//
//   - Linux DeviceBackend → openipmi ioctl (struct ipmi_req + ipmi_addr)
//   - Windows COM/PS      → Microsoft_IPMI::RequestResponse parameters
//
// This type deliberately does NOT mirror the Linux UAPI layout; ioctl
// structs live in uapi_linux.go.
type Request struct {
	NetFn uint8
	Cmd   uint8
	// LUN is the responder LUN (low 2 bits of NetFn/LUN on the system
	// interface; ipmi_addr.lun / WMI Lun on IPMB paths).
	LUN uint8
	// Data is the command request body (may be nil or empty).
	Data []byte

	// TargetAddr is the destination slave address. 0 means the local
	// system interface (ipmitool open.c). Non-zero and different from
	// MyAddr routes via IPMB / WMI ResponderAddress.
	TargetAddr uint8
	// TargetChannel is the IPMB channel (low 4 bits). Ignored for the
	// local system interface.
	TargetChannel uint8

	// MyAddr is this open interface's own IPMB address. 0 is treated as
	// BMCAddr. Used by the Linux backend for system-interface vs IPMB
	// routing; ignored by Windows WMI backends.
	MyAddr uint8
}

// EffectiveTarget returns TargetAddr, substituting BMCAddr when zero.
// Used by Windows WMI (ResponderAddress); Linux addressing treats a raw
// TargetAddr of 0 as "system interface" instead.
func (r *Request) EffectiveTarget() uint8 {
	if r == nil || r.TargetAddr == 0 {
		return BMCAddr
	}
	return r.TargetAddr
}

// EffectiveMyAddr returns MyAddr, substituting BMCAddr when zero.
func (r *Request) EffectiveMyAddr() uint8 {
	if r == nil || r.MyAddr == 0 {
		return BMCAddr
	}
	return r.MyAddr
}

// UsesIPMB reports whether this request should be routed over IPMB rather
// than the local system interface (ipmitool open.c: target != 0 && target != my).
func (r *Request) UsesIPMB() bool {
	if r == nil || r.TargetAddr == 0 {
		return false
	}
	return r.TargetAddr != r.EffectiveMyAddr()
}

// Backend names accepted by ResolveBackend (and goipmi's --open-backend).
// They select which Microsoft_IPMI WMI transport Windows uses; Linux always
// uses DeviceBackend and ignores the preference.
const (
	BackendAuto       = "auto"
	BackendCOM        = "wmi-com"
	BackendPowerShell = "wmi-ps"
)

// ResolveBackend picks a Backend from the preference string using two
// factory callbacks (COM and PowerShell). "" and BackendAuto try COM first
// and fall back to PowerShell on failure; an explicit BackendCOM or
// BackendPowerShell never falls back.
//
// onCOMFail, if non-nil, is invoked with the COM error immediately before
// the PowerShell fallback attempt (auto path only).
func ResolveBackend(pref string, tryCOM, tryPS func() (Backend, error), onCOMFail func(error)) (Backend, error) {
	if pref == "" {
		pref = BackendAuto
	}
	switch pref {
	case BackendCOM:
		b, err := tryCOM()
		if err != nil {
			return nil, err
		}
		return b, nil
	case BackendPowerShell:
		b, err := tryPS()
		if err != nil {
			return nil, err
		}
		return b, nil
	case BackendAuto:
		if b, err := tryCOM(); err == nil {
			return b, nil
		} else {
			if onCOMFail != nil {
				onCOMFail(err)
			}
			b, psErr := tryPS()
			if psErr != nil {
				return nil, fmt.Errorf("wmi-com: %v; wmi-ps: %w", err, psErr)
			}
			return b, nil
		}
	default:
		return nil, fmt.Errorf("unsupported open backend %q, supported: %s, %s, %s", pref, BackendCOM, BackendPowerShell, BackendAuto)
	}
}

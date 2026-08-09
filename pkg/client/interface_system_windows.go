//go:build windows
// +build windows

package client

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/open"
)

// On Windows there is no OpenIPMI character device. Local (in-band) access
// to the BMC is provided by the Microsoft IPMI driver (ipmidrv.sys) through
// the Microsoft_IPMI WMI class. pkg/open provides the two transports
// (open.COMBackend and open.PowerShellBackend); the selection semantics and
// the auto fallback behaviour are documented there.

// ConnectOpen selects a backend based on c.openBackendPref and connects it.
// "" or open.BackendAuto (the default) tries the native COM backend first
// and falls back to the PowerShell backend on failure.
func (c *Client) ConnectOpen(ctx context.Context, devnum int32) error {
	tryCom := func() (open.Backend, error) { return open.ConnectCOMBackend(ctx) }
	tryPS := func() (open.Backend, error) { return open.ConnectPowerShellBackend(ctx) }

	backend, err := open.ResolveBackend(c.openBackendPref, tryCom, tryPS, func(err error) {
		c.Debugf("COM backend unavailable, falling back to PowerShell: %v\n", err)
	})
	if err != nil {
		return err
	}
	c.openipmi.backend = backend

	switch backend.(type) {
	case *open.COMBackend:
		c.Debugf("Using Microsoft_IPMI WMI provider via native COM (root\\wmi)\n")
	case *open.PowerShellBackend:
		c.Debugf("Using Microsoft_IPMI WMI provider via PowerShell (root\\wmi)\n")
	default:
		c.Debugf("Using Microsoft_IPMI WMI provider (%T)\n", backend)
	}
	return nil
}

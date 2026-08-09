//go:build linux
// +build linux

package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/open"
)

// ConnectOpen try to initialize the client by open the device of linux ipmi driver.
func (c *Client) ConnectOpen(ctx context.Context, devnum int32) error {
	c.Debugf("Using ipmi device %d\n", devnum)

	b := &open.DeviceBackend{}
	if err := b.Connect(ctx, devnum); err != nil {
		return fmt.Errorf("connect open device failed, err: %w", err)
	}

	c.openipmi.backend = b
	return nil
}

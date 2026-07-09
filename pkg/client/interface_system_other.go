//go:build !linux && !windows
// +build !linux,!windows

package client

import (
	"context"
	"fmt"
	"runtime"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

var errOpenUnsupported = fmt.Errorf("open (local) interface is not supported on %s, only linux and windows are supported", runtime.GOOS)

// ConnectOpen is not supported on this platform.
func (c *Client) ConnectOpen(ctx context.Context, devnum int32) error {
	return errOpenUnsupported
}

// closeOpen is not supported on this platform.
func (c *Client) closeOpen(ctx context.Context) error {
	return errOpenUnsupported
}

func (c *Client) openSendRequest(ctx context.Context, request ipmi.Request) ([]byte, error) {
	return nil, errOpenUnsupported
}

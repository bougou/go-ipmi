//go:build !linux

package main

import (
	"fmt"
	"runtime"

	"github.com/bougou/go-ipmi/pkg/hal"
)

// openConsoleHAL is only meaningful on Linux (PTY/serial devices); elsewhere
// the reference server runs without a console and SOL stays unadvertised.
func openConsoleHAL(spec string) (hal.ConsoleHAL, string, error) {
	return nil, "", fmt.Errorf("console %q not supported on %s", spec, runtime.GOOS)
}

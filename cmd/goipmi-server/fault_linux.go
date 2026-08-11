//go:build linux

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

// consoleFaultInject simulates a broken console link: while set, every
// console read fails, flipping the SOL instance broken so the reconnect
// engine runs its real recovery path. Toggled by SIGUSR1 (break) and
// SIGUSR2 (restore); e2e uses it to drive a full
// outage → reconnect → recovery cycle against a real client.
//
// Linux-only like the console backend itself (openConsoleHAL); the
// SIGUSR1/SIGUSR2 toggle in startConsoleFaultInjection is the only writer,
// console_linux.go the only reader.
var consoleFaultInject atomic.Bool

// startConsoleFaultInjection wires the e2e console-fault toggle: SIGUSR1
// breaks the console link, SIGUSR2 restores it (see consoleFaultInject
// above). Linux-only — SIGUSR1/SIGUSR2 exist on Unix only, matching the
// Linux-only console backend (openConsoleHAL).
func startConsoleFaultInjection() {
	faultCh := make(chan os.Signal, 2)
	signal.Notify(faultCh, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for sig := range faultCh {
			switch sig {
			case syscall.SIGUSR1:
				consoleFaultInject.Store(true)
				fmt.Fprintln(os.Stderr, "console fault injected: reads now fail")
			case syscall.SIGUSR2:
				consoleFaultInject.Store(false)
				fmt.Fprintln(os.Stderr, "console fault cleared: reads restored")
			}
		}
	}()
}

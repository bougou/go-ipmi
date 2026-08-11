//go:build !linux

package main

// startConsoleFaultInjection is a no-op off Linux: SIGUSR1/SIGUSR2 are Unix
// signals, and the console backend itself is Linux-only (see openConsoleHAL),
// so the fault toggle can never be exercised here.
func startConsoleFaultInjection() {}

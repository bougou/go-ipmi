//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bougou/go-ipmi/pkg/hal"
	"golang.org/x/sys/unix"
)

// consoleOpener adapts an open function to [hal.ConsoleHAL].
type consoleOpener func(context.Context) (hal.ConsoleConn, error)

func (f consoleOpener) Open(ctx context.Context) (hal.ConsoleConn, error) { return f(ctx) }

// openConsoleHAL builds the console HAL for GOIPMI_SERVER_CONSOLE:
//
//	"pty"      – allocate a PTY pair; the BMC holds the master, the slave
//	             path is reported for tests to play the system side
//	<path>     – open an existing character device (e.g. /dev/ttyS0)
//
// The returned string describes the console for the startup banner.
func openConsoleHAL(spec string) (hal.ConsoleHAL, string, error) {
	if spec == "pty" {
		master, slavePath, err := openPTY()
		if err != nil {
			return nil, "", err
		}
		return consoleOpener(dupConsoleOpener(master)), "pty slave: " + slavePath, nil
	}

	f, err := os.OpenFile(spec, os.O_RDWR, 0)
	if err != nil {
		return nil, "", fmt.Errorf("open console device: %w", err)
	}
	return consoleOpener(dupConsoleOpener(f)), spec, nil
}

// dupConsoleOpener hands each activation a dup(2) of the master fd so that
// payload deactivation (conn close) does not prevent later re-activation.
func dupConsoleOpener(f *os.File) func(context.Context) (hal.ConsoleConn, error) {
	return func(context.Context) (hal.ConsoleConn, error) {
		fd, err := unix.Dup(int(f.Fd()))
		if err != nil {
			return nil, fmt.Errorf("dup console fd: %w", err)
		}
		return &fileConsoleConn{
			f:  os.NewFile(uintptr(fd), f.Name()+"#dup"),
			fd: fd,
			sendBreak: func() error {
				return unix.IoctlSetInt(fd, unix.TCSBRK, 0)
			},
		}, nil
	}
}

// fileConsoleConn adapts a character device fd (PTY master, serial port) to
// [hal.ConsoleConn].
type fileConsoleConn struct {
	f  *os.File // owns fd: Write/Close go through it
	fd int

	sendBreak func() error
}

func (c *fileConsoleConn) ReadAvailable(p []byte) (int, error) {
	// Fault injection sits in front of the fd: it must make reads fail
	// even when the underlying tty still has data, so the reconnect engine
	// sees the same "console gone" it would with a dead link.
	if consoleFaultInject.Load() {
		return 0, errors.New("console fault injected (SIGUSR1)")
	}
	// poll(2) with a zero timeout, then a raw read. SetReadDeadline on
	// *os.File cannot express this: the runtime poller reports an already
	// elapsed deadline before attempting the syscall, so a "now" deadline
	// loses data that is sitting in the tty buffer.
	fds := []unix.PollFd{{Fd: int32(c.fd), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, 0)
	if err != nil {
		return 0, err
	}
	if n == 0 || fds[0].Revents&unix.POLLIN == 0 {
		return 0, nil
	}
	rn, err := unix.Read(c.fd, p)
	if err != nil {
		if err == unix.EAGAIN { // drained by a concurrent reader between poll and read
			return 0, nil
		}
		return rn, err
	}
	return rn, nil
}

func (c *fileConsoleConn) Write(p []byte) (int, error) { return c.f.Write(p) }

func (c *fileConsoleConn) Close() error { return c.f.Close() }

func (c *fileConsoleConn) SendBreak(context.Context) error {
	if c.sendBreak == nil {
		return hal.ErrNotSupported
	}
	return c.sendBreak()
}

// openPTY allocates a PTY pair via /dev/ptmx and returns the master side
// plus the slave path (grantpt/unlockpt equivalents via TIOCGPTN/TIOCSPTLCK).
func openPTY() (*os.File, string, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, "", fmt.Errorf("open /dev/ptmx: %w", err)
	}
	n, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		_ = master.Close()
		return nil, "", fmt.Errorf("TIOCGPTN: %w", err)
	}
	// TIOCSPTLCK takes a pointer to int (unlock = 0), unlike the value-arg
	// ioctls above.
	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, 0); err != nil {
		_ = master.Close()
		return nil, "", fmt.Errorf("unlockpt: %w", err)
	}
	return master, fmt.Sprintf("/dev/pts/%d", n), nil
}

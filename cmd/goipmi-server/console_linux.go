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
		if err := unix.SetNonblock(fd, true); err != nil {
			_ = unix.Close(fd)
			return nil, fmt.Errorf("set console fd non-blocking: %w", err)
		}
		return &fileConsoleConn{
			f:  os.NewFile(uintptr(fd), f.Name()+"#dup"),
			fd: fd,
			// A PTY master polls POLLHUP whenever no slave is open — a
			// transient, normal state for a shared console whose peer is
			// opened per write — so hangup detection must not apply to it.
			pty: isPty(fd),
			sendBreak: func() error {
				return unix.IoctlSetInt(fd, unix.TCSBRK, 0)
			},
		}, nil
	}
}

// isPty reports whether fd is a pseudo-terminal: TIOCGPTN succeeds only on
// pty masters and slaves.
func isPty(fd int) bool {
	_, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	return err == nil
}

// fileConsoleConn adapts a character device fd (PTY master, serial port) to
// [hal.ConsoleConn].
type fileConsoleConn struct {
	f  *os.File // owns fd: Close goes through it
	fd int

	// pty marks a pseudo-terminal master: poll(2) reports POLLHUP on it
	// whenever no slave is open, which is transient and normal, so HUP/ERR
	// read as idle instead of a dead link. Only character devices get real
	// hangup detection.
	pty bool

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
	for {
		n, err := unix.Poll(fds, 0)
		if err == unix.EINTR {
			// Interrupted by a signal — Go handles signals without
			// SA_RESTART, so raw syscalls see EINTR (e.g. the runtime's
			// SIGURG async preemption); the console is fine, retry.
			continue
		}
		if err != nil {
			return 0, err
		}
		if n > 0 && fds[0].Revents&unix.POLLNVAL != 0 {
			// The fd itself is gone (device unplugged): a hard error for
			// any console.
			return 0, errors.New("console fd invalid")
		}
		if n == 0 || fds[0].Revents&unix.POLLIN == 0 {
			// Nothing readable: HUP/ERR mean the link is dead for character
			// devices — a PTY master polls POLLHUP whenever no slave is
			// open, which is transient and normal (see
			// fileConsoleConn.pty). Reporting the failure lets the
			// reconnect engine take over; ignoring it would leave the
			// payload active but silent, with status bit [5] never
			// reported. POLLIN is checked first: poll reports POLLIN and
			// POLLHUP together when the peer closed with data still
			// buffered, and that output must be drained, not thrown away.
			if n > 0 && !c.pty && fds[0].Revents&(unix.POLLERR|unix.POLLHUP) != 0 {
				return 0, errors.New("console fd hung up")
			}
			return 0, nil
		}
		break
	}
	for {
		rn, err := unix.Read(c.fd, p)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			if err == unix.EAGAIN { // drained by a concurrent reader between poll and read
				return 0, nil
			}
			// unix.Read returns -1 alongside the error; a negative count
			// violates the Reader contract, so it never escapes.
			return 0, err
		}
		return rn, nil
	}
}

func (c *fileConsoleConn) Write(p []byte) (int, error) {
	// The fd is non-blocking: a console that stops consuming (pty slave not
	// reading, serial flow control asserting stop) must not freeze the SOL
	// instance — ProcessPacket holds inst.mu during the write, and
	// teardown waits on the pump. A full buffer reads as "nothing
	// accepted": the SOL protocol models the backpressure via
	// AcceptedCharacterCount + NACK, so the remote console retries.
	n, err := unix.Write(c.fd, p)
	if err == unix.EAGAIN {
		return 0, nil
	}
	if err != nil {
		// unix.Write returns -1 alongside the error; the SOL data plane
		// casts the count to uint8 (AcceptedCharacterCount, pkg/bmc
		// ProcessPacket), and -1 would report 255 characters accepted that
		// never reached the console — silently discarding the keystrokes.
		return 0, err
	}
	return n, nil
}

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

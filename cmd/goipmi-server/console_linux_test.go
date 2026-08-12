//go:build linux

package main

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// TestReadAvailableHangup verifies a hung-up character device (PTY slave
// closed) is reported as a failure instead of an idle read: without it the
// reconnect engine never notices the dead link and the session stays
// silently mute instead of reporting status bit [5]. A PTY master, whose
// POLLHUP is a transient normal state (no slave open), must keep reading as
// idle. Output buffered before the hangup is drained first — poll reports
// POLLIN and POLLHUP together, and the data must not be thrown away.
func TestReadAvailableHangup(t *testing.T) {
	master, slavePath, err := openPTY()
	if err != nil {
		t.Fatalf("open pty: %v", err)
	}
	t.Cleanup(func() { _ = master.Close() })
	if !isPty(int(master.Fd())) {
		t.Fatal("isPty(master) = false, want true")
	}

	slave, err := os.OpenFile(slavePath, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open slave: %v", err)
	}

	// Buffered output + hangup: write to the slave, then close it — the
	// master polls POLLIN|POLLHUP.
	if _, err := slave.WriteString("E2E"); err != nil {
		t.Fatalf("write to slave: %v", err)
	}
	if err := slave.Close(); err != nil {
		t.Fatalf("close slave: %v", err)
	}

	serial := &fileConsoleConn{f: master, fd: int(master.Fd())}
	buf := make([]byte, 64)
	n, err := serial.ReadAvailable(buf)
	if err != nil || string(buf[:n]) != "E2E" {
		t.Fatalf("ReadAvailable with buffered data + hangup: n=%d data=%q err=%v, want 3/E2E/nil", n, buf[:n], err)
	}
	if _, err := serial.ReadAvailable(buf); err == nil {
		t.Fatal("ReadAvailable on hung-up console: nil error, want failure")
	}

	// PTY-master semantics: the same fd state on a fresh pair is a
	// transient idle, not a dead link.
	master2, slavePath2, err := openPTY()
	if err != nil {
		t.Fatalf("open pty: %v", err)
	}
	t.Cleanup(func() { _ = master2.Close() })
	slave2, err := os.OpenFile(slavePath2, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open slave: %v", err)
	}
	if _, err := slave2.WriteString("E2E"); err != nil {
		t.Fatalf("write to slave: %v", err)
	}
	if err := slave2.Close(); err != nil {
		t.Fatalf("close slave: %v", err)
	}
	pty := &fileConsoleConn{f: master2, fd: int(master2.Fd()), pty: true}
	n, err = pty.ReadAvailable(buf)
	if err != nil || string(buf[:n]) != "E2E" {
		t.Fatalf("pty ReadAvailable with buffered data: n=%d data=%q err=%v, want 3/E2E/nil", n, buf[:n], err)
	}
	if _, err := pty.ReadAvailable(buf); err != nil {
		t.Fatalf("pty ReadAvailable with no slave: %v, want idle", err)
	}

	// A live console with no data reads as idle, not broken.
	fresh, freshPath, err := openPTY()
	if err != nil {
		t.Fatalf("open pty: %v", err)
	}
	t.Cleanup(func() { _ = fresh.Close() })
	alive := &fileConsoleConn{f: fresh, fd: int(fresh.Fd()), pty: true}
	if _, err := alive.ReadAvailable(buf); err != nil {
		t.Fatalf("ReadAvailable on live console: %v, want nil", err)
	}
	_ = os.Remove(freshPath)
}

// TestWriteBackpressure verifies a full output buffer reads as "nothing
// accepted" (0, nil) instead of an error: ProcessPacket holds the instance
// lock during the write, so a blocking fd would freeze the whole SOL
// instance and its teardown. A socketpair stands in for the tty — a pty
// master accepts gigabytes without EAGAIN, while the socket's ~200KB send
// buffer fills deterministically; the write path (unix.Write + EAGAIN
// mapping) is fd-agnostic.
func TestWriteBackpressure(t *testing.T) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("socketpair: %v", err)
	}
	t.Cleanup(func() {
		_ = unix.Close(fds[0])
		_ = unix.Close(fds[1])
	})
	if err := unix.SetNonblock(fds[0], true); err != nil {
		t.Fatalf("set nonblock: %v", err) // like dupConsoleOpener, so a full buffer returns EAGAIN
	}
	conn := &fileConsoleConn{fd: fds[0]} // peer fds[1] open, never read

	buf := make([]byte, 65536)
	var total int
	for i := 0; i < 64; i++ {
		n, err := conn.Write(buf)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		total += n
		if n == 0 {
			break // buffer full: EAGAIN surfaced as backpressure, not an error
		}
	}
	if total == 0 {
		t.Fatal("wrote 0 bytes before backpressure")
	}

	// A failing write must report (0, err): unix.Write returns -1 alongside
	// the error, and the SOL data plane casts the count to uint8 — a -1
	// would report 255 characters accepted that never reached the console.
	_ = unix.Close(fds[1])
	n, err := conn.Write([]byte("k"))
	if err == nil {
		t.Fatalf("write after peer close: n=%d err=nil, want error", n)
	}
	if n != 0 {
		t.Fatalf("write after peer close: n=%d, want 0 (negative counts must not leak)", n)
	}
}

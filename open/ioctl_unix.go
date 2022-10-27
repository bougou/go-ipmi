//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package open

import (
	"fmt"
	"syscall"
)

func IOCTL(fd, name, data uintptr) error {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, name, data)
	if ep != 0 {
		return fmt.Errorf("syscall err: (%#02x) %s", uint8(ep), syscall.Errno(ep))
	}
	return nil
}

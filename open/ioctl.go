package open

import (
	"fmt"
	"syscall"
)

// see: https://github.com/torvalds/linux/blob/master/arch/alpha/include/uapi/asm/ioctl.h

const (
	IOC_NRBITS   = 8
	IOC_TYPEBITS = 8
	IOC_SIZEBITS = 14
	IOC_DIRBITS  = 2

	// Direction bits
	// NOTE, if IOC_DIRBITS=3, then IOC_NONE=0, IOC_READ=2, IOC_WRITE=4
	IOC_NONE  = 0x0
	IOC_READ  = 0x1
	IOC_WRITE = 0x2

	IOC_NRMASK   = ((1 << IOC_NRBITS) - 1)
	IOC_TYPEMASK = ((1 << IOC_TYPEBITS) - 1)
	IOC_SIZEMASK = ((1 << IOC_SIZEBITS) - 1)
	IOC_DIRMASK  = ((1 << IOC_DIRBITS) - 1)

	IOC_NRSHIFT   = 0
	IOC_TYPESHIFT = (IOC_NRSHIFT + IOC_NRBITS)
	IOC_SIZESHIFT = (IOC_TYPESHIFT + IOC_TYPEBITS)
	IOC_DIRSHIFT  = (IOC_SIZESHIFT + IOC_SIZEBITS)

	// ...and for the drivers/sound files...

	IOC_IN        = (IOC_WRITE << IOC_DIRSHIFT)
	IOC_OUT       = (IOC_READ << IOC_DIRSHIFT)
	IOC_INOUT     = ((IOC_WRITE | IOC_READ) << IOC_DIRSHIFT)
	IOCSIZE_MASK  = (IOC_SIZEMASK << IOC_SIZESHIFT)
	IOCSIZE_SHIFT = (IOC_SIZESHIFT)
)

func IOC(dir uintptr, typ uintptr, nr uintptr, size uintptr) uintptr {
	// 00000000  00000000  00000000  00000000
	//                               |- NR
	//                     |- TYPE
	//   |- SIZE
	// |- DIR
	return (dir << IOC_DIRSHIFT) | (typ << IOC_TYPESHIFT) | (nr << IOC_NRSHIFT) | (size << IOC_SIZESHIFT)
}

// used to create numbers
func IO(typ, nr uintptr) uintptr {
	return IOC(IOC_NONE, typ, nr, 0)
}

func IOR(typ, nr, size uintptr) uintptr {
	return IOC(IOC_READ, typ, nr, size)
}

func IOW(typ, nr, size uintptr) uintptr {
	return IOC(IOC_WRITE, typ, nr, size)
}

func IOWR(typ, nr, size uintptr) uintptr {
	return IOC(IOC_READ|IOC_WRITE, typ, nr, size)
}

// IOC_DIR is used to decode DIR from nr
func IOC_DIR(nr uintptr) uintptr {
	return (((nr) >> IOC_DIRSHIFT) & IOC_DIRMASK)
}

// IOC_TYPE is used to decode TYPE from nr
func IOC_TYPE(nr uintptr) uintptr {
	return (((nr) >> IOC_TYPESHIFT) & IOC_TYPEMASK)
}

// IOC_NR is used to decode NR from nr
func IOC_NR(nr uintptr) uintptr {
	return (((nr) >> IOC_NRSHIFT) & IOC_NRMASK)
}

// IOC_SIZE is used to decode SIZE from nr
func IOC_SIZE(nr uintptr) uintptr {
	return (((nr) >> IOC_SIZESHIFT) & IOC_SIZEMASK)
}

func IOCTL(fd, name, data uintptr) error {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, name, data)
	if ep != 0 {
		return fmt.Errorf("syscall err: (%#02x) %s", uint8(ep), syscall.Errno(ep))
	}
	return nil
}

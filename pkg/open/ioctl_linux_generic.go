//go:build linux && !(mips || mipsle || mips64 || mips64le || ppc64 || ppc64le)

package open

// Linux include/uapi/asm-generic/ioctl.h.
const (
	IOC_SIZEBITS = 14
	IOC_DIRBITS  = 2
	IOC_NONE     = 0
	IOC_READ     = 2
	IOC_WRITE    = 1
)

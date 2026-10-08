//go:build linux && (mips || mipsle || mips64 || mips64le || ppc64 || ppc64le)

package open

// Linux arch/{mips,powerpc}/include/uapi/asm/ioctl.h.
const (
	IOC_SIZEBITS = 13
	IOC_DIRBITS  = 3
	IOC_NONE     = 1
	IOC_READ     = 2
	IOC_WRITE    = 4
)

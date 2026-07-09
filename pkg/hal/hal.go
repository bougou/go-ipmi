// Package hal defines the Hardware Abstraction Layer interfaces for an IPMI BMC.
//
// Every sub-interface is optional: a HAL implementation may return nil for
// subsystems that do not exist on the target hardware.  Handlers must nil-check
// sub-interfaces and return an appropriate IPMI completion code when the
// hardware capability is absent.
//
// # Portability
//
// The interfaces in this package are deliberately free of OS-specific types.
// The only concrete Go packages used in the interface signatures are from the
// standard library and only primitives (context, error, []byte, basic types).
// This makes it possible to implement HAL for:
//   - Linux via sysfs / hwmon / i2c-dev / libgpiod (see pkg/hal/linux)
//   - Bare-metal Go / TinyGo with direct MMIO or SPI/I2C drivers
//   - Simulation / test via pkg/hal/mock
//   - Bridges to existing daemon APIs (e.g. OpenBMC D-Bus)
package hal

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/types"
)

// HAL is the top-level hardware abstraction.  Implementations may return nil
// for sub-interfaces that are not available on the target.
type HAL interface {
	// Chassis returns chassis power and identification controls, or nil.
	Chassis() ChassisHAL
	// Sensors returns the sensor reading interface, or nil.
	Sensors() SensorHAL
	// Storage returns persistent key-value storage used for SEL/SDR/FRU, or nil.
	Storage() StorageHAL
	// Network returns BMC NIC configuration, or nil.
	Network() NetworkHAL
	// GPIO returns discrete GPIO control (LEDs, buttons), or nil.
	GPIO() GPIOHAL
	// I2C returns raw I2C bus access for sensors or EEPROMs, or nil.
	I2C() I2CHAL
	// Close releases all hardware resources.
	Close() error
}

// ChassisHAL controls physical chassis power, reset, and identity.
type ChassisHAL interface {
	// PowerState returns true when the managed system is powered on.
	PowerState(ctx context.Context) (bool, error)
	// SetPower powers the managed system on or off.
	SetPower(ctx context.Context, on bool) error
	// PowerCycle performs a full power cycle of the managed system
	// (Chassis Control action 0x02, spec Table 28-3). The semantic meaning
	// for a given BMC is defined by the upper-layer HAL implementation.
	PowerCycle(ctx context.Context) error
	// ColdReset performs a hardware cold reset of the managed system.
	ColdReset(ctx context.Context) error
	// WarmReset requests an OS-level warm reboot of the managed system.
	WarmReset(ctx context.Context) error
	// Identify pulses the chassis identification LED for the given duration.
	// seconds == 0 means turn off; [ForcedIdentify] is expressed by the caller
	// passing a large value (e.g., math.MaxUint8).
	Identify(ctx context.Context, seconds uint8) error
	// IntrusionState returns true when the chassis has been opened since last reset.
	// Implementations that lack intrusion detection must return ErrNotSupported.
	IntrusionState(ctx context.Context) (bool, error)
	// SetBootFlags commits the full boot flags structure (spec Table 28-6).
	// The upper layer decides which bits it cares about; HAL implementations
	// must not silently drop fields they ignore. Implementations that do not
	// maintain boot state return ErrNotSupported.
	SetBootFlags(ctx context.Context, flags *types.BootOptionParam_BootFlags) error
	// GetBootFlags reads back the current boot flags, symmetric with
	// [SetBootFlags]. Implementations that cannot read boot flags back must
	// return ErrNotSupported; handlers translate that to the
	// CodeBootParamNotSupported completion code.
	GetBootFlags(ctx context.Context) (*types.BootOptionParam_BootFlags, error)
	// SetBootInfoAcknowledge persists the boot initiator acknowledge data
	// (spec Table 28-14, param #4).  The HAL may implement this as a no-op
	// (return nil) if it does not track boot initiator identity.
	SetBootInfoAcknowledge(ctx context.Context, ack *types.BootOptionParam_BootInfoAcknowledge) error
	// GetBootInfoAcknowledge reads back the stored acknowledge data.
	// Implementations that do not persist this return ErrNotSupported.
	GetBootInfoAcknowledge(ctx context.Context) (*types.BootOptionParam_BootInfoAcknowledge, error)
}

// SensorDescriptor describes a sensor exposed by the hardware.
type SensorDescriptor struct {
	ID   uint8
	Type uint8 // IPMI sensor type code (section 42 of IPMI spec)
	Name string
}

// SensorHAL reads hardware sensor values.
type SensorHAL interface {
	// ReadRaw returns the raw sensor byte that the BMC formula maps to a real value.
	ReadRaw(ctx context.Context, sensorID uint8) (uint8, error)
	// List returns all sensors available on the hardware.
	List(ctx context.Context) ([]SensorDescriptor, error)
}

// StorageHAL provides namespace-isolated key-value persistence.
//
// The server uses the following namespaces:
//   - "sel"    – System Event Log entries (key = decimal record ID)
//   - "sdr"    – SDR records (key = decimal record ID)
//   - "fru"    – FRU inventory (key = decimal device ID)
//   - "config" – BMC configuration (key = parameter name)
type StorageHAL interface {
	Read(ctx context.Context, namespace, key string) ([]byte, error)
	Write(ctx context.Context, namespace, key string, data []byte) error
	Delete(ctx context.Context, namespace, key string) error
	Keys(ctx context.Context, namespace string) ([]string, error)
}

// IPConfig holds the BMC network interface configuration.
type IPConfig struct {
	IP      [4]byte
	Mask    [4]byte
	Gateway [4]byte
	MAC     [6]byte
	DHCP    bool
}

// NetworkHAL configures the BMC's own network interface.
// This is separate from [transport.PacketConn]: transport is how packets arrive;
// NetworkHAL is how the LAN configuration commands read/write NIC parameters.
type NetworkHAL interface {
	GetConfig(ctx context.Context) (*IPConfig, error)
	SetConfig(ctx context.Context, cfg *IPConfig) error
}

// GPIOHAL provides access to discrete GPIO lines (status LEDs, front-panel buttons).
type GPIOHAL interface {
	// Set drives an output GPIO high (true) or low (false).
	Set(ctx context.Context, pin string, high bool) error
	// Get reads the current level of an input GPIO.
	Get(ctx context.Context, pin string) (bool, error)
	// Watch calls callback whenever the input level changes.
	// The returned cancel function stops watching.
	Watch(ctx context.Context, pin string, callback func(high bool)) (cancel func(), err error)
}

// I2CHAL provides raw I2C bus access for sensors or EEPROM-backed FRU devices.
type I2CHAL interface {
	// Read performs a register read from an I2C device.
	Read(ctx context.Context, bus int, addr uint8, reg uint8, length int) ([]byte, error)
	// Write performs a register write to an I2C device.
	Write(ctx context.Context, bus int, addr uint8, reg uint8, data []byte) error
}

// ErrNotSupported is returned by HAL methods when the hardware does not
// support the requested operation.  Handlers translate this to an appropriate
// IPMI completion code: CodeBootParamNotSupported 0x80 (§28.12/§28.13) for
// parameter-level operations, CodeUnspecifiedError 0xFF (§5.2 Table 5-2)
// via codeFromErr for general error paths.
var ErrNotSupported = errNotSupported{}

type errNotSupported struct{}

func (errNotSupported) Error() string { return "operation not supported by hardware" }

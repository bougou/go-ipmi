package ipmi

// see: https://community.infineon.com/t5/Knowledge-Base-Articles/Difference-between-7-bit-vs-8-bit-I2C-addressing/ta-p/798072
//
// I2CAddress7Bit is a 7-bit I2C address, the bit 0 to bit 6 take effect, bit 7 is always 0.
//
// eg:  `var a I2CAddress7Bit = 0x35`
//   - 0 011 0101  = 0x35
//   - 7 654 3210  bit position
type I2CAddress7Bit uint8

// I2CAddress8Bit represents 8-bit I2C address, the 7-bit I2CAddress occupies bit 1 to bit 7, bit 0 indicates for write or read.
//
// eg:  `var a I2CAddress7Bit = 0x35`
//   - 0110 101 0  = 0x6A (for write)
//   - 0110 101 1  = 0x6B (fore read)
//   - 7654 321 0  bit position
type I2CAddress8Bit uint8

// To8BitForWrite convert 7-bit I2C address to 8-bit I2C address, bit 1 to bit 7 for I2C address, bit 0 set to 0
func (a I2CAddress7Bit) To8BitForWrite() I2CAddress8Bit {
	return I2CAddress8Bit(a << 1)
}

// To8BitForRead convert 7-bit I2C address to 8-bit I2C address, bit 1 to bit 7 for I2C address, bit 0 set to 1
func (a I2CAddress7Bit) To8BitForRead() I2CAddress8Bit {
	return I2CAddress8Bit(a<<1 | 0x01)
}

func (a I2CAddress8Bit) To7Bit() I2CAddress7Bit {
	return I2CAddress7Bit(a >> 1)
}

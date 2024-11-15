package ipmi

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/kr/pretty"
)

const timeFormat = time.RFC3339

func bytesForPrint(data []byte) string {
	out := ""
	for k, v := range data {
		if k%8 == 0 && k != 0 {
			out += "\n"
		}
		out += fmt.Sprintf("%02x ", v)
	}
	out += "\n"
	return out
}

func debugBytes(header string, data []byte, width int) {
	fmt.Printf("%s (%d bytes)\n", header, len(data))
	for k, v := range data {
		if k%width == 0 && k != 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%02x ", v)
	}
	fmt.Printf("\n")
}

// debugf pretty print any object
func debugf(format string, object ...interface{}) {
	pretty.Printf(format, object...)
}

func debug(header string, object interface{}) {
	if header == "" {
		pretty.Printf("%# v\n", object)
	} else {
		pretty.Printf("%s: \n%# v\n", header, object)
	}
}

func (c *Client) Debugf(format string, object ...interface{}) {
	if !c.debug {
		return
	}
	debugf(format, object...)
}

func (c *Client) Debug(header string, object interface{}) {
	if !c.debug {
		return
	}
	debug(header, object)
}

// DebugBytes print byte slices with a fixed width of bytes on each line.
func (c *Client) DebugBytes(header string, data []byte, width int) {
	if !c.debug {
		return
	}
	debugBytes(header, data, width)
}

func (c *Client) DebugfRed(format string, object ...interface{}) {
	if !c.debug {
		return
	}
	const colorRed = "\033[0;31m"
	fmt.Printf(colorRed+format+"\033[0m", object...)
}

func (c *Client) DebugfGreen(format string, object ...interface{}) {
	if !c.debug {
		return
	}
	const colorGreen = "\033[0;32m"
	fmt.Printf(colorGreen+format+"\033[0m", object...)
}

func (c *Client) DebugfYellow(format string, object ...interface{}) {
	if !c.debug {
		return
	}
	const colorYellow = "\033[0;33m"
	fmt.Printf(colorYellow+format+"\033[0m", object...)
}

// 37 Timestamp Format
func parseTimestamp(timestamp uint32) time.Time {
	return time.Unix(int64(timestamp), 0)
}

func formatBool(b bool, trueStr string, falseStr string) string {
	if b {
		return trueStr
	}
	return falseStr
}

// padBytes will padding the origin "s" string to fixed "width" length,
// with "pad" as the padding byte.
func padBytes(s string, width int, pad byte) []byte {
	o := []byte(s)
	if len(s) >= width {
		return o[:width]
	}

	for i := 0; i < width-len(s); i++ {
		o = append(o, pad)
	}
	return o
}

func isByteSliceEqual(b1 []byte, b2 []byte) bool {
	// not equal if both are nil
	if b1 == nil || b2 == nil {
		return false
	}
	if len(b1) != len(b2) {
		return false
	}
	for k := range b1 {
		if b1[k] != b2[k] {
			return false
		}
	}
	return true
}

func array16(s []byte) [16]byte {
	var out [16]byte
	copy(out[:16], s[:])
	return out
}

func randomUint32() uint32 {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return r.Uint32()
}

func randomBytes(n int) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	b := make([]byte, n)
	r.Read(b)
	return b
}

// onesComplement returns the signed integer of the input number encoded with 1's complement.
// The lowest significant 'bitSize' bits of the input number i is considered.
func onesComplement(i uint32, bitSize uint8) int32 {
	var leftBitSize uint8 = 32 - bitSize
	var temp uint32 = i << uint32(leftBitSize) >> uint32(leftBitSize)

	var mask uint32 = 1 << (bitSize - 1)
	if temp&mask == 0 {
		// means the bit at `bitSize-1` (from left starting at 0) is 0
		// so the result should be a positive value
		return int32(temp)
	}

	// means the bit at `bitSize-1` (from left starting at 0) is 1
	// so the result should be a negative value
	t := temp ^ 0xffff
	t = t << uint32(leftBitSize) >> uint32(leftBitSize)
	return -int32(t)
}

// twosComplement returns the signed integer of the input number encoded with 2's complement.
// The lowest significant 'bitSize' bits of the input number i is considered.
func twosComplement(i uint32, bitSize uint8) int32 {
	var leftBitSize uint8 = 32 - bitSize
	var temp uint32 = i << uint32(leftBitSize) >> uint32(leftBitSize)

	var mask uint32 = 1 << (bitSize - 1)
	if temp&mask == 0 {
		// means the bit at `bitSize-1` (from left starting at 0) is 0
		// so the result should be a positive value
		return int32(temp)
	}

	// means the bit at `bitSize-1` (from left starting at 0) is 1
	// so the result should be a negative value
	t := temp ^ 0xffff + 1
	t = t << uint32(leftBitSize) >> uint32(leftBitSize)
	return -int32(t)
}

func onesComplementEncode(i int32, bitSize uint8) uint32 {
	if i >= 0 {
		return uint32(i)
	}
	var total int32 = int32(math.Pow(2, float64(bitSize))) - 1
	return uint32(total + i)
}

func twosComplementEncode(i int32, bitSize uint8) uint32 {
	if i >= 0 {
		return uint32(i)
	}
	var total int32 = int32(math.Pow(2, float64(bitSize)))
	return uint32(total + i)
}

// The TCP/IP standard network byte order is big-endian.

var base32HexNoPadEncoding = base32.HexEncoding.WithPadding(base32.NoPadding)

func fromBase32(s []byte) (buf []byte, err error) {
	for i, b := range s {
		if b >= 'a' && b <= 'z' {
			s[i] = b - 32
		}
	}
	buflen := base32HexNoPadEncoding.DecodedLen(len(s))
	buf = make([]byte, buflen)
	n, err := base32HexNoPadEncoding.Decode(buf, s)
	buf = buf[:n]
	return
}

func toBase32(b []byte) string {
	return base32HexNoPadEncoding.EncodeToString(b)
}

func fromBase64(s []byte) (buf []byte, err error) {
	buflen := base64.StdEncoding.DecodedLen(len(s))
	buf = make([]byte, buflen)
	n, err := base64.StdEncoding.Decode(buf, s)
	buf = buf[:n]
	return
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func setBit7(b uint8) uint8 {
	return b | 0x80
}

func setBit6(b uint8) uint8 {
	return b | 0x40
}

func setBit5(b uint8) uint8 {
	return b | 0x20
}

func setBit4(b uint8) uint8 {
	return b | 0x10
}

func setBit3(b uint8) uint8 {
	return b | 0x08
}

func setBit2(b uint8) uint8 {
	return b | 0x04
}

func setBit1(b uint8) uint8 {
	return b | 0x02
}

func setBit0(b uint8) uint8 {
	return b | 0x01
}

func clearBit7(b uint8) uint8 {
	return b & 0x7f
}

func clearBit6(b uint8) uint8 {
	return b & 0xbf
}

func clearBit5(b uint8) uint8 {
	return b & 0xdf
}

func clearBit4(b uint8) uint8 {
	return b & 0xef
}

func clearBit3(b uint8) uint8 {
	return b & 0xf7
}

func clearBit2(b uint8) uint8 {
	return b & 0xfb
}

func clearBit1(b uint8) uint8 {
	return b & 0xfd
}

func clearBit0(b uint8) uint8 {
	return b & 0xfe
}
func isBit7Set(b uint8) bool {
	return b&0x80 == 0x80
}

func isBit6Set(b uint8) bool {
	return b&0x40 == 0x40
}

func isBit5Set(b uint8) bool {
	return b&0x20 == 0x20
}

func isBit4Set(b uint8) bool {
	return b&0x10 == 0x10
}

func isBit3Set(b uint8) bool {
	return b&0x08 == 0x08
}

func isBit2Set(b uint8) bool {
	return b&0x04 == 0x04
}

func isBit1Set(b uint8) bool {
	return b&0x02 == 0x02
}

func isBit0Set(b uint8) bool {
	return b&0x01 == 0x01
}

func unpackUint8(msg []byte, off int) (uint8, int, error) {
	if off+1 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint8")
	}
	return msg[off], off + 1, nil
}

// packUint8 fills an uint8 value (i) into byte slice (msg) at the index of (offset)
func packUint8(i uint8, msg []byte, off int) (int, error) {
	if off+1 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint8")
	}
	msg[off] = i
	return off + 1, nil
}

func packBytes(v []byte, msg []byte, off int) (int, error) {
	if off+len(v) > len(msg) {
		return len(msg), fmt.Errorf("overflow packing byte slice")
	}
	for k, b := range v {
		msg[off+k] = b
	}
	return off + len(v), nil
}

func unpackBytes(msg []byte, off int, length int) ([]byte, int, error) {
	out := []byte{}
	if off+length > len(msg) {
		return out, off, fmt.Errorf("overflow unpacking %d bytes", length)
	}

	out = append(out, msg[off:off+length]...)
	return out, off + length, nil
}

// unpackBytesMost unpacks most length of bytes from msg starting from off index.
// It stops when reaching the msg end or reaching the most length.
// This functions never failed, it always return nil error.
// The caller should check the length of the returned out byte
func unpackBytesMost(msg []byte, off int, length int) ([]byte, int, error) {
	out := make([]byte, length)
	var i int = 0
	for ; i < length; i++ {
		if off+i >= len(msg) {
			break
		}
		out[i] = msg[off+i]
	}

	return out[:i], off + len(out), nil

}

func unpackUint16(msg []byte, off int) (uint16, int, error) {
	if off+2 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.BigEndian.Uint16(msg[off:]), off + 2, nil
}

func unpackUint16L(msg []byte, off int) (uint16, int, error) {
	if off+2 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.LittleEndian.Uint16(msg[off:]), off + 2, nil
}

func packUint16(i uint16, msg []byte, off int) (int, error) {
	if off+2 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint16")
	}
	binary.BigEndian.PutUint16(msg[off:], i)
	return off + 2, nil
}

func packUint16L(i uint16, msg []byte, off int) (int, error) {
	if off+2 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint16")
	}
	binary.LittleEndian.PutUint16(msg[off:], i)
	return off + 2, nil
}

func unpackUint24(msg []byte, off int) (uint32, int, error) {
	if off+3 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32 as uint24")
	}
	i := uint32(msg[off])<<16 | uint32(msg[off+1])<<8 | uint32(msg[off+2])
	off += 3
	return i, off, nil
}

func unpackUint24L(msg []byte, off int) (uint32, int, error) {
	if off+3 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32 as uint24")
	}
	i := uint32(msg[off]) | uint32(msg[off+1])<<8 | uint32(msg[off+2])<<16
	off += 3
	return i, off, nil
}

func packUint24(i uint32, msg []byte, off int) (int, error) {
	if off+3 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32 as uint24")
	}
	msg[off] = byte(i >> 16)
	msg[off+1] = byte(i >> 8)
	msg[off+2] = byte(i)
	off += 3
	return off, nil
}

func packUint24L(i uint32, msg []byte, off int) (int, error) {
	if off+3 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32 as uint24")
	}
	msg[off] = byte(i)
	msg[off+1] = byte(i >> 8)
	msg[off+2] = byte(i >> 16)
	off += 3
	return off, nil
}

func unpackUint32(msg []byte, off int) (uint32, int, error) {
	if off+4 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32")
	}
	return binary.BigEndian.Uint32(msg[off:]), off + 4, nil
}

func unpackUint32L(msg []byte, off int) (uint32, int, error) {
	if off+4 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32")
	}
	return binary.LittleEndian.Uint32(msg[off:]), off + 4, nil
}

func packUint32(i uint32, msg []byte, off int) (int, error) {
	if off+4 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32")
	}
	binary.BigEndian.PutUint32(msg[off:], i)
	return off + 4, nil
}

func packUint32L(i uint32, msg []byte, off int) (int, error) {
	if off+4 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32")
	}
	binary.LittleEndian.PutUint32(msg[off:], i)
	return off + 4, nil
}

func unpackUint48(msg []byte, off int) (uint64, int, error) {
	if off+6 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64 as uint48")
	}
	i := uint64(msg[off])<<40 | uint64(msg[off+1])<<32 | uint64(msg[off+2])<<24 | uint64(msg[off+3])<<16 |
		uint64(msg[off+4])<<8 | uint64(msg[off+5])
	off += 6
	return i, off, nil
}

func unpackUint48L(msg []byte, off int) (uint64, int, error) {
	if off+6 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64 as uint48")
	}
	i := uint64(msg[off]) | uint64(msg[off+1])<<8 | uint64(msg[off+2])<<16 | uint64(msg[off+3])<<24 |
		uint64(msg[off+4])<<32 | uint64(msg[off+5])<<40
	off += 6
	return i, off, nil
}

func packUint48(i uint64, msg []byte, off int) (int, error) {
	if off+6 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64 as uint48")
	}
	msg[off] = byte(i >> 40)
	msg[off+1] = byte(i >> 32)
	msg[off+2] = byte(i >> 24)
	msg[off+3] = byte(i >> 16)
	msg[off+4] = byte(i >> 8)
	msg[off+5] = byte(i)
	off += 6
	return off, nil
}

func packUint48L(i uint64, msg []byte, off int) (int, error) {
	if off+6 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64 as uint48")
	}
	msg[off] = byte(i)
	msg[off+1] = byte(i >> 8)
	msg[off+2] = byte(i >> 16)
	msg[off+3] = byte(i >> 24)
	msg[off+4] = byte(i >> 32)
	msg[off+5] = byte(i >> 40)
	off += 6
	return off, nil
}

func unpackUint64(msg []byte, off int) (uint64, int, error) {
	if off+8 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64")
	}
	return binary.BigEndian.Uint64(msg[off:]), off + 8, nil
}

func unpackUint64L(msg []byte, off int) (uint64, int, error) {
	if off+8 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64")
	}
	return binary.LittleEndian.Uint64(msg[off:]), off + 8, nil
}

func packUint64(i uint64, msg []byte, off int) (int, error) {
	if off+8 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64")
	}
	binary.BigEndian.PutUint64(msg[off:], i)
	off += 8
	return off, nil
}

func packUint64L(i uint64, msg []byte, off int) (int, error) {
	if off+8 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64")
	}
	binary.LittleEndian.PutUint64(msg[off:], i)
	off += 8
	return off, nil
}

// 8421 BCD
// bcdUint8 decodes BCD encoded integer to normal unsigned integer.
func bcdUint8(i uint8) uint8 {
	msb4 := i >> 4
	lsb4 := i & 0x0f
	return msb4*10 + lsb4
}

func parseStringToInt64(s string) (int64, error) {
	if len(s) > 2 {
		if s[0] == '0' {
			return strconv.ParseInt(s, 0, 64)
		}
	}
	return strconv.ParseInt(s, 10, 64)
}

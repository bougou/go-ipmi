package types

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"iter"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

const TimeFormat = time.RFC3339

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
func debugf(format string, object ...any) {
	pretty.Printf(format, object...)
}

func debug(header string, object any) {
	if header == "" {
		pretty.Printf("%# v\n", object)
	} else {
		pretty.Printf("%s: \n%# v\n", header, object)
	}
}

// 37 Timestamp Format
func ParseTimestamp(timestamp uint32) time.Time {
	return time.Unix(int64(timestamp), 0)
}

func FormatBool(b bool, trueStr string, falseStr string) string {
	if b {
		return trueStr
	}
	return falseStr
}

// padBytes will padding the origin "s" string to fixed "width" length,
// with "pad" as the padding byte.
func PadBytes(s string, width int, pad byte) []byte {
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

func Array16(s []byte) [16]byte {
	var out [16]byte
	copy(out[:16], s[:])
	return out
}

func randomUint32() uint32 {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return r.Uint32()
}

func RandomBytes(n int) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	b := make([]byte, n)
	r.Read(b)
	return b
}

// onesComplement returns the signed integer of the input number encoded with 1's complement.
// The lowest significant 'bitSize' bits of the input number i is considered.
func OneSComplement(i uint32, bitSize uint8) int32 {
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
func TwoSComplement(i uint32, bitSize uint8) int32 {
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

func OneSComplementEncode(i int32, bitSize uint8) uint32 {
	if i >= 0 {
		return uint32(i)
	}
	var total int32 = int32(math.Pow(2, float64(bitSize))) - 1
	return uint32(total + i)
}

func TwoSComplementEncode(i int32, bitSize uint8) uint32 {
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

func SetBit7(b uint8) uint8 {
	return b | 0x80
}

func SetBit6(b uint8) uint8 {
	return b | 0x40
}

func SetBit5(b uint8) uint8 {
	return b | 0x20
}

func SetBit4(b uint8) uint8 {
	return b | 0x10
}

func SetBit3(b uint8) uint8 {
	return b | 0x08
}

func SetBit2(b uint8) uint8 {
	return b | 0x04
}

func SetBit1(b uint8) uint8 {
	return b | 0x02
}

func SetBit0(b uint8) uint8 {
	return b | 0x01
}

func ClearBit7(b uint8) uint8 {
	return b & 0x7f
}

func ClearBit6(b uint8) uint8 {
	return b & 0xbf
}

func ClearBit5(b uint8) uint8 {
	return b & 0xdf
}

func ClearBit4(b uint8) uint8 {
	return b & 0xef
}

func ClearBit3(b uint8) uint8 {
	return b & 0xf7
}

func ClearBit2(b uint8) uint8 {
	return b & 0xfb
}

func ClearBit1(b uint8) uint8 {
	return b & 0xfd
}

func ClearBit0(b uint8) uint8 {
	return b & 0xfe
}

func SetOrClearBit7(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x80
	}
	return b & 0x7f
}

func SetOrClearBit6(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x40

	}
	return b & 0xbf
}

func SetOrClearBit5(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x20
	}
	return b & 0xdf
}

func SetOrClearBit4(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x10
	}
	return b & 0xef
}

func SetOrClearBit3(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x08

	}
	return b & 0xf7
}

func SetOrClearBit2(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x04
	}
	return b & 0xfb
}

func SetOrClearBit1(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x02
	}
	return b & 0xfd
}

func SetOrClearBit0(b uint8, cond bool) uint8 {
	if cond {
		return b | 0x01
	}
	return b & 0xfe
}

func IsBit7Set(b uint8) bool {
	return b&0x80 == 0x80
}

func IsBit6Set(b uint8) bool {
	return b&0x40 == 0x40
}

func IsBit5Set(b uint8) bool {
	return b&0x20 == 0x20
}

func IsBit4Set(b uint8) bool {
	return b&0x10 == 0x10
}

func IsBit3Set(b uint8) bool {
	return b&0x08 == 0x08
}

func IsBit2Set(b uint8) bool {
	return b&0x04 == 0x04
}

func IsBit1Set(b uint8) bool {
	return b&0x02 == 0x02
}

func IsBit0Set(b uint8) bool {
	return b&0x01 == 0x01
}

func UnpackUint8(msg []byte, off int) (uint8, int, error) {
	if off+1 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint8")
	}
	return msg[off], off + 1, nil
}

// packUint8 fills an uint8 value (i) into byte slice (msg) at the index of (offset)
func PackUint8(i uint8, msg []byte, off int) (int, error) {
	if off+1 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint8")
	}
	msg[off] = i
	return off + 1, nil
}

func PackBytes(v []byte, msg []byte, off int) (int, error) {
	if off+len(v) > len(msg) {
		return len(msg), fmt.Errorf("overflow packing byte slice")
	}
	for k, b := range v {
		msg[off+k] = b
	}
	return off + len(v), nil
}

func UnpackBytes(msg []byte, off int, length int) ([]byte, int, error) {
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
func UnpackBytesMost(msg []byte, off int, length int) ([]byte, int, error) {
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

func UnpackUint16(msg []byte, off int) (uint16, int, error) {
	if off+2 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.BigEndian.Uint16(msg[off:]), off + 2, nil
}

func UnpackUint16L(msg []byte, off int) (uint16, int, error) {
	if off+2 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.LittleEndian.Uint16(msg[off:]), off + 2, nil
}

func PackUint16(i uint16, msg []byte, off int) (int, error) {
	if off+2 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint16")
	}
	binary.BigEndian.PutUint16(msg[off:], i)
	return off + 2, nil
}

func PackUint16L(i uint16, msg []byte, off int) (int, error) {
	if off+2 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint16")
	}
	binary.LittleEndian.PutUint16(msg[off:], i)
	return off + 2, nil
}

func UnpackUint24(msg []byte, off int) (uint32, int, error) {
	if off+3 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32 as uint24")
	}
	i := uint32(msg[off])<<16 | uint32(msg[off+1])<<8 | uint32(msg[off+2])
	off += 3
	return i, off, nil
}

func UnpackUint24L(msg []byte, off int) (uint32, int, error) {
	if off+3 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32 as uint24")
	}
	i := uint32(msg[off]) | uint32(msg[off+1])<<8 | uint32(msg[off+2])<<16
	off += 3
	return i, off, nil
}

func PackUint24(i uint32, msg []byte, off int) (int, error) {
	if off+3 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32 as uint24")
	}
	msg[off] = byte(i >> 16)
	msg[off+1] = byte(i >> 8)
	msg[off+2] = byte(i)
	off += 3
	return off, nil
}

func PackUint24L(i uint32, msg []byte, off int) (int, error) {
	if off+3 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32 as uint24")
	}
	msg[off] = byte(i)
	msg[off+1] = byte(i >> 8)
	msg[off+2] = byte(i >> 16)
	off += 3
	return off, nil
}

func UnpackUint32(msg []byte, off int) (uint32, int, error) {
	if off+4 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32")
	}
	return binary.BigEndian.Uint32(msg[off:]), off + 4, nil
}

func UnpackUint32L(msg []byte, off int) (uint32, int, error) {
	if off+4 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint32")
	}
	return binary.LittleEndian.Uint32(msg[off:]), off + 4, nil
}

func PackUint32(i uint32, msg []byte, off int) (int, error) {
	if off+4 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32")
	}
	binary.BigEndian.PutUint32(msg[off:], i)
	return off + 4, nil
}

func PackUint32L(i uint32, msg []byte, off int) (int, error) {
	if off+4 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint32")
	}
	binary.LittleEndian.PutUint32(msg[off:], i)
	return off + 4, nil
}

func UnpackUint48(msg []byte, off int) (uint64, int, error) {
	if off+6 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64 as uint48")
	}
	i := uint64(msg[off])<<40 | uint64(msg[off+1])<<32 | uint64(msg[off+2])<<24 | uint64(msg[off+3])<<16 |
		uint64(msg[off+4])<<8 | uint64(msg[off+5])
	off += 6
	return i, off, nil
}

func UnpackUint48L(msg []byte, off int) (uint64, int, error) {
	if off+6 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64 as uint48")
	}
	i := uint64(msg[off]) | uint64(msg[off+1])<<8 | uint64(msg[off+2])<<16 | uint64(msg[off+3])<<24 |
		uint64(msg[off+4])<<32 | uint64(msg[off+5])<<40
	off += 6
	return i, off, nil
}

func PackUint48(i uint64, msg []byte, off int) (int, error) {
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

func PackUint48L(i uint64, msg []byte, off int) (int, error) {
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

func UnpackUint64(msg []byte, off int) (uint64, int, error) {
	if off+8 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64")
	}
	return binary.BigEndian.Uint64(msg[off:]), off + 8, nil
}

func UnpackUint64L(msg []byte, off int) (uint64, int, error) {
	if off+8 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint64")
	}
	return binary.LittleEndian.Uint64(msg[off:]), off + 8, nil
}

func PackUint64(i uint64, msg []byte, off int) (int, error) {
	if off+8 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64")
	}
	binary.BigEndian.PutUint64(msg[off:], i)
	off += 8
	return off, nil
}

func PackUint64L(i uint64, msg []byte, off int) (int, error) {
	if off+8 > len(msg) {
		return len(msg), fmt.Errorf("overflow packing uint64")
	}
	binary.LittleEndian.PutUint64(msg[off:], i)
	off += 8
	return off, nil
}

// 8421 BCD
// bcdUint8 decodes BCD encoded integer to normal unsigned integer.
func BCDUint8(i uint8) uint8 {
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

// RenderTable formats a table from a slice of rows.
// The Headers is a slice of strings, the order of the headers is the order of the columns in the table.
// Each row is represented as a map, the keys of the map are the headers of the table.
func RenderTable(headers []string, rows []map[string]string) string {
	var buf = new(bytes.Buffer)

	table := tablewriter.NewTable(buf)

	table.Options(
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignRight),
		tablewriter.WithRowAutoWrap(0), // Disable auto wrap
		tablewriter.WithFooterAlignmentConfig(tw.CellAlignment{
			Global: tw.AlignCenter, // Center align footer like header
		}),
		tablewriter.WithFooterAutoFormat(tw.On),
	)

	table.Header(headers)
	table.Footer(headers)

	// Add rows
	for _, _row := range rows {
		row := make([]any, len(headers))
		for i, header := range headers {
			row[i] = _row[header]
		}
		table.Append(row)
	}

	table.Render()
	return buf.String()
}

func RenderTableStream(headers []string, rowSeq iter.Seq[map[string]string]) error {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	table.Options(
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignRight),
		tablewriter.WithRowAutoWrap(0), // Disable auto wrap
		tablewriter.WithFooterAlignmentConfig(tw.CellAlignment{
			Global: tw.AlignCenter, // Center align footer like header
		}),
		tablewriter.WithFooterAutoFormat(tw.On),
	)

	// Start streaming
	if err := table.Start(); err != nil {
		return fmt.Errorf("table start failed: %w", err)
	}
	defer table.Close()

	table.Header(headers)

	for _row := range rowSeq {
		row := make([]any, len(headers))
		for i, header := range headers {
			canonicalHeader := strings.TrimSpace(header)
			canonicalHeader = strings.Trim(canonicalHeader, "-")
			canonicalHeader = strings.TrimSpace(canonicalHeader)
			row[i] = _row[canonicalHeader]
		}
		table.Append(row)
	}

	table.Footer(headers)

	return nil

}

type itemToRowFn[T any] func(item *T, options ...any) map[string]string

func formatStream[T any](seq iter.Seq[*Result[T]], headers []string, itemToRowFn itemToRowFn[T], itemToRowFnOptions ...any) error {
	var resultErr error

	// convert channel to sequence.
	rowSeq := func(seq iter.Seq[*Result[T]]) iter.Seq[map[string]string] {
		return func(yield func(map[string]string) bool) {
			for result := range seq {
				if result == nil {
					continue
				}

				if result.Err != nil {
					resultErr = result.Err
					return
				}

				item := result.Ok
				if item != nil {
					row := itemToRowFn(item, itemToRowFnOptions...)
					if !yield(row) {
						return
					}
				}
			}
		}
	}

	if err := RenderTableStream(headers, rowSeq(seq)); err != nil {
		return fmt.Errorf("render table stream failed, err: %w", resultErr)
	}

	if resultErr != nil {
		return fmt.Errorf("got result error, err: %w", resultErr)
	}

	return nil
}

// buildCanIgnoreFn returns a `canIgnore` function that can be used to check if a err
// is a ResponseError with CompletionCode in specified codes.
// If so, the `canIgnore` function returns nil, otherwise it returns the original err.
func buildCanIgnoreFn(codes ...uint8) func(err error) error {
	return func(err error) error {
		if isErrOfCompletionCodes(err, codes...) {
			return nil
		}
		return err
	}
}

func isErrOfCompletionCodes(err error, codes ...uint8) bool {
	if err == nil {
		return false
	}

	if respErr, ok := IsResponseError(err); ok {
		cc := respErr.CompletionCode()
		for _, code := range codes {
			if uint8(cc) == code {
				return true
			}
		}
	}

	return false
}

// Generic function to convert a slice of any type to a slice of any
func convertToInterfaceSlice[T any](input []T) []any {
	result := make([]any, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

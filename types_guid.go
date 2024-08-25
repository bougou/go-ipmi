package ipmi

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GUIDMode is the way how to decode the 16 bytes GUID
type GUIDMode uint8

const (
	GUIDModeRFC4122 GUIDMode = iota
	GUIDModeIPMI
	GUIDModeSMBIOS
)

func (guidMode GUIDMode) String() string {
	m := map[GUIDMode]string{
		GUIDModeRFC4122: "RFC4122",
		GUIDModeIPMI:    "IPMI",
		GUIDModeSMBIOS:  "SMBIOS",
	}

	if s, ok := m[guidMode]; ok {
		return s
	}

	return ""
}

// ParseGUID parses the raw guid data with the specified encoding mode.
// Different GUIDMode would interpret the [16]byte data into different layout of uuid.
//
// see: https://github.com/ipmitool/ipmitool/issues/25
func ParseGUID(data []byte, guidMode GUIDMode) (*uuid.UUID, error) {
	if len(data) != 16 {
		return nil, fmt.Errorf("the length of GUID data must be 16 (%d)", len(data))
	}

	d := array16(data)

	switch guidMode {
	case GUIDModeRFC4122:
		return parseGUID_RFC4122(d)

	case GUIDModeIPMI:
		return parseGUID_IPMI(d)

	case GUIDModeSMBIOS:
		return parseGUID_SMBIOS(d)

	default:
		return nil, fmt.Errorf("unknown GUIDMode: (%s)", guidMode)
	}
}

func parseGUID_RFC4122(data [16]byte) (*uuid.UUID, error) {
	u, err := uuid.FromBytes(data[:])
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// see: 20.8 Get Device GUID Command
func parseGUID_IPMI(data [16]byte) (*uuid.UUID, error) {
	var rfc4122Data [16]byte

	// time_low
	rfc4122Data[0] = data[15]
	rfc4122Data[1] = data[14]
	rfc4122Data[2] = data[13]
	rfc4122Data[3] = data[12]

	// time_mid
	rfc4122Data[4] = data[11]
	rfc4122Data[5] = data[10]

	// time_high_and_version
	rfc4122Data[6] = data[9]
	rfc4122Data[7] = data[8]

	// clock_seq_hi_and_reserved
	rfc4122Data[8] = data[7]

	// clock_seq_low
	rfc4122Data[9] = data[6]

	// node
	rfc4122Data[10] = data[5]
	rfc4122Data[11] = data[4]
	rfc4122Data[12] = data[3]
	rfc4122Data[13] = data[2]
	rfc4122Data[14] = data[1]
	rfc4122Data[15] = data[0]

	return parseGUID_RFC4122(rfc4122Data)
}

// https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.2.0.pdf
//
//	921 7.2.1 System — UUID
//
//	928 Although RFC4122 recommends network byte order for all fields, the PC industry (including the ACPI,
//	929 UEFI, and Microsoft specifications) has consistently used little-endian byte encoding for the first three
//	930 fields: time_low, time_mid, time_hi_and_version. The same encoding, also known as wire format, should
//	931 also be used for the SMBIOS representation of the UUID.
//	932 The UUID {00112233-4455-6677-8899-AABBCCDDEEFF} would thus be represented as:
//	933 33 22 11 00 55 44 77 66 88 99 AA BB CC DD EE FF.
//	934 If the value is all FFh, the ID is not currently present in the system, but it can be set. If the value is all 00h,
//	935 the ID is not present in the system.
func parseGUID_SMBIOS(data [16]byte) (*uuid.UUID, error) {
	var rfc4122Data [16]byte

	// time_low
	rfc4122Data[0] = data[3]
	rfc4122Data[1] = data[2]
	rfc4122Data[2] = data[1]
	rfc4122Data[3] = data[0]

	// time_mid
	rfc4122Data[4] = data[5]
	rfc4122Data[5] = data[4]

	// time_hi_and_version
	rfc4122Data[6] = data[7]
	rfc4122Data[7] = data[6]

	// clock_seq_hi_and_reserved
	rfc4122Data[8] = data[8]

	// clock_seq_low
	rfc4122Data[9] = data[9]

	// node
	rfc4122Data[10] = data[10]
	rfc4122Data[11] = data[11]
	rfc4122Data[12] = data[12]
	rfc4122Data[13] = data[13]
	rfc4122Data[14] = data[14]
	rfc4122Data[15] = data[15]

	return parseGUID_RFC4122(rfc4122Data)
}

// see https://github.com/ipmitool/ipmitool/issues/25#issuecomment-409703163
func IPMILegacyGUIDTime(u *uuid.UUID) time.Time {
	sec := int64(binary.BigEndian.Uint32(u[0:4]))
	return time.Unix(sec, 0)
}

// see: https://uuid.ramsey.dev/en/stable/rfc4122.html
func UUIDVersionString(u *uuid.UUID) string {
	v := u.Version()

	var name string
	switch v {
	case 1:
		name = "Time-based Gregorian Time"
	case 2:
		name = "Time-based DCE Security with POSIX UIDs"
	case 3:
		name = "Name-based MD5"
	case 4:
		name = "Random"
	case 5:
		name = "Name-based SHA-1"
	case 6:
		name = "Reordered Time"
	case 7:
		name = "Unix Epoch Time"
	case 8:
		name = "Custom"
	default:
		name = "Unknown"
	}

	return fmt.Sprintf("%s (%s)", v.String(), name)
}

func ShowDetailGUID(guid [16]byte) string {
	formatGUID := func(u *uuid.UUID, mode GUIDMode) string {
		out := fmt.Sprintf("GUID              : %s\n", u.String())
		out += fmt.Sprintf("UUID Encoding     : %s\n", mode)
		out += fmt.Sprintf("UUID Version      : %s\n", UUIDVersionString(u))
		out += fmt.Sprintf("UUID Variant      : %s\n", u.Variant().String())
		sec, nsec := u.Time().UnixTime()
		out += fmt.Sprintf("Timestamp         : %s\n", time.Unix(sec, nsec).Format(timeFormat))
		out += fmt.Sprintf("Timestamp(Legacy) : %s", IPMILegacyGUIDTime(u).Format(timeFormat))
		return out
	}

	out := ""

	u, err := ParseGUID(guid[:], GUIDModeSMBIOS)
	if err != nil {
		return ""
	}
	out += formatGUID(u, GUIDModeSMBIOS)

	u, err = ParseGUID(guid[:], GUIDModeIPMI)
	if err != nil {
		return ""
	}
	out += "\n"
	out += formatGUID(u, GUIDModeIPMI)

	u, err = ParseGUID(guid[:], GUIDModeRFC4122)
	if err != nil {
		return ""
	}
	out += "\n"
	out += formatGUID(u, GUIDModeRFC4122)

	return out
}

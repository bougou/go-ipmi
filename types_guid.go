package ipmi

import (
	"fmt"

	"github.com/google/uuid"
)

// Table 20-10, GUID Format, Version 1 based GUID
//
// Note that the individual fields within the GUID are stored least-significant byte first
// and in the order illustrated in the following table.
//
// So, for [16]byte
//	                M      M      M       M                M    Most Significant Byte
//	[0][1][2][3][4][5] [6][7] [8][9] [10][11] [12][13][14][15]
//	 |  |  |  |  |  |                                           6 bytes, node mac address
//	                    |  |                                    2 bytes, clock seq and reserverd
//	                           |  |                             2 bytes, time high and version
//	                                   |   |                    2 bytes, time mid
//	                                            |   |   |   |   4 bytes, time low
//
// So, to unpack [16]byte to the fields of GUIDFormat:
// GUIDFormat.Node = [5][4][3][2][1][0]
// GUIDFormat.ClockSeqReserved = [7][6]
// GUIDFormat.TimeHighVersion = [9][8]
// GUIDFormat.TimeMid = [11][10]
// GUIDFormat.TimeLow = [15][14][13][12]
//
// The GUID string representation of GUIDFormat is
//	xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//	||||||||                             4 bytes, time low
//	         ||||                        2 bytes, time mid
//	              ||||                   2 bytes, time high and version
//	                   ||||              2 bytes, clock seq and reserved
//	                        |||||||||||| 6 bytes, mac address of node
type GUIDFormat struct {
	Node             []byte
	ClockSeqReserved []byte
	TimeHighVersion  []byte
	TimeMid          []byte
	TimeLow          []byte
}

type GUIDMode uint8

const (
	GUIDModeRFC4122 GUIDMode = iota
	GUIDModeIPMI
	GUIDModeSMBIOS
	GUIDModeAuto
	GUIDModeDump

	GUIDRealModes  int = 3
	GUIDTotalModes int = 5
)

// ParseGUIDWithMode parses the raw guid data according to the requested encoding mode.
// see: https://github.com/ipmitool/ipmitool/issues/25
func ParseGUIDWithMode(data []byte, guidMode GUIDMode) error {
	// Todo
	return nil
}

func ParseGUID(data []byte) (uuid.UUID, error) {
	z := uuid.UUID([16]byte{})

	if len(data) < 16 {
		return z, fmt.Errorf("the length must be not less than 16")
	}

	uuidRFC4122MSB := make([]byte, 16)
	for i := 0; i < 16; i++ {
		uuidRFC4122MSB[i] = data[:][15-i]
	}
	u, err := uuid.FromBytes(uuidRFC4122MSB)
	if err != nil {
		return z, fmt.Errorf("invalid UUID Bytes")
	}
	return u, nil
}

// https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.2.0.pdf
// 921 7.2.1 System — UUID
//
//	GUID byte | Field     | MSbyte
//	1         | time low  |
//	2         | time low  |
//	3         | time low  |
//	4         | time low  | LSbyte
//	5         | time mid  |
//	6         | time mid  | LSbyte
//	7         | time high |
//	8         | time high | LSbyte
//	9         | clock seq |
//	10        | clock seq | MSbyte
//	11        | node      |
//	12        | node      |
//	13        | node      |
//	14        | node      |
//	15        | node      |
//	16        | node      | MSbyte
func parseGUIDFormatForSMBIOS(data [16]byte) GUIDFormat {
	guid := GUIDFormat{}

	// For SMBIOS time fields are little-endian (as in IPMI), the rest is in network order (as in RFC4122)
	guid.TimeLow = data[0:4]
	guid.TimeMid = data[4:6]
	guid.TimeHighVersion = data[6:8]
	guid.ClockSeqReserved = data[8:10]
	guid.Node = data[10:16]

	return guid
}

// For plain GUID bytes [16]byte, IPMI mode
//
//	GUID byte | Field     | MSbyte
//	1         | node      |
//	2         | node      |
//	3         | node      |
//	4         | node      |
//	5         | node      |
//	6         | node      | MSbyte
//	7         | clock seq |
//	8         | clock seq | MSbyte
//	9         | time high |
//	10        | time high | MSbyte
//	11        | time mid  |
//	12        | time mid  | MSbyte
//	13        | time low  |
//	14        | time low  |
//	15        | time low  |
//	16        | time low  | MSbyte
func parseGUIDForamtForIPMI(data [16]byte) GUIDFormat {
	guid := GUIDFormat{}

	// For IPMI, all fields are little-endian (LSB first)
	guid.TimeLow = data[0:4]
	guid.TimeMid = data[4:6]
	guid.TimeHighVersion = data[6:8]
	guid.ClockSeqReserved = data[8:10]
	guid.Node = data[10:16]

	return guid
}

// RFC4122 UUID string representation is:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//
// time-low "-" time-mid "-" time-high-and-version "-" clock-seq-and-reserved clock-seq-low "-" node
// 4bytes       2bytes       2bytes                    1byte                  1byte             6bytes
// xxxxxxxx  -  xxxx      -  xxxx                   -  xxxx                                  -  xxxxxxxxxxxx
//
func parseGUIDFormatForRFC4122(data [16]byte) GUIDFormat {
	guid := GUIDFormat{}

	// For RFC4122, all fields are network byte order (MSB first)
	uuidRFC4122MSB := make([]byte, 16)
	for i := 0; i < 16; i++ {
		// uuidRFC4122MSB[i] = data[:][15-i]
		uuidRFC4122MSB[i] = data[i]
	}

	guid.TimeLow = uuidRFC4122MSB[0:4]
	guid.TimeMid = uuidRFC4122MSB[4:6]
	guid.TimeHighVersion = uuidRFC4122MSB[6:8]
	guid.ClockSeqReserved = uuidRFC4122MSB[8:10]
	guid.Node = uuidRFC4122MSB[10:16]

	return guid
}

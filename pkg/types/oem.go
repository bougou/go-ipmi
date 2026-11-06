package types

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// OEM represents Manufacturer ID, that is IANA Private Enterprise Number
type OEM uint32

//go:generate go run oem_generate.go

// The source registry stays reviewable in the repository; only the compressed
// ID/organization extract is linked into binaries.
//
//go:embed data/enterprise-names.tsv.gz
var oemNamesGZIP []byte

var (
	oemNamesOnce sync.Once
	oemNames     map[OEM]string
	oemNamesErr  error
)

// cSpell: disable
const (
	OEM_UNKNOWN                      = 0
	OEM_DEBUG                        = 0xFFFFFE /* Hoping IANA won't hit this soon */
	OEM_RESERVED                     = 0x0FFFFF /* As per IPMI 2.0 specification */
	OEM_IBM_2                        = 2        /* 2 for [IBM] */
	OEM_HP                           = 11
	OEM_SUN                          = 42
	OEM_NOKIA                        = 94
	OEM_BULL                         = 107
	OEM_HITACHI_116                  = 116
	OEM_NEC                          = 119
	OEM_TOSHIBA                      = 186
	OEM_ERICSSON                     = 193
	OEM_INTEL                        = 343
	OEM_TATUNG                       = 373
	OEM_HITACHI_399                  = 399
	OEM_DELL                         = 674
	OEM_HUAWEI                       = 2011
	OEM_LMC                          = 2168
	OEM_ASUSTEK                      = 2623
	OEM_RADISYS                      = 4337
	OEM_BROADCOM                     = 4413
	OEM_IBM_4769                     = 4769 /* 4769 for [IBM Corporation] */
	OEM_MAGNUM                       = 5593
	OEM_TYAN                         = 6653
	OEM_QUANTA                       = 7244
	OEM_VIKING                       = 9237
	OEM_ADVANTECH                    = 10297
	OEM_FUJITSU_SIEMENS              = 10368
	OEM_AVOCENT                      = 10418
	OEM_PEPPERCON                    = 10437
	OEM_SUPERMICRO                   = 10876
	OEM_OSA                          = 11102
	OEM_GOOGLE                       = 11129
	OEM_PICMG                        = 12634
	OEM_RARITAN                      = 13742
	OEM_KONTRON                      = 15000
	OEM_GIGABYTE                     = 15370
	OEM_PPS                          = 16394
	OEM_IBM_20301                    = 20301 /* 20301 for [IBM eServer X] */
	OEM_AMI                          = 20974
	OEM_FOXCONN                      = 22238
	OEM_ADLINK_24339                 = 24339 /* 24339 for [ADLINK TECHNOLOGY INC.] */
	OEM_H3C                          = 25506
	OEM_NOKIA_28458                  = 28458
	OEM_NOKIA_SOLUTIONS_AND_NETWORKS = 28458
	OEM_VITA                         = 33196
	OEM_INSPUR                       = 37945
	OEM_TENCENT                      = 41475
	OEM_BYTEDANCE                    = 46045
	// Retained for compatibility with ipmitool; registry lookups resolve 47488 to Nedam.
	OEM_SUPERMICRO_47488 = 47488
	OEM_YADRO            = 49769
)

func (oem OEM) String() string {
	// ipmitool puts these entries before the loaded IANA registry so they take
	// precedence over registrations with the same numeric value.
	switch oem {
	case OEM_UNKNOWN:
		return "Unknown"
	case OEM_RESERVED:
		return "Unspecified"
	}

	names, err := cachedOEMNames()
	if err != nil {
		panic(fmt.Sprintf("types: invalid embedded IANA PEN registry: %v", err))
	}
	if name, ok := names[oem]; ok {
		return name
	}

	// ipmitool puts DEBUG after the registry, so a real IANA assignment wins.
	if oem == OEM_DEBUG {
		return "A Debug Assisting Company, Ltd."
	}
	return "Unknown"
}

func cachedOEMNames() (map[OEM]string, error) {
	oemNamesOnce.Do(func() {
		oemNames, oemNamesErr = loadOEMNames()
	})
	return oemNames, oemNamesErr
}

func loadOEMNames() (map[OEM]string, error) {
	r, err := gzip.NewReader(bytes.NewReader(oemNamesGZIP))
	if err != nil {
		return nil, err
	}
	names, parseErr := parseOEMNames(r)
	closeErr := r.Close()
	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return names, nil
}

func parseOEMNames(r io.Reader) (map[OEM]string, error) {
	names := make(map[OEM]string)
	scanner := bufio.NewScanner(r)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		idText, name, ok := strings.Cut(scanner.Text(), "\t")
		if !ok || name == "" {
			return nil, fmt.Errorf("invalid name record at line %d", lineNumber)
		}
		id, err := strconv.ParseUint(idText, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid PEN at line %d: %w", lineNumber, err)
		}
		oem := OEM(id)
		if _, exists := names[oem]; exists {
			return nil, fmt.Errorf("duplicate PEN %d at line %d", id, lineNumber)
		}
		names[oem] = name
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

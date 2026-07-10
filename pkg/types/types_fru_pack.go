package types

import "time"

// FRUPackConfig describes fields for [PackFRU].
type FRUPackConfig struct {
	Chassis *FRUPackChassis
	Board   *FRUPackBoard
	Product *FRUPackProduct
}

// FRUPackChassis holds chassis info area fields for [PackFRU] (FRU/10).
type FRUPackChassis struct {
	Type       uint8
	PartNumber string
	Serial     string
}

// FRUPackBoard holds board info area fields for [PackFRU] (FRU/11).
type FRUPackBoard struct {
	MfgDate    time.Time
	Mfg        string
	Product    string
	Serial     string
	PartNumber string
}

// FRUPackProduct holds product info area fields for [PackFRU] (FRU/12).
type FRUPackProduct struct {
	Manufacturer string
	Name         string
	PartModel    string
	Version      string
	Serial       string
}

// PackFRU serialises a FRU inventory area per Platform Management FRU v1.0.
func PackFRU(cfg FRUPackConfig) []byte {
	var out []byte
	cursor := uint16(FRUCommonHeaderSize)
	header := FRUCommonHeader{FormatVersion: FRUFormatVersion}

	appendArea := func(offsetField *uint8, data []byte) {
		if data == nil {
			return
		}
		if cursor%8 != 0 {
			pad := 8 - int(cursor%8)
			out = append(out, make([]byte, pad)...)
			cursor += uint16(pad)
		}
		*offsetField = uint8(cursor / 8)
		out = append(out, data...)
		cursor += uint16(len(data))
	}

	if cfg.Chassis != nil {
		appendArea(&header.ChassisOffset8B, packFRUChassisArea(cfg.Chassis))
	}
	if cfg.Board != nil {
		appendArea(&header.BoardOffset8B, packFRUBoardArea(cfg.Board))
	}
	if cfg.Product != nil {
		appendArea(&header.ProductOffset8B, packFRUProductArea(cfg.Product))
	}

	header.Checksum = fruPackChecksum(header.Pack()[:7])
	final := header.Pack()
	final = append(final, out...)
	return final
}

func packFRUChassisArea(c *FRUPackChassis) []byte {
	body := []byte{FRUFormatVersion, 0, c.Type}
	body = append(body, packFRUASCIIField(c.PartNumber)...)
	body = append(body, packFRUASCIIField(c.Serial)...)
	body = append(body, FRUAreaFieldsEndMark)
	return finalizeFRUArea(body)
}

func packFRUBoardArea(b *FRUPackBoard) []byte {
	body := []byte{FRUFormatVersion, 0, 0}
	const secsFrom1970To1996 uint32 = 820454400
	minutes := uint32(0)
	if !b.MfgDate.IsZero() {
		minutes = uint32(b.MfgDate.Unix()-int64(secsFrom1970To1996)) / 60
	}
	body = append(body, byte(minutes), byte(minutes>>8), byte(minutes>>16))
	body = append(body, packFRUASCIIField(b.Mfg)...)
	body = append(body, packFRUASCIIField(b.Product)...)
	body = append(body, packFRUASCIIField(b.Serial)...)
	body = append(body, packFRUASCIIField(b.PartNumber)...)
	body = append(body, packFRUASCIIField("")...)
	body = append(body, FRUAreaFieldsEndMark)
	return finalizeFRUArea(body)
}

func packFRUProductArea(p *FRUPackProduct) []byte {
	body := []byte{FRUFormatVersion, 0, 0}
	body = append(body, packFRUASCIIField(p.Manufacturer)...)
	body = append(body, packFRUASCIIField(p.Name)...)
	body = append(body, packFRUASCIIField(p.PartModel)...)
	body = append(body, packFRUASCIIField(p.Version)...)
	body = append(body, packFRUASCIIField(p.Serial)...)
	body = append(body, packFRUASCIIField("")...)
	body = append(body, packFRUASCIIField("")...)
	body = append(body, FRUAreaFieldsEndMark)
	return finalizeFRUArea(body)
}

func packFRUASCIIField(s string) []byte {
	if s == "" {
		return []byte{0xC0}
	}
	if len(s) > 0x3f {
		s = s[:0x3f]
	}
	out := make([]byte, 1+len(s))
	out[0] = 0xC0 | uint8(len(s))
	copy(out[1:], []byte(s))
	return out
}

func finalizeFRUArea(body []byte) []byte {
	padded := append([]byte{}, body...)
	for (len(padded)+1)%8 != 0 {
		padded = append(padded, 0)
	}
	padded = append(padded, 0)
	padded[1] = uint8(len(padded) / 8)
	padded[len(padded)-1] = fruPackChecksum(padded[:len(padded)-1])
	return padded
}

func fruPackChecksum(data []byte) uint8 {
	sum := 0
	for _, b := range data {
		sum = (sum + int(b)) % 256
	}
	return uint8(-sum)
}

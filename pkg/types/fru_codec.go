package types

import "fmt"

// ParseFRU parses raw FRU inventory bytes into a [FRU] struct.
func ParseFRU(data []byte) (*FRU, error) {
	if len(data) < int(FRUCommonHeaderSize) {
		return nil, ErrUnpackedDataTooShortWith(len(data), int(FRUCommonHeaderSize))
	}

	hdr := &FRUCommonHeader{}
	if err := hdr.Unpack(data[:FRUCommonHeaderSize]); err != nil {
		return nil, fmt.Errorf("parse fru common header: %w", err)
	}
	if !hdr.Valid() {
		return nil, fmt.Errorf("parse fru: invalid common header checksum")
	}
	if hdr.FormatVersion != FRUFormatVersion {
		return nil, fmt.Errorf("parse fru: unsupported format version %#02x", hdr.FormatVersion)
	}

	fru := &FRU{CommonHeader: hdr}

	sliceArea := func(offset8B uint8) ([]byte, error) {
		if offset8B == 0 {
			return nil, nil
		}
		off := int(offset8B) * 8
		if off+2 > len(data) {
			return nil, fmt.Errorf("area at offset %#x: header too short", off)
		}
		areaLen := int(data[off+1]) * 8
		if areaLen < 8 {
			return nil, fmt.Errorf("area at offset %#x: invalid length %d", off, areaLen)
		}
		if off+areaLen > len(data) {
			return nil, fmt.Errorf("area at offset %#x: want %d bytes, have %d", off, areaLen, len(data)-off)
		}
		return data[off : off+areaLen], nil
	}

	if hdr.InternalOffset8B > 0 {
		end, err := fruAreaEnd(data, hdr, hdr.InternalOffset8B)
		if err != nil {
			return nil, fmt.Errorf("parse fru internal area: %w", err)
		}
		off := int(hdr.InternalOffset8B) * 8
		area := &FRUInternalUseArea{}
		if err := area.Unpack(data[off:end]); err != nil {
			return nil, fmt.Errorf("parse fru internal area: %w", err)
		}
		fru.InternalUseArea = area
	}

	if b, err := sliceArea(hdr.ChassisOffset8B); err != nil {
		return nil, fmt.Errorf("parse fru chassis area: %w", err)
	} else if b != nil {
		area := &FRUChassisInfoArea{}
		if err := area.Unpack(b); err != nil {
			return nil, fmt.Errorf("parse fru chassis area: %w", err)
		}
		fru.ChassisInfoArea = area
	}

	if b, err := sliceArea(hdr.BoardOffset8B); err != nil {
		return nil, fmt.Errorf("parse fru board area: %w", err)
	} else if b != nil {
		area := &FRUBoardInfoArea{}
		if err := area.Unpack(b); err != nil {
			return nil, fmt.Errorf("parse fru board area: %w", err)
		}
		fru.BoardInfoArea = area
	}

	if b, err := sliceArea(hdr.ProductOffset8B); err != nil {
		return nil, fmt.Errorf("parse fru product area: %w", err)
	} else if b != nil {
		area := &FRUProductInfoArea{}
		if err := area.Unpack(b); err != nil {
			return nil, fmt.Errorf("parse fru product area: %w", err)
		}
		fru.ProductInfoArea = area
	}

	if hdr.MultiRecordsOffset8B > 0 {
		off := int(hdr.MultiRecordsOffset8B) * 8
		for off < len(data) {
			if off+5 > len(data) {
				break
			}
			recLen := int(data[off+2])
			total := 5 + recLen
			if off+total > len(data) {
				return nil, fmt.Errorf("parse fru multi-record at %#x: short record", off)
			}
			rec := &FRUMultiRecord{}
			if err := rec.Unpack(data[off : off+total]); err != nil {
				return nil, fmt.Errorf("parse fru multi-record at %#x: %w", off, err)
			}
			fru.MultiRecords = append(fru.MultiRecords, rec)
			off += total
			if rec.EndOfList {
				break
			}
		}
	}

	return fru, nil
}

// PackFRUInventory serialises a parsed [FRU] to wire bytes per Platform Management FRU v1.0.
func PackFRUInventory(fru *FRU) ([]byte, error) {
	if fru == nil {
		return nil, fmt.Errorf("types: PackFRUInventory: nil fru")
	}

	header := FRUCommonHeader{FormatVersion: FRUFormatVersion}
	if fru.CommonHeader != nil && fru.CommonHeader.FormatVersion != 0 {
		header.FormatVersion = fru.CommonHeader.FormatVersion
	}

	var out []byte
	cursor := uint16(FRUCommonHeaderSize)

	if fru.InternalUseArea != nil {
		out = appendFRUArea(out, &cursor, &header.InternalOffset8B, fru.InternalUseArea.Pack())
	}
	if fru.ChassisInfoArea != nil {
		out = appendFRUArea(out, &cursor, &header.ChassisOffset8B, fru.ChassisInfoArea.Pack())
	}
	if fru.BoardInfoArea != nil {
		out = appendFRUArea(out, &cursor, &header.BoardOffset8B, fru.BoardInfoArea.Pack())
	}
	if fru.ProductInfoArea != nil {
		out = appendFRUArea(out, &cursor, &header.ProductOffset8B, fru.ProductInfoArea.Pack())
	}
	if len(fru.MultiRecords) > 0 {
		out = appendFRUArea(out, &cursor, &header.MultiRecordsOffset8B, PackMultiRecords(fru.MultiRecords))
	}

	header.Checksum = fruPackChecksum(header.Pack()[:7])
	final := header.Pack()
	final = append(final, out...)
	return final, nil
}

// packFRUASCIIField encodes an 8-bit ASCII FRU field (type/length 11b + Latin-1).
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

func fruASCIITypeLength(data []byte) TypeLength {
	if len(data) == 0 {
		return TypeLength(0xC0)
	}
	return TypeLength(0xC0 | byte(len(data)))
}

// packFRUField serialises a FRU type/length field from parsed area data.
func packFRUField(tl TypeLength, data []byte) []byte {
	if tl == 0 && len(data) == 0 {
		return []byte{0xC0}
	}
	if tl == 0 {
		return packFRUASCIIField(string(data))
	}
	length := int(tl.Length())
	out := make([]byte, 1+length)
	out[0] = byte(tl)
	if length > 0 {
		n := length
		if n > len(data) {
			n = len(data)
		}
		copy(out[1:], data[:n])
	}
	return out
}

func finalizeFRUArea(body []byte, unused []byte) []byte {
	padded := append([]byte{}, body...)
	if len(unused) > 0 {
		padded = append(padded, unused...)
	} else {
		for (len(padded)+1)%8 != 0 {
			padded = append(padded, 0)
		}
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

// fruAreaEnd returns the byte offset where the area at start8B ends, bounded by
// the next present area offset in the common header (fru§8 / fru§9).
func fruAreaEnd(data []byte, hdr *FRUCommonHeader, start8B uint8) (int, error) {
	off := int(start8B) * 8
	if off >= len(data) {
		return 0, fmt.Errorf("area at offset %#x: beyond data", off)
	}
	end := len(data)
	for _, next8B := range []uint8{
		hdr.ChassisOffset8B,
		hdr.BoardOffset8B,
		hdr.ProductOffset8B,
		hdr.MultiRecordsOffset8B,
	} {
		if next8B > start8B {
			candidate := int(next8B) * 8
			if candidate < end {
				end = candidate
			}
		}
	}
	size := end - off
	if size < 8 {
		return 0, fmt.Errorf("area at offset %#x: size %d below minimum 8", off, size)
	}
	if size%8 != 0 {
		return 0, fmt.Errorf("area at offset %#x: size %d not multiple of 8", off, size)
	}
	return end, nil
}

func appendFRUArea(out []byte, cursor *uint16, offsetField *uint8, data []byte) []byte {
	if data == nil {
		return out
	}
	if *cursor%8 != 0 {
		pad := 8 - int(*cursor%8)
		out = append(out, make([]byte, pad)...)
		*cursor += uint16(pad)
	}
	*offsetField = uint8(*cursor / 8)
	out = append(out, data...)
	*cursor += uint16(len(data))
	return out
}

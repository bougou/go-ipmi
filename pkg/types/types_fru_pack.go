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
	fru := &FRU{CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion}}

	if cfg.Chassis != nil {
		c := cfg.Chassis
		fru.ChassisInfoArea = &FRUChassisInfoArea{
			FormatVersion:          FRUFormatVersion,
			ChassisType:            ChassisType(c.Type),
			PartNumberTypeLength:   fruASCIITypeLength([]byte(c.PartNumber)),
			PartNumber:             []byte(c.PartNumber),
			SerialNumberTypeLength: fruASCIITypeLength([]byte(c.Serial)),
			SerialNumber:           []byte(c.Serial),
		}
	}
	if cfg.Board != nil {
		b := cfg.Board
		fru.BoardInfoArea = &FRUBoardInfoArea{
			FormatVersion:          FRUFormatVersion,
			MfgDateTime:            b.MfgDate,
			ManufacturerTypeLength: fruASCIITypeLength([]byte(b.Mfg)),
			Manufacturer:           []byte(b.Mfg),
			ProductNameTypeLength:  fruASCIITypeLength([]byte(b.Product)),
			ProductName:            []byte(b.Product),
			SerialNumberTypeLength: fruASCIITypeLength([]byte(b.Serial)),
			SerialNumber:           []byte(b.Serial),
			PartNumberTypeLength:   fruASCIITypeLength([]byte(b.PartNumber)),
			PartNumber:             []byte(b.PartNumber),
			FRUFileIDTypeLength:    TypeLength(0xC0),
		}
	}
	if cfg.Product != nil {
		p := cfg.Product
		fru.ProductInfoArea = &FRUProductInfoArea{
			FormatVersion:          FRUFormatVersion,
			ManufacturerTypeLength: fruASCIITypeLength([]byte(p.Manufacturer)),
			Manufacturer:           []byte(p.Manufacturer),
			NameTypeLength:         fruASCIITypeLength([]byte(p.Name)),
			Name:                   []byte(p.Name),
			PartModelTypeLength:    fruASCIITypeLength([]byte(p.PartModel)),
			PartModel:              []byte(p.PartModel),
			VersionTypeLength:      fruASCIITypeLength([]byte(p.Version)),
			Version:                []byte(p.Version),
			SerialNumberTypeLength: fruASCIITypeLength([]byte(p.Serial)),
			SerialNumber:           []byte(p.Serial),
			AssetTagTypeLength:     TypeLength(0xC0),
			FRUFileIDTypeLength:    TypeLength(0xC0),
		}
	}

	out, err := PackFRUInventory(fru)
	if err != nil {
		return nil
	}
	return out
}

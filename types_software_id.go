package ipmi

// todo: SoftwareID actually 7bits, 8bits for SensorOwnerID
type SoftwareID uint8 // SWID
type SoftwareType string

// section 5.5

const (
	SoftwareTypeBIOS        = SoftwareType("BIOS")
	SoftwareTypeSMIHandler  = SoftwareType("SMI Handler")
	SoftwareTypeSMS         = SoftwareType("System Management Software")
	SoftwareTypeOEM         = SoftwareType("OEM")
	SoftwareTypeRCS         = SoftwareType("Remote Console Software")
	SoftwareTypeTerminalRCS = SoftwareType("Terminal Mode Remote Console Softeware")
	SoftwareTypeReserved    = SoftwareType("Reserved")
)

// section 5.5
func (i SoftwareID) Type() SoftwareType {
	var t SoftwareType
	if i >= SoftwareID(0x01) && i <= SoftwareID(0x1F) {
		t = SoftwareTypeBIOS
	} else if i >= SoftwareID(0x21) && i <= SoftwareID(0x3F) {
		t = SoftwareTypeSMIHandler
	} else if i >= SoftwareID(0x41) && i <= SoftwareID(0x5F) {
		t = SoftwareTypeSMS
	} else if i >= SoftwareID(0x61) && i <= SoftwareID(0x7F) {
		t = SoftwareTypeOEM
	} else if i >= SoftwareID(0x81) && i <= SoftwareID(0x8D) {
		t = SoftwareTypeRCS
	} else if i == SoftwareID(0x8F) {
		t = SoftwareTypeTerminalRCS
	} else {
		t = SoftwareTypeReserved
	}
	return t
}

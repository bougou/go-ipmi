package ipmi

// 13.27.3
// The Get Channel Payload Support command returns which standard payload type numbers and OEM payload
// type handles are available on a given channel of a BMC.
type PayloadType uint8

const (
	// Standard Payload Types
	// used to identify payloads that are specified by the IPMI specifications

	PayloadTypeIPMI PayloadType = 0x00
	PayloadTypeSOL  PayloadType = 0x01
	PayloadTypeOEM  PayloadType = 0x02

	// Session Setup Payload Types
	// used to identify payloads that are for session setup messages specified by the IPMI specifications

	PayloadTypeRmcpOpenSessionRequest  PayloadType = 0x10
	PayloadTypeRmcpOpenSessionResponse PayloadType = 0x11
	PayloadTypeRAKPMessage1            PayloadType = 0x12
	PayloadTypeRAKPMessage2            PayloadType = 0x13
	PayloadTypeRAKPMessage3            PayloadType = 0x14
	PayloadTypeRAKPMessage4            PayloadType = 0x15

	// OEM Payload Type Handles
	// used to identify payloads that are specified by a given OEM

	PayloadTypeOEM0 PayloadType = 0x20
	PayloadTypeOEM1 PayloadType = 0x21
	PayloadTypeOEM2 PayloadType = 0x22
	PayloadTypeOEM3 PayloadType = 0x23
	PayloadTypeOEM4 PayloadType = 0x24
	PayloadTypeOEM5 PayloadType = 0x25
	PayloadTypeOEM6 PayloadType = 0x26
	PayloadTypeOEM7 PayloadType = 0x27
)

func (pt PayloadType) String() string {
	m := map[PayloadType]string{
		0x00: "ipmi",
		0x01: "sol",
		0x02: "oem",

		0x10: "rmcp-open-session-request",
		0x11: "rmcp-open-session-response",
		0x12: "rakp-message-1",
		0x13: "rakp-message-2",
		0x14: "rakp-message-3",
		0x15: "rakp-message-4",

		0x20: "oem0",
		0x21: "oem1",
		0x22: "oem2",
		0x23: "oem3",
		0x24: "oem4",
		0x25: "oem5",
		0x26: "oem6",
		0x27: "oem7",
	}

	s, ok := m[pt]
	if ok {
		return s
	}

	return "reserved"
}

// 13.28
type AuthAlg uint8

const (
	AuthAlgRAKP_None        AuthAlg = 0x00 // Mandatory
	AuthAlgRAKP_HMAC_SHA1   AuthAlg = 0x01 // Mandatory
	AuthAlgRAKP_HMAC_MD5    AuthAlg = 0x02 // Optional
	AuthAlgRAKP_HMAC_SHA256 AuthAlg = 0x03 // Optional
)

func (authAlg AuthAlg) String() string {
	m := map[AuthAlg]string{
		0x00: "none",
		0x01: "hmac_sha1",
		0x02: "hmac_md5",
		0x03: "hmac_sha256",
	}
	s, ok := m[authAlg]
	if ok {
		return s
	}
	return ""
}

// 13.28.4
type IntegrityAlg uint8

const (
	IntegrityAlg_None            IntegrityAlg = 0x00 // Mandatory
	IntegrityAlg_HMAC_SHA1_96    IntegrityAlg = 0x01 // Mandatory
	IntegrityAlg_HMAC_MD5_128    IntegrityAlg = 0x02 // Optional
	IntegrityAlg_MD5_128         IntegrityAlg = 0x03 // Optional
	IntegrityAlg_HMAC_SHA256_128 IntegrityAlg = 0x04 // Optional
)

func (integrityAlg IntegrityAlg) String() string {
	m := map[IntegrityAlg]string{
		0x00: "none",
		0x01: "hmac_sha1_96",
		0x02: "hmac_md5_128",
		0x03: "md5_128",
		0x04: "hmac_sha256_128",
	}
	s, ok := m[integrityAlg]
	if ok {
		return s
	}
	return ""
}

// 13.28.5
// Confidentiality (Encryption) Algorithms
// AES is more secure than RC4
// RC4 is cryptographically broken and should not be used for secure applications.
type CryptAlg uint8

const (
	CryptAlg_None        CryptAlg = 0x00 // Mandatory
	CryptAlg_AES_CBC_128 CryptAlg = 0x01 // Mandatory
	CryptAlg_xRC4_128    CryptAlg = 0x02 // Optional
	CryptAlg_xRC4_40     CryptAlg = 0x03 // Optional

	Encryption_AES_CBS_128_BlockSize uint8 = 0x10
)

func (cryptAlg CryptAlg) String() string {
	m := map[CryptAlg]string{
		0x00: "none",
		0x01: "aes_cbc_128",
		0x02: "xrc4_128",
		0x03: "xrc4_40",
	}
	s, ok := m[cryptAlg]
	if ok {
		return s
	}
	return ""
}

package types

// 22.15.2 Cipher Suite IDs
type CipherSuiteID uint8

const (
	CipherSuiteID0        CipherSuiteID = 0
	CipherSuiteID1        CipherSuiteID = 1
	CipherSuiteID2        CipherSuiteID = 2
	CipherSuiteID3        CipherSuiteID = 3
	CipherSuiteID4        CipherSuiteID = 4
	CipherSuiteID5        CipherSuiteID = 5
	CipherSuiteID6        CipherSuiteID = 6
	CipherSuiteID7        CipherSuiteID = 7
	CipherSuiteID8        CipherSuiteID = 8
	CipherSuiteID9        CipherSuiteID = 9
	CipherSuiteID10       CipherSuiteID = 10
	CipherSuiteID11       CipherSuiteID = 11
	CipherSuiteID12       CipherSuiteID = 12
	CipherSuiteID13       CipherSuiteID = 13
	CipherSuiteID14       CipherSuiteID = 14
	CipherSuiteID15       CipherSuiteID = 15
	CipherSuiteID16       CipherSuiteID = 16
	CipherSuiteID17       CipherSuiteID = 17
	CipherSuiteID18       CipherSuiteID = 18
	CipherSuiteID19       CipherSuiteID = 19
	CipherSuiteIDReserved CipherSuiteID = 0xff
)

const (
	StandardCipherSuite uint8 = 0xc0
	OEMCipherSuite      uint8 = 0xc1

	CipherAlgMask       uint8 = 0x3f // [5:0]=111111b
	CipherAlgTagBitMask uint8 = 0xc0 // [7:6]=11b

	CipherAlgTagBitAuthMask       uint8 = 0x00 // [7:6]=00b
	CipherAlgTagBitIntegrityMask  uint8 = 0x40 // [7:6]=01b
	CipherAlgTagBitEncryptionMask uint8 = 0x80 // [7:6]=10b

	LIST_ALGORITHMS_BY_CIPHER_SUITE uint8 = 0x80
)

// getCipherSuiteAlgorithms returns AuthAlg, IntegrityAlg and CryptAlg of the specified cipherSuiteID.
func GetCipherSuiteAlgorithms(cipherSuiteID CipherSuiteID) (authAlg AuthAlg, integrity IntegrityAlg, encryptionAlg CryptAlg, ok bool) {
	switch cipherSuiteID {
	case CipherSuiteID0:
		return AuthAlg_None, IntegrityAlg_None, CryptAlg_None, true
	case CipherSuiteID1:
		return AuthAlg_HMAC_SHA1, IntegrityAlg_None, CryptAlg_None, true
	case CipherSuiteID2:
		return AuthAlg_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_None, true
	case CipherSuiteID3:
		return AuthAlg_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_AES_CBC_128, true
	case CipherSuiteID4:
		return AuthAlg_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_xRC4_128, true
	case CipherSuiteID5:
		return AuthAlg_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_xRC4_40, true
	case CipherSuiteID6:
		return AuthAlg_HMAC_MD5, IntegrityAlg_None, CryptAlg_None, true
	case CipherSuiteID7:
		return AuthAlg_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_None, true
	case CipherSuiteID8:
		return AuthAlg_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_AES_CBC_128, true
	case CipherSuiteID9:
		return AuthAlg_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_xRC4_128, true
	case CipherSuiteID10:
		return AuthAlg_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_xRC4_40, true
	case CipherSuiteID11:
		return AuthAlg_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_None, true
	case CipherSuiteID12:
		return AuthAlg_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_AES_CBC_128, true
	case CipherSuiteID13:
		return AuthAlg_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_xRC4_128, true
	case CipherSuiteID14:
		return AuthAlg_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_xRC4_40, true
	case CipherSuiteID15:
		return AuthAlg_HMAC_SHA256, IntegrityAlg_None, CryptAlg_None, true
	case CipherSuiteID16:
		return AuthAlg_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_None, true
	case CipherSuiteID17:
		return AuthAlg_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_AES_CBC_128, true
	case CipherSuiteID18:
		return AuthAlg_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_xRC4_128, true
	case CipherSuiteID19:
		return AuthAlg_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_xRC4_40, true
	case CipherSuiteIDReserved:
		return 0, 0, 0, false
	default:
		return 0, 0, 0, false
	}
}

// 22.15.1 Cipher Suite Records
// The size of a CipherSuiteRecord is
type CipherSuiteRecord struct {
	// If StartOfRecord is C0h, indicating that the Start Of Record byte is followed by an Cipher Suite ID
	// If StartOfRecord is C1h, indicating that the Start Of Record byte is followed  by a OEM Cipher Suite ID plus OEM IANA
	StartOfRecord uint8

	// a numeric way of identifying the Cipher Suite on the platform
	CipherSuitID CipherSuiteID
	OEMIanaID    uint32 // Least significant byte first. 3-byte IANA for the OEM or body that defined the Cipher Suite.

	// an authentication algorithm number is required for all Cipher Suites.
	// It is possible that a given Cipher Suite may not specify use of an integrity or confidentiality algorithm.
	AuthAlg       uint8   // Tag bits: [7:6]=00b
	IntegrityAlgs []uint8 // Tag bits: [7:6]=01b
	CryptAlgs     []uint8 // Tag bits: [7:6]=10b
}

// sortCipherSuites return cipher suites in order of preference.
// the cipher suite not in the PreferredCiphers list would be excluded.
func SortCipherSuites(cipherSuites []CipherSuiteID) []CipherSuiteID {
	sorted := make([]CipherSuiteID, 0)
	for _, preferredCipher := range PreferredCiphers {
		for _, cipherSuiteID := range cipherSuites {
			if preferredCipher == cipherSuiteID {
				sorted = append(sorted, cipherSuiteID)
			}
		}
	}

	return sorted
}

var PreferredCiphers = []CipherSuiteID{
	// Todo
	// cipher suite best order is chosen with this criteria:
	// xRC4 is bad
	// AES128 is required
	// HMAC-SHA256 > HMAC-SHA1
	// secure authentication > encrypted content

	// With xRC4 out, all cipher suites with MD5 out, and cipher suite 3
	// being required by the spec, the only better defined standard cipher
	// suite is 17. So if SHA256 is available, we should try to use that,
	// otherwise, fall back to 3.

	CipherSuiteID17,

	// IPMI 2.0 spec requires that cipher suite 3 is implemented
	// so we should always be able to fall back to that if better
	// options are not available.
	// CipherSuiteID3 -> 01h, 01h, 01h
	CipherSuiteID3,

	CipherSuiteID15,
	CipherSuiteID16,
	CipherSuiteID18,
	CipherSuiteID19,

	CipherSuiteID6,
	CipherSuiteID7,
	CipherSuiteID8,
	CipherSuiteID11,
	CipherSuiteID12,

	// xRC4 is bad, so we don't use it

	// CipherSuiteID4,
	// CipherSuiteID5,
	// CipherSuiteID9,
	// CipherSuiteID10,
	// CipherSuiteID13,
	// CipherSuiteID14,
}

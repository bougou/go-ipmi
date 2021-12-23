package ipmi

// 22.15.2 Cipher Suite IDs
type CipherSuiteID uint8

const (
	CipherSuiteID0        uint8 = 0
	CipherSuiteID1        uint8 = 1
	CipherSuiteID2        uint8 = 2
	CipherSuiteID3        uint8 = 3
	CipherSuiteID4        uint8 = 4
	CipherSuiteID5        uint8 = 5
	CipherSuiteID6        uint8 = 6
	CipherSuiteID7        uint8 = 7
	CipherSuiteID8        uint8 = 8
	CipherSuiteID9        uint8 = 9
	CipherSuiteID10       uint8 = 10
	CipherSuiteID11       uint8 = 11
	CipherSuiteID12       uint8 = 12
	CipherSuiteID13       uint8 = 13
	CipherSuiteID14       uint8 = 14
	CipherSuiteID15       uint8 = 15
	CipherSuiteID16       uint8 = 16
	CipherSuiteID17       uint8 = 17
	CipherSuiteID18       uint8 = 18
	CipherSuiteID19       uint8 = 19
	CipherSuiteIDReserved uint8 = 0xff
)

const (
	StandardCipherSuite uint8 = 0xc0
	OEMCipherSuite      uint8 = 0xc1

	CipherAlgMask       uint8 = 0x3f // [5:0]=111111b
	CipherAlgTagBitMask uint8 = 0xc0 // [7:6]=11b

	CipherAlgTagBitAuthMask       uint8 = 0x00 // [7:6]=00b
	CipherAlgTagBitInegrityMask   uint8 = 0x40 // [7:6]=01b
	CipherAlgTagBitEncryptionMask uint8 = 0x80 // [7:6]=10b

	LIST_ALGORITHMS_BY_CIPHER_SUITE uint8 = 0x80
)

// getCipherSuiteAlgorithms returns AuthAlg, IntegrityAlg and CryptAlg of the specified cipherSuiteID.
func getCipherSuiteAlgorithms(cipherSuiteID uint8) (authAlg AuthAlg, integrity IntegrityAlg, encryptionAlg CryptAlg, err error) {
	switch cipherSuiteID {
	case CipherSuiteID0:
		return AuthAlgRAKP_None, IntegrityAlg_None, CryptAlg_None, nil
	case CipherSuiteID1:
		return AuthAlgRAKP_HMAC_SHA1, IntegrityAlg_None, CryptAlg_None, nil
	case CipherSuiteID2:
		return AuthAlgRAKP_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_None, nil
	case CipherSuiteID3:
		return AuthAlgRAKP_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_AES_CBC_128, nil
	case CipherSuiteID4:
		return AuthAlgRAKP_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_xRC4_128, nil
	case CipherSuiteID5:
		return AuthAlgRAKP_HMAC_SHA1, IntegrityAlg_HMAC_SHA1_96, CryptAlg_xRC4_40, nil
	case CipherSuiteID6:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_None, CryptAlg_None, nil
	case CipherSuiteID7:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_None, nil
	case CipherSuiteID8:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_AES_CBC_128, nil
	case CipherSuiteID9:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_xRC4_128, nil
	case CipherSuiteID10:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_HMAC_MD5_128, CryptAlg_xRC4_40, nil
	case CipherSuiteID11:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_None, nil
	case CipherSuiteID12:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_AES_CBC_128, nil
	case CipherSuiteID13:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_xRC4_128, nil
	case CipherSuiteID14:
		return AuthAlgRAKP_HMAC_MD5, IntegrityAlg_MD5_128, CryptAlg_xRC4_40, nil
	case CipherSuiteID15:
		return AuthAlgRAKP_HMAC_SHA256, IntegrityAlg_None, CryptAlg_None, nil
	case CipherSuiteID16:
		return AuthAlgRAKP_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_None, nil
	case CipherSuiteID17:
		return AuthAlgRAKP_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_AES_CBC_128, nil
	case CipherSuiteID18:
		return AuthAlgRAKP_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_xRC4_128, nil
	case CipherSuiteID19:
		return AuthAlgRAKP_HMAC_SHA256, IntegrityAlg_HMAC_SHA256_128, CryptAlg_xRC4_40, nil
	case CipherSuiteIDReserved:
		return 0, 0, 0, nil
	default:
		return 0, 0, 0, nil
	}
}

// 22.15.1 Cipher Suite Records
// The size of a CipherSuiteRecord is
type CipherSuiteRecord struct {
	// If StartOfRecord is C0h, indicating that the Start Of Record byte is followed by an Cipher Suite ID
	// If StartOfRecord is C1h, iddicating that the Start Of Record byte is followed  by a OEM Cipher Suite ID plus OEM IANA
	StartOfRecord uint8

	// a numeric way of identifying the Cipher Suite on the platform
	CipherSuitID uint8
	OEMIanaID    uint32 // Least significant byte first. 3-byte IANA for the OEM or body that defined the Cipher Suite.

	// an authentication algorithm number is required for all Cipher Suites.
	// It is possible that a given Cipher Suite may not specify use of an integrity or confidentiality algorithm.
	AuthAlg       uint8   // Tag bits: [7:6]=00b
	IntegrityAlgs []uint8 // Tag bits: [7:6]=01b
	CryptAlgs     []uint8 // Tag bits: [7:6]=10b
}

func findBestCipherSuite() uint8 {
	var bestSuite = CipherSuiteIDReserved

	// Todo
	// cipher suite best order is chosen with this criteria:
	// HMAC-MD5 and MD5 are bad
	// xRC4 is bad
	// AES128 is required
	// HMAC-SHA256 > HMAC-SHA1
	// secure authentication > encrypted content

	// With xRC4 out, all cipher suites with MD5 out, and cipher suite 3
	// being required by the spec, the only better defined standard cipher
	// suite is 17. So if SHA256 is available, we should try to use that,
	// otherwise, fall back to 3.
	// var cipherSuitesOrder = []byte{
	// 	CipherSuiteID17,
	// 	CipherSuiteID3,
	// }

	// supportedCipherSuites = c.GetChannelCipherSuite

	if bestSuite == CipherSuiteIDReserved {
		// IPMI 2.0 spec requires that cipher suite 3 is implemented
		// so we should always be able to fall back to that if better
		// options are not available.
		// CipherSuiteID3 -> 01h, 01h, 01h
		bestSuite = CipherSuiteID3
	}
	return bestSuite
}

package app

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

const (
	MaxCipherSuiteListIndex uint8 = 0x3f
)

// 22.15 Get Channel Cipher Suites Command
type GetChannelCipherSuitesRequest struct {
	// 0h-Bh, Fh = channel numbers
	// Eh = retrieve information for channel this request was issued on
	ChannelNumber uint8
	PayloadType   ipmi.PayloadType
	ListIndex     uint8
}

type GetChannelCipherSuitesResponse struct {
	ChannelNumber      uint8
	CipherSuiteRecords []byte
}

func (req *GetChannelCipherSuitesRequest) Command() ipmi.Command {
	return ipmi.CommandGetChannelCipherSuites
}

func (req *GetChannelCipherSuitesRequest) Pack() []byte {
	var msg = make([]byte, 3)
	ipmi.PackUint8(req.ChannelNumber, msg, 0)
	ipmi.PackUint8(uint8(req.PayloadType), msg, 1)
	ipmi.PackUint8(ipmi.LIST_ALGORITHMS_BY_CIPHER_SUITE|req.ListIndex, msg, 2)
	return msg
}

func (res *GetChannelCipherSuitesResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ChannelNumber, _, _ = ipmi.UnpackUint8(msg, 0)
	if len(msg) > 1 {
		res.CipherSuiteRecords, _, _ = ipmi.UnpackBytesMost(msg, 1, 16)
	}
	return nil
}

func (*GetChannelCipherSuitesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetChannelCipherSuitesResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// This command can be executed prior to establishing a session with the BMC.
// The command is used to look up what authentication, integrity, and confidentiality algorithms are supported.
// The algorithms are used in combination as 'Cipher Suites'.
// This command only applies to implementations that support IPMI v2.0/RMCP+ sessions.

func ParseCipherSuitesData(cipherSuitesData []byte) ([]ipmi.CipherSuiteRecord, error) {
	offset := 0
	records := []ipmi.CipherSuiteRecord{}

	for offset < len(cipherSuitesData) {
		csRecord := ipmi.CipherSuiteRecord{}
		startOfRecord := cipherSuitesData[offset]
		csRecord.StartOfRecord = startOfRecord

		switch startOfRecord {
		case ipmi.StandardCipherSuite:
			// Per §22.15.1 the record is tag-delimited, not fixed-length: the
			// start byte is followed by the cipher suite id and then 1..3
			// algorithm bytes (auth is always present; integrity/confidentiality
			// are omitted when the suite does not use them, e.g. suites 0/1/15).
			// Only require the id byte here; the tag loop below handles the rest
			// and its own end-of-data bounds check.
			if offset+1 > len(cipherSuitesData)-1 {
				return records, fmt.Errorf("incomplete cipher suite data")
			}
			offset++
			csRecord.CipherSuitID = ipmi.CipherSuiteID(cipherSuitesData[offset])

		case ipmi.OEMCipherSuite:
			// id + iana (3) + 3 algs (7 bytes)
			if offset+7 > len(cipherSuitesData)-1 {
				return records, fmt.Errorf("incomplete cipher suite data")
			}
			offset++
			csRecord.CipherSuitID = ipmi.CipherSuiteID(cipherSuitesData[offset])
			offset++
			csRecord.OEMIanaID, _, _ = ipmi.UnpackUint24L(cipherSuitesData, offset)

		default:
			return records, fmt.Errorf("bad start of record byte in the cipher suite data, value %x", startOfRecord)
		}

		for {
			offset++
			if offset > len(cipherSuitesData)-1 {
				break
			}

			algByte := cipherSuitesData[offset]
			if algByte == ipmi.StandardCipherSuite || algByte == ipmi.OEMCipherSuite {
				break
			}

			algTag := algByte & ipmi.CipherAlgTagBitMask // clear lowest 6 bits
			algNumber := algByte & ipmi.CipherAlgMask    // clear highest 2 bits
			switch algTag {
			case ipmi.CipherAlgTagBitAuthMask:
				csRecord.AuthAlg = algNumber
			case ipmi.CipherAlgTagBitIntegrityMask:
				csRecord.IntegrityAlgs = append(csRecord.IntegrityAlgs, algNumber)
			case ipmi.CipherAlgTagBitEncryptionMask:
				csRecord.CryptAlgs = append(csRecord.CryptAlgs, algNumber)
			}
		}
		records = append(records, csRecord)
	}

	return records, nil
}

package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	MaxCipherSuiteListIndex uint8 = 0x3f
)

// 22.15 Get Channel Cipher Suites Command
type GetChannelCipherSuitesRequest struct {
	// 0h-Bh, Fh = channel numbers
	// Eh = retrieve information for channel this request was issued on
	ChannelNumber uint8
	PayloadType   types.PayloadType
	ListIndex     uint8
}

type GetChannelCipherSuitesResponse struct {
	ChannelNumber      uint8
	CipherSuiteRecords []byte
}

func (req *GetChannelCipherSuitesRequest) Command() types.Command {
	return types.CommandGetChannelCipherSuites
}

func (req *GetChannelCipherSuitesRequest) Pack() []byte {
	var msg = make([]byte, 3)
	types.PackUint8(req.ChannelNumber, msg, 0)
	types.PackUint8(uint8(req.PayloadType), msg, 1)
	types.PackUint8(types.LIST_ALGORITHMS_BY_CIPHER_SUITE|req.ListIndex, msg, 2)
	return msg
}

func (res *GetChannelCipherSuitesResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ChannelNumber, _, _ = types.UnpackUint8(msg, 0)
	if len(msg) > 1 {
		res.CipherSuiteRecords, _, _ = types.UnpackBytesMost(msg, 1, 16)
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

func ParseCipherSuitesData(cipherSuitesData []byte) ([]types.CipherSuiteRecord, error) {
	offset := 0
	records := []types.CipherSuiteRecord{}

	for offset < len(cipherSuitesData) {
		csRecord := types.CipherSuiteRecord{}
		startOfRecord := cipherSuitesData[offset]
		csRecord.StartOfRecord = startOfRecord

		switch startOfRecord {
		case types.StandardCipherSuite:
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
			csRecord.CipherSuitID = types.CipherSuiteID(cipherSuitesData[offset])

		case types.OEMCipherSuite:
			// id + iana (3) + 3 algs (7 bytes)
			if offset+7 > len(cipherSuitesData)-1 {
				return records, fmt.Errorf("incomplete cipher suite data")
			}
			offset++
			csRecord.CipherSuitID = types.CipherSuiteID(cipherSuitesData[offset])
			offset++
			csRecord.OEMIanaID, _, _ = types.UnpackUint24L(cipherSuitesData, offset)

		default:
			return records, fmt.Errorf("bad start of record byte in the cipher suite data, value %x", startOfRecord)
		}

		for {
			offset++
			if offset > len(cipherSuitesData)-1 {
				break
			}

			algByte := cipherSuitesData[offset]
			if algByte == types.StandardCipherSuite || algByte == types.OEMCipherSuite {
				break
			}

			algTag := algByte & types.CipherAlgTagBitMask // clear lowest 6 bits
			algNumber := algByte & types.CipherAlgMask    // clear highest 2 bits
			switch algTag {
			case types.CipherAlgTagBitAuthMask:
				csRecord.AuthAlg = algNumber
			case types.CipherAlgTagBitIntegrityMask:
				csRecord.IntegrityAlgs = append(csRecord.IntegrityAlgs, algNumber)
			case types.CipherAlgTagBitEncryptionMask:
				csRecord.CryptAlgs = append(csRecord.CryptAlgs, algNumber)
			}
		}
		records = append(records, csRecord)
	}

	return records, nil
}

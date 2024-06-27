package ipmi

import (
	"fmt"
)

const (
	MaxCipherSuiteListIndex uint8 = 0x3f
)

// 22.15 Get Channel Cipher Suites Command
type GetChannelCipherSuitesRequest struct {
	ChannelNumber uint8 // Eh = retrieve information for channel this request was issued on
	PayloadType   PayloadType
	ListIndex     uint8
}

type GetChannelCipherSuitesResponse struct {
	ChannelNumber      uint8
	CipherSuiteRecords []byte
}

func (req *GetChannelCipherSuitesRequest) Command() Command {
	return CommandGetChannelCipherSuites
}

func (req *GetChannelCipherSuitesRequest) Pack() []byte {
	var msg = make([]byte, 3)
	packUint8(req.ChannelNumber, msg, 0)
	packUint8(uint8(req.PayloadType), msg, 1)
	packUint8(LIST_ALGORITHMS_BY_CIPHER_SUITE|req.ListIndex, msg, 2)
	return msg
}

func (res *GetChannelCipherSuitesResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}
	res.ChannelNumber, _, _ = unpackUint8(msg, 0)
	if len(msg) > 1 {
		res.CipherSuiteRecords, _, _ = unpackBytesMost(msg, 1, 16)
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
func (c *Client) GetChannelCipherSuites(channelNumber uint8, index uint8) (response *GetChannelCipherSuitesResponse, err error) {
	request := &GetChannelCipherSuitesRequest{
		ChannelNumber: channelNumber,
		PayloadType:   PayloadTypeIPMI,
		ListIndex:     index,
	}
	response = &GetChannelCipherSuitesResponse{}
	err = c.Exchange(request, response)
	return
}

// GetAllChannelCipherSuites initiates 64 (MaxCipherSuiteListIndex) requests
func (c *Client) GetAllChannelCipherSuites(channelNumber uint8) ([]CipherSuiteRecord, error) {
	var index uint8 = 0
	var cipherSuitesData = make([]byte, 0)
	for ; index < MaxCipherSuiteListIndex; index++ {
		res, err := c.GetChannelCipherSuites(channelNumber, index)
		if err != nil {
			return nil, fmt.Errorf("cmd GetChannelCipherSuites failed, err: %s", err)
		}
		cipherSuitesData = append(cipherSuitesData, res.CipherSuiteRecords...)
		if len(res.CipherSuiteRecords) < 16 {
			break
		}
	}

	c.DebugBytes("cipherSuitesData", cipherSuitesData, 16)
	return parseCipherSuitesData(cipherSuitesData)
}

func parseCipherSuitesData(cipherSuitesData []byte) ([]CipherSuiteRecord, error) {
	offset := 0
	records := []CipherSuiteRecord{}

	for offset < len(cipherSuitesData) {
		csRecord := CipherSuiteRecord{}
		startOfRecord := cipherSuitesData[offset]
		csRecord.StartOfRecord = startOfRecord

		switch startOfRecord {
		case StandardCipherSuite:
			// id + 3 algs (4 bytes)
			if offset+4 > len(cipherSuitesData)-1 {
				return records, fmt.Errorf("incomplete cipher suite data")
			}
			offset++
			csRecord.CipherSuitID = cipherSuitesData[offset]

		case OEMCipherSuite:
			// id + iana (3) + 3 algs (7 bytes)
			if offset+7 > len(cipherSuitesData)-1 {
				return records, fmt.Errorf("incomplete cipher suite data")
			}
			offset++
			csRecord.CipherSuitID = cipherSuitesData[offset]
			offset++
			csRecord.OEMIanaID, _, _ = unpackUint24L(cipherSuitesData, offset)

		default:
			return records, fmt.Errorf("bad start of record byte in the cipher suite data, value %x", startOfRecord)
		}

		for {
			offset++
			if offset > len(cipherSuitesData)-1 {
				break
			}

			algByte := cipherSuitesData[offset]
			if algByte == StandardCipherSuite || algByte == OEMCipherSuite {
				break
			}

			algTag := algByte & CipherAlgTagBitMask // clear lowest 6 bits
			algNumber := algByte & CipherAlgMask    // clear highest 2 bits
			switch algTag {
			case CipherAlgTagBitAuthMask:
				csRecord.AuthAlg = algNumber
			case CipherAlgTagBitIntegrityMask:
				csRecord.IntegrityAlgs = append(csRecord.IntegrityAlgs, algNumber)
			case CipherAlgTagBitEncryptionMask:
				csRecord.CryptAlgs = append(csRecord.CryptAlgs, algNumber)
			}
		}
		records = append(records, csRecord)
	}

	return records, nil
}

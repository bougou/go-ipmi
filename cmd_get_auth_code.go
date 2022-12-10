package ipmi

// GetAuthCodeRequest see 22.21
//
// This command is used to send a block of data to the BMC, whereupon the BMC will
// return a hash of the data together concatenated with the internally stored password for the given channel and user
type GetAuthCodeRequest struct {
	AuthType AuthType

	ChannelNumber uint8

	UserID uint8

	// data to hash (must be 16 bytes)
	Data [16]byte
}

type GetAuthCodeReponse struct {
	CompletionCode

	// For IPMI v1.5 AuthCode Number:
	AuthCode [16]byte

	// ForIPMI v2.0 Integrity Algorithum Number
	// Resultant hash, per selected Integrity algorithm. Up to 20 bytes. An
	// implementation can elect to return a variable length field based on the size of
	// the hash for the given integrity algorithm, or can return a fixed field where the
	// hash data is followed by 00h bytes as needed to pad the data to 20 bytes.
	Hash []byte
}

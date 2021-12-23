package ipmi

// 22.20
type GetSessionInfoRequest struct {
	*SessionHeader15

	SessionIndex uint8

	SessionHandle uint8

	SessionID uint32
}

type GetSessionInfoResponse struct {
	CompletionCode

	SessionHandle uint8

	PossbileActiveSessions uint8

	CurrentActiveSessions uint8

	UserID uint8

	OperatingPrivilegeLevel PrivilegeLevel

	SessionAuxiliaryData uint8 //  4bits

	ChannelNumber uint8 // 4bits

	// The following bytes 8:18 are optionally returned if Channel Type = 802.3 LAN:

	IPAddr uint32

	MacAddr []byte

	Port uint16

	// The following bytes 8:13 are returned if Channel Type = asynch. serial/modem:

	ActivityType uint8

	Destination uint8

	IPaddrPPP uint32

	// The following additional bytes 14:15 are returned if Channel Type = asynch.
	// serial/modem and connection is PPP:
	PortPPP uint16
}

func GetSession() {

}

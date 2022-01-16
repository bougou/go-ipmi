package ipmi

import (
	"fmt"
	"net"
)

// 22.20 Get Session Info Command
type GetSessionInfoRequest struct {
	// 00h = Return info for active session associated with session this command was received over.
	// N = get info for Nth active session
	// FEh = Look up session info according to Session Handle passed in this request.
	// FFh = Look up session info according to Session ID passed in this request.
	SessionIndex uint8

	SessionHandle uint8

	SessionID uint32
}

type GetSessionInfoResponse struct {
	SessionHandle          uint8 // Session Handle presently assigned to active session.
	PossbileActiveSessions uint8 // This value reflects the number of possible entries (slots) in the sessions table.
	CurrentActiveSessions  uint8 // Number of currently active sessions on all channels on this controller

	UserID                  uint8
	OperatingPrivilegeLevel PrivilegeLevel

	// [7:4] - Session protocol auxiliary data
	// For Channel Type = 802.3 LAN:
	// 0h = IPMI v1.5
	// 1h = IPMI v2.0/RMCP+
	AuxiliaryData uint8 // 4bits
	ChannelNumber uint8 // 4bits

	// if Channel Type = 802.3 LAN:
	RemoteConsoleIPAddr  net.IP           // IP Address of remote console (MS-byte first).
	RemoteConsoleMacAddr net.HardwareAddr // 6 bytes, MAC Address (MS-byte first)
	RemoteConsolePort    uint16           // Port Number of remote console (LS-byte first)

	// if Channel Type = asynch. serial/modem
	SessionChannelActivityType uint8
	DestinationSelector        uint8
	RemoteConsoleIPAddr_PPP    uint32 // If PPP connection: IP address of remote console. (MS-byte first) 00h, 00h, 00h, 00h otherwise.

	// if Channel Type = asynch. serial/modem and connection is PPP:
	RemoteConsolePort_PPP uint16
}

func (req *GetSessionInfoRequest) Command() Command {
	return CommandGetSessionInfo
}

func (req *GetSessionInfoRequest) Pack() []byte {
	out := make([]byte, 5)
	packUint8(req.SessionIndex, out, 0)
	if req.SessionIndex == 0xfe {
		packUint8(req.SessionHandle, out, 1)
		return out[0:2]
	}
	if req.SessionIndex == 0xff {
		packUint32L(req.SessionID, out, 1)
		return out[0:5]
	}
	return out[0:1]
}

func (res *GetSessionInfoResponse) Unpack(msg []byte) error {
	// at least 3 bytes
	if len(msg) < 3 {
		return ErrUnpackedDataTooShort
	}
	res.SessionHandle, _, _ = unpackUint8(msg, 0)
	res.PossbileActiveSessions, _, _ = unpackUint8(msg, 1)
	res.CurrentActiveSessions, _, _ = unpackUint8(msg, 2)

	if len(msg) == 3 {
		return nil
	}

	// if len(msg) > 3, then at least 6 bytes
	if len(msg) < 6 {
		return ErrUnpackedDataTooShort
	}
	res.UserID, _, _ = unpackUint8(msg, 3)
	b5, _, _ := unpackUint8(msg, 4)
	res.OperatingPrivilegeLevel = PrivilegeLevel(b5)
	b6, _, _ := unpackUint8(msg, 5)
	res.AuxiliaryData = b6 >> 4
	res.ChannelNumber = b6 & 0x0f

	//  Channel Type = 802.3 LAN:
	if len(msg) >= 18 {
		ipBytes, _, _ := unpackBytes(msg, 6, 4)
		res.RemoteConsoleIPAddr = net.IP(ipBytes)
		macBytes, _, _ := unpackBytes(msg, 10, 6)
		res.RemoteConsoleMacAddr = net.HardwareAddr(macBytes)
		res.RemoteConsolePort, _, _ = unpackUint16L(msg, 16)
	}

	if len(msg) >= 14 {
		res.SessionChannelActivityType, _, _ = unpackUint8(msg, 6)
		res.DestinationSelector, _, _ = unpackUint8(msg, 7)
		res.RemoteConsoleIPAddr_PPP, _, _ = unpackUint32(msg, 8)
		res.RemoteConsolePort_PPP, _, _ = unpackUint16L(msg, 12)
	}

	return nil
}

func (res *GetSessionInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSessionInfoResponse) Format() string {
	var sessionType string
	switch res.AuxiliaryData {
	case 0:
		sessionType = "IPMIv1.5"
	case 1:
		sessionType = "IPMIv2/RMCP+"
	}

	return fmt.Sprintf(`session handle                : %d
slot count                    : %d
active sessions               : %d
user id                       : %d
privilege level               : %s
session type                  : %s
channel number                : %#02x
console ip                    : %s
console mac                   : %s
console port                  : %d
	`,
		res.SessionHandle,
		res.PossbileActiveSessions,
		res.CurrentActiveSessions,
		res.UserID,
		res.OperatingPrivilegeLevel,
		sessionType,
		res.ChannelNumber,
		res.RemoteConsoleIPAddr,
		res.RemoteConsoleMacAddr,
		res.RemoteConsolePort,
	)
}

func (c *Client) GetSessionInfo(request *GetSessionInfoRequest) (response *GetSessionInfoResponse, err error) {
	response = &GetSessionInfoResponse{}
	err = c.Exchange(request, response)
	return
}

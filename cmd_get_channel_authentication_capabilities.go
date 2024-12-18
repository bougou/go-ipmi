package ipmi

import (
	"context"
	"fmt"
)

// 13.14
// 22.13
type GetChannelAuthenticationCapabilitiesRequest struct {
	// [7]
	// - 1b = get IPMI v2.0+ extended data.
	// If the given channel supports authentication but does not support RMCP+
	// (e.g. a serial channel), then the Response data should return with bit [5] of byte 4 = 0b, byte 5 should return 01h,
	//
	// - 0b = Backward compatible with IPMI v1.5. Response data only returns
	// bytes 1:9, bit [7] of byte 3 (Authentication Type Support) and bit [5] of byte 4 returns as 0b, bit [5] of byte byte 5 returns 00h.
	// [6:4] - reserved
	IPMIv20Extended bool
	// [3:0] - channel number.
	// 0h-Bh, Fh = channel numbers
	// Eh = retrieve information for channel this request was issued on
	ChannelNumber uint8

	// Requested Maximum Privilege Level
	MaximumPrivilegeLevel PrivilegeLevel
}

type GetChannelAuthenticationCapabilitiesResponse struct {
	// Channel number that the Authentication Capabilities is being returned for.
	// If the channel number in the request was set to Eh, this will return
	// the channel number for the channel that the request was received on
	ChannelNumber uint8

	// Returns the setting of the Authentication Type Enable field from the
	// configuration parameters for the given channel that corresponds to
	// the Requested Maximum Privilege Level.
	// [7] -
	// 1b = IPMI v2.0+ extended capabilities available. See Extended Capabilities field, below.
	// 0b = IPMI v1.5 support only.
	IPMIv20ExtendedAvailable bool
	// [5:0] - IPMI v1.5 Authentication type(s) enabled for given Requested Maximum Privilege Level
	AuthTypeNoneSupported           bool // bit 0
	AuthTypeMD2Supported            bool // bit 1
	AuthTypeMD5Supported            bool // bit 2
	AuthTypePasswordSupported       bool // bit 4
	AuthTypeOEMProprietarySupported bool // bit 5

	// [5] - Kg status (two-key login status).
	// Applies to v2.0/RMCP+ RAKP Authentication only. Otherwise, ignore as reserved.
	// 0b = Kg is set to default (all 0s).
	// 1b = Kg is set to non-zero value.
	KgStatus bool
	// [4] - Per-message Authentication status
	// 0b = Per-message Authentication is enabled.
	// 1b = Per-message Authentication is disabled.
	// Authentication Type "none" accepted for packets to the BMC after the session has been activated.
	PerMessageAuthenticationDisabled bool
	// [3] - User Level Authentication status
	// 0b = User Level Authentication is enabled.
	// 1b = User Level Authentication is disabled.
	// Authentication Type "none" accepted for User Level commands to the BMC.
	UserLevelAuthenticationDisabled bool
	// [2:0] - Anonymous Login status
	// This parameter returns values that tells the remote console whether
	// there are users on the system that have "null" usernames.
	// This can be used to guide the way the remote console presents login options to the user.
	// (see IPMI v1.5 specification sections 6.9.1, "Anonymous Login" Convention and 6.9.2, Anonymous Login )
	// [2] - 1b = Non-null usernames enabled. (One or more users are enabled that have non-null usernames).
	// [1] - 1b = Null usernames enabled (One or more users that have a null username, but non-null password, are presently enabled)
	// [0] - 1b = Anonymous Login enabled (A user that has a null username and null password is presently enabled)
	NonNullUsernamesEnabled bool
	NullUsernamesEnabled    bool
	AnonymousLoginEnabled   bool

	// For IPMI v1.5: - reserved
	// For IPMI v2.0+: - Extended Capabilities
	// [7:2] - reserved
	// [1] - 1b = channel supports IPMI v2.0 connections.
	// [0] - 1b = channel supports IPMI v1.5 connections.
	SupportIPMIv15 bool
	SupportIPMIv20 bool

	// IANA Enterprise Number for OEM/Organization that specified the particular
	// OEM Authentication Type for RMCP. Least significant byte first.
	// ONLY 3 bytes occupied. Return 00h, 00h, 00h if no OEM authentication type available.
	OEMID uint32

	// Additional OEM-specific information for the OEM Authentication Type for RMCP.
	// Return 00h if no OEM authentication type available.
	OEMAuxiliaryData uint8
}

func (req *GetChannelAuthenticationCapabilitiesRequest) Pack() []byte {
	var msg = make([]byte, 2)
	byte1 := req.ChannelNumber
	if req.IPMIv20Extended {
		byte1 = byte1 | 0x80
	}
	packUint8(byte1, msg, 0)
	packUint8(uint8(req.MaximumPrivilegeLevel), msg, 1)
	return msg
}

func (req *GetChannelAuthenticationCapabilitiesRequest) Command() Command {
	return CommandGetChannelAuthCapabilities
}

func (res *GetChannelAuthenticationCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 8 {
		return ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	res.ChannelNumber, _, _ = unpackUint8(msg, 0)

	b, _, _ := unpackUint8(msg, 1)
	res.IPMIv20ExtendedAvailable = isBit7Set(b)
	res.AuthTypeOEMProprietarySupported = isBit5Set(b)
	res.AuthTypePasswordSupported = isBit4Set(b)
	res.AuthTypeMD5Supported = isBit2Set(b)
	res.AuthTypeMD2Supported = isBit1Set(b)

	c, _, _ := unpackUint8(msg, 2)
	res.KgStatus = isBit5Set(c)
	res.PerMessageAuthenticationDisabled = isBit4Set(c)
	res.UserLevelAuthenticationDisabled = isBit3Set(c)
	res.NonNullUsernamesEnabled = isBit2Set(c)
	res.NullUsernamesEnabled = isBit1Set(c)
	res.AnonymousLoginEnabled = isBit0Set(c)

	d, _, _ := unpackUint8(msg, 3)
	if res.IPMIv20ExtendedAvailable {
		res.SupportIPMIv20 = isBit1Set(d)
		res.SupportIPMIv15 = isBit0Set(d)
	}

	res.OEMID, _, _ = unpackUint24L(msg, 4)
	res.OEMAuxiliaryData, _, _ = unpackUint8(msg, 7)
	return nil
}

func (*GetChannelAuthenticationCapabilitiesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetChannelAuthenticationCapabilitiesResponse) chooseAuthType() AuthType {
	if res.AuthTypeMD2Supported {
		return AuthTypeMD2
	}
	if res.AuthTypeMD5Supported {
		return AuthTypeMD5
	}
	if res.AuthTypePasswordSupported {
		return AuthTypePassword
	}
	if res.AuthTypeOEMProprietarySupported {
		return AuthTypeOEM
	}
	if res.AuthTypeNoneSupported {
		return AuthTypeNone
	}
	return AuthTypeNone
}

func (res *GetChannelAuthenticationCapabilitiesResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// GetChannelAuthenticationCapabilities is used to retrieve capability information
// about the channel that the message is delivered over, or for a particular channel.
// The command returns the authentication algorithm support for the given privilege level.
//
// This command is sent in unauthenticated (clear) format.
//
// When activating a session, the privilege level passed in this command will
// normally be the same Requested Maximum Privilege level that will be used
// for a subsequent Activate Session command.
func (c *Client) GetChannelAuthenticationCapabilities(ctx context.Context, channelNumber uint8, privilegeLevel PrivilegeLevel) (response *GetChannelAuthenticationCapabilitiesResponse, err error) {
	request := &GetChannelAuthenticationCapabilitiesRequest{
		IPMIv20Extended:       true,
		ChannelNumber:         channelNumber,
		MaximumPrivilegeLevel: privilegeLevel,
	}

	response = &GetChannelAuthenticationCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	if err != nil {
		return
	}

	c.session.authType = response.chooseAuthType()

	return
}

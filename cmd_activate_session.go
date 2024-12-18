package ipmi

import (
	"context"
	"fmt"
)

// 22.17
type ActivateSessionRequest struct {
	// Authentication Type for session.
	// The selected type will be used for session activation and for all subsequent authenticated packets under the session, unless "Per-message Authentication" or "User Level Authentication" are disabled.
	// (See 6.12.4, Per-Message and User Level Authentication Disables, for more information.)
	//
	// This value must match with the Authentication Type used in the Get Session Challenge request for the session. In addition, for multi-session channels this value must also match the authentication type used in the Session Header.
	AuthTypeForSession AuthType

	// Maximum privilege level requested. Indicates the highest privilege level that
	// may be requested for this session. This privilege level must be less than
	// or equal to the privilege limit for the channel and the privilege limit for the
	// user in order for the Activate Session command to be successful
	// (completion code = 00h). Once the Activate Session command has been
	// successful, the requested privilege level becomes a 'session limit' that
	// cannot be raised beyond the requested level, even if the user and/or
	// channel privilege level limits would allow it. I.e. it takes precedence over
	// the channel and user privilege level limits.
	//
	// [7:4] - reserved
	// [3:0] - Requested Maximum Privilege Level
	// 0h = reserved
	// 1h = Callback level
	// 2h = User level
	// 3h = Operator level
	// 4h = Administrator level
	// 5h = OEM Proprietary level
	// all other = reserved
	MaxPrivilegeLevel PrivilegeLevel

	// For multi-session channels: (e.g. LAN channel):
	// Challenge String data from corresponding Get Session Challenge response.
	//
	// For single-session channels that lack session header (e.g. serial/modem in Basic Mode):
	// Clear text password or AuthCode. See 22.17.1, AuthCode Algorithms.
	Challenge [16]byte // uint16

	// Initial Outbound Sequence Number = Starting sequence number that remote console wants used for messages from the BMC. (LS byte first). Must be non-null in order to establish a session. 0000_0000h = reserved. Can be any random value.
	//
	// The BMC must increment the outbound session sequence number by one (1) for
	// each subsequent outbound message from the BMC (include ActivateSessionResponse)
	//
	// The BMC sets the incremented number to Sequence field of SessionHeader.
	InitialOutboundSequenceNumber uint32
}

func (req *ActivateSessionRequest) Pack() []byte {
	out := make([]byte, 22)
	packUint8(uint8(req.AuthTypeForSession), out, 0)
	packUint8(uint8(req.MaxPrivilegeLevel), out, 1)
	packBytes(req.Challenge[:], out, 2)
	packUint32L(req.InitialOutboundSequenceNumber, out, 18)
	return out
}

func (req *ActivateSessionRequest) Command() Command {
	return CommandActivateSession
}

type ActivateSessionResponse struct {
	// Authentication Type for remainder of session
	AuthType AuthType

	// use this for remainder of session.
	// While atypical, the BMC is allowed to change the Session ID from the one that passed in the request.
	SessionID uint32

	// Initial inbound seq# = Sequence number that BMC wants remote console to use for subsequent messages in the session. The BMC returns a non-null value for multi-session connections and returns null (all 0s) for single-session connections.
	//
	// The remote console must increment the inbound session sequence number by one (1) for each subsequent message it sends to the BMC.
	InitialInboundSequenceNumber uint32

	// Maximum privilege level allowed for this session
	//  [7:4] - reserved
	//  [3:0] - Maximum Privilege Level allowed
	//  0h = reserved
	//  1h = Callback level
	//  2h = User level
	//  3h = Operator level
	//  4h = Administrator level
	//  5h = OEM Proprietary level
	//  all other = reserved
	MaxPrivilegeLevel uint8
}

func (res *ActivateSessionResponse) Unpack(data []byte) error {
	if len(data) < 10 {
		return ErrUnpackedDataTooShortWith(len(data), 10)
	}
	res.AuthType = AuthType(data[0])
	res.SessionID, _, _ = unpackUint32L(data, 1)
	res.InitialInboundSequenceNumber, _, _ = unpackUint32L(data, 5)
	res.MaxPrivilegeLevel = data[9]
	return nil
}

func (*ActivateSessionResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "No session slot available (BMC cannot accept any more sessions)",
		0x82: "No slot available for given user. (Limit of user sessions allowed under that name has been reached)",
		// (An implementation may only be able to support a certain number of
		// sessions based on what authentication resources are required. For
		// example, if User Level Authentication is disabled, an implementation
		// may be able to allow a larger number of users that are limited to User
		// Level privilege, than users that require higher privilege.)
		0x83: "No slot available to support user due to maximum privilege capability",
		0x84: "session sequence number out-of-range",
		0x85: "invalid Session ID in request",
		0x86: "requested maximum privilege level exceeds user and/or channel privilege limit",
	}
}

func (res *ActivateSessionResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// ActivateSession is only used for IPMI v1.5
func (c *Client) ActivateSession(ctx context.Context) (response *ActivateSessionResponse, err error) {
	request := &ActivateSessionRequest{
		AuthTypeForSession: c.session.authType,
		MaxPrivilegeLevel:  c.maxPrivilegeLevel,
		Challenge:          c.session.v15.challenge,

		// the outbound session sequence number is set by the remote console and can be any random value.
		InitialOutboundSequenceNumber: randomUint32(),
	}
	c.session.v15.outSeq = request.InitialOutboundSequenceNumber

	response = &ActivateSessionResponse{}

	// The Activate Session packet is typically authenticated.
	// We set session to active here to indicate this request should be authenticated
	// but if ActivateSession Command failed, we should set session active to false
	err = c.Exchange(ctx, request, response)
	if err != nil {
		return
	}
	c.session.v15.active = true
	c.session.v15.preSession = false

	// to use for the remainder of the session
	// Todo, validate the SessionID
	c.session.v15.sessionID = response.SessionID

	// The remote console must increment the inbound session sequence number
	// by one (1) for each subsequent message it sends to the BMC
	c.session.v15.inSeq = response.InitialInboundSequenceNumber

	return
}

package handlers

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

// IPMI session-management command IDs (NetFn 0x06 App).
const (
	CmdGetChannelAuthCapabilities uint8 = 0x38
	CmdActivateSession            uint8 = 0x3A
	CmdSetSessionPrivilegeLevel   uint8 = 0x3B
	CmdCloseSession               uint8 = 0x3C
)

// RegisterSessionHandlers adds IPMI 1.5 session and v2.0 RAKP handlers to r.
// Open Session and RAKP messages are dispatched differently (they arrive before
// a session exists); see [HandleOpenSession], [HandleRAKP1], [HandleRAKP3].
func RegisterSessionHandlers(r *Registry) {
	r.Register(NetFnAppRequest, CmdGetChannelAuthCapabilities, HandlerFunc(handleGetChannelAuthCaps))
	r.Register(NetFnAppRequest, CmdSetSessionPrivilegeLevel, HandlerFunc(handleSetSessionPrivilegeLevel))
	r.Register(NetFnAppRequest, CmdCloseSession, HandlerFunc(handleCloseSession))
}

// ---------------------------------------------------------------------------
// Get Channel Authentication Capabilities
// ---------------------------------------------------------------------------

// handleGetChannelAuthCaps implements Get Channel Authentication Capabilities (App 0x38).
// The server currently advertises IPMI 2.0 only (RAKP-HMAC-SHA1 + AES-CBC-128).
func handleGetChannelAuthCaps(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 2 {
		return nil, CodeRequestDataTruncated, nil
	}
	// req[0] bits 5:0 = channel number (0x0E = current)
	// req[1] bits 3:0 = requested privilege level

	resp := make([]byte, 8)
	resp[0] = 0x01 // channel number (1 = LAN)
	resp[1] = 0x00 // auth types supported (none for v2.0-only)
	resp[2] = 0x22 // bit 5: IPMI 2.0 supported; bit 1: user-level auth disabled allowed
	resp[3] = 0x00 // per-message auth support flags
	resp[4] = 0x00 // OEM ID byte 1
	resp[5] = 0x00 // OEM ID byte 2
	resp[6] = 0x00 // OEM ID byte 3
	resp[7] = 0x00 // OEM auxiliary data

	// Bit 5 of byte 2 = IPMI v2.0/RMCP+ support
	resp[2] = resp[2] | 0x20
	return resp, CodeOK, nil
}

// ---------------------------------------------------------------------------
// Set Session Privilege Level
// ---------------------------------------------------------------------------

func handleSetSessionPrivilegeLevel(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 1 {
		return nil, CodeRequestDataTruncated, nil
	}
	if hctx.Session == nil {
		return nil, CodeNotSupportedInState, nil
	}

	requested := bmc.PrivilegeLevel(req[0] & 0x0F)
	// Privilege 0 means "return current level" per spec.
	if requested == 0 {
		return []byte{uint8(hctx.Session.PrivilegeLevel)}, CodeOK, nil
	}
	if requested > hctx.Session.MaxPrivilege {
		return nil, CodeInsufficientPrivilege, nil
	}
	hctx.Session.PrivilegeLevel = requested
	return []byte{uint8(requested)}, CodeOK, nil
}

// ---------------------------------------------------------------------------
// Close Session
// ---------------------------------------------------------------------------

func handleCloseSession(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 4 {
		return nil, CodeRequestDataTruncated, nil
	}
	sessionID := binary.LittleEndian.Uint32(req[0:4])

	if err := hctx.BMC.Sessions.Close(sessionID); err != nil {
		return nil, CodeParamOutOfRange, nil
	}
	return nil, CodeOK, nil
}

// ---------------------------------------------------------------------------
// RMCP+ Open Session (payload type 0x10)
// ---------------------------------------------------------------------------

// OpenSessionRequest holds the fields from an RMCP+ Open Session Request message.
type OpenSessionRequest struct {
	MessageTag      uint8
	MaxPrivilege    uint8
	ConsoleID       uint32
	AuthAlgPayload  [8]byte
	IntAlgPayload   [8]byte
	CryptAlgPayload [8]byte
}

// OpenSessionResponse is the BMC's reply.
type OpenSessionResponse struct {
	MessageTag   uint8
	StatusCode   uint8
	MaxPrivilege uint8
	ConsoleID    uint32
	BMCID        uint32
	AuthAlg      uint8
	IntAlg       uint8
	CryptAlg     uint8
}

// HandleOpenSession processes an RMCP+ Open Session Request and returns the
// raw response payload.  It is called by the server before a session exists.
func HandleOpenSession(ctx context.Context, b *bmc.BMC, data []byte) ([]byte, error) {
	if len(data) < 32 {
		return buildOpenSessionError(0, 0, 0x12), nil // Illegal parameter
	}

	tag := data[0]
	maxPriv := data[1] & 0x0F
	consoleID := binary.LittleEndian.Uint32(data[4:8])

	// Parse algorithm payloads (3 x 8-byte records at offsets 8, 16, 24).
	authAlg := bmc.AuthAlg(data[11])     // byte 3 of auth payload
	intAlg := bmc.IntegrityAlg(data[19]) // byte 3 of integrity payload
	cryptAlg := bmc.CryptAlg(data[27])   // byte 3 of crypt payload

	// Validate algorithm support.  We support RAKP-HMAC-SHA1 (0x01),
	// HMAC-SHA1-96 (0x01), AES-CBC-128 (0x01) as the reference cipher suite.
	if authAlg != bmc.AuthAlgNone && authAlg != bmc.AuthAlgHMACSHA1 {
		return buildOpenSessionError(tag, consoleID, 0x04), nil // Invalid auth alg
	}
	if intAlg != bmc.IntegrityAlgNone && intAlg != bmc.IntegrityAlgHMACSHA1_96 {
		return buildOpenSessionError(tag, consoleID, 0x05), nil // Invalid integrity alg
	}
	if cryptAlg != bmc.CryptAlgNone && cryptAlg != bmc.CryptAlgAESCBC128 {
		return buildOpenSessionError(tag, consoleID, 0x10), nil // Invalid confidentiality alg
	}

	sess, err := b.Sessions.Allocate(consoleID, authAlg, intAlg, cryptAlg)
	if err != nil {
		return buildOpenSessionError(tag, consoleID, 0x01), nil // Insufficient resources
	}
	sess.MaxPrivilege = bmc.PrivilegeLevel(maxPriv)
	if sess.MaxPrivilege == 0 {
		sess.MaxPrivilege = bmc.PrivilegeLevelAdministrator
	}

	resp := make([]byte, 36)
	resp[0] = tag
	resp[1] = 0x00 // no error
	resp[2] = uint8(sess.MaxPrivilege)
	resp[3] = 0x00 // reserved
	binary.LittleEndian.PutUint32(resp[4:8], consoleID)
	binary.LittleEndian.PutUint32(resp[8:12], sess.BMCID)
	// Auth algorithm payload (8 bytes at offset 12)
	resp[12] = 0x00 // payload type = auth
	resp[13] = 0x00
	resp[14] = 0x00
	resp[15] = uint8(authAlg)
	resp[16] = 0x00
	resp[17] = 0x00
	resp[18] = 0x00
	resp[19] = 0x00
	// Integrity algorithm payload (8 bytes at offset 20)
	resp[20] = 0x01 // payload type = integrity
	resp[21] = 0x00
	resp[22] = 0x00
	resp[23] = uint8(intAlg)
	resp[24] = 0x00
	resp[25] = 0x00
	resp[26] = 0x00
	resp[27] = 0x00
	// Confidentiality algorithm payload (8 bytes at offset 28)
	resp[28] = 0x02 // payload type = confidentiality
	resp[29] = 0x00
	resp[30] = 0x00
	resp[31] = uint8(cryptAlg)
	resp[32] = 0x00
	resp[33] = 0x00
	resp[34] = 0x00
	resp[35] = 0x00

	return resp, nil
}

func buildOpenSessionError(tag uint8, consoleID uint32, statusCode uint8) []byte {
	resp := make([]byte, 8)
	resp[0] = tag
	resp[1] = statusCode
	resp[2] = 0x00
	resp[3] = 0x00
	binary.LittleEndian.PutUint32(resp[4:8], consoleID)
	return resp
}

// ---------------------------------------------------------------------------
// RAKP Message 1 → Message 2  (payload types 0x12, 0x13)
// ---------------------------------------------------------------------------

// HandleRAKP1 processes RAKP Message 1 and produces RAKP Message 2.
// It is called before the session is active; the session is identified by the
// BMC session ID embedded in Message 1.
func HandleRAKP1(ctx context.Context, b *bmc.BMC, data []byte) ([]byte, error) {
	if len(data) < 28 {
		return rakp2Error(0, 0, 0x12), nil
	}

	tag := data[0]
	bmcSessionID := binary.LittleEndian.Uint32(data[4:8])

	sess, err := b.Sessions.Get(bmcSessionID)
	if err != nil {
		return rakp2Error(tag, 0, 0x02), nil // Invalid Session ID
	}
	if sess.State != bmc.SessionStatePending {
		return rakp2Error(tag, sess.ConsoleID, 0x08), nil // Inactive Session ID
	}

	// Store the console's random number and requested role.
	copy(sess.ConsoleRand[:], data[8:24])
	sess.Role = data[24] // whole privilege byte including name-only bit

	usernameLen := data[27]
	if int(28+usernameLen) > len(data) {
		return rakp2Error(tag, sess.ConsoleID, 0x0C), nil // Invalid name length
	}
	username := string(data[28 : 28+usernameLen])

	// Look up user.
	user, lookupErr := b.Users.GetByName(username)
	if lookupErr != nil {
		// Spec says we must still generate a valid-looking response to prevent
		// user enumeration; we use a zero password for the HMAC then fail on RAKP3.
		user = nil
	}
	sess.User = user

	// Generate BMC random number.
	if _, err := rand.Read(sess.BMCRand[:]); err != nil {
		return rakp2Error(tag, sess.ConsoleID, 0xFF), fmt.Errorf("generate bmc rand: %w", err)
	}

	// Compute Key Exchange Authentication Code (HMAC over session params).
	authCode, err := computeRAKP2AuthCode(sess, b)
	if err != nil {
		return rakp2Error(tag, sess.ConsoleID, 0xFF), err
	}

	resp := make([]byte, 40+len(authCode))
	resp[0] = tag
	resp[1] = 0x00 // no error
	resp[2] = 0x00
	resp[3] = 0x00
	binary.LittleEndian.PutUint32(resp[4:8], sess.ConsoleID)
	copy(resp[8:24], sess.BMCRand[:])
	copy(resp[24:40], b.GUID[:])
	copy(resp[40:], authCode)
	return resp, nil
}

func rakp2Error(tag uint8, consoleID uint32, status uint8) []byte {
	resp := make([]byte, 8)
	resp[0] = tag
	resp[1] = status
	binary.LittleEndian.PutUint32(resp[4:8], consoleID)
	return resp
}

// ---------------------------------------------------------------------------
// RAKP Message 3 → Message 4  (payload types 0x14, 0x15)
// ---------------------------------------------------------------------------

// HandleRAKP3 processes RAKP Message 3, verifies the console's HMAC, derives
// session keys, marks the session active, and returns RAKP Message 4.
func HandleRAKP3(ctx context.Context, b *bmc.BMC, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return rakp4Error(0, 0, 0x12), nil
	}

	tag := data[0]
	statusCode := data[1]
	bmcSessionID := binary.LittleEndian.Uint32(data[4:8])

	sess, err := b.Sessions.Get(bmcSessionID)
	if err != nil {
		return rakp4Error(tag, 0, 0x02), nil // Invalid Session ID
	}

	// If the console sent a non-zero status in RAKP3, it means the console
	// rejected RAKP2.  Close the session and return an error response.
	if statusCode != 0x00 {
		_ = b.Sessions.Close(bmcSessionID)
		return rakp4Error(tag, sess.ConsoleID, statusCode), nil
	}

	// Verify the auth code sent by the console.
	authCodeLen := rakp3AuthCodeLen(sess.AuthAlg)
	if len(data) < 8+authCodeLen {
		return rakp4Error(tag, sess.ConsoleID, 0x0F), nil // Invalid integrity check value
	}
	consoleAuthCode := data[8 : 8+authCodeLen]

	expected, err := computeRAKP3AuthCode(sess, b)
	if err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}

	if sess.User == nil || !hmacEqual(expected, consoleAuthCode) {
		_ = b.Sessions.Close(bmcSessionID)
		return rakp4Error(tag, sess.ConsoleID, 0x0D), nil // Unauthorized name
	}

	// Derive SIK, K1, K2.
	if err := deriveSessKeys(sess, b); err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}
	sess.State = bmc.SessionStateActive
	sess.PrivilegeLevel = sess.MaxPrivilege

	// Compute RAKP4 auth code using SIK as HMAC key.
	rakp4Code, err := computeRAKP4AuthCode(sess, b)
	if err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}

	resp := make([]byte, 8+len(rakp4Code))
	resp[0] = tag
	resp[1] = 0x00
	resp[2] = 0x00
	resp[3] = 0x00
	binary.LittleEndian.PutUint32(resp[4:8], sess.ConsoleID)
	copy(resp[8:], rakp4Code)
	return resp, nil
}

func rakp4Error(tag uint8, consoleID uint32, status uint8) []byte {
	resp := make([]byte, 8)
	resp[0] = tag
	resp[1] = status
	binary.LittleEndian.PutUint32(resp[4:8], consoleID)
	return resp
}

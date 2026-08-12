package handlers

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

// IPMI session-management command IDs (NetFn 0x06 App).
const (
	CmdGetChannelAuthCapabilities uint8 = 0x38
	CmdGetSessionChallenge        uint8 = 0x39
	CmdActivateSession            uint8 = 0x3A

	lanChannelNumber uint8 = 1
)

// RegisterSessionHandlers adds IPMI 1.5 session and v2.0 RAKP handlers to r.
// Open Session and RAKP messages are dispatched differently (they arrive before
// a session exists); see [HandleOpenSession], [HandleRAKP1], [HandleRAKP3].
func RegisterSessionHandlers(r *Registry) {
	r.RegisterFunc(types.CommandGetChannelAuthCapabilities, handleGetChannelAuthCaps)
	r.RegisterFunc(types.CommandGetChannelCipherSuites, handleGetChannelCipherSuites)
	r.RegisterFunc(types.CommandSetSessionPrivilegeLevel, handleSetSessionPrivilegeLevel)
	r.RegisterFunc(types.CommandCloseSession, handleCloseSession)
	r.RegisterFunc(types.CommandGetSessionInfo, handleGetSessionInfo)
	registerV15SessionHandlers(r)
}

// ---------------------------------------------------------------------------
// Get Channel Authentication Capabilities
// ---------------------------------------------------------------------------

// handleGetChannelAuthCaps implements Get Channel Authentication Capabilities (App 0x38).
// Advertises both IPMI v1.5 (lan) and v2.0/RMCP+ (lanplus) when enabled on the BMC.
func handleGetChannelAuthCaps(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 2 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	// req[0] bits 3:0 = channel number (0x0E = current)
	// req[1] bits 3:0 = requested privilege level

	resp := make([]byte, 8)
	// resp[0] — channel number the capabilities are returned for.  The request
	// may use 0x0E to mean "the channel this request was received on".
	resp[0] = resolveChannelNumber(req[0])
	// resp[1] — auth type support (IPMI spec Table 22-15, byte 3):
	//   bit 7 = IPMI v2.0 extended capabilities available
	//   bits 5:0 = enabled IPMI v1.5 auth types
	resp[1] = 0x80
	if hctx.BMC != nil && hctx.BMC.V15LANEnabled() {
		for _, t := range hctx.BMC.ResolvedV15AuthTypes() {
			resp[1] |= bmc.V15AuthTypeToCapsBit(t)
		}
	}
	chNum := resolveChannelNumber(req[0])
	var ch *bmc.Channel
	if hctx.BMC != nil {
		ch, _ = hctx.BMC.Channels.Get(chNum)
	}
	fillChannelAuthCapsByte4(resp, hctx.BMC, ch)
	// resp[3] — extended capabilities (byte 5):
	//   bit 1 = IPMI v2.0 connections supported; bit 0 = IPMI v1.5 supported
	resp[3] = 0x02 // IPMI v2.0 (lanplus) always available on the reference server
	if hctx.BMC != nil && hctx.BMC.V15LANEnabled() {
		resp[3] |= 0x01
	}
	resp[4] = 0x00 // OEM ID byte 1
	resp[5] = 0x00 // OEM ID byte 2
	resp[6] = 0x00 // OEM ID byte 3
	resp[7] = 0x00 // OEM auxiliary data
	return resp, types.CodeOK, nil
}

// resolveChannelNumber maps the channel number field of a channel-scoped
// request to the concrete channel number.  Per IPMI spec, 0x0E means "the
// channel this request was received on"; the reference server only serves the
// LAN channel, so it resolves to lanChannelNumber.
func resolveChannelNumber(reqByte uint8) uint8 {
	ch := reqByte & 0x0F
	if ch == 0x0E {
		return lanChannelNumber
	}
	return ch
}

// handleGetChannelCipherSuites implements Get Channel Cipher Suites (App 0x54).
// Returns one record per cipher suite configured on the BMC (default
// {3, 17}), encoded per spec §22.15.1. Each standard record is:
//
//	0xC0 <id> 0x00|authAlg [0x40|integAlg] [0x80|cryptAlg]
//
// Records are returned in 16-byte windows addressed by the list index; the
// remote console keeps incrementing the index until fewer than 16 record bytes
// are returned.
func handleGetChannelCipherSuites(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 2 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return []byte{resolveChannelNumber(req[0])}, types.CodeOK, nil
	}
	// Byte 0: channel number (bits 3:0; 0x0E = current channel)
	// Byte 1: payload type (0x00 = IPMI)
	// Byte 2: bits 5:0 = list index; bit 6 = list mode flag (echoed unused here)
	record := cipherSuiteRecords(hctx.BMC)

	var listIndex int
	if len(req) >= 3 {
		listIndex = int(req[2] & 0x3F)
	}

	resp := []byte{resolveChannelNumber(req[0])}
	start := listIndex * 16
	if start < len(record) {
		end := start + 16
		if end > len(record) {
			end = len(record)
		}
		resp = append(resp, record[start:end]...)
	}
	return resp, types.CodeOK, nil
}

// ---------------------------------------------------------------------------
// Set Session Privilege Level
// ---------------------------------------------------------------------------

func handleSetSessionPrivilegeLevel(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 1 {
		return nil, types.CodeRequestDataTruncated, nil
	}

	requested := bmc.PrivilegeLevel(req[0] & 0x0F)

	if hctx.V15Session != nil {
		if requested == 0 {
			return []byte{uint8(hctx.V15Session.PrivilegeLevel)}, types.CodeOK, nil
		}
		if requested > hctx.V15Session.MaxPrivilege {
			return nil, types.CodeInsufficientPrivilege, nil
		}
		hctx.V15Session.PrivilegeLevel = requested
		return []byte{uint8(requested)}, types.CodeOK, nil
	}

	if hctx.Session == nil {
		return nil, types.CodeNotSupported, nil
	}

	// Privilege 0 means "return current level" per spec.
	if requested == 0 {
		return []byte{uint8(hctx.Session.PrivilegeLevel)}, types.CodeOK, nil
	}
	if requested > hctx.Session.MaxPrivilege {
		return nil, types.CodeInsufficientPrivilege, nil
	}
	hctx.Session.PrivilegeLevel = requested
	return []byte{uint8(requested)}, types.CodeOK, nil
}

// ---------------------------------------------------------------------------
// Close Session
// ---------------------------------------------------------------------------

func handleCloseSession(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 4 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	sessionID := binary.LittleEndian.Uint32(req[0:4])

	if err := hctx.BMC.Sessions.Close(sessionID); err == nil {
		return nil, types.CodeOK, nil
	}
	if err := hctx.BMC.V15Sessions.Close(sessionID); err != nil {
		return nil, types.CodeParameterOutOfRange, nil
	}
	return nil, types.CodeOK, nil
}

// ---------------------------------------------------------------------------
// Get Session Info
// ---------------------------------------------------------------------------

// Session-index request forms of Get Session Info (spec §22.20, request byte 1).
const (
	sessionIndexCurrent  uint8 = 0x00 // active session this command was received over
	sessionIndexByHandle uint8 = 0xFE // look up by session handle
	sessionIndexByID     uint8 = 0xFF // look up by session ID
)

// Session protocol auxiliary-data nibble (response byte 6, bits [7:4]) for an
// 802.3 LAN channel: the IPMI version the session negotiated.
const (
	sessionAuxV15 uint8 = 0x00 // IPMI v1.5
	sessionAuxV20 uint8 = 0x01 // IPMI v2.0 / RMCP+
)

// handleGetSessionInfo implements Get Session Info (App 0x3d).
//
// It answers the "current session" form (session index 0), which is what a
// remote console's keepalive polls (see the client's GetCurrentSessionInfo).
// The response carries the session handle, the total session-table capacity and
// occupancy, and the current session's user ID, operating privilege level and
// channel. Both RMCP+ and IPMI v1.5 sessions are supported.
//
// Lookup by session index N, by handle (0xFE) or by session ID (0xFF) is not
// implemented and returns CodeRequestDataFieldInvalid; the reference server has
// no need to describe sessions other than the caller's own.
func handleGetSessionInfo(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 1 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}
	if req[0] != sessionIndexCurrent {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	var (
		userID   uint8
		privLvl  uint8
		channel  uint8
		aux      uint8
		handle   uint8
		possible int
		active   int
		lanAddr  net.Addr
	)

	// The server holds the current session's lock for the whole dispatch, so
	// reading these session fields here needs no additional locking (see the
	// HandlerContext concurrency contract). Possible and active sessions
	// describe the caller's own session pool: the v2.0 and v1.5 tables are
	// independent, so summing them would advertise slots the caller can never
	// occupy. The RMCP+ store reports table occupancy (Count) as its active
	// count because per-session state is guarded by the session lock rather
	// than the store lock and so cannot be scanned race-free from here; outside
	// an in-flight handshake every occupied slot is an active session.
	switch {
	case hctx.Session != nil:
		aux = sessionAuxV20
		channel = hctx.Session.Channel
		privLvl = uint8(hctx.Session.PrivilegeLevel)
		handle = hctx.Session.Handle
		if hctx.Session.User != nil {
			userID = hctx.Session.User.ID
		}
		possible = hctx.BMC.Sessions.Cap()
		active = hctx.BMC.Sessions.Count()
		lanAddr = hctx.Session.GetAddr()
	case hctx.V15Session != nil:
		aux = sessionAuxV15
		channel = hctx.V15Session.Channel
		privLvl = uint8(hctx.V15Session.PrivilegeLevel)
		handle = hctx.V15Session.Handle
		if hctx.V15Session.User != nil {
			userID = hctx.V15Session.User.ID
		}
		possible = hctx.BMC.V15Sessions.Cap()
		active = hctx.BMC.V15Sessions.CountActiveSessions()
	default:
		// Received over the system interface: there is no current session to
		// describe.
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	resp := []byte{
		handle,
		clampSessionCount(possible),
		clampSessionCount(active),
		userID,
		privLvl,
		aux<<4 | channel&0x0F,
	}
	// 802.3 LAN trailer (Table 22-26 bytes 7-18): remote console IP, MAC, and
	// port. The emulator knows the console's UDP address; the MAC is not
	// tracked and real BMCs commonly zero it. The v1.5 path does not record the
	// console address, and clients tolerate the short response there.
	if ua, ok := lanAddr.(*net.UDPAddr); ok {
		if ip4 := ua.IP.To4(); ip4 != nil {
			resp = append(resp, ip4...)
			resp = append(resp, 0, 0, 0, 0, 0, 0) // MAC, not tracked
			resp = append(resp, uint8(ua.Port&0xFF), uint8(ua.Port>>8))
		}
	}
	return resp, types.CodeOK, nil
}

// clampSessionCount saturates a session count to the 6-bit field the Get
// Session Info response carries for possible/active sessions (spec Table 22-26;
// the high two bits are reserved), so an over-large count cannot bleed into
// them.
func clampSessionCount(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > 0x3F {
		return 0x3F
	}
	return uint8(n)
}

// ---------------------------------------------------------------------------
// RMCP+ Open Session (payload type 0x10)
// ---------------------------------------------------------------------------

// HandleOpenSession processes an RMCP+ Open Session Request and returns the
// raw response payload.  It is called by the server before a session exists.
func HandleOpenSession(ctx context.Context, b *bmc.BMC, data []byte) ([]byte, error) {
	var req rmcpplus.OpenSessionRequest
	if err := req.Unpack(data); err != nil {
		return buildOpenSessionError(0, 0, 0x12), nil // Illegal parameter
	}

	authAlg := types.AuthAlg(req.AuthAlg)
	intAlg := types.IntegrityAlg(req.IntegrityAlg)
	cryptAlg := types.CryptAlg(req.CryptAlg)
	tag := req.MessageTag
	consoleID := req.RemoteConsoleSessionID
	maxPriv := uint8(req.RequestedMaximumPrivilegeLevel)

	// Validate that the requested algorithm triple matches a configured
	// cipher suite (spec §22.15.2, §13.17). The triple must appear as a
	// unit — cross-suite recombinations are rejected even when each
	// individual algorithm exists in some configured suite. Error codes per
	// spec Table 13-17: 0x04 invalid auth, 0x05 invalid integrity, 0x10
	// invalid confidentiality.
	if ok, code := isCipherSuiteAllowed(b, authAlg, intAlg, cryptAlg); !ok {
		return buildOpenSessionError(tag, consoleID, code), nil
	}

	maxPrivilege := bmc.PrivilegeLevel(maxPriv)
	if maxPrivilege == 0 {
		maxPrivilege = bmc.PrivilegeLevelAdministrator
	}

	// Allocate fully initializes the session (including MaxPrivilege and
	// Channel) before inserting it into the store, so no lock-free field write
	// happens after it becomes reachable.
	sess, err := b.Sessions.Allocate(consoleID, authAlg, intAlg, cryptAlg, maxPrivilege, lanChannelNumber)
	if err != nil {
		return buildOpenSessionError(tag, consoleID, 0x01), nil // Insufficient resources
	}

	authPayload, integPayload, cryptPayload := rmcpplus.NewAlgorithmPayloads(authAlg, intAlg, cryptAlg)
	resp := &rmcpplus.OpenSessionResponse{
		MessageTag:             tag,
		RmcpStatusCode:         types.RmcpStatusCodeNoErrors,
		MaximumPrivilegeLevel:  uint8(sess.MaxPrivilege),
		RemoteConsoleSessionID: consoleID,
		ManagedSystemSessionID: sess.BMCID,
		AuthenticationPayload:  authPayload,
		IntegrityPayload:       integPayload,
		ConfidentialityPayload: cryptPayload,
	}
	return resp.Pack(), nil
}

func buildOpenSessionError(tag uint8, consoleID uint32, statusCode uint8) []byte {
	resp := &rmcpplus.OpenSessionResponse{
		MessageTag:             tag,
		RmcpStatusCode:         types.RmcpStatusCode(statusCode),
		RemoteConsoleSessionID: consoleID,
	}
	return resp.Pack()
}

// ---------------------------------------------------------------------------
// RAKP Message 1 → Message 2  (payload types 0x12, 0x13)
// ---------------------------------------------------------------------------

// HandleRAKP1 processes RAKP Message 1 and produces RAKP Message 2.
// It is called before the session is active; the session is identified by the
// BMC session ID embedded in Message 1.
func HandleRAKP1(ctx context.Context, b *bmc.BMC, data []byte) ([]byte, error) {
	var req rmcpplus.RAKPMessage1
	if err := req.Unpack(data); err != nil {
		return rakp2Error(0, 0, 0x12), nil
	}

	tag := req.MessageTag
	bmcSessionID := req.ManagedSystemSessionID

	sess, err := b.Sessions.Get(bmcSessionID)
	if err != nil {
		return rakp2Error(tag, 0, 0x02), nil // Invalid Session ID
	}
	// Hold ProcMu across every session field read/write below. Handshake
	// packets are dispatched outside the server's in-session path (they carry
	// no session ID in the RMCP+ header), so this handler must serialize
	// itself: duplicate RAKP1s for one pending session otherwise race on the
	// nonces and derived keys. ProcMu-then-store-lock is the allowed order;
	// the Close calls below take only the store lock.
	//
	// Deliberately no activity refresh here: RAKP1 carries no authenticator,
	// so refreshing would let anyone who saw a session ID keep that session
	// alive with replayed handshake packets. The inactivity budget stamped at
	// allocation bounds the whole handshake instead, and it is orders of
	// magnitude more than three round trips need.
	sess.ProcMu.Lock()
	defer sess.ProcMu.Unlock()
	if sess.State != bmc.SessionStatePending {
		return rakp2Error(tag, sess.ConsoleID, 0x08), nil // Inactive Session ID
	}

	// Store the console's random number and requested role.
	sess.ConsoleRand = req.RemoteConsoleRandomNumber
	sess.Role = req.Role() // whole privilege byte including name-only bit
	username := string(req.Username)

	// Look up user.
	user, lookupErr := b.Users.GetByName(username)
	if lookupErr != nil {
		// Spec says we must still generate a valid-looking response to prevent
		// user enumeration; we use a zero password for the HMAC then fail on RAKP3.
		user = nil
	}
	sess.User = user
	if user != nil {
		if status, ok := authorizeSessionPrivilege(b, sess); !ok {
			_ = b.Sessions.Close(bmcSessionID)
			return rakp2Error(tag, sess.ConsoleID, status), nil
		}
	}

	// Generate BMC random number.
	if _, err := rand.Read(sess.BMCRand[:]); err != nil {
		return rakp2Error(tag, sess.ConsoleID, 0xFF), fmt.Errorf("generate bmc rand: %w", err)
	}

	// Compute Key Exchange Authentication Code (HMAC over session params).
	authCode, err := computeRAKP2AuthCode(sess, b)
	if err != nil {
		return rakp2Error(tag, sess.ConsoleID, 0xFF), err
	}

	resp := &rmcpplus.RAKPMessage2{
		MessageTag:                    tag,
		RmcpStatusCode:                types.RmcpStatusCodeNoErrors,
		RemoteConsoleSessionID:        sess.ConsoleID,
		ManagedSystemRandomNumber:     sess.BMCRand,
		ManagedSystemGUID:             b.GUID,
		KeyExchangeAuthenticationCode: authCode,
	}
	return resp.Pack(), nil
}

func rakp2Error(tag uint8, consoleID uint32, status uint8) []byte {
	resp := &rmcpplus.RAKPMessage2{
		MessageTag:             tag,
		RmcpStatusCode:         types.RmcpStatusCode(status),
		RemoteConsoleSessionID: consoleID,
	}
	return resp.Pack()
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
	// Hold ProcMu across all session field reads/writes and key derivation.
	// See HandleRAKP1 for the lock-order rationale and for why the handshake
	// deliberately never refreshes session activity.
	sess.ProcMu.Lock()
	defer sess.ProcMu.Unlock()

	// If the console sent a non-zero status in RAKP3, it means the console
	// rejected RAKP2.  Close the session and return an error response.
	if statusCode != 0x00 {
		_ = b.Sessions.Close(bmcSessionID)
		return rakp4Error(tag, sess.ConsoleID, statusCode), nil
	}

	var req rmcpplus.RAKPMessage3
	if err := req.Unpack(data, sess.AuthAlg); err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0x0F), nil // Invalid integrity check value
	}

	expected, err := computeRAKP3AuthCode(sess, b)
	if err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}

	if sess.User == nil || !hmacEqual(expected, req.KeyExchangeAuthenticationCode) {
		_ = b.Sessions.Close(bmcSessionID)
		return rakp4Error(tag, sess.ConsoleID, 0x0D), nil // Unauthorized name
	}
	if status, ok := authorizeSessionPrivilege(b, sess); !ok {
		_ = b.Sessions.Close(bmcSessionID)
		return rakp4Error(tag, sess.ConsoleID, status), nil
	}

	// Derive SIK, K1, K2.
	if err := deriveSessKeys(sess, b); err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}
	if err := b.Sessions.Activate(bmcSessionID); err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0x02), nil // Invalid Session ID
	}
	sess.PrivilegeLevel = sess.MaxPrivilege

	// Compute RAKP4 auth code using SIK as HMAC key.
	rakp4Code, err := computeRAKP4AuthCode(sess, b)
	if err != nil {
		return rakp4Error(tag, sess.ConsoleID, 0xFF), err
	}

	resp := &rmcpplus.RAKPMessage4{
		MessageTag:           tag,
		RmcpStatusCode:       types.RmcpStatusCodeNoErrors,
		MgmtConsoleSessionID: sess.ConsoleID,
		IntegrityCheckValue:  rakp4Code,
	}
	return resp.Pack(), nil
}

func rakp4Error(tag uint8, consoleID uint32, status uint8) []byte {
	resp := &rmcpplus.RAKPMessage4{
		MessageTag:           tag,
		RmcpStatusCode:       types.RmcpStatusCode(status),
		MgmtConsoleSessionID: consoleID,
	}
	return resp.Pack()
}

func authorizeSessionPrivilege(b *bmc.BMC, sess *bmc.Session) (uint8, bool) {
	if sess.User == nil || !sess.User.Enabled {
		return 0x0D, false // Unauthorized name
	}

	requested, ok := requestedSessionPrivilege(sess)
	if !ok {
		return 0x09, false // Invalid role
	}
	if requested > sess.MaxPrivilege {
		return 0x0A, false // Unauthorized role or privilege level
	}

	ch, err := b.Channels.Get(sess.Channel)
	if err != nil || ch.AccessMode == bmc.ChannelAccessDisabled {
		return 0x0A, false
	}
	if requested > ch.MaxPrivilege {
		return 0x0A, false
	}

	access, ok := sess.User.ChannelAccess[sess.Channel]
	if !ok || !access.Enabled {
		return 0x0D, false // User is not enabled for this channel.
	}
	if access.CallbackOnly && requested != bmc.PrivilegeLevelCallback {
		return 0x0A, false
	}
	if access.MaxPrivilege == bmc.PrivilegeLevelNoAccess || requested > access.MaxPrivilege {
		return 0x0A, false
	}
	return 0x00, true
}

func requestedSessionPrivilege(sess *bmc.Session) (bmc.PrivilegeLevel, bool) {
	requested := bmc.PrivilegeLevel(sess.Role & 0x0F)
	if requested == 0 {
		requested = sess.MaxPrivilege
	}
	switch requested {
	case bmc.PrivilegeLevelCallback,
		bmc.PrivilegeLevelUser,
		bmc.PrivilegeLevelOperator,
		bmc.PrivilegeLevelAdministrator,
		bmc.PrivilegeLevelOEM:
		return requested, true
	default:
		return 0, false
	}
}

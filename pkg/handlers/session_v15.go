package handlers

import (
	"context"
	"encoding/binary"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

// v1.5 Activate Session / Get Session Challenge completion codes (spec Table 18-16/18-17).
const (
	ccV15InvalidUserName       CompletionCode = 0x81
	ccV15NullUserNotEnabled    CompletionCode = 0x82
	ccV15NoSessionSlot         CompletionCode = 0x81
	ccV15NoSlotForUser         CompletionCode = 0x82
	ccV15NoSlotForPrivilege    CompletionCode = 0x83
	CCV15SeqOutOfRange         CompletionCode = 0x84
	CCV15InvalidSessionID      CompletionCode = 0x85
	ccV15PrivilegeExceedsLimit CompletionCode = 0x86
)

func registerV15SessionHandlers(r *Registry) {
	r.Register(NetFnAppRequest, CmdGetSessionChallenge, HandlerFunc(handleGetSessionChallenge))
	r.Register(NetFnAppRequest, CmdActivateSession, HandlerFunc(handleActivateSession))
}

func handleGetSessionChallenge(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 17 {
		return nil, CodeRequestDataTruncated, nil
	}
	if hctx.BMC == nil {
		return nil, CodeUnspecifiedError, nil
	}

	authType := bmc.V15AuthType(req[0] & 0x0F)
	if !hctx.BMC.V15AuthTypeEnabled(authType) {
		return nil, CodeParamOutOfRange, nil
	}

	user, cc, ok := lookupV15User(hctx.BMC, req[1:17], lanChannelNumber)
	if !ok {
		return nil, cc, nil
	}

	var challenge [16]byte
	if err := bmc.GenerateChallenge(&challenge); err != nil {
		return nil, CodeUnspecifiedError, err
	}

	sess, err := hctx.BMC.V15Sessions.CreatePending(authType, user, challenge, lanChannelNumber)
	if err != nil {
		// Table 18-16 defines only 0x81/0x82; no slot-full code for this command.
		return nil, CodeUnspecifiedError, nil
	}

	resp := make([]byte, 20)
	binary.LittleEndian.PutUint32(resp[0:4], sess.TempSessionID)
	copy(resp[4:20], challenge[:])
	return resp, CodeOK, nil
}

func handleActivateSession(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 22 {
		return nil, CodeRequestDataTruncated, nil
	}
	if hctx.V15Session == nil || hctx.V15Session.State != bmc.V15SessionStatePending {
		return nil, CCV15InvalidSessionID, nil
	}
	sess := hctx.V15Session

	authType := bmc.V15AuthType(req[0] & 0x0F)
	if authType != sess.AuthType {
		return nil, CCV15InvalidSessionID, nil
	}

	requested := bmc.PrivilegeLevel(req[1] & 0x0F)
	if requested == 0 {
		return nil, CodeParamOutOfRange, nil
	}

	var reqChallenge [16]byte
	copy(reqChallenge[:], req[2:18])
	if reqChallenge != sess.Challenge {
		return nil, CCV15InvalidSessionID, nil
	}

	if cc, ok := authorizeV15Session(hctx.BMC, sess, requested); !ok {
		return nil, cc, nil
	}

	initialOutbound := binary.LittleEndian.Uint32(req[18:22])
	if initialOutbound == 0 {
		return nil, CCV15InvalidSessionID, nil
	}

	if hctx.BMC.V15Sessions.CountActiveSessions() >= bmc.MaxSessions {
		return nil, ccV15NoSessionSlot, nil
	}
	if sess.User != nil && hctx.BMC.V15Sessions.CountActiveSessionsForUser(sess.User.ID) >= 1 {
		return nil, ccV15NoSlotForUser, nil
	}
	if requested >= bmc.PrivilegeLevelOperator &&
		hctx.BMC.V15Sessions.CountActiveSessionsWithMaxPrivilegeAtLeast(bmc.PrivilegeLevelOperator) >= 2 {
		return nil, ccV15NoSlotForPrivilege, nil
	}

	inboundSeq, err := bmc.GenerateInboundSeq()
	if err != nil {
		return nil, CodeUnspecifiedError, err
	}

	permanentID, err := bmc.GenerateInboundSeq()
	if err != nil {
		return nil, CodeUnspecifiedError, err
	}

	if err := hctx.BMC.V15Sessions.Activate(sess, permanentID, inboundSeq, initialOutbound, requested); err != nil {
		return nil, ccV15NoSessionSlot, nil
	}

	ch, _ := hctx.BMC.Channels.Get(sess.Channel)
	respAuthType := sess.AuthType
	if ch != nil && !ch.PerMessageAuth {
		respAuthType = bmc.V15AuthTypeNone
	}

	resp := make([]byte, 10)
	resp[0] = uint8(respAuthType)
	binary.LittleEndian.PutUint32(resp[1:5], sess.SessionID)
	binary.LittleEndian.PutUint32(resp[5:9], inboundSeq)
	resp[9] = uint8(requested) // maximum privilege allowed for session
	return resp, CodeOK, nil
}

func lookupV15User(b *bmc.BMC, username []byte, channel uint8) (*bmc.User, CompletionCode, bool) {
	isNull := true
	for _, c := range username {
		if c != 0 {
			isNull = false
			break
		}
	}

	if isNull {
		user, err := b.Users.Get(1)
		if err != nil || !user.Enabled {
			return nil, ccV15NullUserNotEnabled, false
		}
		access, ok := user.ChannelAccess[channel]
		if !ok || !access.Enabled {
			return nil, ccV15NullUserNotEnabled, false
		}
		return user, CodeOK, true
	}

	name := trimV15Username(username)
	user, err := b.Users.FindEnabledByNameOnChannel(name, channel)
	if err != nil {
		return nil, ccV15InvalidUserName, false
	}
	return user, CodeOK, true
}

func trimV15Username(username []byte) string {
	end := len(username)
	for end > 0 && username[end-1] == 0 {
		end--
	}
	return string(username[:end])
}

func authorizeV15Session(b *bmc.BMC, sess *bmc.V15Session, requested bmc.PrivilegeLevel) (CompletionCode, bool) {
	if sess.User == nil || !sess.User.Enabled {
		return CCV15InvalidSessionID, false
	}

	ch, err := b.Channels.Get(sess.Channel)
	if err != nil || ch.AccessMode == bmc.ChannelAccessDisabled {
		return ccV15PrivilegeExceedsLimit, false
	}
	if requested > ch.MaxPrivilege {
		return ccV15PrivilegeExceedsLimit, false
	}

	access, ok := sess.User.ChannelAccess[sess.Channel]
	if !ok || !access.Enabled {
		return ccV15PrivilegeExceedsLimit, false
	}
	if access.CallbackOnly && requested != bmc.PrivilegeLevelCallback {
		return ccV15PrivilegeExceedsLimit, false
	}
	if access.MaxPrivilege == bmc.PrivilegeLevelNoAccess || requested > access.MaxPrivilege {
		return ccV15PrivilegeExceedsLimit, false
	}
	return CodeOK, true
}

// V15ResponseAuthType returns the session-header auth type for outbound v1.5
// packets after activation (spec Table 18-17 response byte 1).
func V15ResponseAuthType(ch *bmc.Channel, sess *bmc.V15Session) bmc.V15AuthType {
	if sess == nil {
		return bmc.V15AuthTypeNone
	}
	if ch != nil && !ch.PerMessageAuth {
		return bmc.V15AuthTypeNone
	}
	return sess.AuthType
}

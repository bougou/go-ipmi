package server

import (
	"context"
	"net"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/protocol"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// handleIPMIv15 dispatches IPMI v1.5 LAN packets (AuthType != 0x06).
func (s *Server) handleIPMIv15(addr net.Addr, pkt []byte) {
	if len(pkt) < 14 {
		return
	}

	var sess ipmi.Session15
	if err := sess.Unpack(pkt[4:]); err != nil {
		return
	}
	hdr := sess.SessionHeader15

	if hdr.AuthType != ipmi.AuthTypeNone && (s.bmc == nil || !s.bmc.V15LANEnabled()) {
		return
	}

	switch hdr.AuthType {
	case ipmi.AuthTypeNone:
		if hdr.SessionID != 0 {
			s.dispatchIPMIv15SessionUnauth(addr, pkt, &sess)
		} else {
			s.dispatchIPMIv15UnAuth(addr, pkt, &sess)
		}
	case ipmi.AuthTypeMD2, ipmi.AuthTypeMD5, ipmi.AuthTypePassword:
		s.dispatchIPMIv15Auth(addr, pkt, &sess)
	default:
		return
	}
}

func (s *Server) dispatchIPMIv15UnAuth(addr net.Addr, pkt []byte, sess *ipmi.Session15) {
	netFn, cmd, data, seq, ok := protocol.ParseIPMIRequest(sess.Payload)
	if !ok {
		return
	}

	ctx := context.Background()
	hctx := &handlers.HandlerContext{BMC: s.bmc}
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)

	ipmiResp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), respData)
	s.sendIPMIv15UnAuth(addr, pkt, ipmiResp)
}

// dispatchIPMIv15SessionUnauth handles AuthType NONE packets on an established
// session when per-message or user-level authentication is disabled (spec §6.11.4).
func (s *Server) dispatchIPMIv15SessionUnauth(addr net.Addr, pkt []byte, sess *ipmi.Session15) {
	hdr := sess.SessionHeader15
	netFn, cmd, data, seq, ok := protocol.ParseIPMIRequest(sess.Payload)
	if !ok {
		return
	}

	v15Sess, err := s.bmc.V15Sessions.Get(hdr.SessionID)
	if err != nil || v15Sess.State != bmc.V15SessionStateActive {
		return
	}

	ch, _ := s.bmc.Channels.Get(v15Sess.Channel)
	if !handlers.V15AllowsAuthTypeNone(ch, netFn, cmd, v15Sess) {
		return
	}
	if !v15Sess.TryAcceptInboundSeq(hdr.Sequence) {
		return
	}

	hctx := &handlers.HandlerContext{
		BMC:        s.bmc,
		V15Session: v15Sess,
		Channel:    ch,
		User:       v15Sess.User,
	}

	ctx := context.Background()
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)
	s.bmc.V15Sessions.Touch(v15Sess)

	ipmiResp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), respData)
	outboundSeq := v15Sess.NextOutboundSeq()
	s.sendIPMIv15Session(addr, pkt, v15Sess, ch, outboundSeq, ipmiResp, false)
}

func (s *Server) dispatchIPMIv15Auth(addr net.Addr, pkt []byte, sess *ipmi.Session15) {
	hdr := sess.SessionHeader15
	netFn, cmd, data, seq, ok := protocol.ParseIPMIRequest(sess.Payload)
	if !ok {
		return
	}

	v15Sess, err := s.bmc.V15Sessions.Get(hdr.SessionID)
	if err != nil {
		return
	}

	authType := bmc.V15AuthType(hdr.AuthType)
	if authType != v15Sess.AuthType {
		if v15Sess.State == bmc.V15SessionStatePending && cmd == handlers.CmdActivateSession {
			s.sendIPMIv15CommandCC(addr, pkt, v15Sess, netFn, cmd, seq, handlers.CCV15InvalidSessionID, true)
		}
		return
	}

	lookupID := hdr.SessionID
	sessionSeq := hdr.Sequence
	pendingActivate := v15Sess.State == bmc.V15SessionStatePending

	if pendingActivate {
		if cmd != handlers.CmdActivateSession {
			return
		}
		if hdr.Sequence != 0 {
			return
		}
		lookupID = v15Sess.TempSessionID
		sessionSeq = 0
	} else if v15Sess.State == bmc.V15SessionStateActive {
		// Non-destructive window check first — do NOT consume the slot
		// until the auth code is verified, otherwise a spoofed packet
		// with an in-window sequence number can exhaust the window (DoS).
		if !bmc.V15InboundSeqValid(v15Sess, hdr.Sequence) {
			if cmd == handlers.CmdActivateSession {
				s.sendIPMIv15CommandCC(addr, pkt, v15Sess, netFn, cmd, seq, handlers.CCV15SeqOutOfRange, true)
			}
			return
		}
	} else {
		return
	}

	password := v15Sess.User.PasswordV15Padded()
	if !handlers.VerifyV15AuthCode(password, authType, lookupID, sess.Payload, sessionSeq, hdr.AuthCode) {
		if pendingActivate {
			s.sendIPMIv15CommandCC(addr, pkt, v15Sess, netFn, cmd, seq, handlers.CCV15InvalidSessionID, true)
		}
		return
	}

	// Auth verified — now commit the sequence-window state.
	if !pendingActivate {
		if !v15Sess.TryAcceptInboundSeq(hdr.Sequence) {
			// Lost a race (e.g. duplicate accepted concurrently); drop.
			return
		}
	}

	ch, _ := s.bmc.Channels.Get(v15Sess.Channel)
	hctx := &handlers.HandlerContext{
		BMC:        s.bmc,
		V15Session: v15Sess,
		Channel:    ch,
		User:       v15Sess.User,
	}

	ctx := context.Background()
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)
	s.bmc.V15Sessions.Touch(v15Sess)

	ipmiResp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), respData)
	outboundSeq := v15Sess.NextOutboundSeq()
	useAuth := cmd == handlers.CmdActivateSession ||
		handlers.V15ResponseAuthType(ch, v15Sess) != bmc.V15AuthTypeNone
	s.sendIPMIv15Session(addr, pkt, v15Sess, ch, outboundSeq, ipmiResp, useAuth)
}

func (s *Server) sendIPMIv15UnAuth(addr net.Addr, reqPkt []byte, ipmiResp []byte) {
	respHdr := ipmi.SessionHeader15{
		AuthType:      ipmi.AuthTypeNone,
		PayloadLength: uint8(len(ipmiResp)),
	}
	rmcp := []byte{reqPkt[0], reqPkt[1], 0xFF, reqPkt[3]}
	out := append(rmcp, respHdr.Pack()...)
	out = append(out, ipmiResp...)
	_, _ = s.conn.WriteTo(out, addr)
}

func (s *Server) sendIPMIv15CommandCC(addr net.Addr, reqPkt []byte, sess *bmc.V15Session, netFn, cmd, seq uint8, cc handlers.CompletionCode, authenticated bool) {
	ipmiResp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), nil)
	var outboundSeq uint32
	if sess != nil && sess.State == bmc.V15SessionStateActive {
		outboundSeq = sess.NextOutboundSeq()
	}
	ch, _ := s.bmc.Channels.Get(sess.Channel)
	s.sendIPMIv15Session(addr, reqPkt, sess, ch, outboundSeq, ipmiResp, authenticated)
}

func (s *Server) sendIPMIv15Session(addr net.Addr, reqPkt []byte, sess *bmc.V15Session, ch *bmc.Channel, outboundSeq uint32, ipmiResp []byte, authenticated bool) {
	var respHdr ipmi.SessionHeader15
	if authenticated && sess != nil {
		authType := handlers.V15ResponseAuthType(ch, sess)
		if authType == bmc.V15AuthTypeNone {
			respHdr = ipmi.SessionHeader15{
				AuthType:      ipmi.AuthTypeNone,
				Sequence:      outboundSeq,
				SessionID:     sess.SessionID,
				PayloadLength: uint8(len(ipmiResp)),
			}
		} else {
			authCode := handlers.GenV15AuthCode(
				sess.User.PasswordV15Padded(),
				sess.AuthType,
				sess.SessionID,
				ipmiResp,
				outboundSeq,
			)
			respHdr = ipmi.SessionHeader15{
				AuthType:      ipmi.AuthType(sess.AuthType),
				Sequence:      outboundSeq,
				SessionID:     sess.SessionID,
				AuthCode:      authCode,
				PayloadLength: uint8(len(ipmiResp)),
			}
		}
	} else {
		sessionID := uint32(0)
		if sess != nil {
			if sess.State == bmc.V15SessionStatePending {
				sessionID = sess.TempSessionID
			} else {
				sessionID = sess.SessionID
			}
		}
		respHdr = ipmi.SessionHeader15{
			AuthType:      ipmi.AuthTypeNone,
			Sequence:      outboundSeq,
			SessionID:     sessionID,
			PayloadLength: uint8(len(ipmiResp)),
		}
	}

	rmcp := []byte{reqPkt[0], reqPkt[1], 0xFF, reqPkt[3]}
	out := append(rmcp, respHdr.Pack()...)
	out = append(out, ipmiResp...)
	_, _ = s.conn.WriteTo(out, addr)
}

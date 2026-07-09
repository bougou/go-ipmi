package server

// server.go is the entry point for the IPMI BMC server.
//
// The Server ties together a [transport.PacketConn] (how packets arrive), a
// [bmc.BMC] (all BMC state), and a [handlers.Registry] (what each command
// does).  Everything is injected via [ServerOption] functional options so the
// same binary can run as a simulator, a virtual-BMC for VMs, or a real
// embedded BMC with a custom HAL.
//
// Composability
//
// - Attach to an existing socket by passing a pre-bound [transport.PacketConn].
// - Override any handler with [WithHandlerRegistry].
// - Substitute hardware via the [hal.HAL] you pass to [bmc.New].
// - Replace time via a custom [clock.Clock] in [bmc.WithClock].

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/transport"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Payload type aliases from pkg/protocol for readability within this file.
const (
	srvPayloadIPMI                = protocol.PayloadIPMI
	srvPayloadOpenSessionResponse = protocol.PayloadOpenSessionResponse
	srvPayloadOpenSessionRequest  = protocol.PayloadOpenSessionRequest
	srvPayloadRAKPMessage1        = protocol.PayloadRAKPMessage1
	srvPayloadRAKPMessage2        = protocol.PayloadRAKPMessage2
	srvPayloadRAKPMessage3        = protocol.PayloadRAKPMessage3
	srvPayloadRAKPMessage4        = protocol.PayloadRAKPMessage4
)

const defaultBufferSize = 4096

// Server is an IPMI BMC server.
//
// Create one with [NewServer], then call [Server.Serve] to start accepting
// packets.  Serve blocks until ctx is cancelled or [Server.Close] is called.
type Server struct {
	bmc     *bmc.BMC
	conn    transport.PacketConn
	reg     *handlers.Registry
	clk     clock.Clock
	bufSize int

	mu     sync.Mutex
	closed bool
}

// ServerOption configures a [Server].
type ServerOption func(*Server)

// WithHandlerRegistry replaces the default handler registry.
// Use this to add OEM commands or override built-in handlers.
func WithHandlerRegistry(r *handlers.Registry) ServerOption {
	return func(s *Server) { s.reg = r }
}

// WithServerBufferSize sets the UDP read buffer size (default 4096).
func WithServerBufferSize(n int) ServerOption {
	return func(s *Server) { s.bufSize = n }
}

// WithCipherSuites configures the RMCP+ cipher suites the server advertises
// and accepts. Each ID must be a suite the reference server implements
// (validated by [bmc.BMC.SetCipherSuites]); passing an unsupported suite
// panics. Use this to advertise only a subset, e.g. only suite 17.
func WithCipherSuites(ids []types.CipherSuiteID) ServerOption {
	return func(s *Server) {
		if s.bmc != nil {
			s.bmc.SetCipherSuites(ids)
		}
	}
}

// WithV15AuthTypes configures IPMI v1.5 authentication types the server
// advertises and accepts (lan / -A MD5). Mirrors [bmc.WithV15AuthTypes].
func WithV15AuthTypes(types []bmc.V15AuthType) ServerOption {
	return func(s *Server) {
		if s.bmc != nil {
			bmc.WithV15AuthTypes(types)(s.bmc)
		}
	}
}

// WithV15Disabled disables IPMI v1.5 LAN sessions. RMCP+ (lanplus) is unaffected.
func WithV15Disabled() ServerOption {
	return func(s *Server) {
		if s.bmc != nil {
			bmc.WithV15Disabled()(s.bmc)
		}
	}
}

// NewServer creates a Server.
//
// b is the BMC state (create with [bmc.New]).
// conn is the packet transport (e.g., from [transport/udp.Listen]).
// opts are applied in order.
//
// A default [handlers.Registry] populated with all standard commands is used
// unless overridden via [WithHandlerRegistry].
func NewServer(b *bmc.BMC, conn transport.PacketConn, opts ...ServerOption) *Server {
	reg := handlers.NewRegistry()
	handlers.RegisterAppHandlers(reg)
	handlers.RegisterSessionHandlers(reg)
	handlers.RegisterChassisHandlers(reg)

	s := &Server{
		bmc:     b,
		conn:    conn,
		reg:     reg,
		clk:     b.Clock(),
		bufSize: defaultBufferSize,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Serve reads packets from the transport and dispatches them until ctx is
// cancelled or [Server.Close] is called.
func (s *Server) Serve(ctx context.Context) error {
	evictCtx, evictCancel := context.WithCancel(ctx)
	defer evictCancel()
	go s.runSessionEviction(evictCtx)

	buf := make([]byte, s.bufSize)
	for {
		// Respect context cancellation between reads.
		if err := ctx.Err(); err != nil {
			return err
		}

		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			if s.isClosed() || errors.Is(err, net.ErrClosed) {
				return nil
			}
			// Timeout or transient error – keep looping.
			continue
		}

		pkt := make([]byte, n)
		copy(pkt, buf[:n])
		go s.handlePacket(ctx, addr, pkt)
	}
}

// Close shuts down the server and its transport.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.conn.Close()
}

func (s *Server) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// runSessionEviction periodically removes v1.5 and RMCP+ sessions that exceeded
// the configured inactivity timeout (spec: 60s default; see bmc.DefaultInactivityTimeout).
func (s *Server) runSessionEviction(ctx context.Context) {
	ticker := s.clk.NewTicker(bmc.DefaultSessionEvictInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C():
			if s.bmc != nil {
				s.bmc.Sessions.EvictExpired()
				s.bmc.V15Sessions.EvictExpired()
			}
		}
	}
}

// handlePacket is the top-level packet dispatcher.
func (s *Server) handlePacket(ctx context.Context, addr net.Addr, pkt []byte) {
	if len(pkt) < 4 {
		return
	}

	// RMCP header: version(1) reserved(1) seq(1) class(1)
	// class 0x07 = IPMI, class 0x06 = ASF
	msgClass := pkt[3] & 0x1F
	switch msgClass {
	case 0x06: // ASF
		s.handleASF(ctx, addr, pkt)
	case 0x07: // IPMI
		s.handleIPMI(ctx, addr, pkt)
	}
}

// handleASF handles RMCP/ASF Presence Ping (used by ipmitool -p 623 ping).
func (s *Server) handleASF(ctx context.Context, addr net.Addr, pkt []byte) {
	if len(pkt) < 12 {
		return
	}
	msgType := pkt[8] // ASF message type
	if msgType != 0x80 {
		return // only handle Presence Ping (0x80)
	}
	tag := pkt[9]

	_, _ = s.conn.WriteTo(protocol.BuildASFPresencePong(tag), addr)
}

// handleIPMI routes a raw IPMI-class RMCP packet.
func (s *Server) handleIPMI(_ context.Context, addr net.Addr, pkt []byte) {
	if len(pkt) < 5 {
		return
	}

	// Byte 4 is AuthType for v1.5 or 0x06 (AuthTypeRMCPPlus) for v2.0.
	authTypeByte := pkt[4]
	if authTypeByte == 0x06 {
		s.handleRMCPPlus(addr, pkt)
	} else {
		s.handleIPMIv15(addr, pkt)
	}
}

// handleRMCPPlus routes RMCP+ (IPMI 2.0) packets.
func (s *Server) handleRMCPPlus(addr net.Addr, pkt []byte) {
	sessionID, inboundSeq, payloadType, flags, payload, ok := protocol.ParseRMCPPlusHeader(pkt)
	if !ok {
		return
	}
	encrypted := flags&protocol.PayloadEncryptedFlag != 0
	authenticated := flags&protocol.PayloadAuthenticatedFlag != 0

	ctx := context.Background()

	switch payloadType {
	case srvPayloadOpenSessionRequest:
		resp, err := handlers.HandleOpenSession(ctx, s.bmc, payload)
		if err != nil || resp == nil {
			return
		}
		s.sendRMCPPlus(addr, srvPayloadOpenSessionResponse, 0, resp)

	case srvPayloadRAKPMessage1:
		resp, err := handlers.HandleRAKP1(ctx, s.bmc, payload)
		if err != nil || resp == nil {
			return
		}
		s.sendRMCPPlus(addr, srvPayloadRAKPMessage2, 0, resp)

	case srvPayloadRAKPMessage3:
		resp, err := handlers.HandleRAKP3(ctx, s.bmc, payload)
		if err != nil || resp == nil {
			return
		}
		s.sendRMCPPlus(addr, srvPayloadRAKPMessage4, 0, resp)

	case srvPayloadIPMI:
		if sessionID == 0 {
			// Pre-session IPMI command (e.g., GetChannelAuthCaps).
			s.dispatchIPMIPreSession(ctx, addr, payload)
			return
		}
		sess, err := s.bmc.Sessions.Get(sessionID)
		if err != nil {
			return
		}
		if !verifyRMCPPlusIntegrity(pkt, sess, authenticated) {
			return
		}
		if !bmc.InboundSeqValid(sess.InboundSeq, inboundSeq) {
			return
		}
		sess.InboundSeq = inboundSeq
		s.dispatchIPMISession(ctx, addr, sess, payload, encrypted)
	}
}

// dispatchIPMIPreSession handles IPMI commands that arrive before a session exists.
func (s *Server) dispatchIPMIPreSession(ctx context.Context, addr net.Addr, payload []byte) {
	netFn, cmd, data, seq, ok := protocol.ParseIPMIRequest(payload)
	if !ok {
		return
	}
	hctx := &handlers.HandlerContext{BMC: s.bmc}
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)
	resp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), respData)
	s.sendRMCPPlus(addr, srvPayloadIPMI, 0, resp)
}

// dispatchIPMISession handles IPMI commands within an authenticated session.
func (s *Server) dispatchIPMISession(ctx context.Context, addr net.Addr, sess *bmc.Session, payload []byte, encrypted bool) {
	ipmiPayload := payload
	if encrypted && len(sess.K2) >= 16 {
		dec, err := decryptPayload(payload, sess.K2)
		if err != nil {
			return
		}
		ipmiPayload = dec
	}

	netFn, cmd, data, seq, ok := protocol.ParseIPMIRequest(ipmiPayload)
	if !ok {
		return
	}

	ch, _ := s.bmc.Channels.Get(sess.Channel)
	hctx := &handlers.HandlerContext{
		BMC:     s.bmc,
		Session: sess,
		Channel: ch,
		User:    sess.User,
	}
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)
	rawResp := protocol.BuildIPMIResponse(netFn, cmd, seq, uint8(cc), respData)

	var finalPayload []byte
	flags := uint8(0)
	if encrypted && len(sess.K2) >= 16 {
		enc, err := encryptPayload(rawResp, sess.K2)
		if err != nil {
			return
		}
		finalPayload = enc
		flags |= protocol.PayloadEncryptedFlag
	} else {
		finalPayload = rawResp
	}

	sess.OutboundSeq++
	s.sendRMCPPlusSession(addr, srvPayloadIPMI, flags, sess, finalPayload)
}

// sendRMCPPlus sends a session-zero (unauthenticated) RMCP+ packet.
func (s *Server) sendRMCPPlus(addr net.Addr, payloadType, flags uint8, payload []byte) {
	pkt := protocol.BuildRMCPPlusPacket(payloadType, flags, 0, 0, payload)
	_, _ = s.conn.WriteTo(pkt, addr)
}

// sendRMCPPlusSession sends an authenticated / optionally encrypted RMCP+ packet.
func (s *Server) sendRMCPPlusSession(addr net.Addr, payloadType, flags uint8, sess *bmc.Session, payload []byte) {
	if sess.IntegrityAlg != types.IntegrityAlg_None {
		flags |= protocol.PayloadAuthenticatedFlag
	}
	pkt := protocol.BuildRMCPPlusPacket(payloadType, flags, sess.ConsoleID, sess.OutboundSeq, payload)
	var ok bool
	pkt, ok = appendRMCPPlusIntegrity(pkt, sess)
	if !ok {
		return
	}
	_, _ = s.conn.WriteTo(pkt, addr)
}

// decryptPayload and encryptPayload delegate to the AES-CBC-128 helpers
// that already exist in helpers_hmac.go in the main package.
// We reproduce small wrappers here to avoid importing ourselves.

// decryptPayload decrypts an AES-CBC-128 confidential payload and strips the
// IPMI 2.0 padding (spec §13.29).  The wire format is:
//
//	IV(16) || AES-CBC( payload || pad bytes || pad-length )
//
// where the final decrypted byte is the number of pad bytes (0..15).  The
// returned slice is the original payload with pad bytes and the pad-length
// byte removed.
func decryptPayload(cipherText, k2 []byte) ([]byte, error) {
	if len(cipherText) < 16 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := cipherText[:16]
	padded, err := decryptAES(cipherText[16:], k2[:16], iv)
	if err != nil {
		return nil, err
	}
	if len(padded) == 0 {
		return nil, fmt.Errorf("decrypted payload is empty")
	}
	padLen := int(padded[len(padded)-1])
	// pad-length must fit within the trailing pad region; a value >= len-1
	// would leave no payload and indicates a corrupted/invalid padding.
	if padLen >= len(padded) {
		return nil, fmt.Errorf("invalid AES pad length %d for %d-byte block", padLen, len(padded))
	}
	return padded[:len(padded)-1-padLen], nil
}

func encryptPayload(plain, k2 []byte) ([]byte, error) {
	padded, _ := aesPad(plain)
	iv := randomBytes(16)
	encrypted, err := encryptAES(padded, k2[:16], iv)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 16+len(encrypted))
	copy(out[:16], iv)
	copy(out[16:], encrypted)
	return out, nil
}

// aesPad pads plain to a multiple of 16 bytes per IPMI 2.0 spec §13.29.
func aesPad(plain []byte) ([]byte, int) {
	padLen := 16 - (len(plain)+1)%16
	if padLen == 16 {
		padLen = 0
	}
	padded := make([]byte, len(plain)+padLen+1)
	copy(padded, plain)
	for i := 0; i < padLen; i++ {
		padded[len(plain)+i] = byte(i + 1)
	}
	padded[len(plain)+padLen] = byte(padLen)
	return padded, padLen
}

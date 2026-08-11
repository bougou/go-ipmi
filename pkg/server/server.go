package server

// Package server is the entry point for the IPMI BMC server.
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
	"os"
	"sync"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/transport"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Payload type aliases from pkg/protocol for readability within this file.
const (
	srvPayloadIPMI                = protocol.PayloadIPMI
	srvPayloadSOL                 = protocol.PayloadSOL
	srvPayloadOpenSessionResponse = protocol.PayloadOpenSessionResponse
	srvPayloadOpenSessionRequest  = protocol.PayloadOpenSessionRequest
	srvPayloadRAKPMessage1        = protocol.PayloadRAKPMessage1
	srvPayloadRAKPMessage2        = protocol.PayloadRAKPMessage2
	srvPayloadRAKPMessage3        = protocol.PayloadRAKPMessage3
	srvPayloadRAKPMessage4        = protocol.PayloadRAKPMessage4
)

const defaultBufferSize = 4096

// WithSOLDebug traces every SOL data-plane packet (decoded fields on
// receipt/send, raw wire bytes on transmit) to stderr. SOL failures are
// otherwise invisible to command tracing: the data plane rides the session
// directly and never passes through the handler registry.
func WithSOLDebug() ServerOption {
	return func(s *Server) { s.solDebug = true }
}

// solStamp prefixes SOL trace lines with a millisecond timestamp: stall and
// retry patterns in the data plane are only readable with per-packet timing.
func solStamp() string { return time.Now().Format("15:04:05.000") }

// Server is an IPMI BMC server.
//
// Create one with [NewServer], then call [Server.Serve] to start accepting
// packets.  Serve blocks until ctx is cancelled or [Server.Close] is called.
type Server struct {
	bmc      *bmc.BMC
	conn     transport.PacketConn
	reg      *handlers.Registry
	clk      clock.Clock
	bufSize  int
	solDebug bool

	// solQueues holds one ordered packet queue per session for the SOL data
	// plane. The Serve loop spawns a goroutine per packet; for SOL that
	// would apply console keystrokes in scheduler order, so SOL packets are
	// funneled through a per-session channel drained by a single worker.
	solMu     sync.Mutex
	solQueues map[uint32]chan solJob
	solDone   chan struct{} // closed by Close; retires all SOL workers

	mu     sync.Mutex
	closed bool
}

// solJob carries one inbound SOL packet through the per-session ordered
// queue.
type solJob struct {
	addr net.Addr
	pkt  []byte // full wire packet; integrity verification needs the trailer
}

// solQueueCapacity bounds queued SOL packets per session; excess packets are
// dropped (UDP semantics — the console retries unacknowledged data).
const solQueueCapacity = 64

// solWorkerIdleTimeout retires a session's SOL worker after this much
// silence; the next packet starts a fresh one.
const solWorkerIdleTimeout = 90 * time.Second

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
	handlers.RegisterAllHandlers(reg)

	s := &Server{
		bmc:       b,
		conn:      conn,
		reg:       reg,
		clk:       b.Clock(),
		bufSize:   defaultBufferSize,
		solQueues: make(map[uint32]chan solJob),
		solDone:   make(chan struct{}),
	}
	for _, o := range opts {
		o(s)
	}
	// Activate Payload reports the port payloads are served on (Table 24-2);
	// learn it from the transport when possible.
	if la, ok := conn.(interface{ LocalAddr() net.Addr }); ok {
		if ua, ok := la.LocalAddr().(*net.UDPAddr); ok && ua.Port != 0 {
			b.SOL.Config().PayloadPort = uint16(ua.Port)
		}
	}
	// SOL data is pushed asynchronously (§15.3); the factory hands each
	// activation a sender that applies the session's encryption and
	// authentication exactly like command responses.
	b.SOL.SetSenderFactory(func(sess *bmc.Session, inst *bmc.SOLInstance) bmc.SOLSendFunc {
		if s.solDebug {
			// Console lifecycle events (reconnect) ride the same trace
			// facility as the packet lines.
			inst.SetTracef(func(format string, args ...any) {
				fmt.Fprintf(os.Stderr, "%s "+format, append([]any{solStamp()}, args...)...)
			})
		}
		return func(pkt *types.SOLPayloadPacket) error {
			if s.solDebug {
				fmt.Fprintf(os.Stderr, "%s sol> sess=%x seq=%d ack=%d accept=%d nack=%v ctrl=%#02x data=%q (async %s)\n",
					solStamp(), sess.BMCID, pkt.SequenceNumber, pkt.AckedSequenceNumber, pkt.AcceptedCharacterCount, pkt.NACK, pkt.ControlByte, pkt.CharacterData, sess.Addr)
			}
			s.respondInSession(sess.Addr, sess, srvPayloadSOL, inst.OutboundEncrypted(), pkt.Pack())
			return nil
		}
	})
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

		// SOL session packets are queued in the read loop itself: dispatch
		// happens from per-packet goroutines, and a keystroke burst handed
		// to the queue in scheduler order would reach the console garbled.
		if sessionID, ok := solSessionPacket(pkt); ok {
			s.enqueueSOL(sessionID, addr, pkt)
			continue
		}
		go s.handlePacket(ctx, addr, pkt)
	}
}

// solSessionPacket reports whether pkt is an in-session RMCP+ SOL payload
// packet, returning its session ID. ParseRMCPPlusHeader bounds-checks the
// header; session-ID 0h packets (pre-session traffic) are not SOL.
func solSessionPacket(pkt []byte) (uint32, bool) {
	sessionID, _, payloadType, _, _, ok := protocol.ParseRMCPPlusHeader(pkt)
	return sessionID, ok && sessionID != 0 && payloadType == protocol.PayloadSOL
}

// Close shuts down the server and its transport.
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	// Retire SOL workers before the transport dies so none sends through a
	// closing conn; their queues go with them.
	s.solMu.Lock()
	close(s.solDone)
	s.solQueues = make(map[uint32]chan solJob)
	s.solMu.Unlock()

	if s.bmc != nil {
		s.bmc.SOL.CloseAll()
	}
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
		// The seq window check and dispatch must be atomic per session:
		// packets of one session arrive in order but are dispatched from
		// per-packet goroutines.
		sess.ProcMu.Lock()
		defer sess.ProcMu.Unlock()
		if !acceptInbound(sess, pkt, addr, inboundSeq, authenticated) {
			return
		}
		s.dispatchIPMISession(ctx, addr, sess, payload, encrypted)

	case srvPayloadSOL:
		// In-session SOL packets never reach here: the Serve loop routes
		// them to the ordered per-session queue (see solSessionPacket).
		// A session-ID-less SOL packet is meaningless; drop it.
		return
	}
}

// enqueueSOL queues an inbound SOL packet for the session's ordered worker,
// starting the worker on first use. Called from the Serve loop so that queue
// order is socket read order. Packets beyond solQueueCapacity are dropped;
// the console retries unacknowledged data (§15.11).
func (s *Server) enqueueSOL(sessionID uint32, addr net.Addr, pkt []byte) {
	s.solMu.Lock()
	defer s.solMu.Unlock()
	select {
	case <-s.solDone:
		return // closing: drop like any packet to an unreachable BMC
	default:
	}
	q, ok := s.solQueues[sessionID]
	if !ok {
		q = make(chan solJob, solQueueCapacity)
		s.solQueues[sessionID] = q
		go s.runSOLWorker(sessionID, q)
	}
	select {
	case q <- solJob{addr: addr, pkt: pkt}:
	default:
		if s.solDebug {
			fmt.Fprintf(os.Stderr, "%s sol! sess=%x queue full, packet dropped\n", solStamp(), sessionID)
		}
	}
}

// runSOLWorker drains one session's SOL queue in arrival order. It retires
// itself after solWorkerIdleTimeout of silence, but only with the queue
// empty and under solMu, so a packet can never land in a worker-less queue.
func (s *Server) runSOLWorker(sessionID uint32, q chan solJob) {
	idle := s.clk.NewTimer(solWorkerIdleTimeout)
	defer idle.Stop()
	for {
		select {
		case <-s.solDone:
			return
		case job := <-q:
			if !idle.Stop() {
				select {
				case <-idle.C():
				default:
				}
			}
			idle.Reset(solWorkerIdleTimeout)
			sess, err := s.bmc.Sessions.Get(sessionID)
			if err != nil {
				continue // session went away; payload auto-deactivates (§24.2)
			}
			s.processSOLJob(sess, job, q)
		case <-idle.C():
			s.solMu.Lock()
			if len(q) == 0 {
				delete(s.solQueues, sessionID)
				s.solMu.Unlock()
				return
			}
			s.solMu.Unlock()
			idle.Reset(solWorkerIdleTimeout)
		}
	}
}

// acceptInbound verifies one in-session packet's integrity and session
// sequence window, recording the sender address on success. sess.ProcMu
// must be held: the window check-then-set is only atomic under it.
func acceptInbound(sess *bmc.Session, pkt []byte, addr net.Addr, inboundSeq uint32, authenticated bool) bool {
	if !verifyRMCPPlusIntegrity(pkt, sess, authenticated) {
		return false
	}
	if !bmc.InboundSeqValid(sess.InboundSeq, inboundSeq) {
		return false
	}
	sess.InboundSeq = inboundSeq
	sess.Addr = addr
	return true
}

// processSOLJob verifies and dispatches one queued SOL packet. A Flush
// Inbound (Table 15-2 bit [1]) additionally drops the packets still queued
// behind it: keystrokes the console asked to discard must not reach the
// system console later.
func (s *Server) processSOLJob(sess *bmc.Session, job solJob, q chan solJob) {
	_, inboundSeq, _, flags, payload, ok := protocol.ParseRMCPPlusHeader(job.pkt)
	if !ok {
		return
	}
	encrypted := flags&protocol.PayloadEncryptedFlag != 0
	authenticated := flags&protocol.PayloadAuthenticatedFlag != 0

	// Unlike IPMI commands, the SOL dispatch runs outside ProcMu: the worker
	// is the only SOL processor per session, and holding it would serialize
	// the data plane against command responses of the same session.
	sess.ProcMu.Lock()
	accepted := acceptInbound(sess, job.pkt, job.addr, inboundSeq, authenticated)
	sess.ProcMu.Unlock()
	if !accepted {
		if s.solDebug {
			fmt.Fprintf(os.Stderr, "%s sol! sess=%x dropped: seq=%d outside window or integrity failed\n",
				solStamp(), sess.BMCID, inboundSeq)
		}
		return
	}

	if s.dispatchSOLSession(context.Background(), job.addr, sess, payload, encrypted) {
		if n := drainSOLQueue(q); n > 0 && s.solDebug {
			fmt.Fprintf(os.Stderr, "%s sol! sess=%x flush inbound: dropped %d queued packet(s)\n",
				solStamp(), sess.BMCID, n)
		}
	}
}

// drainSOLQueue drops all packets currently queued for a session, returning
// how many were dropped.
func drainSOLQueue(q chan solJob) int {
	n := 0
	for {
		select {
		case <-q:
			n++
		default:
			return n
		}
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
	ipmiPayload, ok := decryptSessionPayload(sess, payload, encrypted)
	if !ok {
		return
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

	s.respondInSession(addr, sess, srvPayloadIPMI, encrypted, rawResp)
}

// dispatchSOLSession handles one in-session SOL payload packet (spec v2.0
// §15.9). Inbound packets are processed on their own protection flags — the
// encryption/authentication negotiated at activation (Table 24-2 aux data)
// constrains only BMC→console data, so a console sending cleartext on an
// encrypted activation is served as-is; §13.29 decryption applies whenever
// the packet carries the encrypted flag. Outbound data keeps the negotiated
// outbound protection (§24.4 / Table 24-4). It reports whether the packet
// carried the Flush Inbound operation (Table 15-2 bit [1]).
func (s *Server) dispatchSOLSession(ctx context.Context, addr net.Addr, sess *bmc.Session, payload []byte, encrypted bool) (flushed bool) {
	inst := s.bmc.SOL.InstanceBySession(sess.BMCID)
	if inst == nil {
		return false
	}

	plain, ok := decryptSessionPayload(sess, payload, encrypted)
	if !ok {
		return false
	}
	var in types.SOLPayloadPacket
	if err := in.Unpack(plain); err != nil {
		return false
	}
	flushed = in.ControlByte&0x02 != 0
	if s.solDebug {
		fmt.Fprintf(os.Stderr, "%s sol< sess=%x seq=%d ack=%d accept=%d nack=%v ctrl=%#02x data=%q\n",
			solStamp(), sess.BMCID, in.SequenceNumber, in.AckedSequenceNumber, in.AcceptedCharacterCount, in.NACK, in.ControlByte, in.CharacterData)
	}
	out := s.bmc.SOL.ProcessPacket(ctx, sess.BMCID, &in)
	if out == nil {
		return flushed
	}
	if s.solDebug {
		fmt.Fprintf(os.Stderr, "%s sol> sess=%x seq=%d ack=%d accept=%d nack=%v ctrl=%#02x data=%q\n",
			solStamp(), sess.BMCID, out.SequenceNumber, out.AckedSequenceNumber, out.AcceptedCharacterCount, out.NACK, out.ControlByte, out.CharacterData)
	}
	s.respondInSession(addr, sess, srvPayloadSOL, inst.OutboundEncrypted(), out.Pack())
	return flushed
}

// decryptSessionPayload returns the plaintext of an in-session payload,
// decrypting with K2 when the encrypted flag is set (v2.0 §13.29).
func decryptSessionPayload(sess *bmc.Session, payload []byte, encrypted bool) ([]byte, bool) {
	if !encrypted {
		return payload, true
	}
	if len(sess.K2) < 16 {
		return nil, false
	}
	dec, err := crypto.DecryptAESPayload(payload, sess.K2)
	if err != nil {
		return nil, false
	}
	return dec, true
}

// respondInSession encrypts (when requested), authenticates, and sends one
// in-session payload, advancing the outbound session sequence number shared
// by IPMI commands and SOL packets (§15.5).
func (s *Server) respondInSession(addr net.Addr, sess *bmc.Session, payloadType uint8, encrypt bool, payload []byte) {
	finalPayload := payload
	var flags uint8
	if encrypt && len(sess.K2) >= 16 {
		enc, err := crypto.EncryptAESPayload(payload, sess.K2, crypto.RandomBytes(16))
		if err != nil {
			return
		}
		finalPayload = enc
		flags |= protocol.PayloadEncryptedFlag
	}
	s.sendSessionPayload(addr, sess, payloadType, flags, finalPayload)
}

// sendSessionPayload authenticates (per the session integrity algorithm) and
// sends one in-session payload. Used for command responses and asynchronous
// SOL packets alike.
func (s *Server) sendSessionPayload(addr net.Addr, sess *bmc.Session, payloadType, flags uint8, payload []byte) {
	if sess.IntegrityAlg != types.IntegrityAlg_None {
		flags |= protocol.PayloadAuthenticatedFlag
	}
	pkt := protocol.BuildRMCPPlusPacket(payloadType, flags, sess.ConsoleID, sess.NextOutboundSeq(), payload)
	var ok bool
	pkt, ok = appendRMCPPlusIntegrity(pkt, sess)
	if !ok {
		return
	}
	if s.solDebug && payloadType == srvPayloadSOL {
		fmt.Fprintf(os.Stderr, "%s sol! wire %x\n", solStamp(), pkt)
	}
	_, _ = s.conn.WriteTo(pkt, addr)
}

// sendRMCPPlus sends a session-zero (unauthenticated) RMCP+ packet.
func (s *Server) sendRMCPPlus(addr net.Addr, payloadType, flags uint8, payload []byte) {
	pkt := protocol.BuildRMCPPlusPacket(payloadType, flags, 0, 0, payload)
	_, _ = s.conn.WriteTo(pkt, addr)
}

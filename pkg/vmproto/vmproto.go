// Package vmproto implements a BMC server frontend that speaks QEMU's OpenIPMI
// VM protocol (the ipmi-bmc-extern wire format), a sibling to the RMCP+ UDP
// server in pkg/server. It exists so an emulated machine's in-band system
// interface and an out-of-band LAN client can drive one shared [bmc.BMC], which
// is what the library's simulation / end-to-end use targets: a guest kernel and
// a userland agent talk to the BMC over the VM protocol while tooling talks to
// the same BMC over LAN, and both observe the same state.
//
// Unlike the RMCP+ frontend, the VM protocol is a byte stream, not datagrams,
// so it is built on a [net.Listener] rather than a transport.PacketConn. The
// system interface is unauthenticated by design (there is no session), so
// messages dispatch with a session-less [handlers.HandlerContext] whose channel
// is the system interface, which the handler privilege check treats as locally
// authorized.
//
// [Client] is the console side of the same protocol, standing in for QEMU or
// the in-guest OpenIPMI driver in simulation and end-to-end testing.
package vmproto

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/handlers"
)

// systemInterfaceChannel is the channel number of the in-band system interface
// (spec §6.13.1). The VM protocol is that interface, so requests dispatched
// through it are attributed to this channel.
const systemInterfaceChannel uint8 = 0x0F

// OpenIPMI VM protocol framing (QEMU ipmi-bmc-extern wire format).
const (
	vmMsgChar    = 0xA0 // ends an IPMI message
	vmCmdChar    = 0xA1 // ends a control command
	vmEscapeChar = 0xAA // escape: the next byte has bit 4 set and must be cleared

	// vmMaxMsgSize caps an accumulated message so a peer cannot grow the buffer
	// without bound; an over-long frame is dropped rather than truncated.
	vmMaxMsgSize = 4096

	vmReadBufferSize = 4096
)

// VMServer serves QEMU's OpenIPMI VM protocol on a stream listener, dispatching
// each in-band IPMI message through the same handler registry as the RMCP+
// [Server] against the same [bmc.BMC].
//
// Construct one with [NewVMServer] sharing the [bmc.BMC] you gave [NewServer],
// then call [VMServer.Serve]. It is a distinct frontend from [Server]: the VM
// protocol is a byte stream, so it takes a [net.Listener], and the system
// interface is unauthenticated, so messages carry no session.
type VMServer struct {
	bmc *bmc.BMC
	reg *handlers.Registry
}

// VMServerOption configures a [VMServer].
type VMServerOption func(*VMServer)

// WithVMHandlerRegistry replaces the default handler registry. Pass the same
// registry the RMCP+ [Server] uses to guarantee the two frontends dispatch an
// identical command set. All registration must be complete before
// [VMServer.Serve] is called; the registry is read-only during dispatch.
func WithVMHandlerRegistry(r *handlers.Registry) VMServerOption {
	return func(s *VMServer) { s.reg = r }
}

// NewVMServer creates a VM protocol frontend over the BMC state b.
//
// Share b with the [Server] created by [NewServer] so in-band (VM protocol) and
// out-of-band (LAN) requests operate on the same BMC. By default it builds its
// own copy of the standard registry, identical to the RMCP+ server's; override
// it with [WithVMHandlerRegistry], for example to share a single registry
// instance or to add OEM commands.
func NewVMServer(b *bmc.BMC, opts ...VMServerOption) *VMServer {
	reg := handlers.NewRegistry()
	handlers.RegisterAllHandlers(reg)
	s := &VMServer{
		bmc: b,
		reg: reg,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Serve accepts connections on ln and serves each until ctx is canceled or ln
// is closed. Each accepted connection is one QEMU process; every guest power
// cycle is a fresh QEMU process reconnecting to the same listener, so the loop
// accepts forever. Connections are served one at a time, matching QEMU's single
// ipmi-bmc-extern link.
func (s *VMServer) Serve(ctx context.Context, ln net.Listener) error {
	// Close the listener to unblock Accept when ctx is canceled. Deriving a
	// child context canceled on return guarantees this goroutine exits even when
	// Serve returns for another reason (e.g. the caller closed the listener), so
	// it cannot leak past the server's lifetime.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept VM connection: %w", err)
		}

		s.serveConn(ctx, conn)
	}
}

// serveConn runs the framing decoder for one connection, dispatching each
// message synchronously. It returns when the peer disconnects or ctx is
// canceled. The VM protocol (KCS/ipmi-bmc-extern) carries one outstanding
// transaction at a time and in-band handlers never block on slow hardware, so
// there is no need to dispatch off the read loop.
func (s *VMServer) serveConn(ctx context.Context, conn net.Conn) {
	// Close the connection when the parent context is canceled so a blocked Read
	// unwinds on shutdown. The goroutine is bound to this connection's context
	// and exits on return, so it does not leak per past connection.
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-connCtx.Done()
		_ = conn.Close()
	}()

	var (
		acc      []byte
		inEscape bool
		overflow bool
		buf      = make([]byte, vmReadBufferSize)
	)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		for _, ch := range buf[:n] {
			switch ch {
			case vmMsgChar:
				if !inEscape && !overflow && len(acc) > 0 {
					// The VM protocol is the in-band system interface (channel
					// 0x0F). Attribute the request to it so "this channel"
					// (0x0E) resolves to the system interface rather than
					// falling back to LAN, resolving per message so a channel
					// table reconfigured mid-connection is observed. A missing
					// channel leaves the context channel nil and downstream
					// handlers fall back gracefully.
					sysCh, _ := s.bmc.Channels.Get(systemInterfaceChannel)
					s.handleMessage(connCtx, conn, sysCh, acc)
				}
				acc, inEscape, overflow = acc[:0], false, false
			case vmCmdChar:
				// Control command (QEMU's version and capability announcements
				// on connect). These carry no BMC state, and the BMC never
				// sends hardware control commands back: power authority stays
				// with the hardware layer, so the frame is consumed and
				// ignored.
				acc, inEscape, overflow = acc[:0], false, false
			case vmEscapeChar:
				inEscape = true
			default:
				if inEscape {
					ch &^= 0x10
					inEscape = false
				}
				if len(acc) >= vmMaxMsgSize {
					overflow = true
				}
				if !overflow {
					acc = append(acc, ch)
				}
			}
		}
	}
}

// handleMessage decodes one IPMI request (msgID, netfn<<2|lun, cmd, data...,
// IPMB checksum), dispatches it with an unauthenticated system-interface
// context attributed to ch, and writes the response framed the same way with a
// completion code after the command byte. Malformed frames are dropped,
// matching the RMCP+ frontend's treatment of unparseable packets.
//
// It runs synchronously on the read loop, so msg (the decoder's accumulation
// buffer) is stable for the whole call and handlers must not retain the request
// slice past return, per the [handlers.Handler] contract; there is no need to
// copy it. Being synchronous also means only one write happens at a time per
// connection, so no write lock is needed.
func (s *VMServer) handleMessage(ctx context.Context, conn net.Conn, ch *bmc.Channel, msg []byte) {
	// msgID + netfn/lun + cmd + checksum is the shortest valid message.
	if len(msg) < 4 || vmChecksum(msg) != 0 {
		return
	}

	msgID := msg[0]
	netFn := msg[1] >> 2
	lun := msg[1] & 0x03
	cmd := msg[2]
	data := msg[3 : len(msg)-1]

	hctx := &handlers.HandlerContext{BMC: s.bmc, Channel: ch}
	respData, cc, _ := s.reg.Dispatch(ctx, hctx, netFn, cmd, data)

	resp := make([]byte, 0, 4+len(respData)+1)
	resp = append(resp, msgID, (netFn|1)<<2|lun, cmd, uint8(cc))
	resp = append(resp, respData...)
	resp = append(resp, -vmChecksum(resp))

	_, _ = conn.Write(vmEncode(resp))
}

// vmChecksum is the IPMB two's-complement checksum: a well-formed message,
// trailing checksum included, sums to zero.
func vmChecksum(bs []byte) byte {
	var sum byte
	for _, b := range bs {
		sum += b
	}
	return sum
}

// vmEncode escapes the framing bytes in bs and appends the end-of-message
// marker, producing the on-wire form of one IPMI message.
func vmEncode(bs []byte) []byte {
	out := make([]byte, 0, len(bs)+8)
	for _, b := range bs {
		switch b {
		case vmMsgChar, vmCmdChar, vmEscapeChar:
			out = append(out, vmEscapeChar, b|0x10)
		default:
			out = append(out, b)
		}
	}
	return append(out, vmMsgChar)
}

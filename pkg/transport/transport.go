// Package transport defines the network transport interfaces used by the IPMI server.
//
// The interfaces are intentionally minimal so that any packet-oriented network
// stack – standard Linux UDP, TinyGo, bare-metal Ethernet, Unix sockets, or a
// wrapper around an existing listener – can be used without modification to the
// core server logic.
package transport

import "net"

// PacketConn is a generic, address-based packet transport.
//
// It mirrors the subset of [net.PacketConn] that the server actually needs, so
// callers can wrap an existing [net.PacketConn] with zero overhead, or provide
// a completely custom implementation for embedded targets.
type PacketConn interface {
	// ReadFrom reads a packet into buf and returns the sender's address.
	ReadFrom(buf []byte) (n int, addr net.Addr, err error)

	// WriteTo sends data to addr.
	WriteTo(data []byte, addr net.Addr) (n int, err error)

	// Close releases any resources held by the connection.
	Close() error
}

package server

import (
	"errors"
	"net"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/types"
)

type capturePacketConn struct {
	writes [][]byte
}

func (c *capturePacketConn) ReadFrom(_ []byte) (int, net.Addr, error) {
	return 0, nil, errors.New("not implemented")
}

func (c *capturePacketConn) WriteTo(data []byte, _ net.Addr) (int, error) {
	c.writes = append(c.writes, append([]byte(nil), data...))
	return len(data), nil
}

func (c *capturePacketConn) Close() error {
	return nil
}

type testAddr string

func (a testAddr) Network() string { return "test" }
func (a testAddr) String() string  { return string(a) }

func TestSendRMCPPlusSessionAddsIntegrityTrailer(t *testing.T) {
	conn := &capturePacketConn{}
	srv := &Server{conn: conn}
	sess := &bmc.Session{
		ConsoleID:    0x11223344,
		OutboundSeq:  7,
		IntegrityAlg: types.IntegrityAlg_HMAC_SHA1_96,
		K1:           []byte("0123456789abcdefghij"),
	}
	payload := []byte{0x20, 0x18, 0xc8, 0x81, 0x00}

	srv.sendRMCPPlusSession(testAddr("console"), srvPayloadIPMI, 0, sess, payload)

	if len(conn.writes) != 1 {
		t.Fatalf("want one packet, got %d", len(conn.writes))
	}
	pkt := conn.writes[0]
	if pkt[5]&protocol.PayloadAuthenticatedFlag == 0 {
		t.Fatalf("authenticated bit was not set: payload type byte=0x%02x", pkt[5])
	}
	if !verifyRMCPPlusIntegrity(pkt, sess, true) {
		t.Fatalf("generated integrity trailer did not verify")
	}

	tampered := append([]byte(nil), pkt...)
	tampered[rmcpPlusPayloadOffset] ^= 0xff
	if verifyRMCPPlusIntegrity(tampered, sess, true) {
		t.Fatalf("tampered packet passed integrity verification")
	}
}

func TestVerifyRMCPPlusIntegrityRequiresAuthenticatedFlag(t *testing.T) {
	sess := &bmc.Session{
		IntegrityAlg: types.IntegrityAlg_HMAC_SHA1_96,
		K1:           []byte("0123456789abcdefghij"),
	}
	pkt := protocol.BuildRMCPPlusPacket(srvPayloadIPMI, protocol.PayloadAuthenticatedFlag, 1, 1, []byte{0x01, 0x02})
	pkt, ok := appendRMCPPlusIntegrity(pkt, sess)
	if !ok {
		t.Fatalf("appendRMCPPlusIntegrity failed")
	}

	if verifyRMCPPlusIntegrity(pkt, sess, false) {
		t.Fatalf("packet without authenticated flag should not verify")
	}
}

func TestRMCPPlusIntegrity_SHA256_128(t *testing.T) {
	sess := &bmc.Session{
		ConsoleID:    0x11223344,
		OutboundSeq:  3,
		IntegrityAlg: types.IntegrityAlg_HMAC_SHA256_128,
		// SHA256-128 uses a 128-bit (16-byte) K1; any 16+ byte key works.
		K1: []byte("0123456789abcdef"),
	}
	payload := []byte{0x20, 0x18, 0xc8, 0x81, 0x00}

	pkt, ok := appendRMCPPlusIntegrity(protocol.BuildRMCPPlusPacket(
		srvPayloadIPMI, protocol.PayloadAuthenticatedFlag, sess.ConsoleID, sess.OutboundSeq, payload), sess)
	if !ok {
		t.Fatalf("appendRMCPPlusIntegrity failed for SHA256-128")
	}

	if !verifyRMCPPlusIntegrity(pkt, sess, true) {
		t.Fatalf("SHA256-128 integrity trailer did not verify")
	}

	// The auth code length must be 16 bytes (128 bits), not 12 (SHA1-96).
	authCodeLen, ok := rmcpPlusIntegrityAuthCodeLen(sess.IntegrityAlg)
	if !ok || authCodeLen != 16 {
		t.Fatalf("want authCodeLen 16, got %d (ok=%v)", authCodeLen, ok)
	}
	padLen := rmcpPlusIntegrityPadLen(rmcpPlusHeaderSize, len(payload))
	authCodeStart := rmcpPlusPayloadOffset + len(payload) + padLen + 2
	if got := len(pkt) - authCodeStart; got != 16 {
		t.Fatalf("trailer auth code length: want 16, got %d", got)
	}

	tampered := append([]byte(nil), pkt...)
	tampered[rmcpPlusPayloadOffset] ^= 0xff
	if verifyRMCPPlusIntegrity(tampered, sess, true) {
		t.Fatalf("tampered SHA256-128 packet passed integrity verification")
	}
}

package server

// IPMI v1.5 per-session data-race regression test driven over raw UDP.

import (
	"encoding/binary"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/types"
)

// raceBuildIPMIRequest builds a raw IPMI LAN request message (Table 13-8).
func raceBuildIPMIRequest(netFn, cmd, seq uint8, data []byte) []byte {
	msg := make([]byte, 7+len(data))
	msg[0] = protocol.BMCAddr
	msg[1] = netFn << 2
	msg[2] = protocol.Checksum(msg[0:2])
	msg[3] = protocol.RemoteConsoleAddr
	msg[4] = seq << 2
	msg[5] = cmd
	copy(msg[6:], data)
	msg[6+len(data)] = protocol.Checksum(msg[3 : 6+len(data)])
	return msg
}

func raceWrapV15(authType types.AuthType, seq, sessionID uint32, authCode, payload []byte) []byte {
	hdr := types.SessionHeader15{
		AuthType:      authType,
		Sequence:      seq,
		SessionID:     sessionID,
		AuthCode:      authCode,
		PayloadLength: uint8(len(payload)),
	}
	out := []byte{0x06, 0x00, 0xff, 0x07} // RMCP header, class IPMI
	out = append(out, hdr.Pack()...)
	out = append(out, payload...)
	return out
}

// TestV15SessionFieldRace establishes a real IPMI v1.5 (-A MD5) session, then
// fires several authenticated command packets for the SAME session
// concurrently. Before the per-session lock, the server mutated the shared
// v1.5 session sequence fields lock-free once the pointer was fetched.
func TestV15SessionFieldRace(t *testing.T) {
	const netFnApp, cmdGetDeviceID, cmdGetSessionChallenge, cmdActivateSession = 0x06, 0x01, 0x39, 0x3a

	b := raceNewBMC(t)

	// The seeded admin password from raceNewBMC, zero-padded to 16 bytes per the
	// v1.5 AuthCode algorithms.
	var padded [16]byte
	copy(padded[:], racePass)
	pass16 := padded[:]

	port, _, stop := raceStartServer(t, b)
	defer stop()

	srvAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port}
	cl, err := net.DialUDP("udp", nil, srvAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	roundtrip := func(pkt []byte) []byte {
		if _, err := cl.Write(pkt); err != nil {
			t.Fatal(err)
		}
		_ = cl.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 4096)
		n, err := cl.Read(buf)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		return buf[:n]
	}

	// ---- Get Session Challenge (AuthType NONE, session 0, seq 0) ----
	var uname [16]byte
	copy(uname[:], raceUser)
	gscData := append([]byte{byte(bmc.V15AuthTypeMD5)}, uname[:]...)
	gscReq := raceBuildIPMIRequest(netFnApp, cmdGetSessionChallenge, 1, gscData)
	resp := roundtrip(raceWrapV15(types.AuthTypeNone, 0, 0, nil, gscReq))

	var s15 types.Session15
	if err := s15.Unpack(resp[4:]); err != nil {
		t.Fatalf("unpack GSC resp: %v", err)
	}
	_, _, gscBody, _, ok := protocol.ParseIPMIRequest(s15.Payload)
	if !ok || len(gscBody) < 21 || gscBody[0] != 0 {
		t.Fatalf("bad GSC body len=%d ok=%v cc=%#x", len(gscBody), ok, gscBody[0])
	}
	tempSessionID := binary.LittleEndian.Uint32(gscBody[1:5])
	var challenge [16]byte
	copy(challenge[:], gscBody[5:21])

	// ---- Activate Session (AuthType MD5, temp session id, seq 0) ----
	actData := make([]byte, 22)
	actData[0] = byte(bmc.V15AuthTypeMD5)
	actData[1] = byte(bmc.PrivilegeLevelAdministrator)
	copy(actData[2:18], challenge[:])
	binary.LittleEndian.PutUint32(actData[18:22], 1)
	actReq := raceBuildIPMIRequest(netFnApp, cmdActivateSession, 2, actData)
	actAuth := handlers.GenV15AuthCode(pass16, bmc.V15AuthTypeMD5, tempSessionID, actReq, 0)
	resp = roundtrip(raceWrapV15(types.AuthTypeMD5, 0, tempSessionID, actAuth, actReq))

	if err := s15.Unpack(resp[4:]); err != nil {
		t.Fatalf("unpack Activate resp: %v", err)
	}
	_, _, actBody, _, ok := protocol.ParseIPMIRequest(s15.Payload)
	if !ok || len(actBody) < 11 || actBody[0] != 0 {
		t.Fatalf("bad Activate body len=%d ok=%v cc=%#x", len(actBody), ok, actBody[0])
	}
	permSessionID := binary.LittleEndian.Uint32(actBody[2:6])
	inboundSeq := binary.LittleEndian.Uint32(actBody[6:10])

	// ---- Fire N authenticated GetDeviceID packets on the SAME session ----
	const n = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range n {
		seq := inboundSeq + uint32(i)
		wg.Add(1)
		go func(seq uint32) {
			defer wg.Done()
			req := raceBuildIPMIRequest(netFnApp, cmdGetDeviceID, uint8(seq&0x3f), nil)
			auth := handlers.GenV15AuthCode(pass16, bmc.V15AuthTypeMD5, permSessionID, req, seq)
			pkt := raceWrapV15(types.AuthTypeMD5, seq, permSessionID, auth, req)
			<-start
			s, err := net.DialUDP("udp", nil, srvAddr)
			if err != nil {
				return
			}
			defer s.Close()
			_, _ = s.Write(pkt)
			_ = s.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			_, _ = s.Read(make([]byte, 4096))
		}(seq)
	}
	close(start)
	wg.Wait()
}

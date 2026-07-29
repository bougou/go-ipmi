package crypto

import (
	"bytes"
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestAuthCodeMultiSession_MD5(t *testing.T) {
	password := []byte("ADMIN")
	sessionID := uint32(0xAABBCCDD)
	sessionSeq := uint32(0)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x3a}

	viaFunc := AuthCodeMultiSession(types.AuthTypeMD5, password, sessionID, sessionSeq, ipmiData)
	viaStruct := (&MultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: sessionSeq,
		IPMIData:   ipmiData,
	}).AuthCode(types.AuthTypeMD5)

	if !bytes.Equal(viaFunc, viaStruct) {
		t.Fatalf("MD5 mismatch:\n func=%x\n struct=%x", viaFunc, viaStruct)
	}
	if len(viaFunc) != 16 {
		t.Fatalf("want 16-byte AuthCode, got %d", len(viaFunc))
	}
	if !VerifyMultiSessionAuthCode(types.AuthTypeMD5, password, sessionID, sessionSeq, ipmiData, viaFunc) {
		t.Fatal("VerifyMultiSessionAuthCode rejected matching code")
	}
}

func TestAuthCodeMultiSession_MD2(t *testing.T) {
	password := []byte("secret")
	sessionID := uint32(0x11223344)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x38}

	viaFunc := AuthCodeMultiSession(types.AuthTypeMD2, password, sessionID, 1, ipmiData)
	viaStruct := (&MultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: 1,
		IPMIData:   ipmiData,
	}).AuthCode(types.AuthTypeMD2)

	if !bytes.Equal(viaFunc, viaStruct) {
		t.Fatalf("MD2 mismatch:\n func=%x\n struct=%x", viaFunc, viaStruct)
	}
}

func TestAuthCodeMultiSession_Password(t *testing.T) {
	password := []byte("straight-pass")
	sessionID := uint32(0x55667788)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x01}

	viaFunc := AuthCodeMultiSession(types.AuthTypePassword, password, sessionID, 2, ipmiData)
	viaStruct := (&MultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: 2,
		IPMIData:   ipmiData,
	}).AuthCode(types.AuthTypePassword)

	if !bytes.Equal(viaFunc, viaStruct) {
		t.Fatalf("password mismatch:\n func=%x\n struct=%x", viaFunc, viaStruct)
	}
	// Straight password AuthCode is the 16-byte padded password.
	var padded [16]byte
	copy(padded[:], password)
	if !bytes.Equal(viaFunc, padded[:]) {
		t.Fatalf("password AuthCode want padded password %x, got %x", padded[:], viaFunc)
	}
}

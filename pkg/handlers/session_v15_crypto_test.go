package handlers

import (
	"bytes"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/client"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

func TestGenV15AuthCodeMD2MatchesClient(t *testing.T) {
	password := []byte("secret")
	sessionID := uint32(0x11223344)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x38}

	serverCode := GenV15AuthCode(password, bmc.V15AuthTypeMD2, sessionID, ipmiData, 1)
	clientCode := (&client.AuthCodeMultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: 1,
		IPMIData:   ipmiData,
	}).AuthCode(ipmi.AuthTypeMD2)

	if !bytes.Equal(serverCode, clientCode) {
		t.Fatalf("MD2 auth code mismatch:\n server=%x\n client=%x", serverCode, clientCode)
	}
}

func TestGenV15AuthCodePasswordMatchesClient(t *testing.T) {
	password := []byte("straight-pass")
	sessionID := uint32(0x55667788)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x01}

	serverCode := GenV15AuthCode(password, bmc.V15AuthTypePassword, sessionID, ipmiData, 2)
	clientCode := (&client.AuthCodeMultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: 2,
		IPMIData:   ipmiData,
	}).AuthCode(ipmi.AuthTypePassword)

	if !bytes.Equal(serverCode, clientCode) {
		t.Fatalf("password auth code mismatch:\n server=%x\n client=%x", serverCode, clientCode)
	}
}

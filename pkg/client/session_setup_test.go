package client

import (
	"bytes"
	"errors"
	"net"
	"testing"

	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestOpenSessionAcceptsOnlyRequestedAlgorithms(t *testing.T) {
	for _, tc := range []struct {
		name                string
		requested, returned types.CipherSuiteID
	}{
		{"explicit suite zero", 0, 0},
		{"matching authenticated suite", 3, 3},
		{"unrequested suite zero", 3, 0},
		{"authentication differs", 1, 6},
		{"integrity differs", 8, 12},
		{"confidentiality differs", 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, "user", "password")
			s := &c.session.v20
			s.cipherSuiteID = tc.requested
			auth, integ, crypt, _ := types.GetCipherSuiteAlgorithms(tc.returned)
			res := rmcpplus.OpenSessionResponse{ManagedSystemSessionID: 42}
			res.AuthenticationPayload, res.IntegrityPayload, res.ConfidentialityPayload = rmcpplus.NewAlgorithmPayloads(auth, integ, crypt)
			conn, peer := net.Pipe()
			t.Cleanup(func() {
				if err := errors.Join(conn.Close(), peer.Close()); err != nil {
					t.Errorf("close test pipe: %v", err)
				}
			})
			c.udpClient.conn = conn
			go func() {
				if _, err := peer.Read(make([]byte, c.bufferSize)); err == nil {
					packet := protocol.BuildRMCPPlusPacket(uint8(types.PayloadTypeRmcpOpenSessionResponse), 0, 0, 0, res.Pack())
					_, _ = peer.Write(packet)
				}
			}()
			response, err := c.OpenSession(t.Context())
			if response == nil || response.ManagedSystemSessionID != res.ManagedSystemSessionID {
				t.Fatalf("OpenSession did not decode the test response: %v", err)
			}
			if tc.requested == tc.returned {
				if err != nil || s.state != types.SessionStateOpenSessionReceived || s.bmcSessionID != res.ManagedSystemSessionID {
					t.Fatalf("matching response: state %v, session ID %d, error %v", s.state, s.bmcSessionID, err)
				}
			} else if err == nil || s.state != types.SessionStateOpenSessionSent || s.bmcSessionID != 0 {
				t.Fatalf("mismatched response: state %v, session ID %d, error %v", s.state, s.bmcSessionID, err)
			}
		})
	}
}

func TestRAKPNoneRejectsUnexpectedAuthCode(t *testing.T) {
	if rakpAuthCodeMatches(types.AuthAlg_None, nil, []byte{1}) {
		t.Fatal("accepted nonempty authentication code")
	}
}

func TestSessionKeysWithoutSIK(t *testing.T) {
	c := newTestClient(t, "user", "password")
	for _, tc := range []struct {
		name    string
		keyByte byte
		derive  func() ([]byte, error)
	}{
		{"K1", 0x01, c.generate_k1},
		{"K2", 0x02, c.generate_k2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c.session.v20.authAlg = types.AuthAlg_None
			want := bytes.Repeat([]byte{tc.keyByte}, 20)
			if key, err := tc.derive(); err != nil || !bytes.Equal(key, want) {
				t.Errorf("RAKP-none: got key %x, error %v; want %x, nil", key, err, want)
			}
			c.session.v20.authAlg = types.AuthAlg_HMAC_SHA1
			if _, err := tc.derive(); err == nil {
				t.Error("authenticated session accepted missing SIK")
			}
		})
	}
}

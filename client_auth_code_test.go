package ipmi

import "testing"

func Test_ActivateSessionAuthCode(t *testing.T) {
	tests := []struct {
		name           string
		commandRequest Request
		ipmiRequest    *IPMIRequest
		input          *AuthCodeMultiSessionInput
		expect         []byte
	}{
		{
			name: "test1",
			commandRequest: &ActivateSessionRequest{
				AuthTypeForSession:            0x02,
				MaxPrivilegeLevel:             0x04,
				Challenge:                     [16]byte{0x82, 0x8f, 0xa9, 0xbf, 0x25, 0x51, 0x6b, 0x2a, 0xf5, 0xf8, 0xfb, 0x3f, 0x37, 0xae, 0x6e, 0x69},
				InitialOutboundSequenceNumber: 0xa2605e12,
			},
			ipmiRequest: &IPMIRequest{
				ResponderAddr:     0x20,
				ResponderLUN:      0x0,
				NetFn:             NetFnAppRequest,
				RequesterAddr:     0x81,
				RequesterLUN:      0x0,
				RequesterSequence: 0x03,
				Command:           0x3a,
			},
			input: &AuthCodeMultiSessionInput{
				// cSpell:disable
				Password: "vtA9kBPODBPBy",
				// cSpell:enable
				SessionID:  0xb215d500,
				SessionSeq: 0x00000000,
			},
			expect: []byte{0xec, 0xb1, 0x65, 0xeb, 0xdc, 0xf7, 0x9f, 0xd9, 0x96, 0xa3, 0xfa, 0x6b, 0xae, 0x18, 0x69, 0x54},
		},
		{
			name: "test2",
			commandRequest: &SetSessionPrivilegeLevelRequest{
				PrivilegeLevel: PrivilegeLevelAdministrator,
			},
			ipmiRequest: &IPMIRequest{
				ResponderAddr:     0x20,
				ResponderLUN:      0x0,
				NetFn:             CommandSetSessionPrivilegeLevel.NetFn,
				RequesterAddr:     0x81,
				RequesterLUN:      0x0,
				RequesterSequence: 0x04,
				Command:           CommandSetSessionPrivilegeLevel.ID,
			},
			input: &AuthCodeMultiSessionInput{
				// cSpell:disable
				Password: "vtA9kBPODBPBy",
				// cSpell:enable
				SessionID:  0xa26f8e00,
				SessionSeq: 0xdabbb496,
			},
			expect: []byte{0x69, 0xe8, 0x3e, 0x2b, 0x99, 0xe3, 0xf6, 0xa9, 0x3d, 0x1c, 0xf0, 0x47, 0x8b, 0x0e, 0xfe, 0xba},
		},
	}

	for _, tt := range tests {

		commandData := tt.commandRequest.Pack()

		tt.ipmiRequest.CommandData = commandData
		tt.ipmiRequest.ComputeChecksum()
		ipmiData := tt.ipmiRequest.Pack()

		tt.input.IPMIData = ipmiData

		got := tt.input.AuthCode(AuthTypeMD5)
		expected := tt.expect

		if !isByteSliceEqual(got, expected) {
			t.Errorf("test %s failed, not equal, got: %v, expected: %v", tt.name, got, expected)
		}
	}
}

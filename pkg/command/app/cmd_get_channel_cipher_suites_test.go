package app

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

// TestParseCipherSuitesData_VariableLengthRecords exercises tag-delimited
// cipher suite records per IPMI 2.0 §22.15.1, where integrity/confidentiality
// entries are omitted when the suite does not use them. This previously broke
// ParseCipherSuitesData, which assumed a fixed id+3-algs (4-byte) payload after
// each 0xC0 start byte and rejected short records.
func TestParseCipherSuitesData_VariableLengthRecords(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want []types.CipherSuiteRecord
	}{
		{
			// Suite 1 alone: auth only (RAKP-HMAC-SHA1), no integ, no crypt.
			name: "suite1_only",
			data: []byte{0xC0, 0x01, 0x01},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID1, AuthAlg: 0x01},
			},
		},
		{
			// Suite 15 alone: auth only (RAKP-HMAC-SHA256), no integ, no crypt.
			name: "suite15_only",
			data: []byte{0xC0, 0x0F, 0x03},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID15, AuthAlg: 0x03},
			},
		},
		{
			// Full 5-byte record (suite 3): auth + integ + crypt.
			name: "suite3_only",
			data: []byte{0xC0, 0x03, 0x01, 0x41, 0x81},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID3, AuthAlg: 0x01, IntegrityAlgs: []uint8{0x01}, CryptAlgs: []uint8{0x01}},
			},
		},
		{
			// Short record LAST: previously rejected by the fixed 4-byte
			// bounds check ("incomplete cipher suite data").
			name: "suite3_then_suite1_short_last",
			data: []byte{
				0xC0, 0x03, 0x01, 0x41, 0x81, // suite 3
				0xC0, 0x01, 0x01, // suite 1 (short, last)
			},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID3, AuthAlg: 0x01, IntegrityAlgs: []uint8{0x01}, CryptAlgs: []uint8{0x01}},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID1, AuthAlg: 0x01},
			},
		},
		{
			// Short record FIRST, then full record.
			name: "suite1_short_first_then_suite3",
			data: []byte{
				0xC0, 0x01, 0x01, // suite 1 (short, first)
				0xC0, 0x03, 0x01, 0x41, 0x81, // suite 3
			},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID1, AuthAlg: 0x01},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID3, AuthAlg: 0x01, IntegrityAlgs: []uint8{0x01}, CryptAlgs: []uint8{0x01}},
			},
		},
		{
			// Mixed omit-None set as advertised by the extended e2e server.
			name: "mixed_1_2_15_16_3_17",
			data: []byte{
				0xC0, 0x01, 0x01, // 1
				0xC0, 0x02, 0x01, 0x41, // 2
				0xC0, 0x0F, 0x03, // 15
				0xC0, 0x10, 0x03, 0x44, // 16
				0xC0, 0x03, 0x01, 0x41, 0x81, // 3
				0xC0, 0x11, 0x03, 0x44, 0x81, // 17
			},
			want: []types.CipherSuiteRecord{
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID1, AuthAlg: 0x01},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID2, AuthAlg: 0x01, IntegrityAlgs: []uint8{0x01}},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID15, AuthAlg: 0x03},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID16, AuthAlg: 0x03, IntegrityAlgs: []uint8{0x04}},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID3, AuthAlg: 0x01, IntegrityAlgs: []uint8{0x01}, CryptAlgs: []uint8{0x01}},
				{StartOfRecord: 0xC0, CipherSuitID: types.CipherSuiteID17, AuthAlg: 0x03, IntegrityAlgs: []uint8{0x04}, CryptAlgs: []uint8{0x01}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCipherSuitesData(tc.data)
			if err != nil {
				t.Fatalf("ParseCipherSuitesData returned unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("record count: got %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if !cipherSuiteRecordEqual(got[i], tc.want[i]) {
					t.Errorf("record %d: got %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func cipherSuiteRecordEqual(a, b types.CipherSuiteRecord) bool {
	if a.StartOfRecord != b.StartOfRecord || a.CipherSuitID != b.CipherSuitID {
		return false
	}
	if a.AuthAlg != b.AuthAlg {
		return false
	}
	if !uint8SliceEqual(a.IntegrityAlgs, b.IntegrityAlgs) {
		return false
	}
	if !uint8SliceEqual(a.CryptAlgs, b.CryptAlgs) {
		return false
	}
	return true
}

func uint8SliceEqual(a, b []uint8) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

package client

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestSOLOutputStreamSuppressesRetransmittedData(t *testing.T) {
	var out bytes.Buffer
	stream := solStream{output: &out}

	for _, response := range []*types.SOLPayloadResponse{
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("abc")}},
		{},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("abc")}},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("abcde")}},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("f")}},
	} {
		if err := stream.processOutput(response); err != nil {
			t.Fatalf("process() error = %v", err)
		}
	}

	if got := out.String(); got != "abcdef" {
		t.Fatalf("SOL output = %q, want %q", got, "abcdef")
	}
}

func TestSOLOutputStreamRejectsInvalidResponse(t *testing.T) {
	for _, tt := range []struct {
		name     string
		response *types.SOLPayloadResponse
		want     string
	}{
		{
			name:     "deactivating",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{ControlByte: solStatusDeactivating}},
			want:     "deactivation",
		},
		{
			name:     "transmit overrun",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{ControlByte: solStatusTransmitOverrun}},
			want:     "output was lost",
		},
		{
			name:     "data with ACK-only sequence",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{CharacterData: []byte("x")}},
			want:     "ACK-only sequence 0",
		},
		{
			name: "shortened retransmission",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{
				SequenceNumber: 1,
				CharacterData:  []byte("ab"),
			}},
			want: "shortened retried",
		},
		{
			name: "sequence gap",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{
				SequenceNumber: 3,
				CharacterData:  []byte("gap"),
			}},
			want: "unexpected BMC SOL sequence",
		},
		{
			name: "oversized data",
			response: &types.SOLPayloadResponse{SOLPayloadPacket: types.SOLPayloadPacket{
				SequenceNumber: 2,
				CharacterData:  bytes.Repeat([]byte("x"), 256),
			}},
			want: "limited to 255",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stream := solStream{output: io.Discard, remoteSequenceNumber: 1, acceptedCharCount: 3}
			err := stream.processOutput(tt.response)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("process() error = %v, want %q", err, tt.want)
			}
		})
	}
}

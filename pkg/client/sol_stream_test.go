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

func TestSOLOutputStreamDropsTrailingRetransmission(t *testing.T) {
	var out bytes.Buffer
	stream := solStream{output: &out}

	for _, response := range []*types.SOLPayloadResponse{
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 14, CharacterData: []byte("abc")}},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("def")}},
		// Retransmissions the BMC sent before our acknowledgement arrived, delayed
		// past newer output. Redelivering them would repeat character data, and
		// treating them as a gap would tear down the stream.
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 14, CharacterData: []byte("abc")}},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 1, CharacterData: []byte("ghi")}},
		{SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("def")}},
	} {
		if err := stream.processOutput(response); err != nil {
			t.Fatalf("processOutput() error = %v", err)
		}
	}

	if got := out.String(); got != "abcdefghi" {
		t.Fatalf("SOL output = %q, want %q", got, "abcdefghi")
	}
	if stream.remoteSequenceNumber != 1 || stream.acceptedCharCount != 3 {
		t.Fatalf("stream acked sequence %d with %d chars, want sequence 1 with 3 chars",
			stream.remoteSequenceNumber, stream.acceptedCharCount)
	}

	// A dropped duplicate carries no news about what the BMC accepted, so it must
	// not schedule an acknowledgement of its own.
	stream.outputAckPending = false
	if err := stream.processOutput(&types.SOLPayloadResponse{
		SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("def")},
	}); err != nil {
		t.Fatalf("processOutput() error = %v", err)
	}
	if stream.outputAckPending {
		t.Fatal("dropped duplicate left an acknowledgement pending")
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

func TestSOLStreamPreservesAsynchronousOutput(t *testing.T) {
	client := &Client{}
	stream := &solStream{client: client}
	if err := client.registerSOLStream(stream); err != nil {
		t.Fatalf("registerSOLStream() error = %v", err)
	}
	defer client.unregisterSOLStream()

	payload := (&types.SOLPayloadPacket{
		SequenceNumber:      1,
		AckedSequenceNumber: 0,
		CharacterData:       []byte("output"),
	}).Pack()
	packet := (&types.Rmcp{
		RmcpHeader: types.NewRmcpHeader(),
		Session20: &types.Session20{
			SessionHeader20: &types.SessionHeader20{
				AuthType:      types.AuthTypeRMCPPlus,
				PayloadType:   types.PayloadTypeSOL,
				PayloadLength: uint16(len(payload)),
			},
			SessionPayload: payload,
		},
	}).Pack()

	matched, err := client.tryMatchSOLResponse(packet, 2)
	if err != nil {
		t.Fatalf("tryMatchSOLResponse() error = %v", err)
	}
	if matched {
		t.Fatal("tryMatchSOLResponse() matched asynchronous output")
	}

	responses := stream.takePendingOutput()
	if len(responses) != 1 || string(responses[0].CharacterData) != "output" {
		t.Fatalf("pending SOL responses = %#v, want asynchronous output", responses)
	}
}

func TestRegisterSOLStreamRejectsActiveStream(t *testing.T) {
	client := &Client{}
	if err := client.registerSOLStream(&solStream{}); err != nil {
		t.Fatalf("registerSOLStream() error = %v", err)
	}
	defer client.unregisterSOLStream()

	if err := client.registerSOLStream(&solStream{}); err == nil {
		t.Fatal("registerSOLStream() accepted a second active stream")
	}
}

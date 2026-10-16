package client

import (
	"bytes"
	"context"
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

	// A trailing duplicate means the BMC never saw our acknowledgement (or its
	// retry raced it): §15.11 requires acknowledging every non-zero packet, so
	// the duplicate must schedule a fresh ACK of the current state — a
	// Table 15-3 "Completion ACK" stops the BMC's retry engine.
	stream.ackPending = false
	if err := stream.processOutput(&types.SOLPayloadResponse{
		SOLPayloadPacket: types.SOLPayloadPacket{SequenceNumber: 15, CharacterData: []byte("def")},
	}); err != nil {
		t.Fatalf("processOutput() error = %v", err)
	}
	if !stream.ackPending {
		t.Fatal("dropped duplicate did not re-schedule the acknowledgement")
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

func TestSOLStreamDispatcherDeliversAsynchronousOutput(t *testing.T) {
	client := &Client{}
	stream := &solStream{outputCh: make(chan *types.SOLPayloadResponse, 16)}
	client.activeSOLStream = stream

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

	client.dispatchInboundDatagram(packet)
	select {
	case got := <-stream.outputCh:
		if string(got.CharacterData) != "output" {
			t.Fatalf("dispatched output = %q, want %q", got.CharacterData, "output")
		}
	default:
		t.Fatal("dispatcher did not deliver asynchronous output to the stream")
	}

	// Command responses claimed by an exchange also pass through the
	// dispatcher; non-SOL datagrams must return early without touching the
	// stream.
	ipmiPacket := (&types.Rmcp{
		RmcpHeader: types.NewRmcpHeader(),
		Session20: &types.Session20{
			SessionHeader20: &types.SessionHeader20{
				AuthType:      types.AuthTypeRMCPPlus,
				PayloadType:   types.PayloadTypeIPMI,
				PayloadLength: 8,
			},
			SessionPayload: []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x01, 0x51, 0x9f},
		},
	}).Pack()
	client.dispatchInboundDatagram(ipmiPacket)
	select {
	case got := <-stream.outputCh:
		t.Fatalf("dispatcher delivered a non-SOL datagram to the stream: %+v", got)
	default:
	}
}

func TestRegisterSOLStreamRejectsActiveStream(t *testing.T) {
	client := &Client{udpClient: NewUDPClient("127.0.0.1", 1)}
	defer client.udpClient.Close()
	if err := client.registerSOLStream(context.Background(), &solStream{}); err != nil {
		t.Fatalf("registerSOLStream() error = %v", err)
	}
	defer client.unregisterSOLStream()

	if err := client.registerSOLStream(context.Background(), &solStream{}); err == nil {
		t.Fatal("registerSOLStream() accepted a second active stream")
	}
}

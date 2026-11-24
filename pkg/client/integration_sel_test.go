package client

import (
	"context"
	"encoding/binary"
	"net"
	"sync/atomic"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

// selTestRecords is the number of records the test SEL serves, IDs 1..n.
const selTestRecords = 5

// newSELTestServer starts the reference BMC with Get SEL Info and Get SEL
// Entry handlers that serve selTestRecords standard records and count the
// requests they receive. The reference BMC has no SEL of its own.
func newSELTestServer(t *testing.T) (c *Client, infoCalls, entryCalls *atomic.Int32) {
	t.Helper()
	infoCalls, entryCalls = &atomic.Int32{}, &atomic.Int32{}

	reg := handlers.NewRegistry()
	handlers.RegisterAllHandlers(reg)
	reg.RegisterFunc(types.CommandGetSELInfo, func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
		infoCalls.Add(1)
		resp := make([]byte, 14)
		resp[0] = 0x51 // SEL version 1.5 (2.0 compatible)
		binary.LittleEndian.PutUint16(resp[1:3], selTestRecords)
		return resp, types.CodeOK, nil
	})
	reg.RegisterFunc(types.CommandGetSELEntry, func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
		entryCalls.Add(1)
		if len(data) < 6 {
			return nil, types.CodeRequestDataLengthInvalid, nil
		}
		id := binary.LittleEndian.Uint16(data[2:4])
		if id == 0 { // 0000h: first entry
			id = 1
		}
		if id > selTestRecords {
			return nil, types.CodeRequestedDataNotPresent, nil
		}
		next := id + 1
		if id == selTestRecords {
			next = 0xffff
		}
		resp := make([]byte, 18)
		binary.LittleEndian.PutUint16(resp[0:2], next)
		rec := resp[2:]
		binary.LittleEndian.PutUint16(rec[0:2], id)
		rec[2] = 0x02                                       // standard record
		binary.LittleEndian.PutUint32(rec[3:7], 0x60000000) // timestamp
		binary.LittleEndian.PutUint16(rec[7:9], 0x0020)     // generator: BMC
		rec[9] = 0x04                                       // EvM rev 2.0
		rec[10] = 0x04                                      // sensor type: fan
		rec[11] = uint8(id)                                 // sensor number
		rec[12] = 0x01                                      // threshold, assertion
		return resp, types.CodeOK, nil
	})

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, mock.New(), bmc.WithClock(clock.Real))
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	srv := server.NewServer(b, udp.Wrap(pc), server.WithHandlerRegistry(reg))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = srv.Close() })
	go func() { _ = srv.Serve(ctx) }()

	addr := pc.LocalAddr().(*net.UDPAddr)
	c, err = NewClient(addr.IP.String(), addr.Port, "ADMIN", "ADMIN")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLanplus)
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c, infoCalls, entryCalls
}

// TestGetSELEntriesSendsSELInfoOnce checks that reading the whole SEL sends
// Get SEL Info once, not once per record, both with GetSELEntries and with
// GetSELEntriesStream.
func TestGetSELEntriesSendsSELInfoOnce(t *testing.T) {
	t.Run("GetSELEntries", func(t *testing.T) {
		c, info, entry := newSELTestServer(t)
		sels, err := c.GetSELEntries(context.Background(), 0)
		if err != nil {
			t.Fatalf("GetSELEntries: %v", err)
		}
		if len(sels) != selTestRecords {
			t.Fatalf("got %d records, want %d", len(sels), selTestRecords)
		}
		if got := info.Load(); got != 1 {
			t.Errorf("Get SEL Info sent %d times, want 1", got)
		}
		if got := entry.Load(); got != selTestRecords {
			t.Errorf("Get SEL Entry sent %d times, want %d", got, selTestRecords)
		}
	})
	t.Run("GetSELEntriesStream", func(t *testing.T) {
		c, info, entry := newSELTestServer(t)
		n := 0
		for r := range c.GetSELEntriesStream(context.Background(), 0) {
			if r.Err != nil {
				t.Fatalf("stream: %v", r.Err)
			}
			n++
		}
		if n != selTestRecords {
			t.Fatalf("got %d records, want %d", n, selTestRecords)
		}
		if got := info.Load(); got != 1 {
			t.Errorf("Get SEL Info sent %d times, want 1", got)
		}
		if got := entry.Load(); got != selTestRecords {
			t.Errorf("Get SEL Entry sent %d times, want %d", got, selTestRecords)
		}
	})
	t.Run("GetSELEntry still sends its own", func(t *testing.T) {
		c, info, _ := newSELTestServer(t)
		if _, err := c.GetSELEntry(context.Background(), 0, 0); err != nil {
			t.Fatalf("GetSELEntry: %v", err)
		}
		if got := info.Load(); got != 1 {
			t.Errorf("Get SEL Info sent %d times, want 1", got)
		}
	})
}

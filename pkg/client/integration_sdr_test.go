package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Sensor numbers in the test repository and how their reading is answered.
const (
	sdrTestOK1     = 0x01 // answers
	sdrTestCCError = 0x02 // answers with completion code 0xff
	sdrTestSilent  = 0x03 // answers too late (after the client's timeout)
	sdrTestOK2     = 0x04 // answers
)

const (
	sdrTestClientTimeout = 400 * time.Millisecond
	sdrTestSilentDelay   = 600 * time.Millisecond
)

// newSDRTestServer starts the reference BMC with four compact threshold
// sensors in its SDR repository and a Get Sensor Reading handler that
// answers them as the constants above say. The reference BMC has no sensor
// reading handler of its own.
func newSDRTestServer(t *testing.T) *Client {
	t.Helper()

	m := mock.New()
	for i, num := range []uint8{sdrTestOK1, sdrTestCCError, sdrTestSilent, sdrTestOK2} {
		name := []byte{'S', '0' + num}
		raw := (&types.SDRCompact{
			GeneratorID:            0x0020,
			SensorNumber:           types.SensorNumber(num),
			SensorType:             types.SensorTypeTemperature,
			SensorEventReadingType: types.EventReadingTypeThreshold,
			SensorUnit:             types.SensorUnit{BaseUnit: types.SensorUnitType_DegreesC},
			IDStringTypeLength:     types.TypeLength(0xC0 | len(name)),
			IDStringBytes:          name,
		}).Pack(uint16(i + 1))
		if err := m.Storage().SDR().Write(context.Background(), uint16(i+1), raw); err != nil {
			t.Fatalf("write SDR: %v", err)
		}
	}

	reg := handlers.NewRegistry()
	handlers.RegisterAllHandlers(reg)
	reg.RegisterFunc(types.CommandGetSensorReading, func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
		if len(data) < 1 {
			return nil, types.CodeRequestDataLengthInvalid, nil
		}
		switch data[0] {
		case sdrTestCCError:
			return nil, types.CompletionCode(0xff), nil
		case sdrTestSilent:
			time.Sleep(sdrTestSilentDelay)
		}
		// reading 0x30; event messages and scanning enabled; no threshold crossed
		return []byte{0x30, 0xc0, 0x00}, types.CodeOK, nil
	})

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, bmc.WithClock(clock.Real))
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
	c, err := NewClient(addr.IP.String(), addr.Port, "ADMIN", "ADMIN")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.WithInterface(InterfaceLanplus)
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	// Short per-request timeout, no retry, so the silent sensor times out
	// once; set after Connect so session set-up keeps the defaults.
	c.WithTimeout(sdrTestClientTimeout).WithRetry(0)
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c
}

// TestGetSDRsSurvivesOneSensor checks that a sensor whose reading fails,
// by an error completion code or by no answer in time, is listed as N/A
// instead of failing GetSDRs (and GetSensors) for every sensor.
func TestGetSDRsSurvivesOneSensor(t *testing.T) {
	c := newSDRTestServer(t)

	sdrs, err := c.GetSDRs(context.Background(), types.SDRRecordTypeCompactSensor)
	if err != nil {
		t.Fatalf("GetSDRs: %v", err)
	}
	if len(sdrs) != 4 {
		t.Fatalf("GetSDRs returned %d records, want 4", len(sdrs))
	}
	for _, s := range sdrs {
		num := uint8(s.Compact.SensorNumber)
		failed := num == sdrTestCCError || num == sdrTestSilent
		if got := s.Compact.SensorStatus; (got == "N/A") != failed {
			t.Errorf("sensor %#02x: status %q, want N/A only for the failing sensors", num, got)
		}
	}

	sensors, err := c.GetSensors(context.Background())
	if err != nil {
		t.Fatalf("GetSensors: %v", err)
	}
	if len(sensors) != 4 {
		t.Fatalf("GetSensors returned %d sensors, want 4", len(sensors))
	}
	for _, s := range sensors {
		failed := s.Number == sdrTestCCError || s.Number == sdrTestSilent
		if s.IsReadingValid() == failed {
			t.Errorf("sensor %#02x: reading valid %v", s.Number, s.IsReadingValid())
		}
	}
}

// A cancelled context still stops the walk.
func TestGetSDRsStopsOnCancelledContext(t *testing.T) {
	c := newSDRTestServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.GetSDRs(ctx, types.SDRRecordTypeCompactSensor); err == nil {
		t.Fatal("GetSDRs succeeded with a cancelled context")
	}
}

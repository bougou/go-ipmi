package client

import (
	"context"
	"errors"
	"iter"
	"net"
	"slices"
	"sync/atomic"
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

// newSDRTestServer starts the reference BMC with three Compact threshold
// sensors, a Full threshold sensor, an Event-Only record and a handler that
// answers them as the constants above say. The reference BMC has no sensor
// reading handler of its own.
func newSDRTestServer(t *testing.T) (*Client, *[256]atomic.Int32) {
	t.Helper()

	m := mock.New()
	for i, num := range []uint8{sdrTestOK1, sdrTestCCError, sdrTestSilent} {
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

	full := types.SDRFull{
		GeneratorID:            0x20,
		SensorNumber:           sdrTestOK2,
		SensorType:             types.SensorTypeTemperature,
		SensorEventReadingType: types.EventReadingTypeThreshold,
		ReadingFactors:         types.ReadingFactors{M: 1},
	}
	if err := m.Storage().SDR().Write(context.Background(), 4, full.Pack(4)); err != nil {
		t.Fatal(err)
	}
	event := types.SDREventOnly{SensorNumber: 5}
	if err := m.Storage().SDR().Write(context.Background(), 5, event.Pack(5)); err != nil {
		t.Fatal(err)
	}

	reads := new([256]atomic.Int32)
	reg := handlers.NewRegistry()
	handlers.RegisterAllHandlers(reg)
	reg.RegisterFunc(types.CommandGetSensorReading, func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
		if len(data) < 1 {
			return nil, types.CodeRequestDataLengthInvalid, nil
		}
		n := reads[data[0]].Add(1)
		switch data[0] {
		case sdrTestCCError:
			return nil, types.CompletionCode(0xff), nil
		case sdrTestSilent:
			// A second sample would succeed and hide the first timeout.
			if n == 1 {
				time.Sleep(sdrTestSilentDelay)
			}
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
	return c, reads
}

func TestGetSDRsSurvivesOneSensor(t *testing.T) {
	c, _ := newSDRTestServer(t)
	sdrs, err := c.GetSDRs(context.Background(), types.SDRRecordTypeCompactSensor, types.SDRRecordTypeFullSensor)
	if err != nil || len(sdrs) != 4 {
		t.Fatalf("GetSDRs: got %d records, err %v", len(sdrs), err)
	}
	if sdrs[0].Compact.SensorValue != 48 || sdrs[0].Compact.SensorStatus != "ok" {
		t.Fatal("missing enhanced Compact reading")
	}
	if sdrs[1].Compact.SensorStatus != "N/A" || sdrs[2].Compact.SensorStatus != "N/A" {
		t.Fatal("failed reads must have status N/A")
	}
	if sdrs[3].Full.SensorValue != 48 || sdrs[3].Full.SensorStatus != "ok" {
		t.Fatal("missing enhanced Full reading")
	}
}

func TestSensorListsSampleEachSensorOnce(t *testing.T) {
	for _, mode := range []string{"batch", "stream"} {
		t.Run(mode, func(t *testing.T) {
			c, reads := newSDRTestServer(t)
			ctx := context.Background()
			var sensors []*types.Sensor
			if mode == "batch" {
				var err error
				sensors, err = c.GetSensors(ctx)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				for result := range c.GetSensorsStream(ctx) {
					if result.Err != nil {
						t.Fatal(result.Err)
					}
					sensors = append(sensors, result.Ok)
				}
			}
			if len(sensors) != 4 {
				t.Fatalf("got %d sensors, want 4", len(sensors))
			}
			for i, s := range sensors {
				num := uint8(i + 1)
				if s.Number != num {
					t.Fatalf("got sensor %d at index %d, want %d", s.Number, i, num)
				}
				if n := reads[num].Load(); n != 1 {
					t.Errorf("sensor %d: %d reading requests, want 1", num, n)
				}
				if num == sdrTestCCError || num == sdrTestSilent {
					if s.Status() != "N/A" || s.IsReadingValid() {
						t.Errorf("sensor %d: expected an unavailable reading", num)
					}
				} else if !s.IsReadingValid() || s.Value != 48 {
					t.Errorf("sensor %d: expected valid reading 48, got %s", num, s.ReadingStr())
				}
			}
			if n := reads[5].Load(); n != 0 {
				t.Errorf("Event-Only sensor: %d reading requests", n)
			}
		})
	}
}

func TestSensorStreamStopsAfterFirstMatch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stream func(*Client, context.Context, ...SensorFilterOption) iter.Seq[*types.Result[types.Sensor]]
	}{{"AND", (*Client).GetSensorsStream}, {"OR", (*Client).GetSensorsAnyStream}} {
		t.Run(tc.name, func(t *testing.T) {
			c, reads := newSDRTestServer(t)
			var count int
			for result := range tc.stream(c, context.Background(), SensorFilterOptionIsReadingValid) {
				if result.Err != nil {
					t.Fatal(result.Err)
				}
				if result.Ok.Number != sdrTestOK1 {
					t.Fatalf("unexpected sensor %d", result.Ok.Number)
				}
				count++
				break
			}
			if count != 1 || reads[sdrTestOK1].Load() != 1 {
				t.Fatal("expected one sample and one result for the first sensor")
			}
			for num := sdrTestCCError; num <= 5; num++ {
				if reads[num].Load() != 0 {
					t.Fatalf("sampled sensor %d after early exit", num)
				}
			}
		})
	}
}

// A cancelled context still stops the walk.
func TestSensorDiscoveryStopsOnCancelledContext(t *testing.T) {
	c, _ := newSDRTestServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.GetSensors(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetSensors with cancelled context: %v", err)
	}
	if _, err := c.GetSDRs(ctx, types.SDRRecordTypeCompactSensor); err == nil {
		t.Fatal("GetSDRs succeeded with a cancelled context")
	}
}

func TestGetSensorsFiltersAfterSampling(t *testing.T) {
	c, reads := newSDRTestServer(t)
	full := func(s *types.Sensor) bool { return s.SDRRecordType == types.SDRRecordTypeFullSensor }
	sensors, err := c.GetSensors(context.Background(), SensorFilterOptionIsReadingValid, full)
	if err != nil {
		t.Fatal(err)
	}
	if len(sensors) != 1 || sensors[0].Number != sdrTestOK2 {
		t.Fatalf("want only the healthy Full sensor, got %v", sensors)
	}
	for n := sdrTestOK1; n <= sdrTestOK2; n++ {
		if got := reads[n].Load(); got != 1 {
			t.Errorf("sensor %d: got %d samples before filtering, want 1", n, got)
		}
	}
}

func TestAnySensorLists(t *testing.T) {
	for _, stream := range []bool{false, true} {
		mode := "batch"
		if stream {
			mode = "stream"
		}
		t.Run(mode, func(t *testing.T) {
			c, reads := newSDRTestServer(t)
			ctx := context.Background()
			list := func(filters ...SensorFilterOption) []*types.Sensor {
				t.Helper()
				if !stream {
					sensors, err := c.GetSensorsAny(ctx, filters...)
					if err != nil {
						t.Fatal(err)
					}
					return sensors
				}
				var sensors []*types.Sensor
				for result := range c.GetSensorsAnyStream(ctx, filters...) {
					if result.Err != nil {
						t.Fatal(result.Err)
					}
					sensors = append(sensors, result.Ok)
				}
				return sensors
			}
			first := func(s *types.Sensor) bool { return s.Number == sdrTestOK1 || s.Number == sdrTestSilent }
			var checked []uint8
			second := func(s *types.Sensor) bool {
				checked = append(checked, s.Number)
				return s.IsReadingValid()
			}
			var got []uint8
			for _, s := range list(first, second) {
				got = append(got, s.Number)
				if s.Number == sdrTestSilent && (s.IsReadingValid() || s.Status() != "N/A") {
					t.Error("second sample masked the timeout")
				}
			}
			if !slices.Equal(got, []uint8{sdrTestOK1, sdrTestSilent, sdrTestOK2}) {
				t.Errorf("OR matches in repository order: %v", got)
			}
			if !slices.Equal(checked, []uint8{sdrTestCCError, sdrTestOK2}) {
				t.Errorf("second filter called for sensors %v", checked)
			}
			for n := sdrTestOK1; n <= sdrTestOK2; n++ {
				if count := reads[n].Load(); count != 1 {
					t.Errorf("sensor %d: got %d reading requests, want 1", n, count)
				}
			}
			if count := reads[5].Load(); count != 0 {
				t.Errorf("Event-Only record: got %d reading requests, want 0", count)
			}
			if sensors := list(); len(sensors) != 0 {
				t.Errorf("empty OR filter returned %d sensors", len(sensors))
			}
		})
	}
}

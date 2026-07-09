package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/cmd/chassis"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// newTestBMCWithMock builds a BMC backed by the supplied mock HAL so tests can
// inspect and drive chassis state directly.
func newTestBMCWithMock(m *mock.HAL) *bmc.BMC {
	info := bmc.DeviceInfo{
		DeviceID:       1,
		IPMIVersion:    0x20,
		ManufacturerID: 0x000157,
		ProductID:      0x0001,
	}
	var guid [16]byte
	return bmc.New(info, guid, m, bmc.WithClock(clock.Real))
}

func TestHandleGetChassisStatus_RoundTrip(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	ch.On = true
	ch.Intruded = false
	hctx := &HandlerContext{BMC: b}

	resp, cc, err := handleGetChassisStatus(context.Background(), hctx, nil)
	if err != nil || cc != CodeOK {
		t.Fatalf("unexpected cc=%d err=%v", cc, err)
	}
	// Re-encode through the typed codec and back to validate round-trip.
	var decoded chassis.GetChassisStatusResponse
	if err := decoded.Unpack(resp); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if decoded.PowerIsOn != true {
		t.Fatalf("PowerIsOn: want true, got %v", decoded.PowerIsOn)
	}
	if decoded.ChassisIdentifySupported != true {
		t.Fatalf("ChassisIdentifySupported: want true, got %v", decoded.ChassisIdentifySupported)
	}
	if decoded.ChassisIntrusionActive != false {
		t.Fatalf("ChassisIntrusionActive: want false, got %v", decoded.ChassisIntrusionActive)
	}
}

func TestHandleChassisControl_PowerCycle(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	hctx := &HandlerContext{BMC: b}

	req := (&chassis.ChassisControlRequest{ChassisControl: chassis.ChassisControlPowerCycle}).Pack()
	_, cc, err := handleChassisControl(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("PowerCycle: want CodeOK, got cc=%d err=%v", cc, err)
	}
	if ch.PowerCycles != 1 {
		t.Fatalf("PowerCycle dispatch: want 1 call, got %d", ch.PowerCycles)
	}

	// Unknown action still returns CodeParamOutOfRange (regression guard for
	// the previous behaviour where PowerCycle fell through to the default).
	_, cc, _ = handleChassisControl(context.Background(), hctx, []byte{0x0F})
	if cc != CodeParamOutOfRange {
		t.Fatalf("unknown action: want CodeParamOutOfRange, got %d", cc)
	}
}

// TestHandleChassisControl_NodeBusy verifies that when a HAL method returns a
// [CompletionCode] as an error (e.g. [CodeNodeBusy]), [codeFromErr] extracts
// the specific code rather than falling back to [CodeUnspecifiedError].
func TestHandleChassisControl_NodeBusy(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	// SetPowerHook allows the test to inject a CompletionCode as error.
	ch.SetPowerHook = func(on bool) error {
		return CodeNodeBusy
	}
	hctx := &HandlerContext{BMC: b}

	req := (&chassis.ChassisControlRequest{ChassisControl: chassis.ChassisControlPowerDown}).Pack()
	_, cc, _ := handleChassisControl(context.Background(), hctx, req)
	if cc != CodeNodeBusy {
		t.Fatalf("want CodeNodeBusy (0xC0), got cc=%d", cc)
	}
}

func TestHandleChassisControl_TypedDispatch(t *testing.T) {
	cases := []struct {
		name   string
		action chassis.ChassisControl
		check  func(*mock.Chassis) bool
	}{
		{"PowerDown", chassis.ChassisControlPowerDown, func(c *mock.Chassis) bool { return !c.On }},
		{"PowerUp", chassis.ChassisControlPowerUp, func(c *mock.Chassis) bool { return c.On }},
		{"HardReset", chassis.ChassisControlHardReset, func(c *mock.Chassis) bool { return c.ColdResets == 1 }},
		{"SoftShutdown", chassis.ChassisControlSoftShutdown, func(c *mock.Chassis) bool { return c.WarmResets == 1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mock.New()
			b := newTestBMCWithMock(m)
			ch := b.HAL().Chassis().(*mock.Chassis)
			hctx := &HandlerContext{BMC: b}

			req := (&chassis.ChassisControlRequest{ChassisControl: tc.action}).Pack()
			_, cc, err := handleChassisControl(context.Background(), hctx, req)
			if err != nil || cc != CodeOK {
				t.Fatalf("%s: want CodeOK, got cc=%d err=%v", tc.name, cc, err)
			}
			if !tc.check(ch) {
				t.Fatalf("%s: HAL state check failed (On=%v ColdResets=%d WarmResets=%d PowerCycles=%d)",
					tc.name, ch.On, ch.ColdResets, ch.WarmResets, ch.PowerCycles)
			}
		})
	}
}

func TestHandleSetSystemBootOptions_BootFlags(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	hctx := &HandlerContext{BMC: b}

	flags := &types.BootOptionParam_BootFlags{
		BootFlagsValid:     true,
		Persist:            true,
		BootDeviceSelector: types.BootDeviceSelectorForcePXE,
	}
	req := append([]byte{byte(types.BootOptionParamSelector_BootFlags)}, flags.Pack()...)

	_, cc, err := handleSetSystemBootOptions(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("want CodeOK, got cc=%d err=%v", cc, err)
	}
	if ch.BootFlags == nil {
		t.Fatalf("SetBootFlags not invoked on HAL")
	}
	if !ch.BootFlags.Persist {
		t.Fatalf("Persist bit lost: HAL received %+v", ch.BootFlags)
	}
	if ch.BootFlags.BootDeviceSelector != types.BootDeviceSelectorForcePXE {
		t.Fatalf("BootDeviceSelector: want PXE, got %v", ch.BootFlags.BootDeviceSelector)
	}
}

func TestHandleGetSystemBootOptions_BootFlags(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	ch.BootFlags = &types.BootOptionParam_BootFlags{
		BootFlagsValid:     true,
		BootDeviceSelector: types.BootDeviceSelectorForcePXE,
	}
	hctx := &HandlerContext{BMC: b}

	req := []byte{byte(types.BootOptionParamSelector_BootFlags)}
	resp, cc, err := handleGetSystemBootOptions(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("want CodeOK, got cc=%d err=%v", cc, err)
	}
	if len(resp) < 2 {
		t.Fatalf("response too short: %d", len(resp))
	}
	// resp[0]=version, resp[1]=selector, resp[2:]=param data (5 bytes).
	var decoded types.BootOptionParam_BootFlags
	if err := decoded.Unpack(resp[2:]); err != nil {
		t.Fatalf("Unpack boot flags: %v", err)
	}
	if decoded.BootDeviceSelector != types.BootDeviceSelectorForcePXE {
		t.Fatalf("BootDeviceSelector round-trip: want PXE, got %v", decoded.BootDeviceSelector)
	}
}

func TestHandleGetSystemBootOptions_NotSupported(t *testing.T) {
	// Use a HAL whose GetBootFlags returns ErrNotSupported. The mock returns
	// ErrNotSupported when no boot flags have been stored.
	m := mock.New()
	b := newTestBMCWithMock(m)
	hctx := &HandlerContext{BMC: b}

	req := []byte{byte(types.BootOptionParamSelector_BootFlags)}
	_, cc, err := handleGetSystemBootOptions(context.Background(), hctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CodeBootParamNotSupported {
		t.Fatalf("want CodeBootParamNotSupported for ErrNotSupported, got %d", cc)
	}
}

func TestHandleSetSystemBootOptions_OtherParam(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	ch := b.HAL().Chassis().(*mock.Chassis)
	hctx := &HandlerContext{BMC: b}

	// Set In Progress (0x00) is not implemented; spec §28.12 requires 80h (CodeBootParamNotSupported)
	// for unsupported parameters.
	req := []byte{byte(types.BootOptionParamSelector_SetInProgress), 0x01}
	_, cc, err := handleSetSystemBootOptions(context.Background(), hctx, req)
	if err != nil || cc != CodeBootParamNotSupported {
		t.Fatalf("want CodeBootParamNotSupported (80h) for unimplemented param, got cc=%d err=%v", cc, err)
	}
	if ch.BootFlags != nil {
		t.Fatalf("HAL SetBootFlags must not be called for non-BootFlags params")
	}
}

func TestHandleSetSystemBootOptions_Truncated(t *testing.T) {
	m := mock.New()
	b := newTestBMCWithMock(m)
	hctx := &HandlerContext{BMC: b}

	// BootFlags param selector but only 2 bytes of data (need 5).
	req := []byte{byte(types.BootOptionParamSelector_BootFlags), 0x01, 0x02}
	_, cc, _ := handleSetSystemBootOptions(context.Background(), hctx, req)
	if cc != CodeRequestDataTruncated {
		t.Fatalf("want CodeRequestDataTruncated, got %d", cc)
	}

	// Empty request.
	_, cc, _ = handleSetSystemBootOptions(context.Background(), hctx, nil)
	if cc != CodeRequestDataTruncated {
		t.Fatalf("empty request: want CodeRequestDataTruncated, got %d", cc)
	}
}

func TestCodeFromHalErr(t *testing.T) {
	if got := codeFromHalErr(nil); got != CodeOK {
		t.Fatalf("nil: want CodeOK, got %d", got)
	}
	if got := codeFromHalErr(hal.ErrNotSupported); got != CodeBootParamNotSupported {
		t.Fatalf("ErrNotSupported: want CodeBootParamNotSupported, got %d", got)
	}
	if got := codeFromHalErr(errors.New("boom")); got != CodeUnspecifiedError {
		t.Fatalf("generic error: want CodeUnspecifiedError, got %d", got)
	}
}

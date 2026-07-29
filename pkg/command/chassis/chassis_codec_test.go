package chassis

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestGetChassisStatusResponse_PackUnpackRoundTrip(t *testing.T) {
	original := &GetChassisStatusResponse{
		PowerRestorePolicy:           PowerRestorePolicyAlwaysOn,
		PowerControlFault:            true,
		PowerFault:                   false,
		InterLock:                    true,
		PowerOverload:                false,
		PowerIsOn:                    true,
		LastPowerOnByCommand:         true,
		LastPowerDownByPowerFault:    false,
		ACFailed:                     true,
		ChassisIdentifySupported:     true,
		ChassisIdentifyState:         ChassisIdentifyStateTemporaryOn,
		CollingFanFault:              true,
		DriveFault:                   false,
		FrontPanelLockoutActive:      true,
		ChassisIntrusionActive:       false,
		SleepButtonDisableAllowed:    true,
		ResetButtonDisabled:          true,
		PoweroffButtonDisableAllowed: true,
	}

	packed := original.Pack()
	if len(packed) != 4 {
		t.Fatalf("Pack: want 4 bytes (front-panel bits set), got %d", len(packed))
	}

	var decoded GetChassisStatusResponse
	if err := decoded.Unpack(packed); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if decoded.PowerRestorePolicy != original.PowerRestorePolicy {
		t.Fatalf("PowerRestorePolicy: want %v, got %v", original.PowerRestorePolicy, decoded.PowerRestorePolicy)
	}
	if decoded.PowerIsOn != original.PowerIsOn {
		t.Fatalf("PowerIsOn: want %v, got %v", original.PowerIsOn, decoded.PowerIsOn)
	}
	if decoded.ChassisIdentifyState != original.ChassisIdentifyState {
		t.Fatalf("ChassisIdentifyState: want %v, got %v", original.ChassisIdentifyState, decoded.ChassisIdentifyState)
	}
	if decoded.SleepButtonDisableAllowed != original.SleepButtonDisableAllowed {
		t.Fatalf("SleepButtonDisableAllowed: want %v, got %v", original.SleepButtonDisableAllowed, decoded.SleepButtonDisableAllowed)
	}
	if decoded.PoweroffButtonDisableAllowed != original.PoweroffButtonDisableAllowed {
		t.Fatalf("PoweroffButtonDisableAllowed: want %v, got %v", original.PoweroffButtonDisableAllowed, decoded.PoweroffButtonDisableAllowed)
	}
}

func TestGetChassisStatusResponse_PackAlwaysFourBytes(t *testing.T) {
	// Per §28.2 Table 28-3, byte 3 is optional but "Return as 00h if the panel
	// button disable function is not supported." Pack always emits 4 bytes.
	res := &GetChassisStatusResponse{PowerIsOn: true}
	packed := res.Pack()
	if len(packed) != 4 {
		t.Fatalf("Pack without front-panel bits: want 4 bytes (spec says return as 00h), got %d", len(packed))
	}
	// Byte 0 bit 0 must reflect PowerIsOn.
	if packed[0]&0x01 == 0 {
		t.Fatalf("PowerIsOn not encoded: byte0=0x%02x", packed[0])
	}
	// Byte 3 must be 0x00 when no front-panel button disable bits are set.
	if packed[3] != 0x00 {
		t.Fatalf("byte 3 with no front-panel bits: want 0x00, got 0x%02x", packed[3])
	}
}

func TestChassisControlRequest_Unpack(t *testing.T) {
	cases := []struct {
		in   byte
		want ChassisControl
	}{
		{0x00, ChassisControlPowerDown},
		{0x01, ChassisControlPowerUp},
		{0x02, ChassisControlPowerCycle},
		{0x03, ChassisControlHardReset},
		{0x05, ChassisControlSoftShutdown},
		// Upper nibble is reserved and must be ignored.
		{0xF2, ChassisControlPowerCycle},
	}
	for _, tc := range cases {
		var req ChassisControlRequest
		if err := req.Unpack([]byte{tc.in}); err != nil {
			t.Fatalf("Unpack(0x%02x): %v", tc.in, err)
		}
		if req.ChassisControl != tc.want {
			t.Fatalf("Unpack(0x%02x): want %v, got %v", tc.in, tc.want, req.ChassisControl)
		}
	}

	var req ChassisControlRequest
	if err := req.Unpack(nil); err == nil {
		t.Fatalf("Unpack(nil) should error")
	}

	// Pack/Unpack symmetry.
	orig := &ChassisControlRequest{ChassisControl: ChassisControlHardReset}
	var round ChassisControlRequest
	if err := round.Unpack(orig.Pack()); err != nil {
		t.Fatalf("round-trip Unpack: %v", err)
	}
	if round.ChassisControl != orig.ChassisControl {
		t.Fatalf("round-trip: want %v, got %v", orig.ChassisControl, round.ChassisControl)
	}
	// Touch ipmi to keep the import meaningful for future helper-based cases.
	_ = types.CommandChassisControl
}

package app

import (
	"strings"
	"testing"
)

// TestGetWatchdogTimerResponseUnpack exercises the decoding of the Get Watchdog
// Timer response (IPMI v2.0 27.7), in particular the units of the countdown
// values: the pre-timeout interval is expressed in seconds, but the initial and
// present countdown values are expressed in 100ms units.
func TestGetWatchdogTimerResponseUnpack(t *testing.T) {
	cases := []struct {
		name string
		data []byte

		wantTimerUse            TimerUse
		wantTimerIsStarted      bool
		wantDontLog             bool
		wantTimeoutAction       TimeoutAction
		wantPreTimeoutInterrupt PreTimeoutInterrupt
		wantPreTimeoutSec       uint8
		wantExpirationFlags     uint8
		wantInitialCountdown    uint16
		wantPresentCountdown    uint16
		wantInitialSec          float64
		wantPresentSec          float64
	}{
		{
			// SMS/OS timer running with a 60s timeout and 51.5s left, hard
			// reset on timeout, 30s pre-timeout interval, no interrupt.
			name: "running sms/os timer",
			data: []byte{0x44, 0x01, 0x1e, 0x00, 0x58, 0x02, 0x03, 0x02},

			wantTimerUse:         TimerUseSMSOS,
			wantTimerIsStarted:   true,
			wantTimeoutAction:    TimeoutActionHardReset,
			wantPreTimeoutSec:    30,
			wantInitialCountdown: 600,
			wantPresentCountdown: 515,
			wantInitialSec:       60,
			wantPresentSec:       51.5,
		},
		{
			// Stopped BIOS/FRB2 timer, logging disabled, NMI pre-timeout
			// interrupt, expiration flag for BIOS/FRB2 set, expired countdown.
			name: "stopped bios/frb2 timer",
			data: []byte{0x81, 0x22, 0x00, 0x02, 0xb0, 0x04, 0x00, 0x00},

			wantTimerUse:            TimerUseBIOSFRB2,
			wantDontLog:             true,
			wantTimeoutAction:       TimeoutActionPowerDown,
			wantPreTimeoutInterrupt: PreTimeoutInterruptNMI,
			wantExpirationFlags:     0x02,
			wantInitialCountdown:    1200,
			wantInitialSec:          120,
		},
		{
			// Maximum countdown values, 6553.5s.
			name: "maximum countdown",
			data: []byte{0x44, 0x01, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff},

			wantTimerUse:         TimerUseSMSOS,
			wantTimerIsStarted:   true,
			wantTimeoutAction:    TimeoutActionHardReset,
			wantPreTimeoutSec:    255,
			wantInitialCountdown: 65535,
			wantPresentCountdown: 65535,
			wantInitialSec:       6553.5,
			wantPresentSec:       6553.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := &GetWatchdogTimerResponse{}
			if err := res.Unpack(tc.data); err != nil {
				t.Fatalf("Unpack failed: %v", err)
			}

			if res.TimerUse != tc.wantTimerUse {
				t.Errorf("TimerUse = %#02x, want %#02x", uint8(res.TimerUse), uint8(tc.wantTimerUse))
			}
			if res.TimerIsStarted != tc.wantTimerIsStarted {
				t.Errorf("TimerIsStarted = %v, want %v", res.TimerIsStarted, tc.wantTimerIsStarted)
			}
			if res.DontLog != tc.wantDontLog {
				t.Errorf("DontLog = %v, want %v", res.DontLog, tc.wantDontLog)
			}
			if res.TimeoutAction != tc.wantTimeoutAction {
				t.Errorf("TimeoutAction = %#02x, want %#02x", uint8(res.TimeoutAction), uint8(tc.wantTimeoutAction))
			}
			if res.PreTimeoutInterrupt != tc.wantPreTimeoutInterrupt {
				t.Errorf("PreTimeoutInterrupt = %#02x, want %#02x", uint8(res.PreTimeoutInterrupt), uint8(tc.wantPreTimeoutInterrupt))
			}
			if res.PreTimeoutIntervalSec != tc.wantPreTimeoutSec {
				t.Errorf("PreTimeoutIntervalSec = %d, want %d", res.PreTimeoutIntervalSec, tc.wantPreTimeoutSec)
			}
			if res.ExpirationFlags != tc.wantExpirationFlags {
				t.Errorf("ExpirationFlags = %#02x, want %#02x", res.ExpirationFlags, tc.wantExpirationFlags)
			}
			if res.InitialCountdown != tc.wantInitialCountdown {
				t.Errorf("InitialCountdown = %d, want %d", res.InitialCountdown, tc.wantInitialCountdown)
			}
			if res.PresentCountdown != tc.wantPresentCountdown {
				t.Errorf("PresentCountdown = %d, want %d", res.PresentCountdown, tc.wantPresentCountdown)
			}
			if got := res.InitialCountdownSec(); got != tc.wantInitialSec {
				t.Errorf("InitialCountdownSec() = %v, want %v", got, tc.wantInitialSec)
			}
			if got := res.PresentCountdownSec(); got != tc.wantPresentSec {
				t.Errorf("PresentCountdownSec() = %v, want %v", got, tc.wantPresentSec)
			}
		})
	}
}

// TestGetWatchdogTimerResponseTooShort checks that a truncated response is
// rejected rather than silently decoded.
func TestGetWatchdogTimerResponseTooShort(t *testing.T) {
	res := &GetWatchdogTimerResponse{}
	if err := res.Unpack([]byte{0x44, 0x01, 0x1e, 0x00, 0x58, 0x02, 0x03}); err == nil {
		t.Error("Unpack of a 7 byte response succeeded, want error")
	}
}

// TestGetWatchdogTimerResponseFormat checks that the countdown values are
// formatted in the unit the output claims, that is, seconds.
func TestGetWatchdogTimerResponseFormat(t *testing.T) {
	res := &GetWatchdogTimerResponse{}
	if err := res.Unpack([]byte{0x44, 0x01, 0x1e, 0x00, 0x58, 0x02, 0x03, 0x02}); err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	out := res.Format()
	for _, want := range []string{
		"Pre-timeout interval   : 30 seconds",
		"Initial Countdown      : 60.0 sec",
		"Present Countdown      : 51.5 sec",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() output does not contain %q:\n%s", want, out)
		}
	}
}

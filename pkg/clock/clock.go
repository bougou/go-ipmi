// Package clock abstracts time so that embedded targets can inject a hardware
// RTC or monotonic counter, and tests can use a deterministic fake clock.
//
// All server and BMC state-machine code must obtain time exclusively through
// this interface – never via [time.Now] directly.
package clock

import "time"

// Clock provides the current time and factory methods for timers/tickers.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
	NewTicker(d time.Duration) Ticker
}

// Timer mirrors the relevant subset of [time.Timer].
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// Ticker mirrors the relevant subset of [time.Ticker].
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// Real is the standard [Clock] implementation backed by the time package.
var Real Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time                   { return time.Now() }
func (realClock) NewTimer(d time.Duration) Timer   { return &realTimer{t: time.NewTimer(d)} }
func (realClock) NewTicker(d time.Duration) Ticker { return &realTicker{t: time.NewTicker(d)} }

type realTimer struct{ t *time.Timer }

func (r *realTimer) C() <-chan time.Time        { return r.t.C }
func (r *realTimer) Stop() bool                 { return r.t.Stop() }
func (r *realTimer) Reset(d time.Duration) bool { return r.t.Reset(d) }

type realTicker struct{ t *time.Ticker }

func (r *realTicker) C() <-chan time.Time { return r.t.C }
func (r *realTicker) Stop()               { r.t.Stop() }

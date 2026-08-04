package handlers

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

// TestWireConstantsMatchCommandTable pins the raw NetFn/command bytes this
// package compares against to the [types] command table.
//
// The two cannot be collapsed: types.Command entries are struct values, so no
// untyped constant can be derived from them, and the policy switches below want
// constants for the duplicate-case check the compiler gives them. Asserting the
// values here turns a silent drift into a failing test instead.
func TestWireConstantsMatchCommandTable(t *testing.T) {
	netFns := map[string]struct {
		got  uint8
		want types.NetFn
	}{
		"NetFnAppRequest":     {NetFnAppRequest, types.NetFnAppRequest},
		"NetFnChassisRequest": {NetFnChassisRequest, types.NetFnChassisRequest},
	}
	for name, tc := range netFns {
		if tc.got != uint8(tc.want) {
			t.Errorf("%s = 0x%02x, command table says 0x%02x", name, tc.got, uint8(tc.want))
		}
	}

	commands := map[string]struct {
		got  uint8
		want types.Command
	}{
		"CmdColdReset":                  {CmdColdReset, types.CommandColdReset},
		"CmdWarmReset":                  {CmdWarmReset, types.CommandWarmReset},
		"CmdGetChannelCipherSuites":     {CmdGetChannelCipherSuites, types.CommandGetChannelCipherSuites},
		"CmdChassisControl":             {CmdChassisControl, types.CommandChassisControl},
		"CmdGetChannelAuthCapabilities": {CmdGetChannelAuthCapabilities, types.CommandGetChannelAuthCapabilities},
		"CmdGetSessionChallenge":        {CmdGetSessionChallenge, types.CommandGetSessionChallenge},
		"CmdActivateSession":            {CmdActivateSession, types.CommandActivateSession},
	}
	for name, tc := range commands {
		if tc.got != tc.want.ID {
			t.Errorf("%s = 0x%02x, %q in the command table is 0x%02x", name, tc.got, tc.want.Name, tc.want.ID)
		}
	}
}

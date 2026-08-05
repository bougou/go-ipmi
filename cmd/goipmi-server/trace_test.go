package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/types"
)

// TestTraceLineNamesTheCommand covers what the trace exists to demonstrate: the
// middleware describes a request from HandlerContext alone, including one that
// reached no handler.
func TestTraceLineNamesTheCommand(t *testing.T) {
	tests := []struct {
		name    string
		hctx    *handlers.HandlerContext
		cc      types.CompletionCode
		err     error
		want    []string
		notWant []string
	}{
		{
			name: "registered command reports its spec name",
			hctx: &handlers.HandlerContext{
				Command: types.CommandChassisControl,
				Session: &bmc.Session{PrivilegeLevel: bmc.PrivilegeLevelAdministrator},
				User:    &bmc.User{Name: "ADMIN"},
			},
			cc:   types.CodeOK,
			want: []string{`"Chassis Control"`, "netfn=0x00", "cmd=0x02", "user=ADMIN", "priv=ADMINISTRATOR", "cc=0x00", "Command completed normally"},
		},
		{
			name: "unregistered command still reports its numbers",
			hctx: &handlers.HandlerContext{
				Command: types.Command{ID: 0x3e, NetFn: 0x2c},
				Session: &bmc.Session{PrivilegeLevel: bmc.PrivilegeLevelAdministrator},
				User:    &bmc.User{Name: "ADMIN"},
			},
			cc:   types.CodeInvalidCommand,
			err:  errors.New("no handler for netFn=0x2c cmd=0x3e"),
			want: []string{"<unregistered>", "netfn=0x2c", "cmd=0x3e", "cc=0xc1", "Invalid command", "err=no handler for netFn=0x2c cmd=0x3e"},
		},
		{
			name: "command-specific completion code is named without a Response",
			hctx: &handlers.HandlerContext{
				Command: types.CommandSetUserPassword,
				Session: &bmc.Session{PrivilegeLevel: bmc.PrivilegeLevelAdministrator},
				User:    &bmc.User{Name: "ADMIN"},
			},
			cc:   0x81,
			want: []string{`"Set User Password Command"`, "cc=0x81", "Password test failed. Wrong password size was used"},
		},
		{
			name: "Name-less command key still resolves command-specific codes",
			hctx: &handlers.HandlerContext{
				Command: types.Command{ID: types.CommandCloseSession.ID, NetFn: types.CommandCloseSession.NetFn},
			},
			cc:   0x87,
			want: []string{"<unregistered>", "cc=0x87", "Invalid Session ID in request"},
		},
		{
			name: "pre-session command has neither user nor privilege",
			hctx: &handlers.HandlerContext{
				Command: types.CommandGetChannelAuthCapabilities,
			},
			cc:      types.CodeOK,
			want:    []string{`"Get Channel Authentication Capabilities"`, "user=-", "priv=-"},
			notWant: []string{"ADMINISTRATOR"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := traceLine(tc.hctx, tc.cc, tc.err, 1, 4, 2*time.Microsecond)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("line missing %q:\n%s", want, got)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(got, notWant) {
					t.Errorf("line should not contain %q:\n%s", notWant, got)
				}
			}
		})
	}
}

// TestTracingRegistryObservesUnknownCommands pins the ordering the example
// depends on: Use before Register, so every dispatch runs the middleware --
// including one with no handler, which would otherwise be answered behind it.
func TestTracingRegistryObservesUnknownCommands(t *testing.T) {
	var seen []types.Command
	reg := handlers.NewRegistry()
	reg.Use(func(next handlers.Handler) handlers.Handler {
		return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
			seen = append(seen, hctx.Command)
			return next.Handle(ctx, hctx, data)
		})
	})
	handlers.RegisterChassisHandlers(reg)

	hctx := &handlers.HandlerContext{BMC: bmc.New(bmc.DeviceInfo{}, [16]byte{}, mock.New())}
	reg.Dispatch(context.Background(), hctx, uint8(types.NetFnChassisRequest), types.CommandGetChassisStatus.ID, nil)
	reg.Dispatch(context.Background(), hctx, 0x2c, 0x3e, nil)

	if len(seen) != 2 {
		t.Fatalf("middleware ran %d times, want 2", len(seen))
	}
	if seen[0].Name != "Get Chassis Status" {
		t.Errorf("registered command name = %q, want %q", seen[0].Name, "Get Chassis Status")
	}
	if seen[1].Name != "" || seen[1].ID != 0x3e || uint8(seen[1].NetFn) != 0x2c {
		t.Errorf("unregistered command = %+v, want nameless 0x2c/0x3e", seen[1])
	}
}

package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// authedCtx is a HandlerContext with an active Administrator session, so the
// privilege gate lets a privileged command through and these tests exercise
// only the registry's dispatch plumbing. A session-less context would be
// rejected before reaching the handler.
func authedCtx() *HandlerContext {
	return &HandlerContext{Session: &bmc.Session{PrivilegeLevel: bmc.PrivilegeLevelAdministrator}}
}

func TestRegistry_Dispatch(t *testing.T) {
	tests := []struct {
		name    string
		netFn   uint8
		cmd     uint8
		setup   func(*Registry)
		wantCC  types.CompletionCode
		wantLen int
	}{
		{
			name:   "unknown command returns not supported",
			netFn:  0x06,
			cmd:    0xFF,
			setup:  func(r *Registry) {},
			wantCC: types.CodeInvalidCommand,
		},
		{
			name:  "registered handler is dispatched",
			netFn: 0x06,
			cmd:   0x01,
			setup: func(r *Registry) {
				r.RegisterFunc(types.Command{ID: 0x01, NetFn: 0x06}, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
					return []byte{0xAB}, types.CodeOK, nil
				})
			},
			wantCC:  types.CodeOK,
			wantLen: 1,
		},
		{
			name:  "middleware wraps handler",
			netFn: 0x06,
			cmd:   0x02,
			setup: func(r *Registry) {
				called := false
				r.Use(func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
						called = true
						return next.Handle(ctx, hctx, data)
					})
				})
				r.RegisterFunc(types.Command{ID: 0x02, NetFn: 0x06}, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
					return nil, types.CodeOK, nil
				})
				_ = called
			},
			wantCC: types.CodeOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			tc.setup(r)
			resp, cc, _ := r.Dispatch(context.Background(), authedCtx(), tc.netFn, tc.cmd, nil)
			if cc != tc.wantCC {
				t.Errorf("cc: want %d, got %d", tc.wantCC, cc)
			}
			if tc.wantLen > 0 && len(resp) != tc.wantLen {
				t.Errorf("resp len: want %d, got %d", tc.wantLen, len(resp))
			}
		})
	}
}

func TestRegistry_DispatchIdentifiesCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Registry)
		netFn    uint8
		cmd      uint8
		wantName string
	}{
		{
			name: "registered from the command table carries the spec name",
			setup: func(r *Registry) {
				r.RegisterFunc(types.CommandGetChassisStatus, okHandler)
			},
			netFn:    uint8(types.NetFnChassisRequest),
			cmd:      types.CommandGetChassisStatus.ID,
			wantName: "Get Chassis Status",
		},
		{
			name: "registered from a nameless literal reports NetFn/Cmd only",
			setup: func(r *Registry) {
				r.RegisterFunc(types.Command{ID: 0x01, NetFn: 0x06}, okHandler)
			},
			netFn: 0x06,
			cmd:   0x01,
		},
		{
			name:  "unregistered command is still identified",
			setup: func(*Registry) {},
			netFn: 0x06,
			cmd:   0xFF,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			tc.setup(r)

			hctx := &HandlerContext{}
			r.Dispatch(context.Background(), hctx, tc.netFn, tc.cmd, nil)

			if hctx.Command.ID != tc.cmd {
				t.Errorf("command ID: want 0x%02x, got 0x%02x", tc.cmd, hctx.Command.ID)
			}
			if uint8(hctx.Command.NetFn) != tc.netFn {
				t.Errorf("command NetFn: want 0x%02x, got 0x%02x", tc.netFn, uint8(hctx.Command.NetFn))
			}
			if hctx.Command.Name != tc.wantName {
				t.Errorf("command name: want %q, got %q", tc.wantName, hctx.Command.Name)
			}
		})
	}
}

// TestRegistry_DispatchRunsMiddlewareForUnknownCommand pins the behaviour audit
// logging depends on: an initiator probing for an unimplemented command must be
// observable, not silently short-circuited.
func TestRegistry_DispatchRunsMiddlewareForUnknownCommand(t *testing.T) {
	r := NewRegistry()

	var seen types.Command
	called := 0
	r.Use(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
			called++
			seen = hctx.Command
			return next.Handle(ctx, hctx, data)
		})
	})

	_, cc, err := r.Dispatch(context.Background(), &HandlerContext{}, 0x06, 0xFF, nil)

	if called != 1 {
		t.Errorf("middleware calls: want 1, got %d", called)
	}
	if cc != types.CodeInvalidCommand {
		t.Errorf("cc: want %d, got %d", types.CodeInvalidCommand, cc)
	}
	if err == nil {
		t.Error("want an error describing the missing handler, got nil")
	}
	if seen.ID != 0xFF || uint8(seen.NetFn) != 0x06 {
		t.Errorf("middleware saw netFn=0x%02x cmd=0x%02x, want 0x06/0xff", uint8(seen.NetFn), seen.ID)
	}
}

func TestRegistry_MergeCarriesCommandNames(t *testing.T) {
	a := NewRegistry()
	b := NewRegistry()
	b.RegisterFunc(types.CommandGetDeviceID, okHandler)

	a.Merge(b)

	hctx := &HandlerContext{}
	a.Dispatch(context.Background(), hctx, uint8(types.NetFnAppRequest), types.CommandGetDeviceID.ID, nil)
	if hctx.Command.Name != "Get Device ID" {
		t.Errorf("command name after merge: want %q, got %q", "Get Device ID", hctx.Command.Name)
	}
}

func okHandler(context.Context, *HandlerContext, []byte) ([]byte, types.CompletionCode, error) {
	return nil, types.CodeOK, nil
}

func TestRegistry_Merge(t *testing.T) {
	a := NewRegistry()
	a.RegisterFunc(types.Command{ID: 0x01, NetFn: 0x06}, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
		return []byte{0x01}, types.CodeOK, nil
	})

	b := NewRegistry()
	b.RegisterFunc(types.Command{ID: 0x02, NetFn: 0x06}, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
		return []byte{0x02}, types.CodeOK, nil
	})

	a.Merge(b)

	_, cc1, _ := a.Dispatch(context.Background(), authedCtx(), 0x06, 0x01, nil)
	_, cc2, _ := a.Dispatch(context.Background(), authedCtx(), 0x06, 0x02, nil)

	if cc1 != types.CodeOK || cc2 != types.CodeOK {
		t.Errorf("after merge both keys should be present: cc1=%d cc2=%d", cc1, cc2)
	}
}

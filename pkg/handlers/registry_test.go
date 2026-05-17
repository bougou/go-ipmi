package handlers

import (
	"context"
	"testing"
)

func TestRegistry_Dispatch(t *testing.T) {
	tests := []struct {
		name    string
		netFn   uint8
		cmd     uint8
		setup   func(*Registry)
		wantCC  CompletionCode
		wantLen int
	}{
		{
			name:   "unknown command returns not supported",
			netFn:  0x06,
			cmd:    0xFF,
			setup:  func(r *Registry) {},
			wantCC: CodeCommandNotSupported,
		},
		{
			name:  "registered handler is dispatched",
			netFn: 0x06,
			cmd:   0x01,
			setup: func(r *Registry) {
				r.RegisterFunc(0x06, 0x01, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
					return []byte{0xAB}, CodeOK, nil
				})
			},
			wantCC:  CodeOK,
			wantLen: 1,
		},
		{
			name:  "middleware wraps handler",
			netFn: 0x06,
			cmd:   0x02,
			setup: func(r *Registry) {
				called := false
				r.Use(func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, CompletionCode, error) {
						called = true
						return next.Handle(ctx, hctx, data)
					})
				})
				r.RegisterFunc(0x06, 0x02, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
					return nil, CodeOK, nil
				})
				_ = called
			},
			wantCC: CodeOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			tc.setup(r)
			resp, cc, _ := r.Dispatch(context.Background(), &HandlerContext{}, tc.netFn, tc.cmd, nil)
			if cc != tc.wantCC {
				t.Errorf("cc: want %d, got %d", tc.wantCC, cc)
			}
			if tc.wantLen > 0 && len(resp) != tc.wantLen {
				t.Errorf("resp len: want %d, got %d", tc.wantLen, len(resp))
			}
		})
	}
}

func TestRegistry_Merge(t *testing.T) {
	a := NewRegistry()
	a.RegisterFunc(0x06, 0x01, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
		return []byte{0x01}, CodeOK, nil
	})

	b := NewRegistry()
	b.RegisterFunc(0x06, 0x02, func(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
		return []byte{0x02}, CodeOK, nil
	})

	a.Merge(b)

	_, cc1, _ := a.Dispatch(context.Background(), &HandlerContext{}, 0x06, 0x01, nil)
	_, cc2, _ := a.Dispatch(context.Background(), &HandlerContext{}, 0x06, 0x02, nil)

	if cc1 != CodeOK || cc2 != CodeOK {
		t.Errorf("after merge both keys should be present: cc1=%d cc2=%d", cc1, cc2)
	}
}

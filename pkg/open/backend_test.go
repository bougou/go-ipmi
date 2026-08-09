package open

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeBackend is a test double for Backend that records calls and
// returns scripted results. It implements Backend without touching any
// real transport.
type fakeBackend struct {
	name       string
	connectErr error
	sendResult []byte
	sendErr    error
	closeErr   error

	connectCalled bool
	closeCalled   bool
	sendCalled    int
}

func (f *fakeBackend) Connect(ctx context.Context, devnum int32) error {
	f.connectCalled = true
	return f.connectErr
}

func (f *fakeBackend) Close(ctx context.Context) error {
	f.closeCalled = true
	return f.closeErr
}

func (f *fakeBackend) Send(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	f.sendCalled++
	return f.sendResult, f.sendErr
}

func TestResolveBackend(t *testing.T) {
	comOK := &fakeBackend{name: "com"}
	psOK := &fakeBackend{name: "ps"}
	comFail := errors.New("com init failed")
	psFail := errors.New("ps init failed")

	tests := []struct {
		name       string
		pref       string
		comFn      func() (Backend, error)
		psFn       func() (Backend, error)
		want       string
		wantErr    string
		wantOnFail bool
	}{
		{
			name:  "explicit wmi-com success",
			pref:  BackendCOM,
			comFn: func() (Backend, error) { return comOK, nil },
			psFn:  func() (Backend, error) { return psOK, nil },
			want:  "com",
		},
		{
			name:  "explicit wmi-ps success",
			pref:  BackendPowerShell,
			comFn: func() (Backend, error) { return comOK, nil },
			psFn:  func() (Backend, error) { return psOK, nil },
			want:  "ps",
		},
		{
			name:    "explicit wmi-com failure does not fall back",
			pref:    BackendCOM,
			comFn:   func() (Backend, error) { return nil, comFail },
			psFn:    func() (Backend, error) { return psOK, nil },
			wantErr: "com init failed",
		},
		{
			name:       "auto falls back to ps when com fails",
			pref:       BackendAuto,
			comFn:      func() (Backend, error) { return nil, comFail },
			psFn:       func() (Backend, error) { return psOK, nil },
			want:       "ps",
			wantOnFail: true,
		},
		{
			name:  "auto uses com when available",
			pref:  BackendAuto,
			comFn: func() (Backend, error) { return comOK, nil },
			psFn:  func() (Backend, error) { return psOK, nil },
			want:  "com",
		},
		{
			name:       "auto fails when both fail includes both errors",
			pref:       BackendAuto,
			comFn:      func() (Backend, error) { return nil, comFail },
			psFn:       func() (Backend, error) { return nil, psFail },
			wantErr:    "wmi-com: com init failed; wmi-ps: ps init failed",
			wantOnFail: true,
		},
		{
			name:  "empty pref defaults to auto",
			pref:  "",
			comFn: func() (Backend, error) { return comOK, nil },
			psFn:  func() (Backend, error) { return psOK, nil },
			want:  "com",
		},
		{
			name:    "invalid pref rejected",
			pref:    "bogus",
			comFn:   func() (Backend, error) { return comOK, nil },
			psFn:    func() (Backend, error) { return psOK, nil },
			wantErr: `unsupported open backend "bogus"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var onFailCalled bool
			got, err := ResolveBackend(tt.pref, tt.comFn, tt.psFn, func(error) {
				onFailCalled = true
			})
			if onFailCalled != tt.wantOnFail {
				t.Fatalf("onCOMFail called=%v, want %v", onFailCalled, tt.wantOnFail)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("want error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			fake, ok := got.(*fakeBackend)
			if !ok {
				t.Fatalf("want *fakeBackend, got %T", got)
			}
			if fake.name != tt.want {
				t.Fatalf("want backend %q, got %q", tt.want, fake.name)
			}
		})
	}
}

func TestRequestEffectiveDefaults(t *testing.T) {
	r := &Request{}
	if r.EffectiveTarget() != BMCAddr {
		t.Fatalf("EffectiveTarget: got %#02x", r.EffectiveTarget())
	}
	if r.EffectiveMyAddr() != BMCAddr {
		t.Fatalf("EffectiveMyAddr: got %#02x", r.EffectiveMyAddr())
	}
	if r.UsesIPMB() {
		t.Fatal("default request should use system interface")
	}
}

func TestRequestUsesIPMB(t *testing.T) {
	r := &Request{TargetAddr: 0x2c, MyAddr: 0x20}
	if !r.UsesIPMB() {
		t.Fatal("expected IPMB routing")
	}
	r.TargetAddr = 0x20
	if r.UsesIPMB() {
		t.Fatal("same as MyAddr should be system interface")
	}
	r.TargetAddr = 0
	if r.UsesIPMB() {
		t.Fatal("TargetAddr 0 means local system interface")
	}
}

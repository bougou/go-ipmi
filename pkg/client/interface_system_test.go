package client

import (
	"context"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/open"
	"github.com/bougou/go-ipmi/pkg/types"
)

// captureBackend records the last open.Request handed to Send.
type captureBackend struct {
	last *open.Request
}

func (b *captureBackend) Connect(context.Context, int32) error { return nil }
func (b *captureBackend) Close(context.Context) error          { return nil }
func (b *captureBackend) Send(_ context.Context, req *open.Request, _ time.Duration) ([]byte, error) {
	// Copy so later mutations (if any) cannot race assertions.
	cp := *req
	if req.Data != nil {
		cp.Data = append([]byte(nil), req.Data...)
	}
	b.last = &cp
	return []byte{0x00}, nil
}

func newOpenTestClient(t *testing.T) (*Client, *captureBackend) {
	t.Helper()
	c, err := NewOpenClient()
	if err != nil {
		t.Fatalf("NewOpenClient: %v", err)
	}
	b := &captureBackend{}
	c.openipmi.backend = b
	c.timeout = time.Second
	return c, b
}

func TestBuildOpenRequestSystemInterfaceDefault(t *testing.T) {
	c, b := newOpenTestClient(t)
	if _, err := c.openSendRequest(context.Background(), &app.GetDeviceIDRequest{}); err != nil {
		t.Fatalf("openSendRequest: %v", err)
	}
	if b.last.UsesIPMB() {
		t.Fatal("default should use system interface")
	}
	if b.last.TargetAddr != types.BMC_SA {
		t.Fatalf("TargetAddr: got %#02x, want BMC_SA", b.last.TargetAddr)
	}
	if b.last.LUN != 0 {
		t.Fatalf("LUN: got %d, want 0", b.last.LUN)
	}
}

func TestBuildOpenRequestCommandContextIPMB(t *testing.T) {
	c, b := newOpenTestClient(t)
	cmdCtx := (&CommandContext{}).WithResponderAddr(0x2c).WithResponderLUN(0x01)
	ctx := WithCommandContext(context.Background(), cmdCtx)

	if _, err := c.openSendRequest(ctx, &app.GetDeviceIDRequest{}); err != nil {
		t.Fatalf("openSendRequest: %v", err)
	}
	if !b.last.UsesIPMB() {
		t.Fatal("expected IPMB routing")
	}
	if b.last.TargetAddr != 0x2c || b.last.LUN != 0x01 {
		t.Fatalf("target=%#02x lun=%d", b.last.TargetAddr, b.last.LUN)
	}
	if b.last.MyAddr != types.BMC_SA {
		t.Fatalf("MyAddr: got %#02x", b.last.MyAddr)
	}
}

func TestBuildOpenRequestCommandContextBMCStaysSystemInterface(t *testing.T) {
	c, b := newOpenTestClient(t)
	cmdCtx := (&CommandContext{}).WithResponderAddr(types.BMC_SA).WithResponderLUN(2)
	ctx := WithCommandContext(context.Background(), cmdCtx)

	if _, err := c.openSendRequest(ctx, &app.GetDeviceIDRequest{}); err != nil {
		t.Fatalf("openSendRequest: %v", err)
	}
	if b.last.UsesIPMB() {
		t.Fatal("BMC_SA should stay on system interface")
	}
	if b.last.LUN != 2 {
		t.Fatalf("LUN: got %d, want 2", b.last.LUN)
	}
}

func TestBuildOpenRequestSessionTargetIPMB(t *testing.T) {
	c, b := newOpenTestClient(t)
	c.openipmi.targetAddr = 0x2c
	c.openipmi.targetChannel = 0x01

	if _, err := c.openSendRequest(context.Background(), &app.GetDeviceIDRequest{}); err != nil {
		t.Fatalf("openSendRequest: %v", err)
	}
	if !b.last.UsesIPMB() {
		t.Fatal("expected IPMB from session target")
	}
	if b.last.TargetAddr != 0x2c || b.last.TargetChannel != 0x01 {
		t.Fatalf("target=%#02x ch=%d", b.last.TargetAddr, b.last.TargetChannel)
	}
}

func TestBuildOpenRequestCommandContextOverridesSessionTarget(t *testing.T) {
	c, b := newOpenTestClient(t)
	c.openipmi.targetAddr = 0x2c
	cmdCtx := (&CommandContext{}).WithResponderAddr(0x30)
	ctx := WithCommandContext(context.Background(), cmdCtx)

	if _, err := c.openSendRequest(ctx, &app.GetDeviceIDRequest{}); err != nil {
		t.Fatalf("openSendRequest: %v", err)
	}
	if b.last.TargetAddr != 0x30 {
		t.Fatalf("target: got %#02x, want CommandContext 0x30", b.last.TargetAddr)
	}
}

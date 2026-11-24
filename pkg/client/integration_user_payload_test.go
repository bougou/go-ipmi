package client

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/command/app"
)

// TestSetUserPayloadAccess checks that the request handed to
// SetUserPayloadAccess reaches the BMC: disabling and then enabling SOL for
// user 2 on channel 1 is read back by GetUserPayloadAccess.
func TestSetUserPayloadAccess(t *testing.T) {
	c, _, _ := newSOLTestServer(t)
	ctx := context.Background()

	set := func(op app.SetUserPayloadAccessOperation) {
		t.Helper()
		if _, err := c.SetUserPayloadAccess(ctx, &app.SetUserPayloadAccessRequest{
			ChannelNumber:  1,
			UserID:         2,
			Operation:      op,
			PayloadTypeSOL: true,
		}); err != nil {
			t.Fatalf("SetUserPayloadAccess: %v", err)
		}
	}
	sol := func() bool {
		t.Helper()
		res, err := c.GetUserPayloadAccess(ctx, 1, 2)
		if err != nil {
			t.Fatalf("GetUserPayloadAccess: %v", err)
		}
		return res.PayloadTypeSOL
	}

	set(app.SetUserPayloadAccessOperationDisable)
	if sol() {
		t.Fatal("SOL access still enabled after disable")
	}
	set(app.SetUserPayloadAccessOperationEnable)
	if !sol() {
		t.Fatal("SOL access not enabled after enable")
	}
}

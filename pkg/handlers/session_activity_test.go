package handlers

import (
	"context"
	"encoding/binary"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// advanceableClock is a Clock whose current time can be moved forward by tests.
type advanceableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *advanceableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *advanceableClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func (c *advanceableClock) NewTimer(d time.Duration) clock.Timer   { return clock.Real.NewTimer(d) }
func (c *advanceableClock) NewTicker(d time.Duration) clock.Ticker { return clock.Real.NewTicker(d) }

func newAdminBMC(clk clock.Clock) (*bmc.BMC, error) {
	info := bmc.DeviceInfo{
		DeviceID:                1,
		DeviceRevision:          1,
		FirmwareMajor:           2,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x39,
	}
	b := bmc.New(info, [16]byte{}, mock.New(), bmc.WithClock(clk))
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		return nil, err
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[lanChannelNumber] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	return b, nil
}

// TestHandleRAKP1DoesNotRefreshLastActivity asserts RAKP Message 1 leaves the
// session's LastActivity alone. RAKP messages carry no authenticator, so a
// refresh would let anyone who saw a session ID keep that session alive with
// replayed handshake packets; the budget stamped at allocation bounds the
// handshake instead.
func TestHandleRAKP1DoesNotRefreshLastActivity(t *testing.T) {
	clk := &advanceableClock{now: time.Now()}
	b, err := newAdminBMC(clk)
	if err != nil {
		t.Fatal(err)
	}

	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, bmc.PrivilegeLevelAdministrator, lanChannelNumber)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	created := sess.LastActivity

	clk.advance(30 * time.Second)

	resp, err := HandleRAKP1(context.Background(), b, rakp1Payload(sess.BMCID, bmc.PrivilegeLevelAdministrator, "ADMIN"))
	if err != nil {
		t.Fatalf("HandleRAKP1: %v", err)
	}
	if len(resp) < 2 || resp[1] != 0x00 {
		t.Fatalf("want successful RAKP2 response, got %x", resp)
	}
	if !sess.LastActivity.Equal(created) {
		t.Fatalf("LastActivity changed by RAKP1: created=%v after=%v", created, sess.LastActivity)
	}
}

// TestHandleRAKP3DoesNotRefreshLastActivity asserts a malformed RAKP Message 3
// naming a pending session leaves LastActivity alone: refreshing before the
// message is authenticated would let garbage packets hold the pending slot open
// forever, which is exactly the keepalive the touch-on-validated-packets design
// removes.
func TestHandleRAKP3DoesNotRefreshLastActivity(t *testing.T) {
	clk := &advanceableClock{now: time.Now()}
	b, err := newAdminBMC(clk)
	if err != nil {
		t.Fatal(err)
	}

	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, bmc.PrivilegeLevelAdministrator, lanChannelNumber)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	created := sess.LastActivity

	clk.advance(30 * time.Second)

	// Minimal RAKP3 payload: tag, statusCode=0, bmc session ID. The unpack
	// fails, so this is exactly the garbage packet that must not count as
	// activity.
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[4:8], sess.BMCID)
	if _, err := HandleRAKP3(context.Background(), b, payload); err != nil {
		t.Fatalf("HandleRAKP3: %v", err)
	}
	if !sess.LastActivity.Equal(created) {
		t.Fatalf("LastActivity changed by malformed RAKP3: created=%v after=%v", created, sess.LastActivity)
	}
}

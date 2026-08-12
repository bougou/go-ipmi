package server

// End-to-end user-management tests driven through the real pkg/client, proving
// the create-then-authenticate round trip the metal-agent relies on and that
// the client's GetUsers enumerator walks every slot without error.

import (
	"context"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/types"
)

// adminClient returns a connected admin client against the server on port.
func adminClient(t *testing.T, ctx context.Context, port int) *client.Client {
	t.Helper()
	cl, err := client.NewClient("127.0.0.1", port, raceUser, racePass)
	if err != nil {
		t.Fatal(err)
	}
	cl = cl.WithTimeout(2 * time.Second).WithCipherSuiteID(types.CipherSuiteID3)
	if err := cl.Connect(ctx); err != nil {
		t.Fatalf("admin connect: %v", err)
	}
	return cl
}

// createUser provisions an enabled admin-level user on channel 1 via the client
// command set (Set User Name, Set User Access, Set User Password, Enable User).
func createUser(t *testing.T, ctx context.Context, cl *client.Client, id uint8, name, pass string, stored20 bool) {
	t.Helper()
	if _, err := cl.SetUsername(ctx, id, name); err != nil {
		t.Fatalf("SetUsername: %v", err)
	}
	if _, err := cl.SetUserAccess(ctx, &app.SetUserAccessRequest{
		EnableChanging:      true,
		EnableIPMIMessaging: true,
		ChannelNumber:       1,
		UserID:              id,
		MaxPrivLevel:        uint8(bmc.PrivilegeLevelAdministrator),
	}); err != nil {
		t.Fatalf("SetUserAccess: %v", err)
	}
	if _, err := cl.SetUserPassword(ctx, id, pass, stored20); err != nil {
		t.Fatalf("SetUserPassword: %v", err)
	}
	if err := cl.EnableUser(ctx, id); err != nil {
		t.Fatalf("EnableUser: %v", err)
	}
}

// authAs opens a fresh RMCP+ (suite 3) session as the given user and returns any
// connect error.
func authAs(ctx context.Context, port int, name, pass string) error {
	cl, err := client.NewClient("127.0.0.1", port, name, pass)
	if err != nil {
		return err
	}
	cl = cl.WithTimeout(2 * time.Second).WithCipherSuiteID(types.CipherSuiteID3)
	if err := cl.Connect(ctx); err != nil {
		return err
	}
	return cl.Close(ctx)
}

// TestUserCreateThenAuthenticate creates a user over the client command set and
// then authenticates a brand-new RMCP+ session as that user. It runs the round
// trip for both the 16-byte and 20-byte password-set forms of the same secret,
// which must authenticate identically because RAKP derives keys from the
// 20-byte padded password.
func TestUserCreateThenAuthenticate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		stored20 bool
	}{
		{"password16", false},
		{"password20", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := raceNewBMC(t)
			port, ctx, stop := raceStartServer(t, b)
			defer stop()

			admin := adminClient(t, ctx, port)
			defer admin.Close(ctx) //nolint:errcheck

			const (
				newUser = "operator"
				newPass = "operatorpass"
			)
			createUser(t, ctx, admin, 3, newUser, newPass, tc.stored20)

			if err := authAs(ctx, port, newUser, newPass); err != nil {
				t.Fatalf("authenticate as new user: %v", err)
			}

			// A wrong password must still be rejected.
			if err := authAs(ctx, port, newUser, "wrongpassword"); err == nil {
				t.Fatalf("authentication succeeded with wrong password")
			}
		})
	}
}

// TestUserSetCommandsRequireAdministrator is the privilege-escalation
// regression: an Operator-privilege session must be rejected from all three Set
// User* commands (security-restriction completion code), while Get User Access
// and Get User Name stay reachable at Operator.
func TestUserSetCommandsRequireAdministrator(t *testing.T) {
	b := raceNewBMC(t)

	// Seed an operator-limited user (construction-time seeding, before Serve).
	op, err := b.Users.Add(3, "operator")
	if err != nil {
		t.Fatal(err)
	}
	op.SetPassword([]byte("operatorpass1234"))
	op.Enabled = true
	op.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelOperator, Enabled: true}

	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	cl, err := client.NewClient("127.0.0.1", port, "operator", "operatorpass1234")
	if err != nil {
		t.Fatal(err)
	}
	cl = cl.WithTimeout(2 * time.Second).WithCipherSuiteID(types.CipherSuiteID3).
		WithMaxPrivilegeLevel(types.PrivilegeLevelOperator)
	if err := cl.Connect(ctx); err != nil {
		t.Fatalf("operator connect: %v", err)
	}
	defer cl.Close(ctx) //nolint:errcheck

	wantRestricted := func(name string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: succeeded at operator privilege, want security restriction", name)
		}
		respErr, ok := types.IsResponseError(err)
		if !ok {
			t.Fatalf("%s: err = %v, want IPMI response error", name, err)
		}
		if respErr.CompletionCode() != types.CodeInsufficientPrivilege {
			t.Fatalf("%s: cc = 0x%02x, want 0xD4 (security restriction)", name, uint8(respErr.CompletionCode()))
		}
	}

	_, err = cl.SetUserPassword(ctx, 4, "irrelevant", false)
	wantRestricted("Set User Password", err)

	_, err = cl.SetUsername(ctx, 4, "evil")
	wantRestricted("Set User Name", err)

	_, err = cl.SetUserAccess(ctx, &app.SetUserAccessRequest{
		EnableChanging:      true,
		EnableIPMIMessaging: true,
		ChannelNumber:       1,
		UserID:              4,
		MaxPrivLevel:        uint8(bmc.PrivilegeLevelAdministrator),
	})
	wantRestricted("Set User Access", err)

	// Get User Access / Get User Name require only Operator, so they succeed.
	if _, err := cl.GetUserAccess(ctx, 1, 3); err != nil {
		t.Errorf("Get User Access at operator privilege: %v", err)
	}
	if _, err := cl.GetUsername(ctx, 3); err != nil {
		t.Errorf("Get User Name at operator privilege: %v", err)
	}
}

// TestGetUsersEnumerates proves the bougou client's own GetUsers walks every
// slot from 1 to the reported maximum without error, returning empty names for
// unpopulated slots (Get User Name answers 0xCC there).
func TestGetUsersEnumerates(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	admin := adminClient(t, ctx, port)
	defer admin.Close(ctx) //nolint:errcheck

	users, err := admin.GetUsers(ctx, 1)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if got := len(users); got != int(bmc.MaxUsers) {
		t.Fatalf("enumerated %d users, want %d", got, bmc.MaxUsers)
	}
	// Slot 2 is the seeded admin; an unpopulated slot reports an empty name.
	if users[1].Name != raceUser {
		t.Errorf("slot 2 name = %q, want %q", users[1].Name, raceUser)
	}
	if users[9].Name != "" {
		t.Errorf("empty slot name = %q, want empty", users[9].Name)
	}
}

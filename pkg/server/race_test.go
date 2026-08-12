package server

// Data-race regression tests. Each spins up a real server and drives concurrent
// traffic that, before the per-session lock / copy-out / config-lock fixes,
// tripped the race detector. Run with: go test -race ./pkg/server/...

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	raceUser = "admin"
	racePass = "adminpass1234567"
)

// raceNewBMC builds a BMC with an enabled admin user on the LAN channel.
func raceNewBMC(t *testing.T, opts ...bmc.Option) *bmc.BMC {
	t.Helper()

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, mock.New(),
		append([]bmc.Option{bmc.WithClock(clock.Real)}, opts...)...)

	admin, err := b.Users.Add(2, raceUser)
	if err != nil {
		t.Fatal(err)
	}
	admin.SetPassword([]byte(racePass))
	admin.Enabled = true
	admin.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	return b
}

// raceStartServer starts b on a loopback UDP socket and returns its port, a
// context, and a stop func.
func raceStartServer(t *testing.T, b *bmc.BMC) (int, context.Context, func()) {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port //nolint:forcetypeassert

	srv := NewServer(b, udp.Wrap(conn, udp.WithReadTimeout(time.Second)))
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx) //nolint:errcheck

	// Poll for readiness with an ASF Presence Ping instead of a fixed sleep:
	// under -race on a loaded CI runner, a fixed delay is both too long on a
	// fast machine and too short on a slow one.
	probe, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Close()
	deadline := time.Now().Add(5 * time.Second)
	buf := make([]byte, 64)
	for {
		// RMCP/ASF Presence Ping (class 0x06, type 0x80).
		_, _ = probe.Write([]byte{0x06, 0x00, 0xff, 0x06, 0x00, 0x00, 0x11, 0xbe, 0x80, 0x00, 0x00, 0x00})
		_ = probe.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err := probe.Read(buf); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server did not answer ASF presence ping within 5s")
		}
	}

	return port, ctx, func() { cancel(); _ = srv.Close() }
}

// raceConnectLoop runs four goroutines that each establish and close iters
// RMCP+ (suite 3) sessions, exercising the server's read of user/channel/cipher
// state on the handshake hot path.
func raceConnectLoop(ctx context.Context, wg *sync.WaitGroup, port, iters int) {
	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range iters {
				cl, err := client.NewClient("127.0.0.1", port, raceUser, racePass)
				if err != nil {
					continue
				}

				cl = cl.WithTimeout(2 * time.Second).WithCipherSuiteID(types.CipherSuiteID3)

				if err := cl.Connect(ctx); err == nil {
					_ = cl.Close(ctx)
				}
			}
		}()
	}
}

// TestUserMutationRace mutates a user at runtime via the race-free
// UserStore.Update path while the server authenticates concurrent LAN sessions.
// It guards the copy-out in UserStore.Get/GetByName: the handshake reads users
// through those snapshot accessors without holding the store lock, so reverting
// them to hand out live *User pointers would race this concurrent writer.
func TestUserMutationRace(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for range 200 {
			_ = b.Users.Update(2, func(u *bmc.User) error {
				u.SetPassword([]byte(racePass))
				u.Enabled = true
				return nil
			})
		}
	}()

	raceConnectLoop(ctx, &wg, port, 25)

	wg.Wait()
}

// TestCipherSuitesMutationRace reconfigures the advertised cipher suites at
// runtime (BMC.SetCipherSuites) while the server reads them via
// ResolvedCipherSuites on every Open Session Request and Get Channel Cipher
// Suites. The config RWMutex serializes writer and readers.
func TestCipherSuitesMutationRace(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for range 500 {
			b.SetCipherSuites([]types.CipherSuiteID{types.CipherSuiteID3, types.CipherSuiteID17})
		}
	}()

	raceConnectLoop(ctx, &wg, port, 25)

	wg.Wait()
}

// TestChannelReconfigRace republishes a channel at runtime via the race-free
// ChannelStore.Set path while the server reads channel config (via
// ChannelStore.Get snapshots) on every RAKP3. Set stores a private copy and Get
// returns one, so writer and readers never share a Channel.
func TestChannelReconfigRace(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for range 500 {
			b.Channels.Set(&bmc.Channel{
				Number:         1,
				Medium:         bmc.ChannelMediumLAN,
				AccessMode:     bmc.ChannelAccessAlways,
				MaxPrivilege:   bmc.PrivilegeLevelAdministrator,
				PerMessageAuth: true,
				UserLevelAuth:  true,
			})
		}
	}()

	raceConnectLoop(ctx, &wg, port, 25)

	wg.Wait()
}

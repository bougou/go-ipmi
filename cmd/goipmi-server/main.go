// goipmi-server is a minimal IPMI BMC server for e2e testing.
//
// It uses the in-memory mock HAL and serves on UDP port 623.  By default it
// creates a single user "ADMIN" with password "ADMIN" so that external tools
// such as ipmitool can connect via lanplus.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "goipmi-server: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// BMC identity (used in Get Device ID).
	info := bmc.DeviceInfo{
		DeviceID:                32,
		DeviceRevision:          1,
		FirmwareMajor:           1,
		FirmwareMinor:           0,
		IPMIVersion:             0x20, // IPMI 2.0
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x3D,
	}
	var guid [16]byte
	copy(guid[:], "go-ipmi-e2e\x00\x00\x00\x00")

	b := bmc.New(info, guid, mock.New(), bmc.WithClock(clock.Real))

	// Optional cipher suite override. GOIPMI_SERVER_CIPHER_SUITES is a
	// comma-separated list of RMCP+ cipher suite IDs (e.g. "3,17" or
	// "1,2,15,16"). Empty/unset falls back to bmc.DefaultCipherSuites.
	if raw := strings.TrimSpace(os.Getenv("GOIPMI_SERVER_CIPHER_SUITES")); raw != "" {
		ids, err := parseCipherSuites(raw)
		if err != nil {
			return err
		}
		b.SetCipherSuites(ids)
	}

	// User credentials for IPMI session authentication.
	userName := os.Getenv("GOIPMI_SERVER_USER")
	if userName == "" {
		userName = "ADMIN"
	}
	userPass := os.Getenv("GOIPMI_SERVER_PASS")
	if userPass == "" {
		userPass = "ADMIN"
	}

	// Add the user with Administrator privilege on LAN channel 1.
	user, err := b.Users.Add(2, userName)
	if err != nil {
		return fmt.Errorf("add user: %w", err)
	}
	user.SetPassword([]byte(userPass))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}

	// Bind UDP on the configured port.
	port := os.Getenv("GOIPMI_SERVER_PORT")
	if port == "" {
		port = "623"
	}
	addr := ":" + port
	conn, err := udp.Listen(addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	srv := server.NewServer(b, conn)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Close the server when the process receives a signal to unblock Serve().
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	fmt.Printf("goipmi-server: listening on %s (lanplus, user=%s, pass=%s)\n", addr, userName, userPass)
	if len(b.ResolvedCipherSuites()) != 0 {
		fmt.Printf("goipmi-server: cipher suites: %v\n", b.ResolvedCipherSuites())
	}
	if err := srv.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("serve: %w", err)
	}
	fmt.Println("goipmi-server: stopped")
	return nil
}

// parseCipherSuites parses a comma-separated list of cipher suite IDs from the
// GOIPMI_SERVER_CIPHER_SUITES env var and validates each against the reference
// server's supported set. Returns a descriptive error for bad input rather than
// letting bmc.SetCipherSuites panic.
func parseCipherSuites(raw string) ([]bmc.CipherSuiteID, error) {
	parts := strings.Split(raw, ",")
	ids := make([]bmc.CipherSuiteID, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil, fmt.Errorf("invalid cipher suite id %q: expected an integer 0..255", p)
		}
		id := bmc.CipherSuiteID(n)
		if !bmc.SupportedCipherSuite(id) {
			return nil, fmt.Errorf("cipher suite %d is not implemented by the reference server", id)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("GOIPMI_SERVER_CIPHER_SUITES contained no valid ids")
	}
	return ids, nil
}

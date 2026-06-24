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
	if err := srv.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("serve: %w", err)
	}
	fmt.Println("goipmi-server: stopped")
	return nil
}

// goipmi-server is a reference IPMI BMC server for development and E2E testing.
//
// By default it serves both IPMI protocol generations on the same UDP port:
//   - IPMI v2.0 / RMCP+ (ipmitool -I lanplus, goipmi -I lanplus)
//   - IPMI v1.5       (ipmitool -I lan -A MD5, goipmi -I lan)
//
// Environment variables:
//
//	GOIPMI_SERVER_PORT            – UDP listen port (default: 623)
//	GOIPMI_SERVER_USER            – BMC username (default: ADMIN)
//	GOIPMI_SERVER_PASS            – BMC password (default: ADMIN)
//	GOIPMI_SERVER_CIPHER_SUITES   – RMCP+ cipher suite IDs, comma-separated (default: 3,17)
//	GOIPMI_SERVER_V15_AUTH_TYPES  – v1.5 auth types: none,md2,md5,password,oem (default: md5)
//	GOIPMI_SERVER_V15             – set to 0/false to disable v1.5 while keeping lanplus (default: 1)
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
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/bougou/go-ipmi/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "goipmi-server: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadRuntimeConfig()
	if err != nil {
		return err
	}

	info := bmc.DeviceInfo{
		DeviceID:                32,
		DeviceRevision:          1,
		FirmwareMajor:           1,
		FirmwareMinor:           0,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x3D,
	}
	var guid [16]byte
	copy(guid[:], "go-ipmi-e2e\x00\x00\x00\x00")

	halImpl := mock.New()
	seedReferenceStorage(context.Background(), halImpl)

	b := bmc.New(info, guid, halImpl, bmc.WithClock(clock.Real))
	applyRuntimeConfig(b, cfg)

	user, err := b.Users.Add(2, cfg.User)
	if err != nil {
		return fmt.Errorf("add user: %w", err)
	}
	user.SetPassword([]byte(cfg.Password))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}

	addr := ":" + cfg.Port
	conn, err := udp.Listen(addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	srv := server.NewServer(b, conn)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	printRuntimeBanner(cfg, b)

	if err := srv.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("serve: %w", err)
	}
	fmt.Println("goipmi-server: stopped")
	return nil
}

func seedReferenceStorage(ctx context.Context, h hal.HAL) {
	store := h.Storage()
	if store == nil {
		return
	}
	if fru := store.FRU(); fru != nil {
		_ = fru.Write(ctx, 0, types.PackFRU(types.FRUPackConfig{
			Product: &types.FRUPackProduct{
				Manufacturer: "go-ipmi",
				Name:         "reference-bmc",
				Version:      "1.0",
				Serial:       "e2e",
			},
		}))
	}
	if sdr := store.SDR(); sdr != nil {
		_ = sdr.Write(ctx, 1, types.PackMCLocator(types.MCLocatorPackOpts{
			RecordID: 1,
		}))
	}
}

package main

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

func TestParseCipherSuites(t *testing.T) {
	ids, err := parseCipherSuites("3,17")
	if err != nil {
		t.Fatalf("parseCipherSuites: %v", err)
	}
	if len(ids) != 2 || ids[0] != ipmi.CipherSuiteID3 || ids[1] != ipmi.CipherSuiteID17 {
		t.Fatalf("unexpected ids: %v", ids)
	}
}

func TestParseBoolEnv(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"0", false},
		{"1", true},
		{"off", false},
		{"yes", true},
	} {
		got, err := parseBoolEnv(tc.in)
		if err != nil {
			t.Fatalf("parseBoolEnv(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("parseBoolEnv(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestLoadRuntimeConfigDefaultsDualStack(t *testing.T) {
	t.Setenv("GOIPMI_SERVER_PORT", "")
	t.Setenv("GOIPMI_SERVER_CIPHER_SUITES", "")
	t.Setenv("GOIPMI_SERVER_V15", "")
	t.Setenv("GOIPMI_SERVER_V15_AUTH_TYPES", "")

	cfg, err := loadRuntimeConfig()
	if err != nil {
		t.Fatalf("loadRuntimeConfig: %v", err)
	}
	if cfg.V15Disabled {
		t.Fatal("v1.5 should be enabled by default")
	}
	if len(cfg.V15AuthTypes) != 0 {
		t.Fatalf("expected nil v15 auth types (use BMC default), got %v", cfg.V15AuthTypes)
	}

	b := bmc.New(bmc.DeviceInfo{}, [16]byte{}, mock.New())
	applyRuntimeConfig(b, cfg)
	if !b.V15LANEnabled() {
		t.Fatal("expected v1.5 enabled after apply")
	}
	if b.ResolvedV15AuthTypes()[0] != bmc.V15AuthTypeMD5 {
		t.Fatalf("default v1.5 auth: %v", b.ResolvedV15AuthTypes())
	}
}

func TestLoadRuntimeConfigCustomV15Auth(t *testing.T) {
	t.Setenv("GOIPMI_SERVER_V15_AUTH_TYPES", "md5,md2")

	cfg, err := loadRuntimeConfig()
	if err != nil {
		t.Fatalf("loadRuntimeConfig: %v", err)
	}
	if len(cfg.V15AuthTypes) != 2 {
		t.Fatalf("want 2 auth types, got %v", cfg.V15AuthTypes)
	}
}

func TestLoadRuntimeConfigDisableV15(t *testing.T) {
	t.Setenv("GOIPMI_SERVER_V15", "0")
	t.Setenv("GOIPMI_SERVER_V15_AUTH_TYPES", "")

	cfg, err := loadRuntimeConfig()
	if err != nil {
		t.Fatalf("loadRuntimeConfig: %v", err)
	}
	if !cfg.V15Disabled {
		t.Fatal("expected v1.5 disabled")
	}

	b := bmc.New(bmc.DeviceInfo{}, [16]byte{}, mock.New())
	applyRuntimeConfig(b, cfg)
	if b.V15LANEnabled() {
		t.Fatal("expected v1.5 disabled after apply")
	}
}

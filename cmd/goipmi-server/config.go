package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bougou/go-ipmi/pkg/bmc"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// runtimeConfig holds goipmi-server settings from the environment.
type runtimeConfig struct {
	Port     string
	User     string
	Password string

	CipherSuites []ipmi.CipherSuiteID
	V15AuthTypes []bmc.V15AuthType // nil = default (md5)
	V15Disabled  bool
}

func loadRuntimeConfig() (runtimeConfig, error) {
	cfg := runtimeConfig{
		Port:     envOr("GOIPMI_SERVER_PORT", "623"),
		User:     envOr("GOIPMI_SERVER_USER", "ADMIN"),
		Password: envOr("GOIPMI_SERVER_PASS", "ADMIN"),
	}

	if raw := strings.TrimSpace(os.Getenv("GOIPMI_SERVER_CIPHER_SUITES")); raw != "" {
		ids, err := parseCipherSuites(raw)
		if err != nil {
			return cfg, err
		}
		cfg.CipherSuites = ids
	}

	if v := strings.TrimSpace(os.Getenv("GOIPMI_SERVER_V15")); v != "" {
		enabled, err := parseBoolEnv(v)
		if err != nil {
			return cfg, fmt.Errorf("GOIPMI_SERVER_V15: %w", err)
		}
		cfg.V15Disabled = !enabled
	}

	if raw := strings.TrimSpace(os.Getenv("GOIPMI_SERVER_V15_AUTH_TYPES")); raw != "" {
		types, err := bmc.ParseV15AuthTypes(raw)
		if err != nil {
			return cfg, fmt.Errorf("GOIPMI_SERVER_V15_AUTH_TYPES: %w", err)
		}
		cfg.V15AuthTypes = types
		cfg.V15Disabled = false
	}

	return cfg, nil
}

func applyRuntimeConfig(b *bmc.BMC, cfg runtimeConfig) {
	if len(cfg.CipherSuites) > 0 {
		b.SetCipherSuites(cfg.CipherSuites)
	}
	if cfg.V15Disabled {
		bmc.WithV15Disabled()(b)
	} else if len(cfg.V15AuthTypes) > 0 {
		bmc.WithV15AuthTypes(cfg.V15AuthTypes)(b)
	}
}

func printRuntimeBanner(cfg runtimeConfig, b *bmc.BMC) {
	fmt.Printf("goipmi-server: listening on :%s (user=%s)\n", cfg.Port, cfg.User)
	fmt.Printf("goipmi-server: IPMI v2.0 (lanplus) cipher suites: %v\n", b.ResolvedCipherSuites())
	if b.V15LANEnabled() {
		fmt.Printf("goipmi-server: IPMI v1.5 (lan) auth types: %s\n", bmc.FormatV15AuthTypes(b.ResolvedV15AuthTypes()))
	} else {
		fmt.Println("goipmi-server: IPMI v1.5 (lan) disabled")
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseBoolEnv(v string) (bool, error) {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected 0/1, true/false, yes/no, or on/off, got %q", v)
	}
}

// parseCipherSuites parses a comma-separated list of cipher suite IDs from the
// GOIPMI_SERVER_CIPHER_SUITES env var and validates each against the reference
// server's supported set. Returns a descriptive error for bad input rather than
// letting bmc.SetCipherSuites panic.
func parseCipherSuites(raw string) ([]ipmi.CipherSuiteID, error) {
	parts := strings.Split(raw, ",")
	ids := make([]ipmi.CipherSuiteID, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil, fmt.Errorf("invalid cipher suite id %q: expected an integer 0..255", p)
		}
		id := ipmi.CipherSuiteID(n)
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

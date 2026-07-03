package bmc

import (
	"fmt"
	"strings"
)

// V15AuthTypeName returns a human-readable name for t.
func V15AuthTypeName(t V15AuthType) string {
	switch t {
	case V15AuthTypeNone:
		return "none"
	case V15AuthTypeMD2:
		return "md2"
	case V15AuthTypeMD5:
		return "md5"
	case V15AuthTypePassword:
		return "password"
	case V15AuthTypeOEM:
		return "oem"
	default:
		return fmt.Sprintf("0x%02x", uint8(t))
	}
}

// FormatV15AuthTypes formats auth types for logging (e.g. "md5,md2").
func FormatV15AuthTypes(types []V15AuthType) string {
	if len(types) == 0 {
		return ""
	}
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = V15AuthTypeName(t)
	}
	return strings.Join(names, ",")
}

// ParseV15AuthType parses a single auth type name (case-insensitive).
func ParseV15AuthType(name string) (V15AuthType, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "none", "0", "00":
		return V15AuthTypeNone, nil
	case "md2", "1", "01":
		return V15AuthTypeMD2, nil
	case "md5", "2", "02":
		return V15AuthTypeMD5, nil
	case "password", "straight", "4", "04":
		return V15AuthTypePassword, nil
	case "oem", "5", "05":
		return V15AuthTypeOEM, nil
	default:
		return 0, fmt.Errorf("unknown v1.5 auth type %q (want none, md2, md5, password, or oem)", name)
	}
}

// ParseV15AuthTypes parses a comma-separated list of v1.5 auth type names.
func ParseV15AuthTypes(raw string) ([]V15AuthType, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("v1.5 auth type list is empty")
	}
	parts := strings.Split(raw, ",")
	out := make([]V15AuthType, 0, len(parts))
	seen := make(map[V15AuthType]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		t, err := ParseV15AuthType(p)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("v1.5 auth type list contained no valid entries")
	}
	return out, nil
}

package commands

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestResolvePassword(t *testing.T) {
	prompted := errors.New("prompted")
	prompt := func() (string, error) {
		return "", prompted
	}

	tests := []struct {
		name    string
		in      passwordSpec
		want    string
		wantErr error
		warns   []string
	}{
		{
			name: "explicit -P including empty",
			in: passwordSpec{
				passSet: true,
				pass:    "",
				fromEnv: true,
				remote:  true,
				lookup: func(string) (string, bool) {
					t.Fatal("lookup called")
					return "", false
				},
				prompt: prompt,
			},
		},
		{
			name: "-P wins over -E",
			in: passwordSpec{
				passSet: true,
				pass:    "from-flag",
				fromEnv: true,
				remote:  true,
				lookup: func(key string) (string, bool) {
					return "from-env", true
				},
				prompt: prompt,
			},
			want: "from-flag",
		},
		{
			name: "IPMITOOL_PASSWORD before IPMI_PASSWORD",
			in: passwordSpec{
				fromEnv: true,
				remote:  true,
				lookup: func(key string) (string, bool) {
					switch key {
					case ipmitoolPasswordEnv:
						return "tool", true
					case ipmiPasswordEnv:
						return "ipmi", true
					default:
						return "", false
					}
				},
				prompt: prompt,
			},
			want: "tool",
		},
		{
			name: "IPMI_PASSWORD when IPMITOOL_PASSWORD is unset",
			in: passwordSpec{
				fromEnv: true,
				lookup: func(key string) (string, bool) {
					if key == ipmiPasswordEnv {
						return "ipmi", true
					}
					return "", false
				},
				prompt: prompt,
			},
			want: "ipmi",
		},
		{
			name: "empty IPMITOOL_PASSWORD is explicit",
			in: passwordSpec{
				fromEnv: true,
				remote:  true,
				lookup: func(key string) (string, bool) {
					if key == ipmitoolPasswordEnv {
						return "", true
					}
					return "ipmi", true
				},
				prompt: prompt,
			},
		},
		{
			name: "-E with neither variable warns then prompts",
			in: passwordSpec{
				fromEnv: true,
				remote:  true,
				lookup: func(string) (string, bool) {
					return "", false
				},
				prompt: func() (string, error) {
					return "typed", nil
				},
			},
			want:  "typed",
			warns: []string{"Unable to read password from environment"},
		},
		{
			name: "-E with neither variable on a local interface warns and stays empty",
			in: passwordSpec{
				fromEnv: true,
				lookup: func(string) (string, bool) {
					return "", false
				},
				prompt: prompt,
			},
			warns: []string{"Unable to read password from environment"},
		},
		{
			name: "unset password on a remote interface prompts",
			in: passwordSpec{
				remote: true,
				lookup: func(string) (string, bool) {
					t.Fatal("environment read without -E")
					return "", false
				},
				prompt: func() (string, error) {
					return "typed", nil
				},
			},
			want: "typed",
		},
		{
			name: "unset password on a local interface does not prompt",
			in: passwordSpec{
				lookup: func(string) (string, bool) {
					t.Fatal("environment read without -E")
					return "", false
				},
				prompt: prompt,
			},
		},
		{
			name: "prompt error",
			in: passwordSpec{
				remote: true,
				prompt: func() (string, error) {
					return "", errors.New("no terminal")
				},
			},
			wantErr: errors.New("no terminal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var warns []string
			tt.in.warn = func(msg string) {
				warns = append(warns, msg)
			}
			got, err := resolvePassword(tt.in)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvePassword: %v", err)
			}
			if got != tt.want {
				t.Fatalf("password = %q, want %q", got, tt.want)
			}
			if strings.Join(warns, "\n") != strings.Join(tt.warns, "\n") {
				t.Fatalf("warnings = %q, want %q", warns, tt.warns)
			}
		})
	}
}

func TestPasswordFlags(t *testing.T) {
	origPass := password
	origEnv := passwordFromEnv
	origIntf := intf
	origRoot := rootCommand
	t.Cleanup(func() {
		password = origPass
		passwordFromEnv = origEnv
		intf = origIntf
		rootCommand = origRoot
	})

	cmd := NewRootCommand()
	if err := cmd.ParseFlags([]string{"-P", "secret", "-E"}); err != nil {
		t.Fatal(err)
	}
	if password != "secret" {
		t.Fatalf("password = %q", password)
	}
	if !passwordFromEnv {
		t.Fatal("-E was not set")
	}
	if !cmd.PersistentFlags().Changed("pass") {
		t.Fatal("-P was not marked as set")
	}

	got, err := resolveSessionPassword()
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret" {
		t.Fatalf("resolved = %q, want secret", got)
	}

	// -E alone reads the environment. A fresh command avoids the -P set above.
	password = ""
	passwordFromEnv = false
	cmd = NewRootCommand()
	if err := cmd.ParseFlags([]string{"-I", "lanplus", "-E"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ipmiPasswordEnv, "from-env")
	// Register restoration, then unset so IPMITOOL_PASSWORD does not hide IPMI_PASSWORD.
	t.Setenv(ipmitoolPasswordEnv, "")
	if err := os.Unsetenv(ipmitoolPasswordEnv); err != nil {
		t.Fatal(err)
	}

	got, err = resolveSessionPassword()
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-env" {
		t.Fatalf("resolved = %q, want from-env", got)
	}
}

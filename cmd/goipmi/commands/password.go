package commands

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	ipmiPasswordEnv     = "IPMI_PASSWORD"
	ipmitoolPasswordEnv = "IPMITOOL_PASSWORD"
)

// passwordSpec is the input to resolvePassword.
// lookup reports whether an environment variable is set.
// prompt reads a password from the terminal.
// warn reports a non-fatal problem, such as -E with no environment variable.
type passwordSpec struct {
	passSet bool
	pass    string
	fromEnv bool
	remote  bool
	lookup  func(string) (string, bool)
	prompt  func() (string, error)
	warn    func(string)
}

// resolvePassword selects the remote session password.
//
// Precedence matches ipmitool:
//  1. -P, when that flag was set (an empty value is an explicit NULL password)
//  2. -E, from IPMITOOL_PASSWORD, then IPMI_PASSWORD
//  3. a terminal prompt, for lan/lanplus, when no password was obtained above
//
// -E with neither variable set prints a warning and falls through to the prompt.
func resolvePassword(in passwordSpec) (string, error) {
	if in.passSet {
		return in.pass, nil
	}

	if in.fromEnv {
		if v, ok := in.lookup(ipmitoolPasswordEnv); ok {
			return v, nil
		}
		if v, ok := in.lookup(ipmiPasswordEnv); ok {
			return v, nil
		}
		warn := in.warn
		if warn == nil {
			warn = func(msg string) {
				fmt.Fprintln(os.Stderr, msg)
			}
		}
		warn("Unable to read password from environment")
	}

	if !in.remote {
		return "", nil
	}
	if in.prompt == nil {
		return "", fmt.Errorf("no terminal available to read password; specify -P or -E")
	}
	return in.prompt()
}

// resolveSessionPassword applies the process flags and environment to the
// session password. lan and lanplus are remote sessions and may prompt.
func resolveSessionPassword() (string, error) {
	passSet := false
	if rootCommand != nil {
		passSet = rootCommand.PersistentFlags().Changed("pass")
	}
	remote := intf == "lan" || intf == "lanplus"
	return resolvePassword(passwordSpec{
		passSet: passSet,
		pass:    password,
		fromEnv: passwordFromEnv,
		remote:  remote,
		lookup:  os.LookupEnv,
		prompt:  promptPassword,
	})
}

// promptPassword reads a password from the controlling terminal without echo.
// An empty entry is a NULL password, matching ipmitool.
func promptPassword() (string, error) {
	tty, closeFn, err := openPasswordTTY()
	if err != nil {
		return "", err
	}
	defer closeFn()

	if _, err := fmt.Fprint(tty, "Password: "); err != nil {
		return "", fmt.Errorf("prompt for password: %w", err)
	}
	pw, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	defer func() {
		for i := range pw {
			pw[i] = 0
		}
	}()
	return string(pw), nil
}

// openPasswordTTY returns the controlling terminal, or stdin when it is a
// terminal (Windows consoles have no /dev/tty). The closer releases a
// terminal opened here and does not close stdin.
func openPasswordTTY() (*os.File, func(), error) {
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		return f, func() { f.Close() }, nil
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, func() {}, nil
	}
	return nil, nil, fmt.Errorf("no terminal available to read password; specify -P or -E")
}

package commands

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	ipmiclient "github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/types"
	"github.com/spf13/cobra"
)

const homePage = "https://github.com/bougou/go-ipmi"

var (
	host     string
	port     int
	username string
	password string
	intf     string
	debug    bool

	privilegeLevel string
	showVersion    bool

	timeout int
	retries int

	client *ipmiclient.Client
)

func initClient() error {

	if debug {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("BuildAt: %s\n", BuildAt)
	}

	if debug {
		fmt.Printf("host: %s\n", host)
		fmt.Printf("port: %d\n", port)
		fmt.Printf("username: %s\n", username)
		fmt.Printf("password: %s\n", password)
		fmt.Printf("intf: %s\n", intf)
		fmt.Printf("debug: %t\n", debug)
		fmt.Printf("showVersion: %t\n", showVersion)
		fmt.Printf("timeout: %d\n", timeout)
		fmt.Printf("retries: %d\n", retries)
	}

	switch intf {
	case "", "open":
		c, err := ipmiclient.NewOpenClient()
		if err != nil {
			return fmt.Errorf("create open client failed, err: %w", err)
		}
		client = c // assign to global variable
		client.WithInterface(ipmiclient.InterfaceOpen)

	case "lan", "lanplus":
		c, err := ipmiclient.NewClient(host, port, username, password)
		if err != nil {
			return fmt.Errorf("create lan or lanplus client failed, err: %w", err)
		}
		client = c // assign to global variable
		switch intf {
		case "lan":
			client.WithInterface(ipmiclient.InterfaceLan)
			client.WithTimeout(time.Duration(ipmiclient.DefaultLanTimeoutSec) * time.Second)
			client.WithRetry(ipmiclient.DefaultLanRetries)

		case "lanplus":
			client.WithInterface(ipmiclient.InterfaceLanplus)
			client.WithTimeout(time.Duration(ipmiclient.DefaultLanplusTimeoutSec) * time.Second)
			client.WithRetry(ipmiclient.DefaultLanplusRetries)
		}

		client.WithRetry(retries)
		if timeout > 0 {
			client.WithTimeout(time.Duration(timeout) * time.Second)
		}

	case "tool":
		c, err := ipmiclient.NewToolClient(host)
		if err != nil {
			return fmt.Errorf("create client based on ipmitool (%s) failed, err: %w", host, err)
		}
		client = c // assign to global variable
		client.WithInterface(ipmiclient.InterfaceTool)

	default:
		return fmt.Errorf("unsupported interface")
	}

	client.WithDebug(debug)

	var privLevel types.PrivilegeLevel = types.PrivilegeLevelUnspecified
	switch strings.ToUpper(privilegeLevel) {
	case "CALLBACK":
		privLevel = types.PrivilegeLevelCallback
	case "USER":
		privLevel = types.PrivilegeLevelUser
	case "OPERATOR":
		privLevel = types.PrivilegeLevelOperator
	case "ADMINISTRATOR":
		privLevel = types.PrivilegeLevelAdministrator
	}

	if privLevel != types.PrivilegeLevelUnspecified {
		client.WithMaxPrivilegeLevel(privLevel)
	}

	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("client connect failed, err: %w", err)
	}
	return nil
}

func closeClient() error {
	ctx := context.Background()
	if err := client.Close(ctx); err != nil {
		return fmt.Errorf("close client failed, err: %w", err)
	}
	return nil
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "goipmi",
		Short: "goipmi",
		Long:  fmt.Sprintf("goipmi\n\nFind more information at: %s\n\n", homePage),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Printf("Version: %s\n", Version)
				fmt.Printf("Commit: %s\n", Commit)
				fmt.Printf("BuildAt: %s\n", BuildAt)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Remote host name for LAN interface")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 623, "Remote RMCP port")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "U", "", "Remote session username")
	rootCmd.PersistentFlags().StringVarP(&password, "pass", "P", "", "Remote session password")
	rootCmd.PersistentFlags().StringVarP(&intf, "interface", "I", "open", "Interface to use, supported (open,lan,lanplus)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "V", false, "Show version information")
	rootCmd.PersistentFlags().StringVarP(&privilegeLevel, "priv-level", "L", "ADMINISTRATOR", "Remote session privilege level to use. Can be CALLBACK, USER, OPERATOR, ADMINISTRATOR.")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "", 0, "timeout in seconds for each IPMI request/response cycle (not for entire command execution)"+
		"\n0 means to use the default hard-coded timeout of the interface")
	rootCmd.PersistentFlags().IntVarP(&retries, "retries", "R", 4, "Set the number of retries for lan/lanplus interface")
	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(NewCmdMC())
	rootCmd.AddCommand(NewCmdSEL())
	rootCmd.AddCommand(NewCmdSDR())
	rootCmd.AddCommand(NewCmdChassis())
	rootCmd.AddCommand(NewCmdChannel())
	rootCmd.AddCommand(NewCmdLan())
	rootCmd.AddCommand(NewCmdUser())
	rootCmd.AddCommand(NewCmdSession())
	rootCmd.AddCommand(NewCmdSensor())
	rootCmd.AddCommand(NewCmdFRU())
	rootCmd.AddCommand(NewCmdSOL())
	rootCmd.AddCommand(NewCmdPEF())
	rootCmd.AddCommand(NewCmdDCMI())

	rootCmd.AddCommand(NewCmdX())

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	return rootCmd
}

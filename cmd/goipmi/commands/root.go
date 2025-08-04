package commands

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/bougou/go-ipmi"
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

	client *ipmi.Client
)

func initClient() error {

	if debug {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("BuildAt: %s\n", BuildAt)
	}

	switch intf {
	case "", "open":
		c, err := ipmi.NewOpenClient()
		if err != nil {
			return fmt.Errorf("create open client failed, err: %w", err)
		}
		client = c
		client.WithInterface(ipmi.InterfaceOpen)

	case "lan", "lanplus":
		c, err := ipmi.NewClient(host, port, username, password)
		if err != nil {
			return fmt.Errorf("create lan or lanplus client failed, err: %w", err)
		}
		client = c
		switch intf {
		case "lan":
			client.WithInterface(ipmi.InterfaceLan)
		case "lanplus":
			client.WithInterface(ipmi.InterfaceLanplus)
		}
		client.WithTimeout(time.Duration(timeout) * time.Second)

	case "tool":
		c, err := ipmi.NewToolClient(host)
		if err != nil {
			return fmt.Errorf("create client based on ipmitool (%s) failed, err: %w", host, err)
		}
		client = c
		client.WithInterface(ipmi.InterfaceTool)
	default:
		return fmt.Errorf("unsupported interface")
	}

	client.WithDebug(debug)

	var privLevel ipmi.PrivilegeLevel = ipmi.PrivilegeLevelUnspecified
	switch strings.ToUpper(privilegeLevel) {
	case "CALLBACK":
		privLevel = ipmi.PrivilegeLevelCallback
	case "USER":
		privLevel = ipmi.PrivilegeLevelUser
	case "OPERATOR":
		privLevel = ipmi.PrivilegeLevelOperator
	case "ADMINISTRATOR":
		privLevel = ipmi.PrivilegeLevelAdministrator
	}

	if privLevel != ipmi.PrivilegeLevelUnspecified {
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

	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "host")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 623, "port")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "U", "", "username")
	rootCmd.PersistentFlags().StringVarP(&password, "pass", "P", "", "password")
	rootCmd.PersistentFlags().StringVarP(&intf, "interface", "I", "open", "interface, supported (open,lan,lanplus)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "V", false, "version")
	rootCmd.PersistentFlags().StringVarP(&privilegeLevel, "priv-level", "L", "ADMINISTRATOR", "Force session privilege level. Can be CALLBACK, USER, OPERATOR, ADMINISTRATOR.")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 20, "timeout in seconds. 0 means no timeout.")

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

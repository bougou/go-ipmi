package commands

import (
	"flag"
	"fmt"

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

	client *ipmi.Client
)

func initClient() error {
	c, err := ipmi.NewClient(host, port, username, password)
	if err != nil {
		return fmt.Errorf("create client failed, err: %s", err)
	}
	client = c

	client.WithDebug(debug)
	client.WithInterface(ipmi.Interface(intf))

	if err := client.Connect(); err != nil {
		return fmt.Errorf("client connect failed, err: %s", err)
	}
	return nil
}

func closeClient() error {
	if err := client.Close(); err != nil {
		return fmt.Errorf("close client failed, err: %s", err)
	}
	return nil
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "goipmi",
		Short: "goipmi",
		Long:  fmt.Sprintf("goipmi\n\nFind more information at: %s\n\n", homePage),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("goipmi run ...")
		},
	}

	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "host")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 623, "port")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "U", "", "username")
	rootCmd.PersistentFlags().StringVarP(&password, "pass", "P", "", "password")
	rootCmd.PersistentFlags().StringVarP(&intf, "interface", "I", "lanplus", "interface")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")

	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(NewCmdMC())
	rootCmd.AddCommand(NewCmdSEL())
	rootCmd.AddCommand(NewCmdSDR())
	rootCmd.AddCommand(NewCmdChassis())
	rootCmd.AddCommand(NewCmdChannel())
	rootCmd.AddCommand(NewCmdLan())
	rootCmd.AddCommand(NewCmdUser())
	rootCmd.AddCommand(NewCmdSession())

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	return rootCmd
}

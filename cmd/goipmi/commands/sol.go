package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdSOL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sol",
		Short: "sol",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdSOLInfo())
	cmd.AddCommand(NewCmdSOLActivate())

	return cmd
}

func NewCmdSOLInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			solConfigParams, err := client.GetSOLConfigParams(ctx, 0x0e)
			if err != nil {
				CheckErr(fmt.Errorf("GetDeviceID failed, err: %w", err))
			}
			fmt.Println(solConfigParams.Format())
		},
	}
	return cmd
}

func NewCmdSOLActivate() *cobra.Command {
	usage := `
	activate [<payload instance>]
	`

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "activate SOL payload",
		Run: func(cmd *cobra.Command, args []string) {
			payloadInstance := uint8(1)
			if len(args) >= 1 {
				id, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid payload instance, err: %w", err))
				}
				if id < 0 || id > 0x3f {
					CheckErr(fmt.Errorf("invalid payload instance, usage: %s", usage))
				}
				payloadInstance = uint8(id)
			}

			ctx := context.Background()
			var opts *ipmi.SOLActivateOptions
			if payloadInstance != 1 {
				opts = &ipmi.SOLActivateOptions{
					PayloadInstance: payloadInstance,
				}
			}

			if err := client.SOLActivate(ctx, os.Stdin, os.Stdout, opts); err != nil {
				CheckErr(err)
			}
		},
	}
	return cmd
}

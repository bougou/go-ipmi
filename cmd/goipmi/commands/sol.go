package commands

import (
	"context"
	"fmt"

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

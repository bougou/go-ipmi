package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdMC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mc",
		Short: "mc",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.AddCommand(NewCmdMCInfo())

	return cmd
}

func NewCmdMCInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetDeviceID()
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

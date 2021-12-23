package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdSDR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdr",
		Short: "sdr",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sdr ...")
		},
	}
	cmd.AddCommand(NewCmdSDRInfo())
	return cmd
}

func NewCmdSDRInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetSDRRepoInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRRepoInfo failed, err: %s", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

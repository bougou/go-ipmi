package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdSEL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sel",
		Short: "sel",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.AddCommand(NewCmdSELInfo())
	return cmd
}

func NewCmdSELInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			selInfo, err := client.GetSELInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}
			fmt.Println(selInfo.Format())

			selAllocInfo, err := client.GetSELAllocInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}
			fmt.Println(selAllocInfo.Format())
		},
	}
	return cmd
}
